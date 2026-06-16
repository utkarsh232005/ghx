package screens

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/KDM-cli/ghx/internal/ai"
	"github.com/KDM-cli/ghx/internal/components"
	"github.com/KDM-cli/ghx/internal/gh"
	"github.com/KDM-cli/ghx/internal/git"
	"github.com/KDM-cli/ghx/styles"
)

type prState int

const (
	prDashboard prState = iota
	prEnterTitle
	prEnterDescription
	prSelectBase
	prReview
	prConfigUpstream
	prCreating
	prDone
	prEditTitle
	prEditDescription
	prEditing
)

type PRModel struct {
	theme            *styles.Theme
	aiManager        *ai.Manager
	ghClient         *gh.Client
	state            prState
	title            components.TextInputModel
	desc             components.TextAreaModel
	upstreamInput    components.TextInputModel
	baseBranch       string
	headBranch       string
	branches         []string
	draft            bool
	selectedBase     int
	loading          bool
	generating       bool
	prURL            string
	width            int
	height           int
	err              error
	spinner          spinner.Model
	generationStart  time.Time
	elapsedTime      time.Duration
	targetRemote     string
	targetRepoNWO    string
	remotes          []string
	remoteURLs       map[string]string
	prs              []gh.PRInfo
	selectedPRIdx    int
	selectedPRDetail *gh.PRDetails
	loadingDetails   bool
	loadingPRs       bool
}

func NewPRModel(theme *styles.Theme, aiManager *ai.Manager) PRModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = theme.Accent

	return PRModel{
		theme:         theme,
		aiManager:     aiManager,
		ghClient:      gh.NewClient(),
		state:         prDashboard,
		title:         components.NewTextInputModel(theme, "PR title..."),
		desc:          components.NewTextAreaModel(theme, "PR description..."),
		upstreamInput: components.NewTextInputModel(theme, "owner/repo or github-url..."),
		baseBranch:    "main",
		spinner:       s,
		remoteURLs:    make(map[string]string),
		loadingPRs:    true,
	}
}

func (m PRModel) Init() tea.Cmd {
	return tea.Batch(
		m.loadBranches,
		m.loadCurrentBranch,
		m.loadRemotes,
		m.loadPRs,
		m.spinner.Tick,
	)
}

func (m PRModel) loadBranches() tea.Msg {
	branches, err := git.GetRemoteBranches()
	if err != nil {
		return branchesLoadedMsg{err: err}
	}
	return branchesLoadedMsg{branches: branches}
}

func (m PRModel) loadCurrentBranch() tea.Msg {
	branch, err := git.GetCurrentBranch()
	if err != nil {
		return currentBranchMsg{err: err}
	}
	return currentBranchMsg{branch: branch}
}

func (m PRModel) loadRemotes() tea.Msg {
	remotes, err := git.GetRemotes()
	if err != nil {
		return remotesLoadedMsg{err: err}
	}

	urls := make(map[string]string)
	for _, r := range remotes {
		rawURL, err := git.GetRemoteURL(r)
		if err == nil {
			nwo := git.ParseNWOFromURL(rawURL)
			if nwo != "" {
				urls[r] = nwo
			}
		}
	}

	return remotesLoadedMsg{remotes: remotes, remoteURLs: urls}
}

type remotesLoadedMsg struct {
	remotes    []string
	remoteURLs map[string]string
	err        error
}

type upstreamConfiguredMsg struct {
	remotesMsg remotesLoadedMsg
	err        error
}

func (m PRModel) addUpstreamRemote(urlStr string) tea.Cmd {
	return func() tea.Msg {
		err := git.AddRemote("upstream", urlStr)
		if err != nil {
			return upstreamConfiguredMsg{err: err}
		}
		remotesMsg := m.loadRemotes()
		return upstreamConfiguredMsg{remotesMsg: remotesMsg.(remotesLoadedMsg)}
	}
}

type branchesLoadedMsg struct {
	branches []string
	err      error
}

type currentBranchMsg struct {
	branch string
	err    error
}

type prResultMsg struct {
	prURL string
	err   error
}

type prsLoadedMsg struct {
	prs []gh.PRInfo
	err error
}

