// internal/middleware/context_keys.go
package middleware

// contextKey - тип для ключей контекста (предотвращает коллизии)
type contextKey string

const (
	// CorrelationIDKey - ключ для корреляционного идентификатора
	CorrelationIDKey contextKey = "correlation_id"
	// UserIDKey - ключ для ID пользователя
	UserIDKey contextKey = "user_id"
	// RoleKey - ключ для роли пользователя
	RoleKey contextKey = "role"
	// RequestIDKey - ключ для идентификатора запроса
	RequestIDKey contextKey = "request_id"
)
