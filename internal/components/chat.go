package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/KDM-cli/ghx/styles"
)

type ChatMessage struct {
	Role    string
	Content string
}

type ChatModel struct {
	Messages []ChatMessage
	Theme    *styles.Theme
	Width    int
	Height   int
}

func NewChatModel(theme *styles.Theme) ChatModel {
	return ChatModel{
		Theme:    theme,
		Messages: []ChatMessage{},
	}
}

func (m ChatModel) Init() tea.Cmd {
	return nil
}

func (m ChatModel) Update(msg tea.Msg) (ChatModel, tea.Cmd) {
	return m, nil
}

func (m ChatModel) View() string {
	var s strings.Builder

	visibleMessages := m.Messages
	if len(visibleMessages) > m.Height-4 && m.Height > 4 {
		visibleMessages = visibleMessages[len(visibleMessages)-(m.Height-4):]
	}

	for _, msg := range visibleMessages {
		var style lipgloss.Style
		var prefix string

		if msg.Role == "user" {
			style = m.Theme.Accent
			prefix = "You"
		} else {
			style = m.Theme.Success
			prefix = "AI"
		}

		s.WriteString(style.Bold(true).Render(prefix + ": "))
		s.WriteString(m.Theme.Text.Render(msg.Content))
		s.WriteString("\n\n")
	}

	return s.String()
}

func (m ChatModel) AddMessage(role, content string) ChatModel {
	m.Messages = append(m.Messages, ChatMessage{
		Role:    role,
		Content: content,
	})
	return m
}

func (m ChatModel) Clear() ChatModel {
	m.Messages = []ChatMessage{}
	return m
}

func (m ChatModel) SetWidth(width int) ChatModel {
	m.Width = width
	return m
}

func (m ChatModel) SetHeight(height int) ChatModel {
	m.Height = height
	return m
}
