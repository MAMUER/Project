package main

import (
	"context"
	"encoding/json"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"healthfit-platform/internal/pkg/config"
	"healthfit-platform/internal/pkg/database"
	"healthfit-platform/internal/pkg/logger"
	"healthfit-platform/internal/pkg/mq"
	pb "healthfit-platform/proto/gen/ml_classification"
)

type BiometricData struct {
	UserID      string    `json:"user_id"`
	HeartRate   int       `json:"heart_rate"`
	ECG         string    `json:"ecg"`
	BloodPressure struct {
		Systolic  int `json:"systolic"`
		Diastolic int `json:"diastolic"`
	} `json:"blood_pressure"`
	SpO2        int       `json:"spo2"`
	Temperature float64   `json:"temperature"`
	Sleep       struct {
		Duration  int `json:"duration"`
		DeepSleep int `json:"deep_sleep"`
	} `json:"sleep"`
	Timestamp   time.Time `json:"timestamp"`
}

func main() {
	logger.Init("info", "data-processor")
	defer logger.Sync()
	log := logger.Get()

	cfg, err := config.Load("data-processor")
	if err != nil {
		log.Fatal("Failed to load config", zap.Error(err))
	}

	// Подключение к RabbitMQ
	rmq, err := mq.NewClient(cfg.RabbitMQURL)
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ", zap.Error(err))
	}
	defer rmq.Close()

	// Подключение к ML Classification сервису
	conn, err := grpc.Dial(cfg.MLClassificationAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("Failed to connect to ML service", zap.Error(err))
	}
	defer conn.Close()
	mlClient := pb.NewMLClassificationClient(conn)

	// Подключение к PostgreSQL
	db, err := database.NewPostgres(
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Подписка на очередь raw biometric
	msgs, err := rmq.Consume("biometric.raw")
	if err != nil {
		log.Fatal("Failed to consume queue", zap.Error(err))
	}

	log.Info("Data processor started, waiting for messages...")

	for msg := range msgs {
		var data BiometricData
		if err := json.Unmarshal(msg.Body, &data); err != nil {
			log.Error("Failed to unmarshal message", zap.Error(err))
			msg.Ack(false)
			continue
		}

		// Вызываем ML сервис для классификации
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		resp, err := mlClient.Classify(ctx, &pb.ClassifyRequest{
			HeartRate:   int32(data.HeartRate),
			Ecg:         data.ECG,
			Systolic:    int32(data.BloodPressure.Systolic),
			Diastolic:   int32(data.BloodPressure.Diastolic),
			Spo2:        int32(data.SpO2),
			Temperature: float32(data.Temperature),
			SleepDuration: int32(data.Sleep.Duration),
			DeepSleep:    int32(data.Sleep.DeepSleep),
		})
		cancel()

		if err != nil {
			log.Error("ML classification failed", zap.Error(err))
			msg.Nack(false, true) // requeue
			continue
		}

		// Сохраняем результат в БД
		_, err = db.Exec(
			`INSERT INTO biometric_analysis 
			(user_id, timestamp, training_class, confidence, analysis_data)
			VALUES ($1, $2, $3, $4, $5)`,
			data.UserID, data.Timestamp, resp.Class, resp.Confidence, resp.AnalysisData,
		)
		if err != nil {
			log.Error("Failed to save analysis result", zap.Error(err))
		}

		// Отправляем в очередь обработанных данных
		processed, _ := json.Marshal(map[string]interface{}{
			"user_id": data.UserID,
			"timestamp": data.Timestamp,
			"class": resp.Class,
			"confidence": resp.Confidence,
		})
		rmq.Publish("biometric.processed", processed)

		msg.Ack(false)
		log.Info("Processed biometric data", zap.String("user_id", data.UserID), zap.String("class", resp.Class))
	}
}