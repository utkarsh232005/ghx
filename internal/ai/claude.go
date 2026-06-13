package ai

import (
	"context"
	"fmt"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type ClaudeProvider struct {
	client     *anthropic.Client
	model      string
	configured bool
}

func NewClaudeProvider() *ClaudeProvider {
	return &ClaudeProvider{
		model: anthropic.ModelClaude3_5SonnetLatest,
	}
}

func (c *ClaudeProvider) Name() string {
	return "claude"
}

func (c *ClaudeProvider) Configure(config ProviderConfig) error {
	if config.APIKey == "" {
		return fmt.Errorf("claude requires an API key")
	}

	c.client = anthropic.NewClient(
		option.WithAPIKey(config.APIKey),
	)

	if config.Model != "" {
		c.model = config.Model
	}
	c.configured = true
	return nil
}

func (c *ClaudeProvider) IsConfigured() bool {
	return c.configured && c.client != nil
}

func (c *ClaudeProvider) Models() []string {
	return []string{
		anthropic.ModelClaude3_5SonnetLatest,
		anthropic.ModelClaude3_5HaikuLatest,
		anthropic.ModelClaude3OpusLatest,
	}
}

func (c *ClaudeProvider) Chat(ctx context.Context, messages []Message) (Response, error) {
	if !c.IsConfigured() {
		return Response{}, fmt.Errorf("claude not configured")
	}

	chatMessages := make([]anthropic.MessageParam, len(messages))
	for i, m := range messages {
		if m.Role == "user" {
			chatMessages[i] = anthropic.NewUserMessage(anthropic.NewTextBlock(m.Content))
		} else {
			chatMessages[i] = anthropic.NewAssistantMessage(anthropic.NewTextBlock(m.Content))
		}
	}

	resp, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.F(c.model),
		MaxTokens: anthropic.F(int64(4096)),
		Messages:  anthropic.F(chatMessages),
	})
	if err != nil {
		return Response{}, err
	}

	if len(resp.Content) == 0 {
		return Response{}, fmt.Errorf("no response from claude")
	}

	text := ""
	for _, block := range resp.Content {
		if block.Type == anthropic.ContentBlockTypeText {
			text += block.Text
		}
	}

	return Response{Content: text}, nil
}

func (c *ClaudeProvider) Stream(ctx context.Context, messages []Message) (<-chan StreamResponse, error) {
	if !c.IsConfigured() {
		return nil, fmt.Errorf("claude not configured")
	}

	chatMessages := make([]anthropic.MessageParam, len(messages))
	for i, m := range messages {
		if m.Role == "user" {
			chatMessages[i] = anthropic.NewUserMessage(anthropic.NewTextBlock(m.Content))
		} else {
			chatMessages[i] = anthropic.NewAssistantMessage(anthropic.NewTextBlock(m.Content))
		}
	}

	stream := c.client.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
		Model:     anthropic.F(c.model),
		MaxTokens: anthropic.F(int64(4096)),
		Messages:  anthropic.F(chatMessages),
	})

	ch := make(chan StreamResponse)
	go func() {
		defer close(ch)

		for stream.Next() {
			event := stream.Current()

			switch event.Type {
			case anthropic.MessageStreamEventTypeContentBlockDelta:
				if event.Delta != nil && event.Delta.Type == anthropic.ContentBlockDeltaTypeTextDelta {
					ch <- StreamResponse{
						Content: event.Delta.Text,
						Done:    false,
					}
				}
			case anthropic.MessageStreamEventTypeMessageStop:
				ch <- StreamResponse{Done: true}
				return
			}
		}

		if err := stream.Err(); err != nil {
			ch <- StreamResponse{Error: err}
		}
	}()

	return ch, nil
}
