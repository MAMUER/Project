package main

import (
    "encoding/json"
    "net/http"
    "os"
    "time"

    "github.com/gorilla/mux"
    "github.com/MAMUER/Project/internal/logger"
    "github.com/MAMUER/Project/internal/middleware"
    userpb "github.com/MAMUER/Project/api/gen/user"
    "go.uber.org/zap"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

type gateway struct {
    userClient userpb.UserServiceClient
    log        *logger.Logger
    jwtSecret  string
}

func (g *gateway) registerHandler(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Email    string `json:"email"`
        Password string `json:"password"`
        FullName string `json:"full_name"`
        Role     string `json:"role"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        g.log.Error("Failed to decode register request", zap.Error(err))
        http.Error(w, "invalid request", http.StatusBadRequest)
        return
    }

    resp, err := g.userClient.Register(r.Context(), &userpb.RegisterRequest{
        Email:    req.Email,
        Password: req.Password,
        FullName: req.FullName,
        Role:     req.Role,
    })
    if err != nil {
        g.log.Error("Register failed", zap.Error(err))
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(resp)
}

func (g *gateway) loginHandler(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        g.log.Error("Failed to decode login request", zap.Error(err))
        http.Error(w, "invalid request", http.StatusBadRequest)
        return
    }

    resp, err := g.userClient.Login(r.Context(), &userpb.LoginRequest{
        Email:    req.Email,
        Password: req.Password,
    })
    if err != nil {
        g.log.Error("Login failed", zap.Error(err), zap.String("email", req.Email))
        http.Error(w, err.Error(), http.StatusUnauthorized)
        return
    }

    json.NewEncoder(w).Encode(resp)
}

func (g *gateway) profileHandler(w http.ResponseWriter, r *http.Request) {
    userID, ok := r.Context().Value(middleware.UserIDKey).(string)
    if !ok {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }

    resp, err := g.userClient.GetProfile(r.Context(), &userpb.GetProfileRequest{
        UserId: userID,
    })
    if err != nil {
        g.log.Error("Failed to get profile", zap.Error(err), zap.String("user_id", userID))
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(resp)
}

func (g *gateway) healthHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "status":    "ok",
        "service":   "gateway",
        "timestamp": time.Now().UTC().Format(time.RFC3339),
    })
}

func main() {
    log := logger.New("gateway")
    defer log.Sync()

    port := os.Getenv("GATEWAY_PORT")
    if port == "" {
        port = "8080"
    }

    userServiceAddr := os.Getenv("USER_SERVICE_ADDR")
    if userServiceAddr == "" {
        userServiceAddr = "localhost:50051"
    }

    jwtSecret := os.Getenv("JWT_SECRET")
    if jwtSecret == "" {
        jwtSecret = "default-secret-change-in-production"
        log.Warn("Using default JWT secret")
    }

    // Подключение к User Service
    userConn, err := grpc.Dial(userServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatal("Failed to connect to user service", zap.Error(err))
    }
    defer userConn.Close()

    g := &gateway{
        userClient: userpb.NewUserServiceClient(userConn),
        log:        log,
        jwtSecret:  jwtSecret,
    }

    r := mux.NewRouter()

    // Публичные маршруты
    r.HandleFunc("/api/v1/register", g.registerHandler).Methods("POST")
    r.HandleFunc("/api/v1/login", g.loginHandler).Methods("POST")
    r.HandleFunc("/health", g.healthHandler).Methods("GET")

    // Защищенные маршруты
    protected := r.PathPrefix("/api/v1").Subrouter()
    protected.Use(middleware.AuthMiddleware(jwtSecret, log.Logger))
    protected.HandleFunc("/profile", g.profileHandler).Methods("GET")

    // Статические файлы
    r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/")))

    handler := middleware.RequestID(r)

    log.Info("Gateway starting", zap.String("port", port))
    if err := http.ListenAndServe(":"+port, handler); err != nil {
        log.Fatal("Failed to start server", zap.Error(err))
    }
}