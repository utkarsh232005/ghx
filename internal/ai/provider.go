package ai

import (
	"context"
)

type Message struct {
	Role    string
	Content string
}

type Response struct {
	Content string
	Error   error
}

type StreamResponse struct {
	Content string
	Done    bool
	Error   error
}

type ProviderType string

const (
	ProviderOllama   ProviderType = "ollama"
	ProviderOpenAI   ProviderType = "openai"
	ProviderClaude   ProviderType = "claude"
	ProviderMLX      ProviderType = "mlx"
	ProviderLMStudio ProviderType = "lmstudio"
)

type ProviderConfig struct {
	Type    ProviderType
	Host    string
	APIKey  string
	Model   string
	Options map[string]interface{}
}

type Provider interface {
	Name() string
	Chat(ctx context.Context, messages []Message) (Response, error)
	ChatWithOptions(ctx context.Context, messages []Message, options map[string]interface{}) (Response, error)
	Stream(ctx context.Context, messages []Message) (<-chan StreamResponse, error)
	Configure(config ProviderConfig) error
	IsConfigured() bool
	Models() []string
}
