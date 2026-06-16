package screens

import (
	"encoding/json"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/KDM-cli/ghx/internal/ai"
	"github.com/KDM-cli/ghx/internal/components"
	"github.com/KDM-cli/ghx/internal/db"
	"github.com/KDM-cli/ghx/styles"
)

type settingsState int

const (
	settingsProviderList settingsState = iota
	settingsProviderDetail
	settingsConfigProvider
	settingsCustomModelInput
)

type SettingsModel struct {
	theme            *styles.Theme
	aiManager        *ai.Manager
	db               *db.DB
	state            settingsState
	providers        []ai.ProviderInfo
	selected         int
	configField      int // which config field is being edited
	width            int
	height           int
	err              error
	success          string
	selectedModelIdx int
	providerModels   []string
	customInput      components.TextInputModel
}

func NewSettingsModel(theme *styles.Theme, aiManager *ai.Manager, database *db.DB) SettingsModel {
	return SettingsModel{
		theme:       theme,
		aiManager:   aiManager,
		db:          database,
		state:       settingsProviderList,
		customInput: components.NewTextInputModel(theme, "Enter custom model name..."),
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
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.customInput.SetWidth(m.width - 10)

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
			} else if m.state == settingsConfigProvider && m.selectedModelIdx > 0 {
				m.selectedModelIdx--
			}

		case "down", "j":
			if m.state == settingsProviderList && m.selected < len(m.providers)-1 {
				m.selected++
			} else if m.state == settingsConfigProvider && m.selectedModelIdx < len(m.providerModels) {
				m.selectedModelIdx++
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
			} else if m.state == settingsConfigProvider {
				provider := m.providers[m.selected]
				if m.selectedModelIdx < len(m.providerModels) {
					modelName := m.providerModels[m.selectedModelIdx]
					err := m.saveProviderModel(provider.Type, modelName)
					if err != nil {
						m.err = err
					} else {
						m.success = fmt.Sprintf("Updated %s model to %s", provider.Name, modelName)
						m.state = settingsProviderList
					}
					return m, m.loadProviders
				} else {
					m.state = settingsCustomModelInput
					m.customInput.SetValue("")
					return m, nil
				}
			} else if m.state == settingsCustomModelInput {
				provider := m.providers[m.selected]
				modelName := strings.TrimSpace(m.customInput.Value())
				if modelName != "" {
					err := m.saveProviderModel(provider.Type, modelName)
					if err != nil {
						m.err = err
					} else {
						m.success = fmt.Sprintf("Updated %s model to %s", provider.Name, modelName)
						m.state = settingsProviderList
					}
				} else {
					m.state = settingsConfigProvider
				}
				return m, m.loadProviders
			}

		case "esc", "b":
			if m.state == settingsConfigProvider {
				m.state = settingsProviderList
				return m, nil
			} else if m.state == settingsCustomModelInput {
				m.state = settingsConfigProvider
				return m, nil
			} else {
				return m, func() tea.Msg {
					return Navigate(ScreenHome)
				}
			}

		case "m":
			if m.state == settingsProviderList && m.selected < len(m.providers) {
				provider := m.providers[m.selected]
				m.providerModels = provider.Models
				m.selectedModelIdx = 0
				for i, model := range m.providerModels {
					if model == provider.ConfiguredModel {
						m.selectedModelIdx = i
						break
					}
				}
				m.state = settingsConfigProvider
				return m, nil
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

	if m.state == settingsCustomModelInput {
		m.customInput, cmd = m.customInput.Update(msg)
	}

	return m, cmd
}

func (m *SettingsModel) saveProviderModel(providerType ai.ProviderType, modelName string) error {
	if m.db == nil {
		return fmt.Errorf("database connection not available")
	}

	configJSON, err := m.db.GetAIConfig(string(providerType))
	var config ai.ProviderConfig
	if err == nil && configJSON != "" {
		_ = json.Unmarshal([]byte(configJSON), &config)
	}

	config.Type = providerType
	config.Model = modelName

	return m.aiManager.ConfigureProvider(providerType, config)
}

func (m SettingsModel) View() string {
	var b strings.Builder

	if m.state == settingsCustomModelInput {
		provider := m.providers[m.selected]
		b.WriteString(m.theme.Title.Render(fmt.Sprintf("Settings - Custom Model for %s", provider.Name)))
		b.WriteString("\n\n")

		if m.err != nil {
			b.WriteString(m.theme.Error.Render("Error: " + m.err.Error()))
			b.WriteString("\n\n")
		}

		b.WriteString(m.theme.Header.Render("Enter model name:"))
		b.WriteString("\n\n")

		b.WriteString(m.customInput.View())
		b.WriteString("\n\n")

		b.WriteString(m.theme.Help.Render("Enter Save   Esc Cancel"))
		return lipgloss.NewStyle().Width(m.width).Height(m.height).Render(b.String())
	}

	if m.state == settingsConfigProvider {
		provider := m.providers[m.selected]
		b.WriteString(m.theme.Title.Render(fmt.Sprintf("Settings - %s Models", provider.Name)))
		b.WriteString("\n\n")

		if m.err != nil {
			b.WriteString(m.theme.Error.Render("Error: " + m.err.Error()))
			b.WriteString("\n\n")
		}

		b.WriteString(m.theme.Header.Render("Select model:"))
		b.WriteString("\n\n")

		reservedLines := 11
		visibleCount := m.height - reservedLines
		if visibleCount < 3 {
			visibleCount = 3
		}

		totalItems := len(m.providerModels) + 1
		start := 0
		if m.selectedModelIdx >= visibleCount {
			start = m.selectedModelIdx - visibleCount + 1
		}
		end := start + visibleCount
		if end > totalItems {
			end = totalItems
		}
		if end-start < visibleCount && start > 0 {
			start = end - visibleCount
			if start < 0 {
				start = 0
			}
		}

		for i := start; i < end; i++ {
			if i < len(m.providerModels) {
				model := m.providerModels[i]
				if i == m.selectedModelIdx {
					b.WriteString(m.theme.Selected.Render("> " + model))
					if model == provider.ConfiguredModel {
						b.WriteString(m.theme.Success.Render(" [current]"))
					}
				} else {
					b.WriteString("  ")
					if model == provider.ConfiguredModel {
						b.WriteString(m.theme.Text.Bold(true).Render(model) + m.theme.Success.Render(" [current]"))
					} else {
						b.WriteString(m.theme.Text.Render(model))
					}
				}
			} else {
				// Option for custom model
				if m.selectedModelIdx == len(m.providerModels) {
					b.WriteString(m.theme.Selected.Render("> [Custom Model...]"))
				} else {
					b.WriteString(m.theme.Muted.Render("  [Custom Model...]"))
				}
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")

		b.WriteString(m.theme.Help.Render("↑/↓ Navigate   Enter Select   Esc Cancel"))
		return lipgloss.NewStyle().Width(m.width).Height(m.height).Render(b.String())
	}

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

		// Show configured model
		if p.ConfiguredModel != "" {
			b.WriteString(m.theme.Muted.Render(fmt.Sprintf(" (model: %s)", p.ConfiguredModel)))
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
	b.WriteString(m.theme.Help.Render("↑/↓ Navigate   Enter Select   m Change Model   t Test   b Back"))

	return lipgloss.NewStyle().Width(m.width).Height(m.height).Render(b.String())
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
