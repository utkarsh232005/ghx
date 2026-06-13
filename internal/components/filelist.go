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

func (m FileListModel) View() string {
	var s strings.Builder

	for i, file := range m.files {
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

		if i == m.selected {
			s.WriteString(m.theme.Text.Bold(true).Render(file.Path))
		} else {
			s.WriteString(m.theme.Text.Render(file.Path))
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
