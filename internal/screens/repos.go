package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/KDM-cli/ghx/internal/gh"
	"github.com/KDM-cli/ghx/styles"
)

type reposState int

const (
	reposStateList reposState = iota
	reposStateDetails
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
	state    reposState
}

func NewReposModel(theme *styles.Theme) ReposModel {
	return ReposModel{
		theme:   theme,
		client:  gh.NewClient(),
		loading: true,
		state:   reposStateList,
	}
}

func (m ReposModel) Init() tea.Cmd {
	return m.loadRepos
}

func (m ReposModel) loadRepos() tea.Msg {
	client := gh.NewClient()
	repos, err := client.ListRepos(30)
	if err != nil {
		return reposLoadedMsg{err: err}
	}
	return reposLoadedMsg{repos: repos}
}

type reposLoadedMsg struct {
	repos []gh.RepoInfo
	err   error
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
		case "o":
			if len(m.repos) > 0 && m.selected < len(m.repos) {
				_ = m.client.OpenRepoInBrowser(m.repos[m.selected].Name)
			}
		case "enter":
			if len(m.repos) > 0 && m.selected < len(m.repos) {
				if m.width < 80 {
					m.state = reposStateDetails
				} else {
					_ = m.client.OpenRepoInBrowser(m.repos[m.selected].Name)
				}
			}
		case "b", "esc":
			if m.width < 80 && m.state == reposStateDetails {
				m.state = reposStateList
				return m, nil
			}
			return m, func() tea.Msg {
				return Navigate(ScreenHome)
			}
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
		return lipgloss.NewStyle().Width(m.width).Height(m.height).Render(b.String())
	}

	if m.err != nil {
		b.WriteString(m.theme.Error.Render("Error: " + m.err.Error()))
		b.WriteString("\n\n")
		b.WriteString(m.theme.Help.Render("r Retry   b Back"))
		return lipgloss.NewStyle().Width(m.width).Height(m.height).Render(b.String())
	}

	if len(m.repos) == 0 {
		b.WriteString(m.theme.Muted.Render("No repositories found"))
		b.WriteString("\n\n")
		b.WriteString(m.theme.Help.Render("b Back"))
		return lipgloss.NewStyle().Width(m.width).Height(m.height).Render(b.String())
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
	if end > len(m.repos) {
		end = len(m.repos)
	}
	if end-start < visibleCount && start > 0 {
		start = end - visibleCount
		if start < 0 {
			start = 0
		}
	}

	// Layout presentation (Responsive: wide vs narrow screen)
	if m.width >= 80 {
		// Side-by-side dashboard view
		var leftList strings.Builder
		leftList.WriteString(m.theme.Header.Render("Your Repositories"))
		leftList.WriteString("\n\n")

		leftWidth := 34
		for i := start; i < end; i++ {
			repo := m.repos[i]
			if i == m.selected {
				leftList.WriteString(m.theme.Selected.Render("> "))
			} else {
				leftList.WriteString("  ")
			}

			// Format: name with visibility badge
			visBadge := "P"
			visStyle := m.theme.Success
			if repo.IsPrivate {
				visBadge = "L" // L for Locked/Private
				visStyle = m.theme.Warning
			}

			badge := visStyle.Render(fmt.Sprintf("[%s]", visBadge))

			// Truncate name to fit left panel
			maxNameWidth := leftWidth - 8
			displayName := repo.Name
			if len(displayName) > maxNameWidth {
				displayName = displayName[:maxNameWidth-3] + "..."
			}

			if i == m.selected {
				leftList.WriteString(m.theme.Text.Bold(true).Render(displayName) + " " + badge)
			} else {
				leftList.WriteString(m.theme.Text.Render(displayName) + " " + badge)
			}
			leftList.WriteString("\n")
		}

		// Fill vertical space
		for leftList.Len() < visibleCount {
			leftList.WriteString("\n")
		}

		// Selected repository details
		selectedRepo := m.repos[m.selected]
		var rightCard strings.Builder
		visText := "Public"
		visStyle := m.theme.Success
		if selectedRepo.IsPrivate {
			visText = "Private"
			visStyle = m.theme.Warning
		}

		rightWidth := m.width - leftWidth - 4
		if rightWidth < 30 {
			rightWidth = 30
		}

		rightCard.WriteString(m.theme.Primary.Render(selectedRepo.Name))
		rightCard.WriteString("  ")
		rightCard.WriteString(visStyle.Render("[" + visText + "]"))
		rightCard.WriteString("\n\n")

		if selectedRepo.Description != "" {
			wrappedDesc := lipgloss.NewStyle().Width(rightWidth - 6).Render(selectedRepo.Description)
			rightCard.WriteString(wrappedDesc)
		} else {
			rightCard.WriteString(m.theme.Muted.Render("No description provided."))
		}
		rightCard.WriteString("\n\n")

		rightCard.WriteString(m.theme.Bold.Render("URL:\n"))
		rightCard.WriteString(m.theme.Accent.Render(selectedRepo.URL))
		rightCard.WriteString("\n\n")

		rightCard.WriteString(m.theme.Bold.Render("Clone (HTTPS):\n"))
		rightCard.WriteString(m.theme.Text.Render("git clone " + selectedRepo.URL))
		rightCard.WriteString("\n\n")

		rightCard.WriteString(m.theme.Bold.Render("Clone (CLI):\n"))
		rightCard.WriteString(m.theme.Text.Render("gh repo clone " + selectedRepo.Name))

		rightBox := m.theme.Box.Width(rightWidth).Height(visibleCount + 2).Render(rightCard.String())

		// Join panels horizontally
		mainContent := lipgloss.JoinHorizontal(lipgloss.Top,
			lipgloss.NewStyle().Width(leftWidth).Render(leftList.String()),
			rightBox,
		)
		b.WriteString(mainContent)
		b.WriteString("\n\n")
		b.WriteString(m.theme.Help.Render("↑/↓ Navigate   Enter/o Open in browser   r Refresh   b Back"))

	} else {
		// Narrow screen representation
		if m.state == reposStateList {
			// Scrollable list only
			for i := start; i < end; i++ {
				repo := m.repos[i]
				if i == m.selected {
					b.WriteString(m.theme.Selected.Render("> "))
				} else {
					b.WriteString("  ")
				}

				visText := "pub"
				visStyle := m.theme.Success
				if repo.IsPrivate {
					visText = "priv"
					visStyle = m.theme.Warning
				}

				maxNameWidth := m.width - 12
				displayName := repo.Name
				if len(displayName) > maxNameWidth {
					displayName = displayName[:maxNameWidth-3] + "..."
				}

				b.WriteString(fmt.Sprintf("%s %s\n", displayName, visStyle.Render("["+visText+"]")))
			}
			b.WriteString("\n")
			b.WriteString(m.theme.Help.Render("↑/↓ Navigate   Enter Details   o Open in browser   r Refresh   b Back"))
		} else {
			// Full details card only
			selectedRepo := m.repos[m.selected]
			visText := "Public"
			visStyle := m.theme.Success
			if selectedRepo.IsPrivate {
				visText = "Private"
				visStyle = m.theme.Warning
			}

			var details strings.Builder
			details.WriteString(m.theme.Primary.Render(selectedRepo.Name))
			details.WriteString("  ")
			details.WriteString(visStyle.Render("[" + visText + "]"))
			details.WriteString("\n\n")

			if selectedRepo.Description != "" {
				wrappedDesc := lipgloss.NewStyle().Width(m.width - 6).Render(selectedRepo.Description)
				details.WriteString(wrappedDesc)
			} else {
				details.WriteString(m.theme.Muted.Render("No description provided."))
			}
			details.WriteString("\n\n")

			details.WriteString(m.theme.Bold.Render("URL:\n"))
			details.WriteString(m.theme.Accent.Render(selectedRepo.URL))
			details.WriteString("\n\n")

			details.WriteString(m.theme.Bold.Render("Clone (HTTPS):\n"))
			details.WriteString(m.theme.Text.Render("git clone " + selectedRepo.URL))
			details.WriteString("\n\n")

			details.WriteString(m.theme.Bold.Render("Clone (CLI):\n"))
			details.WriteString(m.theme.Text.Render("gh repo clone " + selectedRepo.Name))

			detailsBox := m.theme.Box.Width(m.width - 2).Render(details.String())
			b.WriteString(detailsBox)
			b.WriteString("\n\n")
			b.WriteString(m.theme.Help.Render("o Open in browser   Esc/b Back to list"))
		}
	}

	return lipgloss.NewStyle().Width(m.width).Height(m.height).Render(b.String())
}
