// internal/middleware/middleware.go
package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/MAMUER/Project/internal/auth"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RequestID добавляет уникальный идентификатор запроса
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AuthMiddleware проверяет JWT токен и добавляет пользователя в контекст
func AuthMiddleware(secret string, log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				log.Debug("Missing authorization header", zap.String("path", r.URL.Path))
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				log.Debug("Invalid authorization format", zap.String("header", authHeader))
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			token := parts[1]
			claims, err := auth.ValidateJWT(token, secret)
			if err != nil {
				log.Debug("Invalid token", zap.Error(err), zap.String("path", r.URL.Path))
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, RoleKey, claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// LoggingMiddleware логирует запросы с корреляционным ID
func LoggingMiddleware(log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			cid := GetCorrelationID(r.Context())

			next.ServeHTTP(w, r)

			log.Info("HTTP request",
				zap.String("correlation_id", cid),
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Duration("duration", time.Since(start)),
				zap.Int("status", 200), // в реальном коде нужно перехватывать статус
			)
		})
	}
}
