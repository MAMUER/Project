// cmd/biometric-service/integration_test.go
//go:build integration

package main

import (
	"context"
	"testing"
	"time"

	pb "github.com/MAMUER/Project/api/gen/biometric"
	"github.com/MAMUER/Project/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ✅ Добавляем require для тестов
func TestBiometricService_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Запускаем PostgreSQL в контейнере
	postgresContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	require.NoError(t, err)
	defer func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate container: %v", err)
		}
	}()

	// Получаем connection string
	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	// Инициализируем конфигурацию
	cfg := &config.Config{
		DBHost:     "localhost", // будет переопределено testcontainers
		DBPort:     "5432",
		DBUser:     "test",
		DBPassword: "test",
		DBName:     "testdb",
		DBSSLMode:  "disable",
	}

	// Создаём подключение к БД
	// Примечание: в реальном коде используйте миграции для создания таблиц
	// Здесь предполагаем, что таблица biometric_data уже создана

	// Запускаем gRPC сервер в отдельной горутине для теста
	// (в реальном тесте лучше использовать in-memory server или mock)

	// Подключаемся к сервису
	// Для тестов можно запустить сервер локально или использовать mock
	// Здесь пример подключения:
	/*
		conn, err := grpc.Dial("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
		require.NoError(t, err)
		defer conn.Close()

		client := pb.NewBiometricServiceClient(conn)

		// Тестовый запрос
		resp, err := client.AddRecord(ctx, &pb.AddRecordRequest{
			UserId:     "test-user",
			MetricType: "heart_rate",
			Value:      75.0,
		})
		assert.NoError(t, err)
		assert.NotNil(t, resp)
	*/

	// ✅ Для интеграционных тестов с testcontainers
	// рекомендуется использовать отдельный helper для инициализации сервиса
	// или запускать сервис в контейнере вместе с БД
}

// ✅ Отдельный тест для валидации при интеграции
func TestBiometricService_Integration_Validation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Запускаем БД
	postgresContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	require.NoError(t, err)
	defer func() { _ = postgresContainer.Terminate(ctx) }()

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)
	t.Logf("PostgreSQL connection string: %s", connStr)

	// ✅ Здесь можно добавить тесты с реальным подключением
	// Например, создание записей, чтение, проверка валидации
}
