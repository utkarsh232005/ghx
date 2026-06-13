package screens

import (
	"context"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/KDM-cli/ghx/internal/ai"
	"github.com/KDM-cli/ghx/internal/components"
	"github.com/KDM-cli/ghx/styles"
)

type AIChatModel struct {
	theme      *styles.Theme
	aiManager  *ai.Manager
	chat       components.ChatModel
	input      components.TextInputModel
	streaming  bool
	streamText strings.Builder
	width      int
	height     int
	err        error
}

func NewAIChatModel(theme *styles.Theme, aiManager *ai.Manager) AIChatModel {
	return AIChatModel{
		theme:     theme,
		aiManager: aiManager,
		chat:      components.NewChatModel(theme),
		input:     components.NewTextInputModel(theme, "Ask AI anything..."),
	}
}

func (m AIChatModel) Init() tea.Cmd {
	return nil
}

type chatResponseMsg struct {
	content string
	done    bool
	err     error
}

type chatStreamMsg struct {
	content string
	done    bool
}

func (m AIChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.input.SetWidth(m.width - 10)
		m.chat = m.chat.SetWidth(m.width)
		m.chat = m.chat.SetHeight(m.height - 8)

	case chatStreamMsg:
		if msg.done {
			m.streaming = false
		}
		if msg.content != "" && len(m.chat.Messages) > 0 {
			m.chat.Messages[len(m.chat.Messages)-1].Content += msg.content
		}

	case chatResponseMsg:
		m.streaming = false
		if msg.err != nil {
			m.err = msg.err
			// Remove empty AI message
			if len(m.chat.Messages) > 0 && m.chat.Messages[len(m.chat.Messages)-1].Content == "" {
				m.chat.Messages = m.chat.Messages[:len(m.chat.Messages)-1]
			}
		}

	case tea.KeyMsg:
		m.err = nil

		switch msg.String() {
		case "enter":
			if !m.streaming && m.input.Value() != "" {
				userMsg := m.input.Value()
				m.input.SetValue("")

				m.chat = m.chat.AddMessage("user", userMsg)
				m.chat = m.chat.AddMessage("ai", "")
				m.streaming = true

				return m, m.sendChat
			}

		case "ctrl+l":
			m.chat = m.chat.Clear()
			m.err = nil
		}
	}

	if !m.streaming {
		m.input, cmd = m.input.Update(msg)
	}

	return m, cmd
}

func (m AIChatModel) sendChat() tea.Msg {
	ctx := context.Background()
	messages := make([]ai.Message, 0)
	for _, msg := range m.chat.Messages {
		messages = append(messages, ai.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Use streaming if available
	stream, err := m.aiManager.Stream(ctx, messages)
	if err != nil {
		// Fallback to non-streaming
		resp, err := m.aiManager.Chat(ctx, messages)
		if err != nil {
			return chatResponseMsg{err: err}
		}
		return chatResponseMsg{content: resp.Content, done: true}
	}

	// Collect stream results
	go func() {
		var content strings.Builder
		for resp := range stream {
			if resp.Error != nil {
				// Send error response
				return
			}
			content.WriteString(resp.Content)
			if resp.Done {
				break
			}
		}
	}()

	// For now, collect all at once
	var content strings.Builder
	for resp := range stream {
		if resp.Error != nil {
			return chatResponseMsg{err: resp.Error}
		}
		content.WriteString(resp.Content)
		if resp.Done {
			break
		}
	}

	return chatResponseMsg{content: content.String(), done: true}
}

func (m AIChatModel) View() string {
	var b strings.Builder

	b.WriteString(m.theme.Title.Render("AI Assistant"))
	b.WriteString(" ")
	provider := m.aiManager.GetActiveProvider()
	b.WriteString(m.theme.Muted.Render("(" + provider.Name() + ")"))

	if m.streaming {
		b.WriteString(" ")
		b.WriteString(m.theme.Accent.Render("●"))
	}
	b.WriteString("\n\n")

	// Chat area
	b.WriteString(m.chat.View())
	b.WriteString("\n")

	// Input
	b.WriteString(m.input.View())

	if m.streaming {
		b.WriteString(" ")
		b.WriteString(m.theme.Accent.Render("Thinking..."))
	}
	b.WriteString("\n\n")

	// Help
	if m.streaming {
		b.WriteString(m.theme.Help.Render("Wait for response..."))
	} else {
		b.WriteString(m.theme.Help.Render("Enter Send   Ctrl+L Clear   b Back"))
	}

	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(m.theme.Error.Render("Error: " + m.err.Error()))
	}

	return lipgloss.NewStyle().Width(m.width).Height(m.height).Render(b.String())
}
