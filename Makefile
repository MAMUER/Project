# Makefile для Windows (без переносов строк)

.PHONY: proto build run test docker-up docker-down clean dev

BIN_DIR := bin

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

test:
	go test -v ./...

docker-up:
	docker-compose -f deployments/docker-compose.yml up -d

docker-down:
	docker-compose -f deployments/docker-compose.yml down

clean:
	@echo "Cleaning..."
	powershell -Command "if (Test-Path '$(BIN_DIR)') { Remove-Item -Recurse -Force '$(BIN_DIR)' }"
	powershell -Command "if (Test-Path 'api/gen') { Remove-Item -Recurse -Force 'api/gen' }"
	@echo "Clean complete."

dev: docker-up
	@echo "Services started. Run 'make run' to start gateway."