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

type OllamaProvider struct {
	host       string
	model      string
	options    map[string]interface{}
	configured bool
}

func NewOllamaProvider() *OllamaProvider {
	return &OllamaProvider{
		host:       "http://localhost:11434",
		model:      "llama3",
		configured: true,
		options: map[string]interface{}{
			"temperature": 0.7,
			"num_ctx":     4096,
		},
	}
}

func (o *OllamaProvider) Name() string {
	return "ollama"
}

func (o *OllamaProvider) Configure(config ProviderConfig) error {
	if config.Host != "" {
		o.host = config.Host
	}
	if config.Model != "" {
		o.model = config.Model
	}
	if config.Options != nil {
		o.options = config.Options
	}
	o.configured = true
	return nil
}

func (o *OllamaProvider) IsConfigured() bool {
	return o.configured
}

func (o *OllamaProvider) Models() []string {
	return []string{"llama3", "llama3:70b", "mistral", "codellama", "phi3", "gemma"}
}

type ollamaRequest struct {
	Model    string                 `json:"model"`
	Messages []ollamaMessage        `json:"messages"`
	Stream   bool                   `json:"stream"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaResponse struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
	Done bool `json:"done"`
}

func (o *OllamaProvider) Chat(ctx context.Context, messages []Message) (Response, error) {
	req := ollamaRequest{
		Model:   o.model,
		Stream:  false,
		Options: o.options,
	}

	for _, m := range messages {
		req.Messages = append(req.Messages, ollamaMessage{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	body, err := json.Marshal(req)
	if err != nil {
		return Response{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", o.host+"/api/chat", strings.NewReader(string(body)))
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
		return Response{}, fmt.Errorf("ollama error: %s", string(bodyBytes))
	}

	var ollamaResp ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return Response{}, err
	}

	return Response{Content: ollamaResp.Message.Content}, nil
}

func (o *OllamaProvider) Stream(ctx context.Context, messages []Message) (<-chan StreamResponse, error) {
	ch := make(chan StreamResponse)

	req := ollamaRequest{
		Model:   o.model,
		Stream:  true,
		Options: o.options,
	}

	for _, m := range messages {
		req.Messages = append(req.Messages, ollamaMessage{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", o.host+"/api/chat", strings.NewReader(string(body)))
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
			var ollamaResp ollamaResponse
			if err := json.Unmarshal(scanner.Bytes(), &ollamaResp); err != nil {
				ch <- StreamResponse{Error: err}
				return
			}

			ch <- StreamResponse{
				Content: ollamaResp.Message.Content,
				Done:    ollamaResp.Done,
			}

			if ollamaResp.Done {
				return
			}
		}
	}()

	return ch, nil
}
