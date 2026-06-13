package app

import (
	"fmt"
	"strings"

	"github.com/utkarshpatrikar/ghx/internal/components"
	"github.com/utkarshpatrikar/ghx/internal/styles"
)

func (m Model) View() string {
	if m.state.Width == 0 {
		return "Starting ghx..."
	}

	var body string
	switch m.state.Screen {
	case ScreenHome:
		body = m.homeView()
	case ScreenStatus:
		body = m.statusView()
	case ScreenCommit:
		body = m.commitView()
	case ScreenDiff:
		body = m.diffView()
	case ScreenPush:
		body = m.pushView()
	case ScreenPR:
		body = m.prView()
	case ScreenIssues:
		body = m.issuesView()
	case ScreenRepos:
		body = m.reposView()
	case ScreenAI:
		body = m.aiView()
	case ScreenHistory:
		body = m.historyView()
	case ScreenSettings:
		body = m.settingsView()
	default:
		body = m.homeView()
	}

	return styles.App.Width(viewWidth(m.state.Width)).Render(body)
}

func (m Model) homeView() string {
	header := styles.Header.Render("ghx - AI-Powered GitHub Assistant")
	body := m.state.Menu.View()
	bar := components.StatusBar{
		Provider: m.state.Config.AI.ActiveProvider,
		Model:    m.state.Config.ActiveProvider().Model,
		Help:     "up/down Navigate  enter Select  q Quit",
		Message:  m.state.Message,
	}.View()

	return styles.Panel.Render(strings.Join([]string{header, body, bar}, "\n"))
}

func (m Model) statusView() string {
	header := styles.Header.Render("Git Status")
	content := m.statusContent()

	bar := components.StatusBar{
		Provider: m.state.Config.AI.ActiveProvider,
		Model:    m.state.Config.ActiveProvider().Model,
		Help:     "r Refresh  b Back  q Quit",
		Message:  m.state.Message,
	}.View()

	return styles.Panel.Render(strings.Join([]string{header, content, bar}, "\n\n"))
}

func (m Model) commitView() string {
	lines := []string{
		styles.Subtle.Render("Files: up/down move  space select  tab message"),
		m.fileListView(true),
		"",
		m.inputLine("Commit message", m.state.CommitMessage, m.state.CommitFocus == 1),
		"",
		m.outputView(),
	}
	return m.workflowPanel("Commit", strings.Join(lines, "\n"), "g Generate  enter Commit  ctrl+u Clear  r Refresh  b Back")
}

func (m Model) diffView() string {
	lines := []string{
		styles.Subtle.Render("Files: up/down choose  enter/r load diff  pgup/pgdown scroll"),
		m.fileListView(false),
		"",
		styles.Header.Render("Patch"),
		m.diffTextView(),
	}
	return m.workflowPanel("Diff", strings.Join(lines, "\n"), "enter Load  pgup/pgdown Scroll  r Refresh  b Back")
}

func (m Model) pushView() string {
	confirm := "Press p to confirm push."
	if m.state.PushConfirm {
		confirm = styles.Error.Render("Press p again to run git push.")
	}
	content := strings.Join([]string{
		fmt.Sprintf("Branch: %s", m.branchLabel()),
		m.remoteLines(),
		"",
		confirm,
		"",
		m.outputView(),
	}, "\n")
	return m.workflowPanel("Push", content, "p Push  r Refresh  b Back")
}

func (m Model) prView() string {
	var content string
	var help string
	if m.state.BranchInputActive {
		content = strings.Join([]string{
			fmt.Sprintf("Head branch: %s", m.branchLabel()),
			"",
			styles.Header.Render("Create & Checkout New Branch"),
			m.inputLine("Branch name", m.state.NewBranchName, true),
			"",
			m.outputView(),
		}, "\n")
		help = "enter Confirm  esc Cancel"
	} else {
		content = strings.Join([]string{
			fmt.Sprintf("Head branch: %s", m.branchLabel()),
			m.inputLine("Title", m.state.PRTitle, m.state.PRFocus == 0),
			"",
			styles.Header.Render("Description"),
			m.textAreaView(m.state.PRBody, m.state.PRFocus == 1),
			"",
			m.outputView(),
		}, "\n")
		help = "tab Field  g Generate  enter Create  c New Branch  ctrl+u Clear  r Refresh  b Back"
	}
	return m.workflowPanel("Create Pull Request", content, help)
}

