package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/MAMUER/Project/internal/auth"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type contextKey string

const (
	RequestIDKey contextKey = "request_id"
	UserIDKey    contextKey = "user_id"
	RoleKey      contextKey = "role"
)

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

func AuthMiddleware(secret string, log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				log.Debug("Missing authorization header")
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
				log.Debug("Invalid token", zap.Error(err))
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, RoleKey, claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
