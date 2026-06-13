package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/KDM-cli/ghx/internal/ai"
	"github.com/KDM-cli/ghx/internal/db"
	"github.com/KDM-cli/ghx/styles"
)

type settingsState int

const (
	settingsProviderList settingsState = iota
	settingsProviderDetail
	settingsConfigProvider
)

type SettingsModel struct {
	theme         *styles.Theme
	aiManager     *ai.Manager
	db            *db.DB
	state         settingsState
	providers     []ai.ProviderInfo
	selected      int
	configField   int // which config field is being edited
	width         int
	height        int
	err           error
	success       string
}

func NewSettingsModel(theme *styles.Theme, aiManager *ai.Manager, database *db.DB) SettingsModel {
	return SettingsModel{
		theme:     theme,
		aiManager: aiManager,
		db:        database,
		state:     settingsProviderList,
	}
}

func (m SettingsModel) Init() tea.Cmd {
	return m.loadProviders
}

func (m SettingsModel) loadProviders() tea.Msg {
	return providersLoadedMsg{providers: m.aiManager.ListProviders()}
}

type providersLoadedMsg struct {
	providers []ai.ProviderInfo
}

type configSavedMsg struct {
	err error
}

func (m SettingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case providersLoadedMsg:
		m.providers = msg.providers

	case configSavedMsg:
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.success = "Configuration saved!"
		}

	case tea.KeyMsg:
		m.err = nil
		m.success = ""

		switch msg.String() {
		case "up", "k":
			if m.state == settingsProviderList && m.selected > 0 {
				m.selected--
			}

		case "down", "j":
			if m.state == settingsProviderList && m.selected < len(m.providers)-1 {
				m.selected++
			}

		case "enter":
			if m.state == settingsProviderList && m.selected < len(m.providers) {
				provider := m.providers[m.selected]
				err := m.aiManager.SetActiveProvider(provider.Type)
				if err != nil {
					m.err = err
				} else {
					m.success = fmt.Sprintf("Switched to %s", provider.Name)
				}
				return m, m.loadProviders
			}

		case "t":
			// Test connection
			if m.state == settingsProviderList && m.selected < len(m.providers) {
				provider := m.providers[m.selected]
				if provider.IsConfigured {
					m.success = fmt.Sprintf("%s is configured and ready", provider.Name)
				} else {
					m.err = fmt.Errorf("%s needs configuration", provider.Name)
				}
			}
		}
	}

	return m, nil
}

func (m SettingsModel) View() string {
	var b strings.Builder

	b.WriteString(m.theme.Title.Render("Settings - AI Providers"))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(m.theme.Error.Render("Error: " + m.err.Error()))
		b.WriteString("\n\n")
	}
	if m.success != "" {
		b.WriteString(m.theme.Success.Render(m.success))
		b.WriteString("\n\n")
	}

	b.WriteString(m.theme.Header.Render("Select active AI provider:"))
	b.WriteString("\n\n")

	if len(m.providers) == 0 {
		b.WriteString(m.theme.Muted.Render("Loading..."))
		return b.String()
	}

	for i, p := range m.providers {
		if i == m.selected {
			b.WriteString(m.theme.Selected.Render("> "))
		} else {
			b.WriteString("  ")
		}

		// Provider name with active indicator
		name := p.Name
		if p.IsActive {
			name = "● " + name
		}

		if p.IsConfigured {
			b.WriteString(m.theme.Text.Bold(true).Render(name))
		} else {
			b.WriteString(m.theme.Muted.Render(name))
		}

		// Status indicators
		if p.IsActive {
			b.WriteString(m.theme.Success.Render(" [active]"))
		}
		if !p.IsConfigured {
			b.WriteString(m.theme.Warning.Render(" [needs config]"))
		}

		b.WriteString("\n")

		// Show models on selected
		if i == m.selected && len(p.Models) > 0 {
			b.WriteString("     ")
			b.WriteString(m.theme.Muted.Render("Models: "))
			b.WriteString(m.theme.Muted.Render(strings.Join(p.Models[:min(3, len(p.Models))], ", ")))
			if len(p.Models) > 3 {
				b.WriteString(m.theme.Muted.Render("..."))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(m.theme.Help.Render("↑/↓ Navigate   Enter Select   t Test   b Back"))

	return lipgloss.NewStyle().Width(m.width).Height(m.height).Render(b.String())
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
