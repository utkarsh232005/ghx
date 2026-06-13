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
	return c.ChatWithOptions(ctx, messages, nil)
}

func (c *ClaudeProvider) ChatWithOptions(ctx context.Context, messages []Message, options map[string]interface{}) (Response, error) {
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

	maxTokens := int64(4096)
	if options != nil {
		if val, ok := options["max_tokens"]; ok {
			if limit, ok := val.(int); ok {
				maxTokens = int64(limit)
			} else if limitF, ok := val.(float64); ok {
				maxTokens = int64(limitF)
			}
		}
	}

	params := anthropic.MessageNewParams{
		Model:     anthropic.F(c.model),
		MaxTokens: anthropic.F(maxTokens),
		Messages:  anthropic.F(chatMessages),
	}

	if options != nil {
		if val, ok := options["temperature"]; ok {
			if temp, ok := val.(float64); ok {
				params.Temperature = anthropic.F(temp)
			} else if tempF, ok := val.(float32); ok {
				params.Temperature = anthropic.F(float64(tempF))
			} else if tempI, ok := val.(int); ok {
				params.Temperature = anthropic.F(float64(tempI))
			}
		}
	}

	resp, err := c.client.Messages.New(ctx, params)
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

			switch eventVariant := event.AsUnion().(type) {
			case anthropic.ContentBlockDeltaEvent:
				if eventVariant.Delta.Type == anthropic.ContentBlockDeltaEventDeltaTypeTextDelta {
					ch <- StreamResponse{
						Content: eventVariant.Delta.Text,
						Done:    false,
					}
				}
			}
		}

		if err := stream.Err(); err != nil {
			ch <- StreamResponse{Error: err}
		} else {
			ch <- StreamResponse{Done: true}
		}
	}()

	return ch, nil
}
