package ai

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/KDM-cli/ghx/internal/db"
)

type Manager struct {
	providers     map[ProviderType]Provider
	active        ProviderType
	db            *db.DB
	mu            sync.RWMutex
}

func NewManager(database *db.DB) *Manager {
	m := &Manager{
		providers: make(map[ProviderType]Provider),
		active:    ProviderOllama,
		db:        database,
	}

	m.providers[ProviderOllama] = NewOllamaProvider()
	m.providers[ProviderOpenAI] = NewOpenAIProvider()
	m.providers[ProviderClaude] = NewClaudeProvider()
	m.providers[ProviderMLX] = NewMLXProvider()
	m.providers[ProviderLMStudio] = NewLMStudioProvider()

	m.loadConfig()

	return m
}

func (m *Manager) loadConfig() {
	if m.db == nil {
		return
	}

	activeProvider, err := m.db.GetSetting("active_provider")
	if err == nil && activeProvider != "" {
		m.active = ProviderType(activeProvider)
	}

	for providerType := range m.providers {
		configJSON, err := m.db.GetAIConfig(string(providerType))
		if err != nil || configJSON == "" {
			continue
		}

		var config ProviderConfig
		if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
			continue
		}

		if p, ok := m.providers[providerType]; ok {
			p.Configure(config)
		}
	}
}

func (m *Manager) GetActiveProvider() Provider {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if p, ok := m.providers[m.active]; ok {
		return p
	}
	return m.providers[ProviderOllama]
}

func (m *Manager) SetActiveProvider(provider ProviderType) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.active = provider
	if m.db != nil {
		return m.db.SetSetting("active_provider", string(provider))
	}
	return nil
}

func (m *Manager) GetProvider(providerType ProviderType) (Provider, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p, ok := m.providers[providerType]
	return p, ok
}

func (m *Manager) ConfigureProvider(providerType ProviderType, config ProviderConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	p, ok := m.providers[providerType]
	if !ok {
		return nil
	}

	if err := p.Configure(config); err != nil {
		return err
	}

	if m.db != nil {
		configJSON, err := json.Marshal(config)
		if err != nil {
			return err
		}
		return m.db.SetAIConfig(string(providerType), string(configJSON))
	}
	return nil
}

type ProviderInfo struct {
	Type         ProviderType
	Name         string
	Models       []string
	IsConfigured bool
	IsActive     bool
}

func (m *Manager) ListProviders() []ProviderInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var providers []ProviderInfo
	for providerType, p := range m.providers {
		info := ProviderInfo{
			Type:         providerType,
			Name:         p.Name(),
			Models:       p.Models(),
			IsConfigured: p.IsConfigured(),
			IsActive:     providerType == m.active,
		}
		providers = append(providers, info)
	}
	return providers
}

func (m *Manager) Chat(ctx context.Context, messages []Message) (Response, error) {
	return m.GetActiveProvider().Chat(ctx, messages)
}

func (m *Manager) Stream(ctx context.Context, messages []Message) (<-chan StreamResponse, error) {
	return m.GetActiveProvider().Stream(ctx, messages)
}
