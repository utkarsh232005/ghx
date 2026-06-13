package components

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/KDM-cli/ghx/styles"
)

type TextInputModel struct {
	Input   textinput.Model
	Theme   *styles.Theme
	Focused bool
}

func NewTextInputModel(theme *styles.Theme, placeholder string) TextInputModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Focus()

	return TextInputModel{
		Input:   ti,
		Theme:   theme,
		Focused: true,
	}
}

func (m TextInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m TextInputModel) Update(msg tea.Msg) (TextInputModel, tea.Cmd) {
	var cmd tea.Cmd
	m.Input, cmd = m.Input.Update(msg)
	return m, cmd
}

func (m TextInputModel) View() string {
	return m.Theme.Input.Render(m.Input.View())
}

func (m TextInputModel) Value() string {
	return m.Input.Value()
}

func (m *TextInputModel) SetValue(val string) {
	m.Input.SetValue(val)
}

func (m *TextInputModel) SetWidth(width int) {
	m.Input.Width = width
}