func (m PRModel) loadPRs() tea.Msg {
	client := gh.NewClient()
	prs, err := client.ListPRs(30)
	return prsLoadedMsg{prs: prs, err: err}
}

type prDetailsLoadedMsg struct {
	details *gh.PRDetails
	number  int
	err     error
}

func (m PRModel) loadPRDetails(number int) tea.Cmd {
	return func() tea.Msg {
		client := gh.NewClient()
		details, err := client.GetPRDetails(number)
		return prDetailsLoadedMsg{details: details, number: number, err: err}
	}
}

type prEditResultMsg struct {
	err error
}

func (m PRModel) doEditPR() tea.Msg {
	if m.selectedPRIdx >= 0 && m.selectedPRIdx < len(m.prs) {
		pr := m.prs[m.selectedPRIdx]
		err := m.ghClient.EditPR(pr.Number, m.title.Value(), m.desc.Value())
		return prEditResultMsg{err: err}
	}
	return prEditResultMsg{err: fmt.Errorf("no pull request selected")}
}

type descGeneratedMsg struct {
	content string
	err     error
}

type titleGeneratedMsg struct {
	content string
	err     error
}

func (m PRModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		var spinCmd tea.Cmd
		m.spinner, spinCmd = m.spinner.Update(msg)
		if m.generating {
			m.elapsedTime = time.Since(m.generationStart)
		}
		return m, spinCmd

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.title.SetWidth(m.width - 10)
		m.desc.SetWidth(m.width - 10)
		m.updateDescHeight()
		m.upstreamInput.SetWidth(m.width - 10)

	case prsLoadedMsg:
		m.loadingPRs = false
		m.prs = msg.prs
		m.err = msg.err
		if msg.err == nil && len(m.prs) > 0 {
			m.selectedPRIdx = 0
			m.loadingDetails = true
			return m, m.loadPRDetails(m.prs[0].Number)
		}

	case prDetailsLoadedMsg:
		if msg.err == nil && len(m.prs) > 0 && m.selectedPRIdx < len(m.prs) && m.prs[m.selectedPRIdx].Number == msg.number {
			m.selectedPRDetail = msg.details
			m.loadingDetails = false
		}

	case prEditResultMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			m.state = prEditDescription
		} else {
			m.state = prDashboard
			m.loadingPRs = true
			m.selectedPRDetail = nil
			return m, m.loadPRs
		}

	case branchesLoadedMsg:
		if msg.err == nil {
			m.branches = msg.branches
			// Find default branch
			for i, b := range m.branches {
				if strings.Contains(b, "/main") || strings.Contains(b, "/master") {
					m.selectedBase = i
					m.baseBranch = b
					break
				}
			}
		}

	case currentBranchMsg:
		if msg.err == nil {
			m.headBranch = msg.branch
		}

	case remotesLoadedMsg:
		if msg.err == nil {
			m.remotes = msg.remotes
			m.remoteURLs = msg.remoteURLs

			// Default targetRemote to "upstream" if it exists, otherwise "origin"
			hasUpstream := false
			for _, r := range m.remotes {
				if r == "upstream" {
					hasUpstream = true
					break
				}
			}
			if hasUpstream {
				m.targetRemote = "upstream"
			} else if len(m.remotes) > 0 {
				m.targetRemote = m.remotes[0]
			} else {
				m.targetRemote = "origin"
			}

			if nwo, ok := m.remoteURLs[m.targetRemote]; ok {
				m.targetRepoNWO = nwo
			}
		}

	case upstreamConfiguredMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.remotes = msg.remotesMsg.remotes
			m.remoteURLs = msg.remotesMsg.remoteURLs
			m.targetRemote = "upstream"
			if nwo, ok := m.remoteURLs[m.targetRemote]; ok {
				m.targetRepoNWO = nwo
			}
			m.state = prReview
		}

	case descGeneratedMsg:
		m.generating = false
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.desc.SetValue(msg.content)
			m.updateDescHeight()
		}

	case titleGeneratedMsg:
		m.generating = false
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.title.SetValue(msg.content)
		}

	case prResultMsg:
		m.state = prDone
		m.loading = false
		m.prURL = msg.prURL
		m.err = msg.err

	case tea.KeyMsg:
		if m.generating || m.loading {
			return m, nil
		}
		m.err = nil

		if m.state == prDashboard {
			switch msg.String() {
			case "up", "k":
				if m.selectedPRIdx > 0 {
					m.selectedPRIdx--
					m.loadingDetails = true
					m.selectedPRDetail = nil
					return m, m.loadPRDetails(m.prs[m.selectedPRIdx].Number)
				}
			case "down", "j":
				if m.selectedPRIdx < len(m.prs)-1 {
					m.selectedPRIdx++
					m.loadingDetails = true
					m.selectedPRDetail = nil
					return m, m.loadPRDetails(m.prs[m.selectedPRIdx].Number)
				}
			case "c":
				m.state = prEnterTitle
				m.title.SetValue("")
				m.desc.SetValue("")
				m.err = nil
			case "e":
				if len(m.prs) > 0 && m.selectedPRDetail != nil {
					m.state = prEditTitle
					m.title.SetValue(m.selectedPRDetail.Title)
					m.desc.SetValue(m.selectedPRDetail.Body)
					m.updateDescHeight()
					m.err = nil
				}
			case "o":
				if len(m.prs) > 0 && m.selectedPRIdx < len(m.prs) {
					_ = exec.Command("gh", "pr", "view", fmt.Sprintf("%d", m.prs[m.selectedPRIdx].Number), "--web").Start()
				}
			case "r":
				m.loadingPRs = true
				m.selectedPRDetail = nil
				return m, m.loadPRs
			case "b", "esc":
				return m, func() tea.Msg {
					return Navigate(ScreenHome)
				}
			}
			return m, nil
		}

		if m.state == prEditTitle {
			switch msg.String() {
			case "tab", "enter":
				m.state = prEditDescription
			case "esc":
				m.state = prDashboard
			case "ctrl+g":
				m.generating = true
				m.generationStart = time.Now()
				m.elapsedTime = 0
				return m, m.generateTitle
			case "ctrl+s":
				m.loading = true
				return m, m.doEditPR
			}
			var editCmd tea.Cmd
			m.title, editCmd = m.title.Update(msg)
			return m, editCmd
		}

		if m.state == prEditDescription {
			switch msg.String() {
			case "tab":
				m.state = prEditTitle
			case "esc":
				m.state = prEditTitle
			case "ctrl+g":
				m.generating = true
				m.generationStart = time.Now()
				m.elapsedTime = 0
				return m, m.generateDescription
			case "ctrl+s":
				m.loading = true
				return m, m.doEditPR
			}
			var editCmd tea.Cmd
			m.desc, editCmd = m.desc.Update(msg)
			m.updateDescHeight()
			return m, editCmd
		}

		switch msg.String() {
		case "tab":
			switch m.state {
			case prEnterTitle:
				if m.title.Value() != "" {
					m.state = prEnterDescription
				}
			case prEnterDescription:
				m.state = prSelectBase
			case prSelectBase:
				m.state = prEnterTitle
			}

		case "shift+tab":
			switch m.state {
			case prEnterDescription:
				m.state = prEnterTitle
			case prSelectBase:
				m.state = prEnterDescription
			case prReview:
				m.state = prSelectBase
			}

		case "enter":
			switch m.state {
			case prEnterTitle:
				if m.title.Value() != "" {
					m.state = prEnterDescription
				}
			case prSelectBase:
				m.state = prReview
			case prReview:
				m.state = prCreating
				return m, m.createPR
			case prConfigUpstream:
				if m.upstreamInput.Value() != "" {
					m.loading = true
					return m, m.addUpstreamRemote(m.upstreamInput.Value())
				}
			case prDone:
				// Reset
				m.state = prEnterTitle
				m.title.SetValue("")
				m.desc.SetValue("")
				m.prURL = ""
			}

		case "b":
			if m.state == prSelectBase || m.state == prReview || m.state == prDone {
				return m, func() tea.Msg {
					return Navigate(ScreenHome)
				}
			}

		case "esc":
			if m.state == prConfigUpstream {
				m.state = prReview
				return m, nil
			}
			if m.state == prEnterTitle || m.state == prEnterDescription || m.state == prSelectBase || m.state == prReview {
				m.state = prDashboard
				return m, nil
			}

		case "u":
			if m.state == prReview {
				hasUpstream := false
				for _, r := range m.remotes {
					if r == "upstream" {
						hasUpstream = true
						break
					}
				}
				if !hasUpstream {
					m.state = prConfigUpstream
					m.upstreamInput.SetValue("")
					return m, nil
				}
			}

		case "ctrl+g":
			if m.state == prEnterTitle && !m.generating {
				m.generating = true
				m.generationStart = time.Now()
				m.elapsedTime = 0
				return m, m.generateTitle
			}
			if m.state == prEnterDescription && !m.generating {
				m.generating = true
				m.generationStart = time.Now()
				m.elapsedTime = 0
				return m, m.generateDescription
			}

		case "up", "k":
			if m.state == prSelectBase && m.selectedBase > 0 {
				m.selectedBase--
				m.baseBranch = m.branches[m.selectedBase]
			}

		case "down", "j":
			if m.state == prSelectBase && m.selectedBase < len(m.branches)-1 {
				m.selectedBase++
				m.baseBranch = m.branches[m.selectedBase]
			}

		case "d":
			if m.state == prReview {
				m.draft = !m.draft
			}

		case "t":
			if m.state == prReview && len(m.remotes) > 1 {
				idx := -1
				for i, r := range m.remotes {
					if r == m.targetRemote {
						idx = i
						break
					}
				}
				if idx != -1 {
					nextIdx := (idx + 1) % len(m.remotes)
					m.targetRemote = m.remotes[nextIdx]
					if nwo, ok := m.remoteURLs[m.targetRemote]; ok {
						m.targetRepoNWO = nwo
					}
				}
			}
		}
	}

	// Update input components
	switch m.state {
	case prEnterTitle:
		m.title, cmd = m.title.Update(msg)
	case prEnterDescription:
		m.desc, cmd = m.desc.Update(msg)
		m.updateDescHeight()
	case prConfigUpstream:
		m.upstreamInput, cmd = m.upstreamInput.Update(msg)
	}

	return m, cmd
}

