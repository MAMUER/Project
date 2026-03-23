package main

import (
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"
	"healthfit-platform/internal/pkg/config"
	"healthfit-platform/internal/pkg/database"
	"healthfit-platform/internal/pkg/logger"
	"healthfit-platform/internal/pkg/mq"
)

type BiometricData struct {
	UserID        string `json:"user_id"`
	DeviceType    string `json:"device_type"`
	HeartRate     int    `json:"heart_rate"`
	ECG           string `json:"ecg"`
	BloodPressure struct {
		Systolic  int `json:"systolic"`
		Diastolic int `json:"diastolic"`
	} `json:"blood_pressure"`
	SpO2        int     `json:"spo2"`
	Temperature float64 `json:"temperature"`
	Sleep       struct {
		Duration  int `json:"duration"`
		DeepSleep int `json:"deep_sleep"`
	} `json:"sleep"`
	Timestamp time.Time `json:"timestamp"`
}

func main() {
	logger.Init("info", "biometric-ingest")
	defer logger.Sync()
	log := logger.Get()

	cfg, err := config.Load("biometric-ingest")
	if err != nil {
		log.Fatal("Failed to load config", zap.Error(err))
	}

	// Подключение к PostgreSQL
	db, err := database.NewPostgres(
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Подключение к RabbitMQ
	rmq, err := mq.NewClient(cfg.RabbitMQURL)
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ", zap.Error(err))
	}
	defer rmq.Close()

	// Объявление очередей
	rmq.DeclareQueue("biometric.raw", true)
	rmq.DeclareQueue("biometric.processed", true)

	// HTTP сервер для приема данных
	http.HandleFunc("/api/v1/biometric", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var data BiometricData
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Валидация
		if data.UserID == "" {
			http.Error(w, "user_id required", http.StatusBadRequest)
			return
		}

		// Сохраняем в БД
		_, err := db.Exec(
			`INSERT INTO biometric_data 
			(user_id, device_type, heart_rate, ecg, systolic, diastolic, spo2, temperature, sleep_duration, deep_sleep, timestamp)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
			data.UserID, data.DeviceType, data.HeartRate, data.ECG,
			data.BloodPressure.Systolic, data.BloodPressure.Diastolic,
			data.SpO2, data.Temperature, data.Sleep.Duration, data.Sleep.DeepSleep,
			data.Timestamp,
		)
		if err != nil {
			log.Error("Failed to save biometric data", zap.Error(err))
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		// Отправляем в очередь для обработки
		body, _ := json.Marshal(data)
		if err := rmq.Publish("biometric.raw", body); err != nil {
			log.Error("Failed to publish message", zap.Error(err))
		}

		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "accepted",
			"message": "Biometric data received for processing",
		})
	})

	// Health check
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	log.Info("Biometric ingest service started", zap.Int("port", 8082))
	if err := http.ListenAndServe(":8082", nil); err != nil {
		log.Fatal("Server failed", zap.Error(err))
	}
}