package main

import (
    "encoding/json"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/gorilla/mux"
    "healthfit-platform/internal/pkg/config"
    "healthfit-platform/internal/pkg/logger"
    "healthfit-platform/internal/pkg/middleware"
)

func main() {
    logger.Init("info", "api-gateway")
    defer logger.Sync()
    log := logger.Get()

    cfg, err := config.Load("api-gateway")
    if err != nil {
        log.Fatal("Failed to load config", zap.Error(err))
    }

    r := mux.NewRouter()

    // Middleware
    r.Use(middleware.Logging)
    r.Use(middleware.CORS)
    r.Use(middleware.Recovery)

    // Health check
    r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
    }).Methods("GET")

    // API Routes
    api := r.PathPrefix("/api/v1").Subrouter()

    // Auth routes
    api.HandleFunc("/auth/login", loginHandler).Methods("POST")
    api.HandleFunc("/auth/verify", verifyHandler).Methods("GET")

    // Biometric routes
    api.HandleFunc("/biometric", biometricHandler).Methods("POST")
    api.HandleFunc("/biometric/{user_id}", getBiometricHandler).Methods("GET")

    // Training routes
    api.HandleFunc("/training/generate", generateProgramHandler).Methods("POST")
    api.HandleFunc("/training/programs/{user_id}", getProgramsHandler).Methods("GET")

    srv := &http.Server{
        Addr:         ":8080",
        Handler:      r,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }

    go func() {
        log.Info("API Gateway started on :8080")
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatal("Server failed", zap.Error(err))
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Info("Shutting down...")
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    srv.Shutdown(ctx)
}

// Handlers (реализация будет позже)
func loginHandler(w http.ResponseWriter, r *http.Request) {}
func verifyHandler(w http.ResponseWriter, r *http.Request) {}
func biometricHandler(w http.ResponseWriter, r *http.Request) {}
func getBiometricHandler(w http.ResponseWriter, r *http.Request) {}
func generateProgramHandler(w http.ResponseWriter, r *http.Request) {}
func getProgramsHandler(w http.ResponseWriter, r *http.Request) {}