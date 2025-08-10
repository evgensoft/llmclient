package llmclient

import "net/http"

// Option определяет функциональную опцию для настройки клиента
type Option func(*Client)

// WithHttpClient устанавливает кастомный HTTP клиент
func WithHttpClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithMaxRetries устанавливает максимальное количество повторов
func WithMaxRetries(maxRetries int) Option {
	return func(c *Client) {
		c.maxRetries = maxRetries
	}
}