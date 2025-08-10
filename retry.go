package llmclient

import (
	"math"
	"net/http"
	"time"
)

// shouldRetry определяет, следует ли повторить запрос
func shouldRetry(err error, resp *http.Response) bool {
	if err != nil {
		return true
	}
	if resp.StatusCode >= 500 || resp.StatusCode == 429 {
		return true
	}
	return false
}

// backoff вычисляет задержку для повторного запроса с экспоненциальным backoff
func backoff(attempt int) time.Duration {
	return time.Duration(math.Pow(2, float64(attempt))) * time.Second
}