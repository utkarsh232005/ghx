package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/KDM-cli/ghx/internal/gh"
	"github.com/KDM-cli/ghx/styles"
)

type ReposModel struct {
	theme    *styles.Theme
	client   *gh.Client
	repos    []gh.RepoInfo
	selected int
	loading  bool
	width    int
	height   int
	err      error
}

func NewReposModel(theme *styles.Theme) ReposModel {
	return ReposModel{
		theme:   theme,
		client:  gh.NewClient(),
		loading: true,
	}
}

func (m ReposModel) Init() tea.Cmd {
	return m.loadRepos
}

func (m ReposModel) loadRepos() tea.Msg {
	client := gh.NewClient()
	repos, err := client.ListRepos(20)
	if err != nil {
		return reposLoadedMsg{err: err}
	}
	return reposLoadedMsg{repos: repos}
}

type reposLoadedMsg struct {
	repos []gh.RepoInfo
	err  error
}

func (m ReposModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case reposLoadedMsg:
		m.loading = false
		m.repos = msg.repos
		m.err = msg.err

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.repos)-1 {
				m.selected++
			}
		case "r":
			m.loading = true
			return m, m.loadRepos
		}
	}

	return m, nil
}

func (m ReposModel) View() string {
	var b strings.Builder

	b.WriteString(m.theme.Title.Render("Repositories"))
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(m.theme.Muted.Render("Loading repositories..."))
		return b.String()
	}

	if m.err != nil {
		b.WriteString(m.theme.Error.Render("Error: " + m.err.Error()))
		b.WriteString("\n\n")
		b.WriteString(m.theme.Help.Render("r Retry   b Back"))
		return b.String()
	}

	if len(m.repos) == 0 {
		b.WriteString(m.theme.Muted.Render("No repositories found"))
		b.WriteString("\n\n")
		b.WriteString(m.theme.Help.Render("b Back"))
		return b.String()
	}

	for i, repo := range m.repos {
		if i == m.selected {
			b.WriteString(m.theme.Selected.Render("> "))
		} else {
			b.WriteString("  ")
		}

		visibility := "public"
		if repo.IsPrivate {
			visibility = "private"
		}

		b.WriteString(fmt.Sprintf("%s [%s] %s\n", repo.Name, visibility, repo.Description))
	}

	b.WriteString("\n")
	b.WriteString(m.theme.Help.Render("↑/↓ Navigate   r Refresh   b Back"))

	return lipgloss.NewStyle().Width(m.width).Height(m.height).Render(b.String())
}
