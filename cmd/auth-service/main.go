package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"healthfit-platform/internal/pkg/config"
	"healthfit-platform/internal/pkg/database"
	"healthfit-platform/internal/pkg/logger"
	pb "healthfit-platform/proto/gen/auth"
)

type authServer struct {
	pb.UnimplementedAuthServiceServer
	db     *sql.DB
	secret []byte
}

func (s *authServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
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

	token, err := generateJWT(user.ID, user.Role, s.secret)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate token")
	}

	return &pb.LoginResponse{
		Token:     token,
		UserId:    user.ID,
		Role:      user.Role,
		ExpiresIn: 86400,
	}, nil
}

func generateJWT(userID, role string, secret []byte) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func main() {
	logger.Init("info", "auth-service")
	defer logger.Sync()
	log := logger.Get()

	cfg, err := config.Load("auth-service")
	if err != nil {
		log.Fatal("Failed to load config", zap.Error(err))
	}

	db, err := database.NewPostgres(
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.ServicePort))
	if err != nil {
		log.Fatal("Failed to listen", zap.Error(err))
	}

	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, &authServer{db: db, secret: []byte(cfg.JWTSecret)})

	log.Info("Auth service started", zap.Int("port", cfg.ServicePort))

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