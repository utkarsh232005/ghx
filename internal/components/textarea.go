package components

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/KDM-cli/ghx/styles"
)

type TextAreaModel struct {
	TextArea textarea.Model
	Theme    *styles.Theme
	Focused  bool
}

func NewTextAreaModel(theme *styles.Theme, placeholder string) TextAreaModel {
	ta := textarea.New()
	ta.Placeholder = placeholder
	ta.Focus()
	ta.CharLimit = 0 // Remove the default character limit

	return TextAreaModel{
		TextArea: ta,
		Theme:    theme,
		Focused:  true,
	}
}

func (m TextAreaModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m TextAreaModel) Update(msg tea.Msg) (TextAreaModel, tea.Cmd) {
	var cmd tea.Cmd
	m.TextArea, cmd = m.TextArea.Update(msg)
	return m, cmd
}

func (m TextAreaModel) View() string {
	return m.Theme.Input.Render(m.TextArea.View())
}

func (m TextAreaModel) Value() string {
	return m.TextArea.Value()
}

func (m *TextAreaModel) SetValue(val string) {
	m.TextArea.SetValue(val)
}

func (m *TextAreaModel) SetWidth(width int) {
	m.TextArea.SetWidth(width)
}

func (m *TextAreaModel) SetHeight(height int) {
	m.TextArea.SetHeight(height)
}