func (m PRModel) generateDescription() tea.Msg {
	base := m.baseBranch
	head := "HEAD"
	if (m.state == prEditDescription || m.state == prEditing) && m.selectedPRDetail != nil {
		base = m.selectedPRDetail.BaseRefName
		head = m.selectedPRDetail.HeadRefName
	}
	commits, _ := git.GetCommitsBetween(base, head, 15)
	var commitStrs []string
	for _, c := range commits {
		commitStrs = append(commitStrs, c.Message)
	}

	diffSummary, _ := git.GetDiffStat(base, head)
	diffSummary = strings.TrimSpace(diffSummary)

	if len(commitStrs) == 0 && diffSummary == "" {
		return descGeneratedMsg{content: "No changes detected between base and head branches. Please commit changes on a feature branch first."}
	}

	resp, err := m.aiManager.ChatWithOptions(context.Background(), []ai.Message{
		{Role: "user", Content: ai.GeneratePRDescriptionPrompt(strings.Join(commitStrs, "\n"), diffSummary)},
	}, map[string]interface{}{
		"max_tokens":  800,
		"num_predict": 800,
		"temperature": 0.3,
	})
	if err != nil {
		return descGeneratedMsg{err: err}
	}

	content := strings.TrimSpace(resp.Content)
	if strings.HasPrefix(content, "```markdown") {
		content = strings.TrimPrefix(content, "```markdown")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	} else if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	}

	return descGeneratedMsg{content: content}
}

