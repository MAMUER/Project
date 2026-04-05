# Загрузка переменных из .env
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

.PHONY: proto build run test test-integration test-cover docker-up docker-down clean dev fmt vet lint

BIN_DIR := bin

# Форматирование кода
fmt:
	@echo "Formatting Go code..."
	go fmt ./...
	@echo "Format complete."

# Проверка кода
vet:
	@echo "Running go vet..."
	go vet ./...
	@echo "Vet complete."

# Запуск всех тестов
test:
	@echo "Running unit tests..."
	go test -v ./...
	@echo "Tests complete."

# Запуск интеграционных тестов (требуют запущенных сервисов)
test-integration:
	@echo "Running integration tests..."
	go test -v -tags=integration ./...
	@echo "Integration tests complete."

# Запуск тестов с покрытием
test-cover:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Запуск линтера
lint:
	@echo "Running golangci-lint..."
	golangci-lint run
	@echo "Lint complete."

# Запуск всех проверок
check: fmt vet lint test
	@echo "All checks passed!"


proto:
	@echo "Generating proto files..."
	powershell -Command "if (!(Test-Path 'api/gen/user')) { New-Item -ItemType Directory -Path 'api/gen/user' -Force }"
	powershell -Command "if (!(Test-Path 'api/gen/biometric')) { New-Item -ItemType Directory -Path 'api/gen/biometric' -Force }"
	powershell -Command "if (!(Test-Path 'api/gen/training')) { New-Item -ItemType Directory -Path 'api/gen/training' -Force }"
	powershell -Command "if (!(Test-Path 'api/gen/ml')) { New-Item -ItemType Directory -Path 'api/gen/ml' -Force }"
	
	@echo "Generating user proto..."
	protoc --proto_path=api/proto --go_out=api/gen/user --go_opt=paths=source_relative --go-grpc_out=api/gen/user --go-grpc_opt=paths=source_relative api/proto/user.proto
	
	@echo "Generating biometric proto..."
	protoc --proto_path=api/proto --go_out=api/gen/biometric --go_opt=paths=source_relative --go-grpc_out=api/gen/biometric --go-grpc_opt=paths=source_relative api/proto/biometric.proto
	
	@echo "Generating training proto..."
	protoc --proto_path=api/proto --go_out=api/gen/training --go_opt=paths=source_relative --go-grpc_out=api/gen/training --go-grpc_opt=paths=source_relative api/proto/training.proto
	
	@echo "Generating ml proto..."
	protoc --proto_path=api/proto --go_out=api/gen/ml --go_opt=paths=source_relative --go-grpc_out=api/gen/ml --go-grpc_opt=paths=source_relative api/proto/ml.proto
	
	@echo "Proto generation complete"

proto-clean:
	@echo "Cleaning generated proto files..."
	powershell -Command "if (Test-Path 'api/gen') { Remove-Item -Recurse -Force 'api/gen' }"
	@echo "Done."

build:
	@echo "Building services..."
	powershell -Command "if (!(Test-Path '$(BIN_DIR)')) { New-Item -ItemType Directory -Path '$(BIN_DIR)' -Force }"
	go build -o $(BIN_DIR)/user-service.exe ./cmd/user-service
	go build -o $(BIN_DIR)/gateway.exe ./cmd/gateway
	go build -o $(BIN_DIR)/biometric-service.exe ./cmd/biometric-service
	go build -o $(BIN_DIR)/training-service.exe ./cmd/training-service
	@echo "Build complete."

run: build
	.\bin\gateway.exe

docker-up:
	docker-compose -f deployments/docker-compose.yml up -d

docker-down:
	docker-compose -f deployments/docker-compose.yml down

clean:
	@echo "Cleaning..."
	powershell -Command "if (Test-Path '$(BIN_DIR)') { Remove-Item -Recurse -Force '$(BIN_DIR)' }"
	powershell -Command "if (Test-Path 'api/gen') { Remove-Item -Recurse -Force 'api/gen' }"
	powershell -Command "if (Test-Path 'coverage.out') { Remove-Item 'coverage.out' }"
	powershell -Command "if (Test-Path 'coverage.html') { Remove-Item 'coverage.html' }"
	@echo "Clean complete."

dev: docker-up
	@echo "Services started. Run 'make run' to start gateway."

help:
	@echo "Available commands:"
	@echo "  make fmt        - Format Go code"
	@echo "  make vet        - Run go vet"
	@echo "  make test       - Run unit tests"
	@echo "  make test-integration - Run integration tests"
	@echo "  make test-cover - Run tests with coverage report"
	@echo "  make check      - Run fmt, vet and test"
	@echo "  make proto      - Generate proto files"
	@echo "  make build      - Build all services"
	@echo "  make run        - Run gateway"
	@echo "  make docker-up  - Start Docker services"
	@echo "  make docker-down - Stop Docker services"
	@echo "  make clean      - Clean generated files"
	@echo "  make dev        - Start Docker services and run gateway"