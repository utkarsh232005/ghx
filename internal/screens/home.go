package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/KDM-cli/ghx/internal/ai"
	"github.com/KDM-cli/ghx/internal/components"
	"github.com/KDM-cli/ghx/styles"
)

type HomeModel struct {
	menu      components.SimpleMenuModel
	theme     *styles.Theme
	aiManager *ai.Manager
	width     int
	height    int
}

func NewHomeModel(theme *styles.Theme, aiManager *ai.Manager) HomeModel {
	items := []components.MenuItem{
		{Title: "Status", Description: "View git status", Screen: "status"},
		{Title: "Commit", Description: "Stage & commit files (AI assist)", Screen: "commit"},
		{Title: "Push", Description: "Push to remote", Screen: "push"},
		{Title: "PR", Description: "Manage pull requests (AI assist)", Screen: "pr"},
		{Title: "Issues", Description: "Manage issues", Screen: "issues"},
		{Title: "Repos", Description: "Browse repositories", Screen: "repos"},
		{Title: "AI Chat", Description: "Ask AI about codebase", Screen: "ai_chat"},
		{Title: "Settings", Description: "Configure providers", Screen: "settings"},
	}

	return HomeModel{
		menu:      components.NewSimpleMenuModel(theme, items),
		theme:     theme,
		aiManager: aiManager,
	}
}

func (m HomeModel) Init() tea.Cmd {
	return nil
}

func (m HomeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		m.menu, cmd = m.menu.Update(msg)
		if msg.String() == "enter" {
			item := m.menu.SelectedItem()
			return m, func() tea.Msg {
				return Navigate(Screen(item.Screen))
			}
		}
	}

	return m, cmd
}

func (m HomeModel) View() string {
	var b strings.Builder

	title := m.theme.Title.Render("ghx - AI-Powered GitHub Assistant")
	b.WriteString(title)
	b.WriteString("\n\n")

	b.WriteString(m.menu.View())
	b.WriteString("\n\n")

	activeProvider := m.aiManager.GetActiveProvider()
	providerInfo := fmt.Sprintf("AI: %s", activeProvider.Name())
	b.WriteString(m.theme.Muted.Render(providerInfo))
	b.WriteString("\n")

	footer := m.theme.Help.Render("/↑↓ Navigate   Enter Select   q Quit   ? Help")
	b.WriteString(footer)

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Render(b.String())
}
