package llmclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_Chat_Success(t *testing.T) {
	// Создаем мок-сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем заголовки
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("Expected Authorization header to be 'Bearer test-key', got %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type to be 'application/json', got %s", r.Header.Get("Content-Type"))
		}

		// Проверяем URL
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("Expected path to be '/v1/chat/completions', got %s", r.URL.Path)
		}

		// Проверяем тело запроса
		var req ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}

		if req.Model != "gpt-3.5-turbo" {
			t.Errorf("Expected model to be 'gpt-3.5-turbo', got %s", req.Model)
		}

		// Отправляем успешный ответ
		resp := ChatResponse{
			Choices: []Choice{
				{
					Message:      Message{Role: "assistant", Content: "Hello! How can I help you?"},
					FinishReason: "stop",
				},
			},
			Usage: Usage{
				PromptTokens:     10,
				CompletionTokens: 5,
				TotalTokens:      15,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Создаем клиента
	client := NewClient(server.URL, "test-key", "model")

	// Выполняем запрос
	req := ChatRequest{
		Model: "gpt-3.5-turbo",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
	}

	resp, err := client.Chat(context.Background(), req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Проверяем ответ
	if len(resp.Choices) != 1 {
		t.Errorf("Expected 1 choice, got %d", len(resp.Choices))
	}

	if resp.Choices[0].Message.Content != "Hello! How can I help you?" {
		t.Errorf("Unexpected response content: %s", resp.Choices[0].Message.Content)
	}

	if resp.Usage.TotalTokens != 15 {
		t.Errorf("Expected total tokens to be 15, got %d", resp.Usage.TotalTokens)
	}
}

func TestClient_Chat_Retry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}

		resp := ChatResponse{
			Choices: []Choice{
				{
					Message:      Message{Role: "assistant", Content: "Success after retry"},
					FinishReason: "stop",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", "model")
	req := ChatRequest{
		Model: "gpt-3.5-turbo",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
	}

	resp, err := client.Chat(context.Background(), req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}

	if resp.Choices[0].Message.Content != "Success after retry" {
		t.Errorf("Unexpected response content: %s", resp.Choices[0].Message.Content)
	}
}

func TestClient_Chat_MaxRetriesExceeded(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", "model")
	req := ChatRequest{
		Model: "gpt-3.5-turbo",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
	}

	_, err := client.Chat(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if attempts != 4 { // 1 initial + 3 retries
		t.Errorf("Expected 4 attempts, got %d", attempts)
	}
}

func TestClient_Chat_ContextTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", "model")
	req := ChatRequest{
		Model: "gpt-3.5-turbo",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := client.Chat(ctx, req)
	if err == nil {
		t.Fatal("Expected context timeout error, got nil")
	}
}

func TestClient_CustomHttpClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ChatResponse{
			Choices: []Choice{
				{
					Message:      Message{Role: "assistant", Content: "Custom client works"},
					FinishReason: "stop",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	customClient := &http.Client{Timeout: 5 * time.Second}
	client := NewClient(server.URL, "test-key", "model", WithHttpClient(customClient))

	req := ChatRequest{
		Model: "gpt-3.5-turbo",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
	}

	resp, err := client.Chat(context.Background(), req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if resp.Choices[0].Message.Content != "Custom client works" {
		t.Errorf("Unexpected response content: %s", resp.Choices[0].Message.Content)
	}
}

func TestClient_CustomMaxRetries(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", "model", WithMaxRetries(2))
	req := ChatRequest{
		Model: "gpt-3.5-turbo",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
	}

	_, err := client.Chat(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if attempts != 3 { // 1 initial + 2 retries
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestClient_SimpleRequest_WithSystemPrompt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}

		// Проверяем, что есть два сообщения: system и user
		if len(req.Messages) != 2 {
			t.Errorf("Expected 2 messages, got %d", len(req.Messages))
		}

		if req.Messages[0].Role != "system" {
			t.Errorf("Expected first message role to be 'system', got %s", req.Messages[0].Role)
		}

		if req.Messages[0].Content != "You are a helpful assistant." {
			t.Errorf("Expected system message content to be 'You are a helpful assistant.', got %s", req.Messages[0].Content)
		}

		if req.Messages[1].Role != "user" {
			t.Errorf("Expected second message role to be 'user', got %s", req.Messages[1].Role)
		}

		if req.Messages[1].Content != "Hello" {
			t.Errorf("Expected user message content to be 'Hello', got %s", req.Messages[1].Content)
		}

		resp := ChatResponse{
			Choices: []Choice{
				{
					Message:      Message{Role: "assistant", Content: "Hello! How can I help you?"},
					FinishReason: "stop",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", "model")
	result, err := client.SimpleRequest(context.Background(), "You are a helpful assistant.", "Hello")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result != "Hello! How can I help you?" {
		t.Errorf("Unexpected response content: %s", result)
	}
}

func TestClient_SimpleRequest_WithoutSystemPrompt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}

		// Проверяем, что есть только одно сообщение: user
		if len(req.Messages) != 1 {
			t.Errorf("Expected 1 message, got %d", len(req.Messages))
		}

		if req.Messages[0].Role != "user" {
			t.Errorf("Expected message role to be 'user', got %s", req.Messages[0].Role)
		}

		if req.Messages[0].Content != "Hello" {
			t.Errorf("Expected user message content to be 'Hello', got %s", req.Messages[0].Content)
		}

		resp := ChatResponse{
			Choices: []Choice{
				{
					Message:      Message{Role: "assistant", Content: "Hello! How can I help you?"},
					FinishReason: "stop",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, "test-key", "model")
	result, err := client.SimpleRequest(context.Background(), "", "Hello")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result != "Hello! How can I help you?" {
		t.Errorf("Unexpected response content: %s", result)
	}
}
