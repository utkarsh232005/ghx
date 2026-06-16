package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/KDM-cli/ghx/internal/db"
	"github.com/KDM-cli/ghx/internal/git"
	"github.com/KDM-cli/ghx/styles"
)

type StatusModel struct {
	theme      *styles.Theme
	db         *db.DB
	status     *git.Status
	fileList   []git.FileStatus
	selected   int
	viewMode   string // "all", "staged", "modified", "untracked"
	width      int
	height     int
	loading    bool
	err        error
}

func NewStatusModel(theme *styles.Theme, database *db.DB) StatusModel {
	return StatusModel{
		theme:    theme,
		db:       database,
		loading:  true,
		viewMode: "all",
	}
}

func (m StatusModel) Init() tea.Cmd {
	return m.loadStatus
}

func (m StatusModel) loadStatus() tea.Msg {
	status, err := git.GetStatus()
	if err != nil {
		return statusLoadedMsg{err: err}
	}
	return statusLoadedMsg{status: status}
}

type statusLoadedMsg struct {
	status *git.Status
	err    error
}

func (m StatusModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case statusLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.status = msg.status
		m.updateFileList()

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.fileList)-1 {
				m.selected++
			}
		case "s":
			if len(m.fileList) > 0 && m.selected < len(m.fileList) {
				file := m.fileList[m.selected]
				if !file.Staged && file.Status != "?" {
					git.Stage([]string{file.Path})
					return m, m.loadStatus
				}
			}
		case "u":
			if len(m.fileList) > 0 && m.selected < len(m.fileList) {
				file := m.fileList[m.selected]
				if file.Staged {
					git.Unstage([]string{file.Path})
					return m, m.loadStatus
				}
			}
		case "a":
			paths := make([]string, len(m.fileList))
			for i, f := range m.fileList {
				paths[i] = f.Path
			}
			git.Stage(paths)
			return m, m.loadStatus
		case "r":
			m.loading = true
			return m, m.loadStatus
		case "1":
			m.viewMode = "all"
			m.selected = 0
			m.updateFileList()
		case "2":
			m.viewMode = "staged"
			m.selected = 0
			m.updateFileList()
		case "3":
			m.viewMode = "modified"
			m.selected = 0
			m.updateFileList()
		case "4":
			m.viewMode = "untracked"
			m.selected = 0
			m.updateFileList()
		case "b":
			return m, func() tea.Msg {
				return Navigate(ScreenHome)
			}
		}
	}

	return m, nil
}

func (m *StatusModel) updateFileList() {
	m.fileList = nil
	if m.status == nil {
		return
	}

	switch m.viewMode {
	case "staged":
		m.fileList = m.status.Staged
	case "modified":
		m.fileList = m.status.Modified
	case "untracked":
		m.fileList = m.status.Untracked
	default:
		m.fileList = append(m.fileList, m.status.Staged...)
		m.fileList = append(m.fileList, m.status.Modified...)
		m.fileList = append(m.fileList, m.status.Untracked...)
	}
}

func (m StatusModel) View() string {
	var b strings.Builder

	b.WriteString(m.theme.Title.Render("Git Status"))
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString(m.theme.Muted.Render("Loading..."))
		return lipgloss.NewStyle().Width(m.width).Height(m.height).Render(b.String())
	}

	if m.err != nil {
		b.WriteString(m.theme.Error.Render("Error: " + m.err.Error()))
		b.WriteString("\n\n")
		b.WriteString(m.theme.Help.Render("r Retry   b Back"))
		return lipgloss.NewStyle().Width(m.width).Height(m.height).Render(b.String())
	}

	// Branch info
	if m.status != nil && m.status.Info.Branch != "" {
		b.WriteString(m.theme.Bold.Render("Branch: "))
		b.WriteString(m.theme.Text.Render(m.status.Info.Branch))
		if m.status.Info.Remote != "" {
			b.WriteString(m.theme.Muted.Render(" → " + m.status.Info.Remote))
		}
		if m.status.Info.Ahead > 0 || m.status.Info.Behind > 0 {
			b.WriteString(m.theme.Accent.Render(fmt.Sprintf(" [↑%d ↓%d]", m.status.Info.Ahead, m.status.Info.Behind)))
		}
		b.WriteString("\n\n")
	}

	// View tabs
	tabs := []string{"All", "Staged", "Modified", "Untracked"}
	tabCounts := []int{
		len(m.status.Staged) + len(m.status.Modified) + len(m.status.Untracked),
		len(m.status.Staged),
		len(m.status.Modified),
		len(m.status.Untracked),
	}
	for i, tab := range tabs {
		if m.viewMode == strings.ToLower(tab) || (m.viewMode == "all" && i == 0) {
			b.WriteString(m.theme.Selected.Render(fmt.Sprintf("[%s:%d]", tab, tabCounts[i])))
		} else {
			b.WriteString(m.theme.Muted.Render(fmt.Sprintf(" %s:%d ", tab, tabCounts[i])))
		}
		b.WriteString(" ")
	}
	b.WriteString("\n\n")

	// File list scrolling/viewport limits
	if len(m.fileList) == 0 {
		b.WriteString(m.theme.Success.Render("No files to display"))
	} else {
		reservedLines := 12
		visibleCount := m.height - reservedLines
		if visibleCount < 3 {
			visibleCount = 3
		}

		start := 0
		if m.selected >= visibleCount {
			start = m.selected - visibleCount + 1
		}
		end := start + visibleCount
		if end > len(m.fileList) {
			end = len(m.fileList)
		}
		if end-start < visibleCount && start > 0 {
			start = end - visibleCount
			if start < 0 {
				start = 0
			}
		}

		for i := start; i < end; i++ {
			file := m.fileList[i]
			if i == m.selected {
				b.WriteString(m.theme.Selected.Render("> "))
			} else {
				b.WriteString("  ")
			}

			b.WriteString(m.theme.StatusIcon(file.Status))
			b.WriteString(" ")

			maxPathWidth := m.width - 8
			displayPath := file.Path
			if maxPathWidth > 10 && len(displayPath) > maxPathWidth {
				displayPath = "..." + displayPath[len(displayPath)-maxPathWidth+3:]
			}

			if file.Staged {
				b.WriteString(m.theme.Staged.Render(displayPath))
			} else {
				b.WriteString(m.theme.Text.Render(displayPath))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(m.theme.Help.Render("s Stage   u Unstage   a Stage All   1-4 Filter   r Refresh   b Back"))

	return lipgloss.NewStyle().Width(m.width).Height(m.height).Render(b.String())
}
