package ai

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

type LMStudioProvider struct {
	client     *openai.Client
	model      string
	configured bool
}

func NewLMStudioProvider() *LMStudioProvider {
	return &LMStudioProvider{
		model: "local-model",
	}
}

func (l *LMStudioProvider) Name() string {
	return "lmstudio"
}

func (l *LMStudioProvider) Configure(config ProviderConfig) error {
	if config.Host == "" {
		config.Host = "http://localhost:1234/v1"
	}

	clientConfig := openai.DefaultConfig("lm-studio")
	clientConfig.BaseURL = config.Host

	l.client = openai.NewClientWithConfig(clientConfig)
	if config.Model != "" {
		l.model = config.Model
	}
	l.configured = true
	return nil
}

func (l *LMStudioProvider) IsConfigured() bool {
	return l.configured && l.client != nil
}

func (l *LMStudioProvider) Models() []string {
	return []string{"local-model"}
}

func (l *LMStudioProvider) Chat(ctx context.Context, messages []Message) (Response, error) {
	return l.ChatWithOptions(ctx, messages, nil)
}

func (l *LMStudioProvider) ChatWithOptions(ctx context.Context, messages []Message, options map[string]interface{}) (Response, error) {
	if !l.IsConfigured() {
		return Response{}, fmt.Errorf("lmstudio not configured")
	}

	chatMessages := make([]openai.ChatCompletionMessage, len(messages))
	for i, m := range messages {
		chatMessages[i] = openai.ChatCompletionMessage{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	req := openai.ChatCompletionRequest{
		Model:    l.model,
		Messages: chatMessages,
	}

	if options != nil {
		if val, ok := options["max_tokens"]; ok {
			if limit, ok := val.(int); ok {
				req.MaxTokens = limit
			} else if limitF, ok := val.(float64); ok {
				req.MaxTokens = int(limitF)
			}
		}
		if val, ok := options["temperature"]; ok {
			if temp, ok := val.(float64); ok {
				req.Temperature = float32(temp)
			} else if tempF, ok := val.(float32); ok {
				req.Temperature = tempF
			} else if tempI, ok := val.(int); ok {
				req.Temperature = float32(tempI)
			}
		}
	}

	resp, err := l.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return Response{}, err
	}

	if len(resp.Choices) == 0 {
		return Response{}, fmt.Errorf("no response from lmstudio")
	}

	return Response{Content: resp.Choices[0].Message.Content}, nil
}

func (l *LMStudioProvider) Stream(ctx context.Context, messages []Message) (<-chan StreamResponse, error) {
	if !l.IsConfigured() {
		return nil, fmt.Errorf("lmstudio not configured")
	}

	chatMessages := make([]openai.ChatCompletionMessage, len(messages))
	for i, m := range messages {
		chatMessages[i] = openai.ChatCompletionMessage{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	stream, err := l.client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:    l.model,
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
