# Fitness Platform - AI-Powered Personalized Training

## Overview
Intelligent web platform for personalized physical activity management and health risk prediction based on biometric data analysis.

## Architecture
- **User Service** (Go, gRPC) - authentication, profiles
- **Biometric Service** (Go, gRPC) - health data collection
- **Training Service** (Go, gRPC) - workout programs
- **ML Classifier** (Python, FastAPI) - state classification (6 inputs)
- **ML Generator** (Python, FastAPI) - GAN for program generation
- **Gateway** (Go) - HTTP API + static files
- **PostgreSQL** - primary database
- **Redis** - caching
- **RabbitMQ** - async messaging

## Quick Start
```bash
# Start dependencies
docker-compose -f deployments/docker-compose.yml up -d

# Run services (in separate terminals)
go run cmd/user-service/main.go
go run cmd/biometric-service/main.go
go run cmd/training-service/main.go
python cmd/ml-classifier/main.py
python cmd/ml-generator/main.py
go run cmd/gateway/main.go
```
API Endpoints
POST /api/v1/register - user registration

POST /api/v1/login - user login

GET /api/v1/profile - user profile

POST /api/v1/biometrics - add biometric data

POST /api/v1/training/generate - generate training plan

GET /api/v1/training/plans - list training plans

GET /api/v1/ml/classify - classify user state

POST /api/v1/ml/generate-plan - generate plan via ML


---

## Запуск Kubernetes

```powershell
# 1. Запустить Minikube
minikube start

# 2. Проверить
minikube status
kubectl cluster-info

# 3. Применить манифесты
kubectl apply -f configs/k8s/namespace.yaml
kubectl apply -f configs/k8s/configmap.yaml
kubectl apply -f configs/k8s/secrets.yaml
kubectl apply -f configs/k8s/pvc.yaml
kubectl apply -f configs/k8s/init-sql.yaml
kubectl apply -f configs/k8s/deployments/
kubectl apply -f configs/k8s/services/

# 4. Проверить
kubectl get pods -n fitness-platform -w
kubectl get svc -n fitness-platform
```