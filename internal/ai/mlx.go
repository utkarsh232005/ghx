package ai

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type MLXProvider struct {
	host       string
	model      string
	configured bool
}

func NewMLXProvider() *MLXProvider {
	return &MLXProvider{
		host:       "http://localhost:8080",
		model:      "default",
		configured: true,
	}
}

func (m *MLXProvider) Name() string {
	return "mlx"
}

func (m *MLXProvider) Configure(config ProviderConfig) error {
	if config.Host != "" {
		m.host = config.Host
	}
	if config.Model != "" {
		m.model = config.Model
	}
	m.configured = true
	return nil
}

func (m *MLXProvider) IsConfigured() bool {
	return m.configured
}

func (m *MLXProvider) Models() []string {
	return []string{"default", "llama-mlx", "mistral-mlx"}
}

type mlxRequest struct {
	Messages []mlxMessage `json:"messages"`
	Stream   bool         `json:"stream"`
}

type mlxMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type mlxResponse struct {
	Text  string `json:"text"`
	Token string `json:"token"`
	Done  bool   `json:"done"`
}

func (m *MLXProvider) Chat(ctx context.Context, messages []Message) (Response, error) {
	return m.ChatWithOptions(ctx, messages, nil)
}

func (m *MLXProvider) ChatWithOptions(ctx context.Context, messages []Message, options map[string]interface{}) (Response, error) {
	req := mlxRequest{
		Stream: false,
	}

	for _, msg := range messages {
		req.Messages = append(req.Messages, mlxMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	body, err := json.Marshal(req)
	if err != nil {
		return Response{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", m.host+"/v1/chat/completions", strings.NewReader(string(body)))
	if err != nil {
		return Response{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return Response{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return Response{}, fmt.Errorf("mlx error: %s", string(bodyBytes))
	}

	var mlxResp mlxResponse
	if err := json.NewDecoder(resp.Body).Decode(&mlxResp); err != nil {
		return Response{}, err
	}

	return Response{Content: mlxResp.Text}, nil
}

func (m *MLXProvider) Stream(ctx context.Context, messages []Message) (<-chan StreamResponse, error) {
	ch := make(chan StreamResponse)

	req := mlxRequest{
		Stream: true,
	}

	for _, msg := range messages {
		req.Messages = append(req.Messages, mlxMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", m.host+"/v1/chat/completions", strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}

	go func() {
		defer close(ch)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				ch <- StreamResponse{Done: true}
				return
			}

			var mlxResp mlxResponse
			if err := json.Unmarshal([]byte(data), &mlxResp); err != nil {
				ch <- StreamResponse{Error: err}
				return
			}

			ch <- StreamResponse{
				Content: mlxResp.Token,
				Done:    mlxResp.Done,
			}
		}
	}()

	return ch, nil
}
