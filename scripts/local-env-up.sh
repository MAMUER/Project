#!/bin/bash

echo "Starting local development environment..."

# Запуск инфраструктуры
docker-compose -f deploy/docker-compose.yml up -d

echo "Waiting for services to be ready..."
sleep 10

# Запуск Go сервисов
echo "Starting Go services..."
go run cmd/auth-service/main.go &
go run cmd/biometric-ingest/main.go &
go run cmd/data-processor/main.go &
go run cmd/notification-service/main.go &
go run cmd/api-gateway/main.go &

# Запуск ML сервисов
echo "Starting ML services..."
cd ml/classification && python app/main.py &
cd ml/generation && python app/main.py &

# Запуск фронтенда
echo "Starting frontend..."
cd web && npm start &

echo "All services started!"
echo "API Gateway: http://localhost:8080"
echo "Frontend: http://localhost:3000"
echo "RabbitMQ UI: http://localhost:15672"
echo "PostgreSQL: localhost:5432"