func (m PRModel) generateTitle() tea.Msg {
	base := m.baseBranch
	head := "HEAD"
	if (m.state == prEditTitle || m.state == prEditing) && m.selectedPRDetail != nil {
		base = m.selectedPRDetail.BaseRefName
		head = m.selectedPRDetail.HeadRefName
	}
	commits, _ := git.GetCommitsBetween(base, head, 15)
	var commitStrs []string
	for _, c := range commits {
		commitStrs = append(commitStrs, c.Message)
	}

	diffSummary, _ := git.GetDiffStat(base, head)
	diffSummary = strings.TrimSpace(diffSummary)

	if len(commitStrs) == 0 && diffSummary == "" {
		return titleGeneratedMsg{content: "No changes detected"}
	}

	resp, err := m.aiManager.ChatWithOptions(context.Background(), []ai.Message{
		{Role: "user", Content: ai.GeneratePRTitlePrompt(strings.Join(commitStrs, "\n"), diffSummary)},
	}, map[string]interface{}{
		"max_tokens":  80,
		"num_predict": 80,
		"temperature": 0.2,
	})
	if err != nil {
		return titleGeneratedMsg{err: err}
	}

	content := strings.Trim(strings.TrimSpace(resp.Content), "\"`'")
	if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	}

	return titleGeneratedMsg{content: content}
}

