package screens

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/KDM-cli/ghx/styles"
)

type HelpInitMsg struct {
	PreviousScreen Screen
}

type HelpModel struct {
	theme          *styles.Theme
	previousScreen Screen
	width          int
	height         int
}

func NewHelpModel(theme *styles.Theme) HelpModel {
	return HelpModel{theme: theme}
}

func (m HelpModel) Init() tea.Cmd {
	return nil
}

func (m HelpModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case HelpInitMsg:
		m.previousScreen = msg.PreviousScreen

	case tea.KeyMsg:
		// Any key closes help
		if m.previousScreen != "" {
			return m, func() tea.Msg {
				return Navigate(m.previousScreen)
			}
		}
		return m, func() tea.Msg {
			return Navigate(ScreenHome)
		}
	}

	return m, nil
}

func (m HelpModel) View() string {
	var b strings.Builder

	b.WriteString(m.theme.Title.Render("Help - Keyboard Shortcuts"))
	b.WriteString("\n\n")

	helpSections := []struct {
		title   string
		content []string
	}{
		{
			title: "Navigation",
			content: []string{
				"↑/k         Move up",
				"↓/j         Move down",
				"Enter       Select/Confirm",
				"Tab         Next field/section",
				"Shift+Tab   Previous field",
				"Esc/q       Back/Quit",
			},
		},
		{
			title: "Global",
			content: []string{
				"?           Show this help",
				"Ctrl+C      Force quit",
				"r           Refresh current view",
			},
		},
		{
			title: "File Selection (Commit)",
			content: []string{
				"Space       Toggle file selection",
				"a           Select all files",
				"n           Deselect all files",
			},
		},
		{
			title: "AI Features",
			content: []string{
				"g           Generate with AI",
				"r           Regenerate response",
				"Ctrl+L      Clear AI chat",
			},
		},
		{
			title: "Git Operations",
			content: []string{
				"s           Stage file",
				"u           Unstage file",
				"d           Toggle draft (PR)",
			},
		},
	}

	for _, section := range helpSections {
		b.WriteString(m.theme.Bold.Render(section.title))
		b.WriteString("\n")

		for _, line := range section.content {
			b.WriteString("  ")
			b.WriteString(m.theme.Text.Render(line))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	b.WriteString(m.theme.Muted.Render("Press any key to close"))

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Render(b.String())
}
