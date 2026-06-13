package screens

import (
	"context"
	"fmt"
	"strings"

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
	prCreating
	prDone
)

type PRModel struct {
	theme        *styles.Theme
	aiManager    *ai.Manager
	ghClient     *gh.Client
	state        prState
	title        components.TextInputModel
	desc         components.TextAreaModel
	baseBranch   string
	headBranch   string
	branches     []string
	draft        bool
	selectedBase int
	loading      bool
	generating   bool
	prURL        string
	width        int
	height       int
	err          error
}

func NewPRModel(theme *styles.Theme, aiManager *ai.Manager) PRModel {
	return PRModel{
		theme:      theme,
		aiManager:  aiManager,
		ghClient:   gh.NewClient(),
		state:      prEnterTitle,
		title:      components.NewTextInputModel(theme, "PR title..."),
		desc:       components.NewTextAreaModel(theme, "PR description..."),
		baseBranch: "main",
	}
}

func (m PRModel) Init() tea.Cmd {
	return tea.Batch(m.loadBranches, m.loadCurrentBranch)
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

func (m PRModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.title.SetWidth(m.width - 10)
		m.desc.SetWidth(m.width - 10)
		m.desc.SetHeight(8)

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

	case descGeneratedMsg:
		m.generating = false
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.desc.SetValue(msg.content)
		}

	case prResultMsg:
		m.state = prDone
		m.loading = false
		m.prURL = msg.prURL
		m.err = msg.err

	case tea.KeyMsg:
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
			case prDone:
				// Reset
				m.state = prEnterTitle
				m.title.SetValue("")
				m.desc.SetValue("")
				m.prURL = ""
			}

		case "g":
			if m.state == prEnterDescription && !m.generating {
				m.generating = true
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
		}
	}

	// Update input components
	switch m.state {
	case prEnterTitle:
		m.title, cmd = m.title.Update(msg)
	case prEnterDescription:
		m.desc, cmd = m.desc.Update(msg)
	}

	return m, cmd
}

func (m PRModel) generateDescription() tea.Msg {
	commits, _ := git.GetLog(10)
	var commitStrs []string
	for _, c := range commits {
		commitStrs = append(commitStrs, c.Message)
	}

	resp, err := m.aiManager.Chat(context.Background(), []ai.Message{
		{Role: "user", Content: ai.GeneratePRDescriptionPrompt(strings.Join(commitStrs, "\n"), "")},
	})
	if err != nil {
		return descGeneratedMsg{err: err}
	}

	return descGeneratedMsg{content: resp.Content}
}

func (m PRModel) createPR() tea.Msg {
	pr, err := m.ghClient.CreatePR(m.title.Value(), m.desc.Value(), extractBranchName(m.baseBranch), m.headBranch, m.draft)
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

		b.WriteString(m.theme.Text.Render("Title:"))
		b.WriteString("\n")
		b.WriteString(m.title.View())
		b.WriteString("\n\n")
		b.WriteString(m.theme.Help.Render("Tab/Enter Next   b Back"))

	case prEnterDescription:
		b.WriteString(m.theme.Header.Render("Step 2: Enter Description (2/3)"))
		b.WriteString("\n\n")

		b.WriteString(m.theme.Text.Render("Description:"))
		b.WriteString("\n")
		b.WriteString(m.desc.View())
		b.WriteString("\n\n")

		if m.generating {
			b.WriteString(m.theme.Accent.Render("Generating with AI..."))
		} else {
			b.WriteString(m.theme.Help.Render("g AI Generate   Tab Next   Shift+Tab Prev   b Back"))
		}

	case prSelectBase:
		b.WriteString(m.theme.Header.Render("Step 3: Select Base Branch (3/3)"))
		b.WriteString("\n\n")

		b.WriteString(m.theme.Bold.Render("Select target branch:"))
		b.WriteString("\n\n")

		for i, branch := range m.branches {
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

		if m.draft {
			b.WriteString(m.theme.Warning.Render("[Draft PR]"))
		} else {
			b.WriteString(m.theme.Success.Render("[Ready for review]"))
		}
		b.WriteString("\n\n")

		b.WriteString(m.theme.Help.Render("d Toggle Draft   Enter Create   Tab Edit   b Back"))

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
