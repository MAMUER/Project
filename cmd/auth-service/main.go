package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"healthfit-platform/internal/pkg/config"
	"healthfit-platform/internal/pkg/database"
	"healthfit-platform/internal/pkg/logger"
	pb "healthfit-platform/proto/gen/auth"
	"golang.org/x/crypto/bcrypt"
)

type authServer struct {
	pb.UnimplementedAuthServiceServer
	db *sql.DB
}

func (s *authServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	// Поиск пользователя в БД
	var user struct {
		ID       string
		Password string
		Role     string
	}
	err := s.db.QueryRowContext(ctx,
		"SELECT id, password_hash, role FROM users WHERE email = $1",
		req.Email).Scan(&user.ID, &user.Password, &user.Role)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	// Генерация JWT
	token := generateJWT(user.ID, user.Role)

	return &pb.LoginResponse{
		Token:     token,
		UserId:    user.ID,
		Role:      user.Role,
		ExpiresIn: 86400,
	}, nil
}

func main() {
	logger.Init("info", "auth-service")
	defer logger.Sync()
	log := logger.Get()

	cfg, err := config.Load("auth-service")
	if err != nil {
		log.Fatal("Failed to load config", zap.Error(err))
	}

	// Подключение к БД
	db, err := database.NewPostgres(
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// gRPC сервер
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.ServicePort))
	if err != nil {
		log.Fatal("Failed to listen", zap.Error(err))
	}

	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, &authServer{db: db})

	log.Info("Auth service started", zap.Int("port", cfg.ServicePort))

	// Graceful shutdown
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal("Failed to serve", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down...")
	grpcServer.GracefulStop()
}