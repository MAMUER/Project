.PHONY: help proto build run test clean docker-up docker-down deploy-k8s deps tree

help:
	@echo "Available commands:"
	@echo "  make proto        - Generate protobuf code"
	@echo "  make build        - Build all Go services"
	@echo "  make run          - Run all services locally"
	@echo "  make test         - Run tests"
	@echo "  make docker-up    - Start Docker Compose environment"
	@echo "  make docker-down  - Stop Docker Compose environment"
	@echo "  make deploy-k8s   - Deploy to Kubernetes"
	@echo "  make deps         - Install all dependencies (Go, Python, Node)"
	@echo "  make tree         - Show project structure"
	@echo "  make clean        - Clean build artifacts and containers"

# Протобуферы
proto:
	@echo "Generating protobuf code..."
	@mkdir -p proto/gen
	@protoc --proto_path=proto \
		--go_out=proto/gen --go_opt=paths=source_relative \
		--go-grpc_out=proto/gen --go-grpc_opt=paths=source_relative \
		proto/auth.proto
	@protoc --proto_path=proto \
		--go_out=proto/gen --go_opt=paths=source_relative \
		--go-grpc_out=proto/gen --go-grpc_opt=paths=source_relative \
		proto/biometrics.proto
	@protoc --proto_path=proto \
		--go_out=proto/gen --go_opt=paths=source_relative \
		--go-grpc_out=proto/gen --go-grpc_opt=paths=source_relative \
		proto/ml_classification.proto
	@protoc --proto_path=proto \
		--go_out=proto/gen --go_opt=paths=source_relative \
		--go-grpc_out=proto/gen --go-grpc_opt=paths=source_relative \
		proto/ml_generation.proto
	@protoc --proto_path=proto \
		--go_out=proto/gen --go_opt=paths=source_relative \
		--go-grpc_out=proto/gen --go-grpc_opt=paths=source_relative \
		proto/java_legacy_bridge.proto
	@echo "Done!"

# Сборка Go сервисов
build: proto
	@echo "Building Go services..."
	@mkdir -p bin
	@go build -o bin/auth-service ./cmd/auth-service
	@go build -o bin/biometric-ingest ./cmd/biometric-ingest
	@go build -o bin/data-processor ./cmd/data-processor
	@go build -o bin/notification-service ./cmd/notification-service
	@go build -o bin/api-gateway ./cmd/api-gateway
	@echo "Go services built"

# Запуск локально
run: build
	@echo "Starting services..."
	@docker-compose -f deploy/docker-compose.yml up -d
	@./bin/auth-service &
	@./bin/biometric-ingest &
	@./bin/data-processor &
	@./bin/notification-service &
	@./bin/api-gateway &
	@cd ml/classification && python app/main.py &
	@cd ml/generation && python app/main.py &
	@echo "All services started"

# Тесты
test:
	@echo "Running tests..."
	@go test -v -race -cover ./...

# Docker
docker-up:
	@docker-compose -f deploy/docker-compose.yml up -d

docker-down:
	@docker-compose -f deploy/docker-compose.yml down

# Kubernetes
deploy-k8s:
	@kubectl apply -k deploy/k8s/overlays/prod

# Чистка
clean:
	@rm -rf bin/
	@rm -rf proto/gen/
	@docker-compose -f deploy/docker-compose.yml down -v
	@echo "Cleaned"

# Установка зависимостей
deps:
	@echo "Installing Go dependencies..."
	@go mod tidy
	@echo "Creating Python virtual environment..."
	@python3 -m venv venv
	@echo "Installing Python dependencies..."
	@. venv/bin/activate && cd ml/classification && pip install -r requirements.txt
	@. venv/bin/activate && cd ml/generation && pip install -r requirements.txt
	@echo "Installing Node dependencies..."
	@cd web && npm install --legacy-peer-deps
	@echo "All dependencies installed"

# Показать структуру проекта
tree:
	@echo "Project structure:"
	@echo ""
	@tree -a -I 'bin|node_modules|__pycache__|*.pyc|.git' --dirsfirst
	@echo ""

# Показать структуру с ограничением глубины
tree-shallow:
	@echo "Project structure (shallow):"
	@echo ""
	@tree -L 2 -a -I 'bin|node_modules|__pycache__|*.pyc|.git' --dirsfirst
	@echo ""

# Проверка кода
lint:
	@echo "Running Go linters..."
	@go vet ./...
	@echo "Running Python linters..."
	@cd ml/classification && flake8 app/ || true
	@cd ml/generation && flake8 app/ || true
	@echo "Done"

# Форматирование кода
fmt:
	@echo "Formatting Go code..."
	@go fmt ./...
	@echo "Done"