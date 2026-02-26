# Создайте новый Makefile без localhost
@"
.PHONY: help build up down logs clean restart test shell psql

GREEN := \033[0;32m
NC := \033[0m

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*##"; printf "\n"} /^[a-zA-Z_-]+:.*##/ { printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

build: ## Build the application and Docker images
	@echo "Building application..."
	docker-compose build --no-cache

up: ## Start all containers in detached mode
	@echo "Starting containers..."
	docker-compose up -d
	@echo "Application is running at http://localhost:8080"
	@echo "PostgreSQL is running at \$$(docker-compose port postgres 5432)"

down: ## Stop all containers
	@echo "Stopping containers..."
	docker-compose down

logs: ## Show logs from all containers
	docker-compose logs -f

restart: down up ## Restart all containers

clean: ## Remove containers, volumes, and built images
	@echo "Cleaning up..."
	docker-compose down -v
	docker system prune -f

test: ## Run tests
	@echo "Running tests..."
	./mvnw test

shell: ## Open a shell in the app container
	docker-compose exec app /bin/bash

psql: ## Connect to PostgreSQL database
	@docker-compose exec postgres psql -U \$$(grep DB_USERNAME .env | cut -d '=' -f2) -d \$$(grep DB_NAME .env | cut -d '=' -f2)

status: ## Show container status
	docker-compose ps

backup-db: ## Backup database to file
	@echo "Backing up database..."
	docker-compose exec -T postgres pg_dump -U \$$(grep DB_USERNAME .env | cut -d '=' -f2) \$$(grep DB_NAME .env | cut -d '=' -f2) > backup_$$(date +%Y%m%d_%H%M%S).sql
"@ | Out-File -FilePath "Makefile" -Encoding UTF8