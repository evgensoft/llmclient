package llmclient

// Message представляет сообщение в чате
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest представляет запрос к API чат-комплишенов
type ChatRequest struct {
	Model            string                 `json:"model"`
	Messages         []Message              `json:"messages"`
	Temperature      float32                `json:"temperature,omitempty"`
	TopP             float32                `json:"top_p,omitempty"`
	MaxTokens        int                    `json:"max_tokens,omitempty"`
	Stop             []string               `json:"stop,omitempty"`
	N                int                    `json:"n,omitempty"`
	PresencePenalty  float32                `json:"presence_penalty,omitempty"`
	FrequencyPenalty float32                `json:"frequency_penalty,omitempty"`
	JSONSchema       map[string]interface{} `json:"json_schema,omitempty"`
}

// Choice представляет один вариант ответа
type Choice struct {
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage представляет информацию об использовании токенов
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatResponse представляет ответ от API
type ChatResponse struct {
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}