func (m Model) issuesView() string {
	lines := []string{styles.Subtle.Render("Uses GitHub CLI: gh issue list --limit 10"), ""}
	if m.state.Loading {
		lines = append(lines, styles.Muted.Render("Loading issues..."))
	} else if m.state.Err != nil {
		lines = append(lines, styles.Error.Render(m.state.Err.Error()))
	} else if len(m.state.Issues) == 0 {
		lines = append(lines, styles.Muted.Render("No issues loaded. Press r."))
	} else {
		lines = append(lines, m.state.Issues...)
	}
	return m.workflowPanel("Issues", strings.Join(lines, "\n"), "r Refresh  enter Reload  b Back")
}

func (m Model) reposView() string {
	lines := []string{
		fmt.Sprintf("Branch: %s", m.branchLabel()),
		"Remotes:",
	}
	if len(m.state.RepoInfo.Remotes) == 0 {
		lines = append(lines, "  none detected")
	} else {
		for _, remote := range m.state.RepoInfo.Remotes {
			lines = append(lines, "  "+remote)
		}
	}
	lines = append(lines, "", m.statusSummary())
	return m.workflowPanel("Repos", strings.Join(lines, "\n"), "r Refresh  b Back  q Quit")
}

func (m Model) aiView() string {
	provider := m.state.Config.ActiveProvider()
	content := strings.Join([]string{
		fmt.Sprintf("Active provider: %s", m.state.Config.AI.ActiveProvider),
		fmt.Sprintf("Model: %s", emptyFallback(provider.Model, "not set")),
		fmt.Sprintf("Host: %s", emptyFallback(provider.Host, "not required")),
		"",
		m.inputLine("Ask", m.state.AIInput, true),
		"",
		styles.Header.Render("Response"),
		emptyFallback(m.state.AIResponse, styles.Muted.Render("Ask about status, branch, or commit messages.")),
	}, "\n")
	return m.workflowPanel("AI Chat", content, "enter Ask  ctrl+u Clear  b Back")
}

func (m Model) historyView() string {
	content := strings.Join([]string{
		"History storage is planned for .ghx/ghx.db.",
		"",
		"Recent in-session context:",
		"  " + emptyFallback(m.state.Message, "No actions yet"),
		"",
		"Next action: add SQLite persistence for commands and AI chats.",
	}, "\n")
	return m.workflowPanel("History", content, "b Back  q Quit")
}

func (m Model) settingsView() string {
	provider := m.state.Config.ActiveProvider()
	content := strings.Join([]string{
		fmt.Sprintf("Active provider: %s", m.state.Config.AI.ActiveProvider),
		fmt.Sprintf("Model: %s", emptyFallback(provider.Model, "not set")),
		fmt.Sprintf("Host: %s", emptyFallback(provider.Host, "not required")),
		fmt.Sprintf("Theme: %s", m.state.Config.UI.Theme),
		fmt.Sprintf("AI suggestions: %t", m.state.Config.UI.ShowAISuggestions),
		"",
		m.outputView(),
	}, "\n")
	return m.workflowPanel("Settings", content, "n Next provider  s Save  b Back")
}

func (m Model) workflowPanel(title string, content string, help string) string {
	header := styles.Header.Render(title)
	bar := components.StatusBar{
		Provider: m.state.Config.AI.ActiveProvider,
		Model:    m.state.Config.ActiveProvider().Model,
		Help:     help,
		Message:  m.state.Message,
	}.View()
	return styles.Panel.Render(strings.Join([]string{header, content, bar}, "\n\n"))
}

func (m Model) workflowContent(_ string, notes []string) string {
	lines := make([]string, 0, len(notes)+6)
	lines = append(lines, notes...)
	lines = append(lines, "", m.statusSummary(), "", m.statusContent())
	return strings.Join(lines, "\n")
}

