package middleware

import (
    "context"
    "net/http"
    "time"

    "github.com/google/uuid"
    "go.uber.org/zap"
    "healthfit-platform/internal/pkg/logger"
)

type contextKey string

const RequestIDKey contextKey = "request_id"

func Logging(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        requestID := r.Header.Get("X-Request-ID")
        if requestID == "" {
            requestID = uuid.New().String()
        }
        w.Header().Set("X-Request-ID", requestID)
        ctx := context.WithValue(r.Context(), RequestIDKey, requestID)

        log := logger.Get()
        log.Info("Request started",
            zap.String("method", r.Method),
            zap.String("path", r.URL.Path),
            zap.String("request_id", requestID),
        )

        next.ServeHTTP(w, r.WithContext(ctx))

        log.Info("Request completed",
            zap.String("method", r.Method),
            zap.String("path", r.URL.Path),
            zap.String("request_id", requestID),
            zap.Duration("duration", time.Since(start)),
        )
    })
}

func CORS(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")

        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }

        next.ServeHTTP(w, r)
    })
}

func Recovery(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                log := logger.Get()
                log.Error("Panic recovered",
                    zap.Any("error", err),
                    zap.String("path", r.URL.Path),
                )
                w.WriteHeader(http.StatusInternalServerError)
                w.Write([]byte(`{"error": "internal server error"}`))
            }
        }()
        next.ServeHTTP(w, r)
    })
}