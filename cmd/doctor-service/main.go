// cmd/doctor-service/main.go — Doctor Consultant Service (Hexagonal Architecture)
//
// This is the composition root — it wires together:
//   - Infrastructure: PostgreSQL database connection
//   - Adapters (Output): PostgreSQL repositories
//   - Application: Doctor service (use cases)
//   - Adapters (Input): gRPC handler
//
// Architecture: Ports & Adapters (Hexagonal)
//   internal/doctor/model    — Domain models
//   internal/doctor/port     — Ports (interfaces)
//   internal/doctor/service  — Application layer (use cases)
//   internal/doctor/postgres — Output adapters (DB repositories)
//   internal/doctor/grpc     — Input adapter (gRPC handlers)

package main

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/MAMUER/Project/api/gen/doctor"
	"github.com/MAMUER/Project/internal/db"
	grpcadapter "github.com/MAMUER/Project/internal/doctor/grpc"
	"github.com/MAMUER/Project/internal/doctor/postgres"
	"github.com/MAMUER/Project/internal/doctor/service"
	"github.com/MAMUER/Project/internal/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	log := logger.New("doctor-service")
	defer func() {
		if syncErr := log.Sync(); syncErr != nil {
			fmt.Fprintf(os.Stderr, "Failed to sync logger: %v\n", syncErr)
		}
	}()

	port := os.Getenv("DOCTOR_SERVICE_PORT")
	if port == "" {
		port = "50054"
	}

	// ===== Infrastructure =====
	dbCfg := db.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
	}

	database, err := db.NewConnection(dbCfg)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer func() {
		if closeErr := database.Close(); closeErr != nil {
			log.Error("Failed to close database connection", zap.Error(closeErr))
		}
	}()

	// ===== Output Adapters (Repositories) =====
	doctorRepo := postgres.NewDoctorRepo(database)
	subscriptionRepo := postgres.NewSubscriptionRepo(database)
	messageRepo := postgres.NewMessageRepo(database)
	prescriptionRepo := postgres.NewPrescriptionRepo(database)
	consultationRepo := postgres.NewConsultationRepo(database)
	modificationRepo := postgres.NewTrainingModificationRepo(database)

	// ===== Application Layer (Use Cases) =====
	svc := service.New(
		doctorRepo,
		subscriptionRepo,
		messageRepo,
		prescriptionRepo,
		consultationRepo,
		modificationRepo,
	)

	// ===== Input Adapter (gRPC Handler) =====
	handler := grpcadapter.New(svc)

	// ===== Server =====
	lc := net.ListenConfig{}
	lis, err := lc.Listen(context.Background(), "tcp", ":"+port)
	if err != nil {
		log.Fatal("Failed to listen", zap.Error(err))
	}

	grpcServer := grpc.NewServer()
	doctor.RegisterDoctorServiceServer(grpcServer, handler)

	log.Info("Doctor service starting (hexagonal architecture)", zap.String("port", port))
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("Failed to serve", zap.Error(err))
	}
}
