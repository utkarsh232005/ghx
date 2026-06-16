package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/KDM-cli/ghx/internal/git"
	"github.com/KDM-cli/ghx/styles"
)

type FileItem struct {
	Path     string
	Status   string
	Selected bool
}

type FileListModel struct {
	files       []FileItem
	selected    int
	theme       *styles.Theme
	multiSelect bool
	Height      int
	Width       int
}

func NewFileListModel(theme *styles.Theme, files []git.FileStatus) FileListModel {
	items := make([]FileItem, len(files))
	for i, f := range files {
		items[i] = FileItem{
			Path:     f.Path,
			Status:   f.Status,
			Selected: false,
		}
	}

	return FileListModel{
		files:       items,
		theme:       theme,
		multiSelect: true,
	}
}

func (m FileListModel) Init() tea.Cmd {
	return nil
}

func (m FileListModel) Update(msg tea.Msg) (FileListModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.files)-1 {
				m.selected++
			}
		case " ":
			if m.multiSelect && m.selected >= 0 && m.selected < len(m.files) {
				m.files[m.selected].Selected = !m.files[m.selected].Selected
			}
		case "a":
			for i := range m.files {
				m.files[i].Selected = true
			}
		case "n":
			for i := range m.files {
				m.files[i].Selected = false
			}
		}
	}

	return m, nil
}

func truncatePath(s string, maxLen int) string {
	if maxLen <= 0 {
		return s
	}
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 4 {
		return "..."
	}
	return "..." + s[len(s)-maxLen+3:]
}

func (m FileListModel) View() string {
	var s strings.Builder

	start := 0
	visibleCount := len(m.files)
	if m.Height > 0 {
		visibleCount = m.Height
		if visibleCount < 3 {
			visibleCount = 3
		}
		if m.selected >= visibleCount {
			start = m.selected - visibleCount + 1
		}
	}
	end := start + visibleCount
	if end > len(m.files) {
		end = len(m.files)
	}
	if end-start < visibleCount && start > 0 {
		start = end - visibleCount
		if start < 0 {
			start = 0
		}
	}

	for i := start; i < end; i++ {
		file := m.files[i]
		if i == m.selected {
			s.WriteString(m.theme.Selected.Render(">"))
		} else {
			s.WriteString(" ")
		}

		s.WriteString(" ")
		s.WriteString(m.theme.Checkbox(file.Selected))
		s.WriteString(" ")

		s.WriteString(m.theme.StatusIcon(file.Status))
		s.WriteString(" ")

		pathWidth := m.Width - 10
		if pathWidth < 10 {
			pathWidth = 30 // fallback default
		}
		displayPath := truncatePath(file.Path, pathWidth)

		if i == m.selected {
			s.WriteString(m.theme.Text.Bold(true).Render(displayPath))
		} else {
			s.WriteString(m.theme.Text.Render(displayPath))
		}

		s.WriteString("\n")
	}

	return s.String()
}

func (m FileListModel) SelectedFiles() []FileItem {
	var selected []FileItem
	for _, file := range m.files {
		if file.Selected {
			selected = append(selected, file)
		}
	}
	return selected
}

func (m FileListModel) SelectedPaths() []string {
	var paths []string
	for _, file := range m.files {
		if file.Selected {
			paths = append(paths, file.Path)
		}
	}
	return paths
}

func (m FileListModel) SetFiles(files []git.FileStatus) FileListModel {
	items := make([]FileItem, len(files))
	for i, f := range files {
		items[i] = FileItem{
			Path:     f.Path,
			Status:   f.Status,
			Selected: false,
		}
	}
	m.files = items
	m.selected = 0
	return m
}

func (m FileListModel) SelectedCount() int {
	count := 0
	for _, file := range m.files {
		if file.Selected {
			count++
		}
	}
	return count
}

func (m FileListModel) Count() int {
	return len(m.files)
}
