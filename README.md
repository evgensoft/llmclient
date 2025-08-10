# LLMClient Go

Библиотека для упрощенного взаимодействия с LLM API, совместимыми с OpenAI форматом.

## Особенности

- Поддержка всех OpenAI-совместимых API (OpenAI, OpenRouter, Groq, LM Studio, Ollama)
- Автоматические повторы при ошибках с экспоненциальным backoff
- Настройка через функциональные опции
- Минимум внешних зависимостей
- Полное покрытие тестами
- Поддержка структурированного вывода через JSON Schema

## Установка

```bash
go get github.com/evgensoft/llmclient
```

## Быстрый старт

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/evgensoft/llmclient"
)

func main() {
    // Создаем клиента
    client := llmclient.NewClient("https://api.openai.com", "your-api-key", "gpt-3.5-turbo")
    
    // Формируем запрос
    req := llmclient.ChatRequest{
        Messages: []llmclient.Message{
            {Role: "user", Content: "Привет! Расскажи о Go"},
        },
        Temperature: 0.7,
        MaxTokens:   100,
    }
    
    // Выполняем запрос
    resp, err := client.Chat(context.Background(), req)
    if err != nil {
        log.Fatal(err)
    }
    
    // Выводим ответ
    if len(resp.Choices) > 0 {
        fmt.Println("Ответ:", resp.Choices[0].Message.Content)
        fmt.Printf("Использовано токенов: %d\n", resp.Usage.TotalTokens)
    }
}
```

## Упрощённый запрос

Для простых случаев можно использовать метод `SimpleRequest`:

```go
client := llmclient.NewClient("https://api.openai.com", "your-api-key", "gpt-3.5-turbo")

response, err := client.SimpleRequest(
    context.Background(), 
    "Ты полезный помощник.", // системный промпт (может быть пустым)
    "Привет! Расскажи о Go", // пользовательский промпт
)
if err != nil {
    log.Fatal(err)
}

fmt.Println("Ответ:", response)
```

## Структурированный вывод

Для получения структурированного JSON-ответа можно использовать `RequestWithSchema`:

```go
type PersonInfo struct {
    Name string `json:"name" schema:"description=Имя человека"`
    Age  int    `json:"age" schema:"description=Возраст человека"`
}

client := llmclient.NewClient("https://api.openai.com", "your-api-key", "gpt-3.5-turbo")

var person PersonInfo
err := client.RequestWithSchema(
    context.Background(),
    "Извлеки информацию о человеке из текста. Ответ должен быть в формате JSON.",
    "Имя Джона - Смит, ему 35 лет",
    &person,
)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Имя: %s, Возраст: %d\n", person.Name, person.Age)
```

## Поддерживаемые провайдеры

### OpenAI
```go
client := llmclient.NewClient("https://api.openai.com", "sk-...", "gpt-3.5-turbo")
```

### OpenRouter
```go
client := llmclient.NewClient("https://openrouter.ai/api", "sk-or-...", "openai/gpt-3.5-turbo")
```

### Ollama (локально)
```go
client := llmclient.NewClient("http://localhost:11434", "ollama", "llama2")
```

## Настройка

### Кастомный HTTP клиент
```go
customClient := &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:    10,
        IdleConnTimeout: 30 * time.Second,
    },
}

client := llmclient.NewClient(
    "https://api.openai.com",
    "your-api-key",
    "gpt-3.5-turbo",
    llmclient.WithHttpClient(customClient),
)
```

### Настройка количества повторов
```go
client := llmclient.NewClient(
    "https://api.openai.com",
    "your-api-key",
    "gpt-3.5-turbo",
    llmclient.WithMaxRetries(5),
)
```

## Параметры запроса

| Параметр | Тип | Описание |
|----------|-----|----------|
| `Model` | string | Название модели |
| `Messages` | []Message | История сообщений |
| `Temperature` | float32 | Температура генерации (0.0-2.0) |
| `TopP` | float32 | Top-p сэмплирование (0.0-1.0) |
| `MaxTokens` | int | Максимальное количество токенов в ответе |
| `Stop` | []string | Стоп-слова для завершения генерации |
| `N` | int | Количество вариантов ответа |
| `PresencePenalty` | float32 | Штраф за повторение тем |
| `FrequencyPenalty` | float32 | Штраф за частоту слов |
| `JSONSchema` | map[string]interface{} | JSON Schema для структурированного вывода |

## Обработка ошибок

Библиотека автоматически обрабатывает следующие типы ошибок:
- `429 Too Many Requests` - rate limit
- `5xx` - серверные ошибки
- Сетевые ошибки

При возникновении таких ошибок запрос будет автоматически повторен с экспоненциальным backoff (1s, 2s, 4s, 8s...).

Максимальное количество повторов по умолчанию - 3, но его можно изменить с помощью опции `WithMaxRetries`.

## Тестирование

```bash
go test -v ./...
```

## Лицензия

MIT