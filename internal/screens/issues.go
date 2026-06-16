package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/KDM-cli/ghx/internal/gh"
	"github.com/KDM-cli/ghx/styles"
)

type IssuesModel struct {
	theme    *styles.Theme
	client   *gh.Client
	issues   []gh.IssueInfo
	selected int
	loading  bool
	width    int
	height   int
	err      error
}

func NewIssuesModel(theme *styles.Theme) IssuesModel {
	return IssuesModel{
		theme:   theme,
		client:  gh.NewClient(),
		loading: true,
	}
}

func (m IssuesModel) Init() tea.Cmd {
	return m.loadIssues
}

func (m IssuesModel) loadIssues() tea.Msg {
	client := gh.NewClient()
	issues, err := client.ListIssues(20)
	if err != nil {
		return issuesLoadedMsg{err: err}
	}
	return issuesLoadedMsg{issues: issues}
}

type issuesLoadedMsg struct {
	issues []gh.IssueInfo
	err    error
}

func (m IssuesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case issuesLoadedMsg:
		m.loading = false
		m.issues = msg.issues
		m.err = msg.err

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.issues)-1 {
				m.selected++
			}
		case "r":
			m.loading = true
			return m, m.loadIssues
		case "b":
			return m, func() tea.Msg {
				return Navigate(ScreenHome)
			}
		}
	}

	return m, nil
}

func (m IssuesModel) View() string {
	var b strings.Builder

	b.WriteString(m.theme.Title.Render("Issues"))
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(m.theme.Muted.Render("Loading issues..."))
		return b.String()
	}

	if m.err != nil {
		b.WriteString(m.theme.Error.Render("Error: " + m.err.Error()))
		b.WriteString("\n\n")
		b.WriteString(m.theme.Help.Render("r Retry   b Back"))
		return b.String()
	}

	if len(m.issues) == 0 {
		b.WriteString(m.theme.Muted.Render("No issues found"))
		b.WriteString("\n\n")
		b.WriteString(m.theme.Help.Render("b Back"))
		return b.String()
	}

	// Dynamic list viewport logic
	reservedLines := 6
	visibleCount := m.height - reservedLines
	if visibleCount < 3 {
		visibleCount = 3
	}

	start := 0
	if m.selected >= visibleCount {
		start = m.selected - visibleCount + 1
	}
	end := start + visibleCount
	if end > len(m.issues) {
		end = len(m.issues)
	}
	if end-start < visibleCount && start > 0 {
		start = end - visibleCount
		if start < 0 {
			start = 0
		}
	}

	for i := start; i < end; i++ {
		issue := m.issues[i]
		if i == m.selected {
			b.WriteString(m.theme.Selected.Render("> "))
		} else {
			b.WriteString("  ")
		}

		stateIcon := "○"
		if issue.State == "closed" {
			stateIcon = "●"
		}

		// Truncate title to fit screen width
		maxTitleLen := m.width - 15
		displayTitle := issue.Title
		if maxTitleLen > 10 && len(displayTitle) > maxTitleLen {
			displayTitle = displayTitle[:maxTitleLen-3] + "..."
		}

		b.WriteString(fmt.Sprintf("#%d %s %s\n", issue.Number, stateIcon, displayTitle))
	}

	b.WriteString("\n")
	b.WriteString(m.theme.Help.Render("↑/↓ Navigate   r Refresh   b Back"))

	return lipgloss.NewStyle().Width(m.width).Height(m.height).Render(b.String())
}
