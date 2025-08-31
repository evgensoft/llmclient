package llmclient

import (
	"fmt"
	"reflect"
	"strings"
)

// GenerateSchema создает JSON Schema для переданного экземпляра структуры.
func GenerateSchema(instance interface{}) (map[string]interface{}, error) {
	// Получаем информацию о типе переданного экземпляра
	t := reflect.TypeOf(instance)

	// Убеждаемся, что работаем с конкретным типом, а не с указателем
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Если это не структура, возвращаем ошибку
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("ожидалась структура, получен %s", t.Kind())
	}

	// Запускаем рекурсивную генерацию
	return generateSchemaForType(t)
}

// generateSchemaForType - рекурсивная функция для построения схемы на основе reflect.Type.
func generateSchemaForType(t reflect.Type) (map[string]interface{}, error) {
	// Используем Kind для определения основного типа данных
	switch t.Kind() {
	case reflect.Struct:
		return generateObjectSchema(t)
	case reflect.Slice, reflect.Array:
		return generateArraySchema(t)
	case reflect.String:
		return map[string]interface{}{"type": "string"}, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return map[string]interface{}{"type": "integer"}, nil
	case reflect.Float32, reflect.Float64:
		return map[string]interface{}{"type": "number"}, nil
	case reflect.Bool:
		return map[string]interface{}{"type": "boolean"}, nil
	case reflect.Ptr:
		// "Разыменовываем" указатель и рекурсивно вызываем для базового типа
		return generateSchemaForType(t.Elem())
	default:
		// Для других типов, таких как map, func и т.д., можно добавить свою логику
		return nil, fmt.Errorf("неподдерживаемый тип: %s", t.Kind())
	}
}

// generateObjectSchema создает схему для объекта (структуры)
func generateObjectSchema(t reflect.Type) (map[string]interface{}, error) {
	schema := map[string]interface{}{
		"type":       "object",
		"properties": make(map[string]interface{}),
	}
	requiredFields := []string{}

	// Итерируемся по всем полям структуры
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Пропускаем неэкспортируемые поля
		if !field.IsExported() {
			continue
		}

		// Обработка встроенных (анонимных) структур
		if field.Anonymous {
			// Рекурсивно получаем схему для встроенной структуры
			embeddedSchema, err := generateSchemaForType(field.Type)
			if err != nil {
				return nil, err
			}
			// Копируем свойства из встроенной схемы в текущую
			for key, value := range embeddedSchema["properties"].(map[string]interface{}) {
				schema["properties"].(map[string]interface{})[key] = value
			}
			// Копируем обязательные поля
			if required, ok := embeddedSchema["required"].([]string); ok {
				requiredFields = append(requiredFields, required...)
			}
			continue
		}

		// Анализируем json тег
		jsonTag := field.Tag.Get("json")
		if jsonTag == "-" {
			continue // Пропускаем поля, помеченные как "-"
		}

		parts := strings.Split(jsonTag, ",")
		jsonName := parts[0]
		if jsonName == "" {
			jsonName = field.Name // Если имя в теге не указано, используем имя поля
		}

		// Проверяем, является ли поле обязательным (отсутствует "omitempty")
		isOptional := false
		for _, part := range parts[1:] {
			if part == "omitempty" {
				isOptional = true
				break
			}
		}
		if !isOptional {
			requiredFields = append(requiredFields, jsonName)
		}

		// Рекурсивно генерируем схему для типа поля
		propSchema, err := generateSchemaForType(field.Type)
		if err != nil {
			return nil, fmt.Errorf("ошибка в поле %s: %w", field.Name, err)
		}

		// Добавляем описание из тега "schema"
		schemaTag := field.Tag.Get("schema")
		if desc := parseSchemaTag(schemaTag, "description"); desc != "" {
			propSchema["description"] = desc
		}

		// Добавляем схему поля в общие свойства
		schema["properties"].(map[string]interface{})[jsonName] = propSchema
	}

	if len(requiredFields) > 0 {
		schema["required"] = requiredFields
	}

	return schema, nil
}

// generateArraySchema создает схему для массива/среза
func generateArraySchema(t reflect.Type) (map[string]interface{}, error) {
	// Получаем схему для типа элементов среза
	elementSchema, err := generateSchemaForType(t.Elem())
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"type":  "array",
		"items": elementSchema,
	}, nil
}

// parseSchemaTag - простой парсер для кастомного тега "schema"
func parseSchemaTag(tag, key string) string {
	parts := strings.Split(tag, ";")
	for _, part := range parts {
		if strings.HasPrefix(part, key+"=") {
			return strings.TrimPrefix(part, key+"=")
		}
	}
	return ""
}

// CleanJSONResponse - очистка лишних символов перед парсингом JSON
func cleanJSONResponse(content string) string {
	// Remove markdown code block markers
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimSuffix(content, "```")

	return strings.TrimSpace(content)
}
