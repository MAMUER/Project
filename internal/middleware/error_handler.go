package middleware

import (
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// ErrorHandler — единая точка обработки ошибок для всех ответов
// Требование #6: Кастомизация страниц ошибок
// Требование #7: Немедленное завершение после отправки ошибки
func ErrorHandler(log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Перехватываем статус код через response writer wrapper
			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(rw, r)

			// Если статус — ошибка, логируем
			if rw.statusCode >= 400 {
				log.Warn("Error response",
					zap.Int("status", rw.statusCode),
					zap.String("path", r.URL.Path),
					zap.String("method", r.Method),
					zap.String("correlation_id", GetRequestID(r.Context())),
				)
			}
		})
	}
}

// responseWriter — обёртка для перехвата статусного кода
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.written = true
	}
	return rw.ResponseWriter.Write(b)
}

// JSONError отправляет стандартизированный JSON ответ с ошибкой
// Требование #4: 403 → 404
// Требование #7: return после вызова
func JSONError(w http.ResponseWriter, r *http.Request, code int, message string) {
	// Требование #4: Заменяем 403 Forbidden на 404 Not Found
	if code == http.StatusForbidden {
		code = http.StatusNotFound
		message = "not found"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	body := map[string]interface{}{
		"success": false,
		"error": map[string]interface{}{
			"code":    http.StatusText(code),
			"message": message,
		},
		"timestamp":     time.Now().UTC().Format(time.RFC3339),
		"correlationId": GetRequestID(r.Context()),
	}

	_ = json.NewEncoder(w).Encode(body)
	// Требование #7: вызывающий код должен сделать return
}

// GetRequestID безопасно получает correlation ID из контекста
func GetRequestID(ctx interface{}) string {
	if ctx == nil {
		return ""
	}
	if s, ok := ctx.(string); ok {
		return s
	}
	return ""
}