func (m Model) fileListView(selectable bool) string {
	if m.state.Loading {
		return styles.Muted.Render("Loading files...")
	}
	if m.state.Err != nil {
		return styles.Error.Render(m.state.Err.Error())
	}
	if len(m.state.StatusFiles) == 0 {
		return styles.Success.Render("No changed files.")
	}

	lines := make([]string, 0, len(m.state.StatusFiles))
	for i, file := range m.state.StatusFiles {
		cursor := "  "
		if i == m.state.FileCursor {
			cursor = styles.Cursor.Render("> ")
		}
		check := ""
		if selectable {
			check = "[ ] "
			if m.state.SelectedFiles[file.Path] {
				check = "[x] "
			}
		}
		line := fmt.Sprintf("%s%s%s  %s", cursor, check, styles.Badge.Render(file.ShortStatus()), file.Path)
		if i == m.state.FileCursor {
			line = styles.Selected.Render(line)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func (m Model) inputLine(label string, value string, focused bool) string {
	cursor := " "
	if focused {
		cursor = styles.Cursor.Render("|")
	}
	if value == "" {
		value = styles.Muted.Render("empty")
	}
	return fmt.Sprintf("%s: %s%s", label, value, cursor)
}

func (m Model) textAreaView(value string, focused bool) string {
	if value == "" {
		value = styles.Muted.Render("empty")
	}
	cursor := ""
	if focused {
		cursor = styles.Cursor.Render("|")
	}
	return value + cursor
}

func (m Model) diffTextView() string {
	if m.state.Loading {
		return styles.Muted.Render("Loading diff...")
	}
	if m.state.Err != nil {
		return styles.Error.Render(m.state.Err.Error())
	}
	text := m.state.DiffText
	if strings.TrimSpace(text) == "" {
		text = "Choose a file and press enter to load its diff."
	}
	lines := strings.Split(text, "\n")
	if m.state.DiffScroll >= len(lines) {
		m.state.DiffScroll = maxView(0, len(lines)-1)
	}
	end := m.state.DiffScroll + 18
	if end > len(lines) {
		end = len(lines)
	}
	return strings.Join(lines[m.state.DiffScroll:end], "\n")
}

func (m Model) outputView() string {
	if m.state.Loading {
		return styles.Muted.Render("Running...")
	}
	if m.state.Output == "" {
		return styles.Muted.Render("No command output yet.")
	}
	if m.state.Err != nil {
		return styles.Error.Render(m.state.Output)
	}
	return styles.Success.Render(m.state.Output)
}

func (m Model) remoteLines() string {
	if len(m.state.RepoInfo.Remotes) == 0 {
		return "Remote: none detected"
	}
	return "Remote: " + strings.Join(m.state.RepoInfo.Remotes, "\n        ")
}

func (m Model) statusContent() string {
	if m.state.Loading {
		return styles.Muted.Render("Loading...")
	}
	if m.state.Err != nil {
		return styles.Error.Render(m.state.Err.Error())
	}
	if len(m.state.StatusFiles) == 0 {
		return styles.Success.Render("Working tree clean")
	}

	lines := make([]string, 0, len(m.state.StatusFiles))
	for _, file := range m.state.StatusFiles {
		lines = append(lines, fmt.Sprintf("%s  %s", styles.Badge.Render(file.ShortStatus()), file.Path))
	}
	return strings.Join(lines, "\n")
}

func (m Model) statusSummary() string {
	if m.state.Err != nil {
		return "Repository status unavailable"
	}
	if len(m.state.StatusFiles) == 0 {
		return "Changed files: 0"
	}
	return fmt.Sprintf("Changed files: %d", len(m.state.StatusFiles))
}

func (m Model) branchLabel() string {
	if m.state.RepoInfo.Branch == "" {
		return "unknown"
	}
	return m.state.RepoInfo.Branch
}

func emptyFallback(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func maxView(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func viewWidth(width int) int {
	if width < 60 {
		return width
	}
	if width > 96 {
		return 96
	}
	return width
}
