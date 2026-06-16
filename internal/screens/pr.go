package screens

import (
	"context"
	"fmt"
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
	prEnterTitle prState = iota
	prEnterDescription
	prSelectBase
	prReview
	prConfigUpstream
	prCreating
	prDone
)

type PRModel struct {
	theme           *styles.Theme
	aiManager       *ai.Manager
	ghClient        *gh.Client
	state           prState
	title           components.TextInputModel
	desc            components.TextAreaModel
	upstreamInput   components.TextInputModel
	baseBranch      string
	headBranch      string
	branches        []string
	draft           bool
	selectedBase    int
	loading         bool
	generating      bool
	prURL           string
	width           int
	height          int
	err             error
	spinner         spinner.Model
	generationStart time.Time
	elapsedTime     time.Duration
	targetRemote    string
	targetRepoNWO   string
	remotes         []string
	remoteURLs      map[string]string
}

func NewPRModel(theme *styles.Theme, aiManager *ai.Manager) PRModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = theme.Accent

	return PRModel{
		theme:         theme,
		aiManager:     aiManager,
		ghClient:      gh.NewClient(),
		state:         prEnterTitle,
		title:         components.NewTextInputModel(theme, "PR title..."),
		desc:          components.NewTextAreaModel(theme, "PR description..."),
		upstreamInput: components.NewTextInputModel(theme, "owner/repo or github-url..."),
		baseBranch:    "main",
		spinner:       s,
		remoteURLs:    make(map[string]string),
	}
}

func (m PRModel) Init() tea.Cmd {
	return tea.Batch(m.loadBranches, m.loadCurrentBranch, m.loadRemotes, m.spinner.Tick)
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
			case prEnterDescription:
				m.state = prSelectBase
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

		case "g":
			if m.state == prEnterTitle && !m.generating && m.title.Value() == "" {
				m.generating = true
				m.generationStart = time.Now()
				m.elapsedTime = 0
				return m, m.generateTitle
			}
			if m.state == prEnterDescription && !m.generating && m.desc.Value() == "" {
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
	commits, _ := git.GetCommitsBetween(base, "HEAD", 15)
	var commitStrs []string
	for _, c := range commits {
		commitStrs = append(commitStrs, c.Message)
	}

	diffSummary, _ := git.GetDiffStat(base, "HEAD")

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
	commits, _ := git.GetCommitsBetween(base, "HEAD", 15)
	var commitStrs []string
	for _, c := range commits {
		commitStrs = append(commitStrs, c.Message)
	}

	diffSummary, _ := git.GetDiffStat(base, "HEAD")

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

	b.WriteString(m.theme.Title.Render("Create Pull Request"))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(m.theme.Error.Render("Error: " + m.err.Error()))
		b.WriteString("\n\n")
	}

	switch m.state {
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
			if m.title.Value() == "" {
				b.WriteString(m.theme.Help.Render("g AI Generate   Tab/Enter Next   b Back"))
			} else {
				b.WriteString(m.theme.Help.Render("Tab/Enter Next   b Back"))
			}
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
			if m.desc.Value() == "" {
				b.WriteString(m.theme.Help.Render("g AI Generate   Tab Next   Shift+Tab Prev   b Back"))
			} else {
				b.WriteString(m.theme.Help.Render("Tab Next   Shift+Tab Prev   b Back"))
			}
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
		b.WriteString(m.theme.Help.Render("↑/↓ Navigate   Tab Next   Shift+Tab Prev   b Back"))

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
				b.WriteString(m.theme.Help.Render("d Toggle Draft   t Toggle Target   Enter Create   Tab Edit   b Back"))
			} else {
				b.WriteString(m.theme.Help.Render("d Toggle Draft   Enter Create   Tab Edit   b Back"))
			}
		} else {
			b.WriteString(m.theme.Help.Render("d Toggle Draft   u Set Upstream   Enter Create   Tab Edit   b Back"))
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
	
	maxHeight := 10
	if m.height > 18 {
		maxHeight = m.height - 12
		if maxHeight > 16 {
			maxHeight = 16
		}
	}
	if height > maxHeight {
		height = maxHeight
	}

	m.desc.SetHeight(height)
}
