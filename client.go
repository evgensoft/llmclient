package llmclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client представляет клиент для взаимодействия с LLM API
type Client struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
	maxRetries int
}

// NewClient создает новый экземпляр клиента
func NewClient(baseURL, apiKey, model string, opts ...Option) *Client {
	c := &Client{
		baseURL:    baseURL,
		apiKey:     apiKey,
		model:      model,
		httpClient: http.DefaultClient,
		maxRetries: 3,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Chat выполняет запрос к API чат-комплишенов
func (c *Client) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	var resp ChatResponse
	var lastErr error

	if req.Model == "" {
		req.Model = c.model
	}

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return resp, ctx.Err()
			case <-time.After(backoff(attempt - 1)):
			}
		}

		apiResp, err := c.doRequest(ctx, req)
		if err != nil {
			lastErr = err
			if !shouldRetry(err, nil) {
				return resp, err
			}
			continue
		}

		if !shouldRetry(nil, apiResp) {
			defer apiResp.Body.Close()
			return parseResponse(apiResp)
		}

		apiResp.Body.Close()
		lastErr = fmt.Errorf("HTTP %d", apiResp.StatusCode)
	}

	return resp, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// SimpleRequest выполняет простой запрос с системным и пользовательским промптом
func (c *Client) SimpleRequest(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	messages := make([]Message, 0, 2)

	if systemPrompt != "" {
		messages = append(messages, Message{Role: "system", Content: systemPrompt})
	}

	messages = append(messages, Message{Role: "user", Content: userPrompt})

	req := ChatRequest{
		Messages: messages,
	}

	resp, err := c.Chat(ctx, req)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return resp.Choices[0].Message.Content, nil
}

// RequestWithSchema выполняет запрос с промптом и схемой JSON
func (c *Client) RequestWithSchema(ctx context.Context, systemPrompt, userPrompt string, schema interface{}) error {
	jsonSchema, err := GenerateSchema(schema)
	if err != nil {
		return err
	}

	req := ChatRequest{
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		JSONSchema: jsonSchema,
	}

	resp, err := c.Chat(ctx, req)
	if err != nil {
		return err
	}

	// todo - добавить парсинг JSON в ответе
	if len(resp.Choices) == 0 {
		return fmt.Errorf("no choices in response")
	}

	err = json.Unmarshal([]byte(resp.Choices[0].Message.Content), schema)
	if err != nil {
		return err
	}

	return nil
}

// doRequest выполняет HTTP запрос к API
func (c *Client) doRequest(ctx context.Context, req ChatRequest) (*http.Response, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	return c.httpClient.Do(httpReq)
}

// parseResponse парсит HTTP ответ в структуру ChatResponse
func parseResponse(resp *http.Response) (ChatResponse, error) {
	var result ChatResponse

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return result, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return result, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Choices) == 0 {
		return result, fmt.Errorf("no choices in response")
	}

	return result, nil
}