func (m PRModel) createPR() tea.Msg {
	headBranchArg := m.headBranch
	if originNWO, ok := m.remoteURLs["origin"]; ok {
		if m.targetRepoNWO != "" && m.targetRepoNWO != originNWO {
			parts := strings.Split(originNWO, "/")
			if len(parts) > 0 && parts[0] != "" {
				headBranchArg = parts[0] + ":" + m.headBranch
			}
		}
	}

	pr, err := m.ghClient.CreatePR(m.title.Value(), m.desc.Value(), extractBranchName(m.baseBranch), headBranchArg, m.draft, m.targetRepoNWO)
	if err != nil {
		return prResultMsg{err: err}
	}
	return prResultMsg{prURL: pr.URL}
}

func extractBranchName(remoteBranch string) string {
	parts := strings.Split(remoteBranch, "/")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return remoteBranch
}

func (m PRModel) View() string {
	var b strings.Builder

	if m.state != prDashboard && m.state != prEditTitle && m.state != prEditDescription && m.state != prEditing {
		b.WriteString(m.theme.Title.Render("Create Pull Request"))
		b.WriteString("\n\n")
	}

	if m.err != nil {
		b.WriteString(m.theme.Error.Render("Error: " + m.err.Error()))
		b.WriteString("\n\n")
	}

	switch m.state {
	case prDashboard:
		b.WriteString(m.theme.Title.Render("Pull Requests Dashboard"))
		b.WriteString("\n\n")

		if m.loadingPRs {
			b.WriteString(m.theme.Muted.Render("Loading pull requests..."))
			return lipgloss.NewStyle().Width(m.width).Height(m.height).Render(b.String())
		}

		if len(m.prs) == 0 {
			b.WriteString(m.theme.Muted.Render("No open pull requests found."))
			b.WriteString("\n\n")
			b.WriteString(m.theme.Help.Render("c Create PR   r Refresh   b Back"))
			return lipgloss.NewStyle().Width(m.width).Height(m.height).Render(b.String())
		}

		// List viewport scrolling
		reservedLines := 6
		visibleCount := m.height - reservedLines
		if visibleCount < 3 {
			visibleCount = 3
		}

		start := 0
		if m.selectedPRIdx >= visibleCount {
			start = m.selectedPRIdx - visibleCount + 1
		}
		end := start + visibleCount
		if end > len(m.prs) {
			end = len(m.prs)
		}
		if end-start < visibleCount && start > 0 {
			start = end - visibleCount
			if start < 0 {
				start = 0
			}
		}

		leftWidth := 34
		var leftList strings.Builder
		leftList.WriteString(m.theme.Header.Render("Open Pull Requests"))
		leftList.WriteString("\n\n")

		for i := start; i < end; i++ {
			pr := m.prs[i]
			if i == m.selectedPRIdx {
				leftList.WriteString(m.theme.Selected.Render("> "))
			} else {
				leftList.WriteString("  ")
			}

			// Format title and number
			displayName := fmt.Sprintf("#%d %s", pr.Number, pr.Title)
			maxLen := leftWidth - 4
			if len(displayName) > maxLen {
				displayName = displayName[:maxLen-3] + "..."
			}

			if i == m.selectedPRIdx {
				leftList.WriteString(m.theme.Text.Bold(true).Render(displayName))
			} else {
				leftList.WriteString(m.theme.Text.Render(displayName))
			}
			leftList.WriteString("\n")
		}

		// Fill vertical space to match right panel height
		renderedLines := end - start
		for i := renderedLines; i < visibleCount; i++ {
			leftList.WriteString("\n")
		}

		// Right card details
		var rightCard strings.Builder
		rightWidth := m.width - leftWidth - 4
		if rightWidth < 30 {
			rightWidth = 30
		}

		if m.loadingDetails || m.selectedPRDetail == nil {
			rightCard.WriteString(m.theme.Muted.Render("Loading pull request details..."))
		} else {
			details := m.selectedPRDetail
			rightCard.WriteString(m.theme.Primary.Render(fmt.Sprintf("%s (#%d)", details.Title, details.Number)))
			rightCard.WriteString("\n\n")

			stateStyle := m.theme.Success
			if details.State == "CLOSED" {
				stateStyle = m.theme.Error
			} else if details.State == "MERGED" {
				stateStyle = m.theme.Accent
			}
			rightCard.WriteString(m.theme.Bold.Render("Status: "))
			rightCard.WriteString(stateStyle.Render(details.State))
			rightCard.WriteString("\n")

			rightCard.WriteString(m.theme.Bold.Render("Mergeable: "))
			if details.Mergeable == "MERGEABLE" {
				rightCard.WriteString(m.theme.Success.Render("Yes"))
			} else if details.Mergeable == "CONFLICTING" {
				rightCard.WriteString(m.theme.Error.Render("Conflicting"))
			} else {
				rightCard.WriteString(m.theme.Warning.Render(details.Mergeable))
			}
			rightCard.WriteString("\n\n")

			rightCard.WriteString(m.theme.Bold.Render("URL:\n"))
			rightCard.WriteString(m.theme.Accent.Render(details.URL))
			rightCard.WriteString("\n\n")

			rightCard.WriteString(m.theme.Bold.Render("Description:\n"))
			if details.Body != "" {
				wrappedDesc := lipgloss.NewStyle().Width(rightWidth - 6).Render(details.Body)
				// Limit text lines to fit viewport height
				lines := strings.Split(wrappedDesc, "\n")
				maxLines := visibleCount - 8
				if maxLines < 3 {
					maxLines = 3
				}
				if len(lines) > maxLines {
					wrappedDesc = strings.Join(lines[:maxLines], "\n") + "\n" + m.theme.Muted.Render("... (truncated)")
				}
				rightCard.WriteString(wrappedDesc)
			} else {
				rightCard.WriteString(m.theme.Muted.Render("No description provided."))
			}
		}

		rightBox := m.theme.Box.Width(rightWidth).Height(visibleCount + 2).Render(rightCard.String())

		mainContent := lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Width(leftWidth).Render(leftList.String()),
			rightBox,
		)
		b.WriteString(mainContent)
		b.WriteString("\n\n")
		b.WriteString(m.theme.Help.Render("↑/↓ Navigate   c Create   e Edit   o Open in browser   r Refresh   b Back"))

	case prEditTitle:
		b.WriteString(m.theme.Header.Render("Edit Pull Request - Title"))
		b.WriteString("\n\n")
		if m.generating {
			loadingContent := fmt.Sprintf(
				"  %s  %s\n\n  %s\n\n  %s",
				m.spinner.View(),
				m.theme.Text.Bold(true).Render("Regenerating PR title using AI..."),
				m.theme.Muted.Render("Analyzing commits and diff summary relative to base branch..."),
				m.theme.Accent.Render(fmt.Sprintf("Elapsed time: %.1fs", m.elapsedTime.Seconds())),
			)
			b.WriteString(m.theme.Box.Render(loadingContent))
			b.WriteString("\n\n")
			b.WriteString(m.theme.Help.Render("Please wait..."))
		} else {
			b.WriteString(m.title.View())
			b.WriteString("\n\n")
			b.WriteString(m.theme.Help.Render("ctrl+g AI Regenerate   ctrl+s Save   Enter/Tab Description   Esc Cancel"))
		}

	case prEditDescription:
		b.WriteString(m.theme.Header.Render("Edit Pull Request - Description"))
		b.WriteString("\n\n")
		if m.generating {
			loadingContent := fmt.Sprintf(
				"  %s  %s\n\n  %s\n\n  %s",
				m.spinner.View(),
				m.theme.Text.Bold(true).Render("Regenerating PR description using AI..."),
				m.theme.Muted.Render("Analyzing commits and diff summary relative to base branch..."),
				m.theme.Accent.Render(fmt.Sprintf("Elapsed time: %.1fs", m.elapsedTime.Seconds())),
			)
			b.WriteString(m.theme.Box.Render(loadingContent))
			b.WriteString("\n\n")
			b.WriteString(m.theme.Help.Render("Please wait..."))
		} else {
			b.WriteString(m.desc.View())
			b.WriteString("\n\n")
			b.WriteString(m.theme.Help.Render("ctrl+g AI Regenerate   ctrl+s Save   Tab Title   Esc Cancel"))
		}

	case prEditing:
		b.WriteString(m.theme.Accent.Render("Saving pull request updates..."))

	case prEnterTitle:
		b.WriteString(m.theme.Header.Render("Step 1: Enter Title (1/3)"))
		b.WriteString("\n\n")

		b.WriteString(m.theme.Bold.Render("Head: "))
		b.WriteString(m.theme.Text.Render(m.headBranch))
		b.WriteString(" → ")
		b.WriteString(m.theme.Bold.Render("Base: "))
		b.WriteString(m.theme.Muted.Render(m.baseBranch))
		b.WriteString("\n\n")

		if m.generating {
			loadingContent := fmt.Sprintf(
				"  %s  %s\n\n  %s\n\n  %s",
				m.spinner.View(),
				m.theme.Text.Bold(true).Render("Generating PR title using AI..."),
				m.theme.Muted.Render("Analyzing commits and diff summary relative to base branch..."),
				m.theme.Accent.Render(fmt.Sprintf("Elapsed time: %.1fs", m.elapsedTime.Seconds())),
			)
			b.WriteString(m.theme.Box.Render(loadingContent))
			b.WriteString("\n\n")
			b.WriteString(m.theme.Help.Render("Please wait..."))
		} else {
			b.WriteString(m.theme.Text.Render("Title:"))
			b.WriteString("\n")
			b.WriteString(m.title.View())
			b.WriteString("\n\n")
			b.WriteString(m.theme.Help.Render("ctrl+g AI Generate   Tab/Enter Next   Esc Cancel"))
		}

	case prEnterDescription:
		b.WriteString(m.theme.Header.Render("Step 2: Enter Description (2/3)"))
		b.WriteString("\n\n")

		if m.generating {
			loadingContent := fmt.Sprintf(
				"  %s  %s\n\n  %s\n\n  %s",
				m.spinner.View(),
				m.theme.Text.Bold(true).Render("Generating PR description using AI..."),
				m.theme.Muted.Render("Analyzing commits and diff summary relative to base branch..."),
				m.theme.Accent.Render(fmt.Sprintf("Elapsed time: %.1fs", m.elapsedTime.Seconds())),
			)
			b.WriteString(m.theme.Box.Render(loadingContent))
			b.WriteString("\n\n")
			b.WriteString(m.theme.Help.Render("Please wait..."))
		} else {
			b.WriteString(m.theme.Text.Render("Description:"))
			b.WriteString("\n")
			b.WriteString(m.desc.View())
			b.WriteString("\n\n")
			b.WriteString(m.theme.Help.Render("ctrl+g AI Generate   Tab Next   Shift+Tab Prev   Esc Cancel"))
		}

	case prSelectBase:
		b.WriteString(m.theme.Header.Render("Step 3: Select Base Branch (3/3)"))
		b.WriteString("\n\n")

		b.WriteString(m.theme.Bold.Render("Select target branch:"))
		b.WriteString("\n\n")

		reservedLines := 12
		visibleCount := m.height - reservedLines
		if visibleCount < 3 {
			visibleCount = 3
		}

		start := 0
		if m.selectedBase >= visibleCount {
			start = m.selectedBase - visibleCount + 1
		}
		end := start + visibleCount
		if end > len(m.branches) {
			end = len(m.branches)
		}
		if end-start < visibleCount && start > 0 {
			start = end - visibleCount
			if start < 0 {
				start = 0
			}
		}

		for i := start; i < end; i++ {
			branch := m.branches[i]
			if i == m.selectedBase {
				b.WriteString(m.theme.Selected.Render("> " + branch))
			} else {
				b.WriteString("  ")
				b.WriteString(m.theme.Text.Render(branch))
			}
			b.WriteString("\n")
		}

		b.WriteString("\n")
		b.WriteString(m.theme.Help.Render("↑/↓ Navigate   Tab Next   Shift+Tab Prev   Esc Cancel"))

	case prReview:
		b.WriteString(m.theme.Header.Render("Review & Create"))
		b.WriteString("\n\n")

		b.WriteString(m.theme.Bold.Render("Title: "))
		b.WriteString(m.theme.Text.Render(m.title.Value()))
		b.WriteString("\n\n")

		b.WriteString(m.theme.Bold.Render("Head → Base: "))
		b.WriteString(m.theme.Accent.Render(m.headBranch + " → " + extractBranchName(m.baseBranch)))
		b.WriteString("\n\n")

		b.WriteString(m.theme.Bold.Render("Target Repo: "))
		if m.targetRepoNWO != "" {
			b.WriteString(m.theme.Accent.Render(fmt.Sprintf("%s (%s)", m.targetRemote, m.targetRepoNWO)))
		} else {
			b.WriteString(m.theme.Accent.Render(m.targetRemote))
		}
		b.WriteString("\n\n")

		if m.draft {
			b.WriteString(m.theme.Warning.Render("[Draft PR]"))
		} else {
			b.WriteString(m.theme.Success.Render("[Ready for review]"))
		}
		b.WriteString("\n\n")

		hasUpstream := false
		for _, r := range m.remotes {
			if r == "upstream" {
				hasUpstream = true
				break
			}
		}

		if hasUpstream {
			if len(m.remotes) > 1 {
				b.WriteString(m.theme.Help.Render("d Toggle Draft   t Toggle Target   Enter Create   Tab Edit   Esc Cancel"))
			} else {
				b.WriteString(m.theme.Help.Render("d Toggle Draft   Enter Create   Tab Edit   Esc Cancel"))
			}
		} else {
			b.WriteString(m.theme.Help.Render("d Toggle Draft   u Set Upstream   Enter Create   Tab Edit   Esc Cancel"))
		}

	case prConfigUpstream:
		b.WriteString(m.theme.Header.Render("Configure Upstream Repository"))
		b.WriteString("\n\n")

		if m.loading {
			b.WriteString(m.theme.Muted.Render("Adding remote repository..."))
		} else {
			b.WriteString(m.theme.Text.Render("Enter upstream repository (e.g. owner/repo or github-url):"))
			b.WriteString("\n")
			b.WriteString(m.upstreamInput.View())
			b.WriteString("\n\n")
			b.WriteString(m.theme.Help.Render("Enter Confirm   Esc Cancel"))
		}

	case prCreating:
		b.WriteString(m.theme.Accent.Render("Creating pull request..."))

	case prDone:
		if m.err != nil {
			b.WriteString(m.theme.Error.Render("Failed to create PR"))
		} else {
			b.WriteString(m.theme.Success.Render("Pull request created!"))
			b.WriteString("\n\n")
			b.WriteString(m.theme.Text.Render(m.prURL))
		}
		b.WriteString("\n\n")
		b.WriteString(m.theme.Help.Render("Enter New PR   b Back"))
	}

	return lipgloss.NewStyle().Width(m.width).Height(m.height).Render(b.String())
}

func (m *PRModel) updateDescHeight() {
	text := m.desc.Value()
	lines := strings.Split(text, "\n")
	
	visualLines := 0
	width := m.desc.TextArea.Width()
	if width <= 0 {
		width = m.width - 10
	}
	if width <= 0 {
		width = 50
	}

	for _, line := range lines {
		lineLen := len(line)
		if lineLen == 0 {
			visualLines++
			continue
		}
		wrapped := (lineLen + width - 1) / width
		visualLines += wrapped
	}

	height := visualLines + 2
	if height < 4 {
		height = 4
	}
	
	maxHeight := m.height - 12
	if maxHeight < 4 {
		maxHeight = 4
	}
	if maxHeight > 16 {
		maxHeight = 16
	}
	if height > maxHeight {
		height = maxHeight
	}

	m.desc.SetHeight(height)
}
