package ai

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

type OpenAIProvider struct {
	client     *openai.Client
	model      string
	configured bool
}

func NewOpenAIProvider() *OpenAIProvider {
	return &OpenAIProvider{
		model: openai.GPT4o,
	}
}

func (o *OpenAIProvider) Name() string {
	return "openai"
}

func (o *OpenAIProvider) Configure(config ProviderConfig) error {
	if config.APIKey == "" {
		return fmt.Errorf("openai requires an API key")
	}

	clientConfig := openai.DefaultConfig(config.APIKey)
	if config.Host != "" {
		clientConfig.BaseURL = config.Host
	}

	o.client = openai.NewClientWithConfig(clientConfig)
	if config.Model != "" {
		o.model = config.Model
	}
	o.configured = true
	return nil
}

func (o *OpenAIProvider) IsConfigured() bool {
	return o.configured && o.client != nil
}

func (o *OpenAIProvider) Models() []string {
	return []string{openai.GPT4o, openai.GPT4Turbo, openai.GPT3Dot5Turbo, "gpt-4o-mini"}
}

func (o *OpenAIProvider) Chat(ctx context.Context, messages []Message) (Response, error) {
	if !o.IsConfigured() {
		return Response{}, fmt.Errorf("openai not configured")
	}

	chatMessages := make([]openai.ChatCompletionMessage, len(messages))
	for i, m := range messages {
		chatMessages[i] = openai.ChatCompletionMessage{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	resp, err := o.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    o.model,
		Messages: chatMessages,
	})
	if err != nil {
		return Response{}, err
	}

	if len(resp.Choices) == 0 {
		return Response{}, fmt.Errorf("no response from openai")
	}

	return Response{Content: resp.Choices[0].Message.Content}, nil
}

func (o *OpenAIProvider) Stream(ctx context.Context, messages []Message) (<-chan StreamResponse, error) {
	if !o.IsConfigured() {
		return nil, fmt.Errorf("openai not configured")
	}

	chatMessages := make([]openai.ChatCompletionMessage, len(messages))
	for i, m := range messages {
		chatMessages[i] = openai.ChatCompletionMessage{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	stream, err := o.client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:    o.model,
		Messages: chatMessages,
		Stream:   true,
	})
	if err != nil {
		return nil, err
	}

	ch := make(chan StreamResponse)
	go func() {
		defer close(ch)
		defer stream.Close()

		for {
			response, err := stream.Recv()
			if err != nil {
				ch <- StreamResponse{Error: err}
				return
			}

			if len(response.Choices) == 0 {
				continue
			}

			delta := response.Choices[0].Delta
			ch <- StreamResponse{
				Content: delta.Content,
				Done:    response.Choices[0].FinishReason == "stop",
			}

			if response.Choices[0].FinishReason == "stop" {
				return
			}
		}
	}()

	return ch, nil
}
