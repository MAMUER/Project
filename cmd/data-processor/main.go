package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/MAMUER/Project/internal/db"
	"github.com/MAMUER/Project/internal/logger"
	"github.com/MAMUER/Project/internal/queue"
	"go.uber.org/zap"
)

func main() {
	log := logger.New("data-processor")
	defer log.Sync() //nolint:errcheck

	log.Info("Data processor service starting")

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
	defer database.Close() //nolint:errcheck

	rabbitURL := os.Getenv("RABBITMQ_URL")
	var consumer queue.Consumer // ← ИНТЕРФЕЙС
	if rabbitURL != "" {
		consumer, err = queue.NewConsumer(rabbitURL, "biometric_events", log.Logger)
		if err != nil {
			log.Warn("Failed to connect to RabbitMQ", zap.Error(err))
		} else {
			defer func() { _ = consumer.Close() }()
			log.Info("Connected to RabbitMQ")
		}
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Data processor shutting down")
}
