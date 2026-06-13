package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/KDM-cli/ghx/styles"
)

type MenuItem struct {
	Title       string
	Description string
	Screen      string
}

type SimpleMenuModel struct {
	items    []MenuItem
	selected int
	theme    *styles.Theme
}

func NewSimpleMenuModel(theme *styles.Theme, items []MenuItem) SimpleMenuModel {
	return SimpleMenuModel{
		items:    items,
		theme:    theme,
		selected: 0,
	}
}

func (m SimpleMenuModel) Init() tea.Cmd {
	return nil
}

func (m SimpleMenuModel) Update(msg tea.Msg) (SimpleMenuModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.items)-1 {
				m.selected++
			}
		}
	}

	return m, nil
}

func (m SimpleMenuModel) View() string {
	var s string
	for i, item := range m.items {
		s += m.theme.MenuItemWithSelector(item.Title, i == m.selected)
		if item.Description != "" {
			s += " " + m.theme.Muted.Render(item.Description)
		}
		s += "\n"
	}
	return s
}

func (m SimpleMenuModel) SelectedIndex() int {
	return m.selected
}

func (m SimpleMenuModel) SelectedItem() MenuItem {
	if m.selected >= 0 && m.selected < len(m.items) {
		return m.items[m.selected]
	}
	return MenuItem{}
}
