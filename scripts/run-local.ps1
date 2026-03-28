# scripts/run-local.ps1
# Local startup for Fitness Platform services

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "   FITNESS PLATFORM - LOCAL STARTUP" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Check Docker
$dockerRunning = docker ps 2>$null
if (-not $dockerRunning) {
    Write-Host "[ERROR] Docker is not running! Start Docker Desktop and try again." -ForegroundColor Red
    exit 1
}

# Check free ports
$ports = @(5432, 6379, 5672, 15672, 50051, 50052, 50053, 8080, 8001, 8002)
$busyPorts = @()
foreach ($port in $ports) {
    $connection = Test-NetConnection -ComputerName localhost -Port $port -WarningAction SilentlyContinue -ErrorAction SilentlyContinue
    if ($connection.TcpTestSucceeded) {
        $busyPorts += $port
    }
}

if ($busyPorts.Count -gt 0) {
    Write-Host "[WARNING] Busy ports: $busyPorts" -ForegroundColor Yellow
    Write-Host "Services may already be running. Run stop-local.ps1 first." -ForegroundColor Yellow
    $response = Read-Host "Continue? (Y/N)"
    if ($response -ne 'Y') {
        exit 1
    }
}

# Environment variables
$env:DB_HOST = "localhost"
$env:DB_PORT = "5432"
$env:DB_USER = "postgres"
$env:DB_PASSWORD = "postgres"
$env:DB_NAME = "fitness"
$env:DB_SSLMODE = "disable"
$env:JWT_SECRET = "my-super-secret-key-change-in-production-2024"
$env:JWT_EXPIRATION_HOURS = "24"

# Service ports
$env:USER_SERVICE_PORT = "50051"
$env:BIOMETRIC_SERVICE_PORT = "50052"
$env:TRAINING_SERVICE_PORT = "50053"
$env:GATEWAY_PORT = "8080"
$env:ML_CLASSIFIER_PORT = "8001"
$env:ML_GENERATOR_PORT = "8002"

# Service addresses
$env:USER_SERVICE_ADDR = "localhost:50051"
$env:BIOMETRIC_SERVICE_ADDR = "localhost:50052"
$env:TRAINING_SERVICE_ADDR = "localhost:50053"
$env:ML_CLASSIFIER_URL = "http://localhost:8001"
$env:ML_GENERATOR_URL = "http://localhost:8002"

# RabbitMQ
$env:RABBITMQ_URL = "amqp://guest:guest@localhost:5672/"

# Create logs directory
New-Item -ItemType Directory -Force -Path "logs" | Out-Null

Write-Host ""
Write-Host "[1/7] Starting Docker containers (PostgreSQL, Redis, RabbitMQ)..." -ForegroundColor Yellow
docker-compose -f deployments/docker-compose.yml up -d 2>&1 | Tee-Object -FilePath "logs/docker-start.log"

Write-Host "Waiting for containers to start..." -ForegroundColor Yellow
Start-Sleep -Seconds 15

# Check container status
Write-Host "Checking container status..." -ForegroundColor Yellow
docker-compose -f deployments/docker-compose.yml ps

Write-Host ""
Write-Host "[2/7] Starting User Service (gRPC :50051)..." -ForegroundColor Yellow
$userService = Start-Process -NoNewWindow -PassThru -FilePath "go" -ArgumentList "run cmd/user-service/main.go" -RedirectStandardOutput "logs/user-service.log" -RedirectStandardError "logs/user-service-error.log"
Start-Sleep -Seconds 3
Write-Host "  * PID: $($userService.Id)"

Write-Host "[3/7] Starting Biometric Service (gRPC :50052)..." -ForegroundColor Yellow
$biometricService = Start-Process -NoNewWindow -PassThru -FilePath "go" -ArgumentList "run cmd/biometric-service/main.go" -RedirectStandardOutput "logs/biometric-service.log" -RedirectStandardError "logs/biometric-service-error.log"
Start-Sleep -Seconds 3
Write-Host "  * PID: $($biometricService.Id)"

Write-Host "[4/7] Starting Training Service (gRPC :50053)..." -ForegroundColor Yellow
$trainingService = Start-Process -NoNewWindow -PassThru -FilePath "go" -ArgumentList "run cmd/training-service/main.go" -RedirectStandardOutput "logs/training-service.log" -RedirectStandardError "logs/training-service-error.log"
Start-Sleep -Seconds 3
Write-Host "  * PID: $($trainingService.Id)"

Write-Host "[5/7] Starting ML Classifier (Python :8001)..." -ForegroundColor Yellow
$mlClassifier = Start-Process -NoNewWindow -PassThru -FilePath "python" -ArgumentList "cmd/ml-classifier/main.py" -RedirectStandardOutput "logs/ml-classifier.log" -RedirectStandardError "logs/ml-classifier-error.log"
Start-Sleep -Seconds 5
Write-Host "  * PID: $($mlClassifier.Id)"

Write-Host "[6/7] Starting ML Generator (Python :8002)..." -ForegroundColor Yellow
$mlGenerator = Start-Process -NoNewWindow -PassThru -FilePath "python" -ArgumentList "cmd/ml-generator/main.py" -RedirectStandardOutput "logs/ml-generator.log" -RedirectStandardError "logs/ml-generator-error.log"
Start-Sleep -Seconds 5
Write-Host "  * PID: $($mlGenerator.Id)"

Write-Host "[7/7] Starting Gateway (HTTP :8080)..." -ForegroundColor Yellow
$gateway = Start-Process -NoNewWindow -PassThru -FilePath "go" -ArgumentList "run cmd/gateway/main.go" -RedirectStandardOutput "logs/gateway.log" -RedirectStandardError "logs/gateway-error.log"
Start-Sleep -Seconds 5
Write-Host "  * PID: $($gateway.Id)"

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "   ALL SERVICES STARTED!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""

# Save PIDs for stopping
$processes = @{
    "user-service" = $userService.Id
    "biometric-service" = $biometricService.Id
    "training-service" = $trainingService.Id
    "ml-classifier" = $mlClassifier.Id
    "ml-generator" = $mlGenerator.Id
    "gateway" = $gateway.Id
}

$processes | ConvertTo-Json | Out-File -FilePath "scripts/.pids.json"

Write-Host "Available endpoints:" -ForegroundColor Cyan
Write-Host "  * Gateway:       http://localhost:8080"
Write-Host "  * ML Classifier: http://localhost:8001"
Write-Host "  * ML Generator:  http://localhost:8002"
Write-Host "  * RabbitMQ UI:   http://localhost:15672 (guest/guest)"
Write-Host "  * PostgreSQL:    localhost:5432 (postgres/postgres)"
Write-Host "  * Redis:         localhost:6379"
Write-Host ""
Write-Host "API Testing:" -ForegroundColor Cyan
Write-Host "  * Health:        curl http://localhost:8080/health"
Write-Host "  * Register:      curl -X POST http://localhost:8080/api/v1/register -H 'Content-Type: application/json' -d '{"""email""":"""test@test.com""","""password""":"""123456""","""full_name""":"""Test User"""}'"
Write-Host "  * Login:         curl -X POST http://localhost:8080/api/v1/login -H 'Content-Type: application/json' -d '{"""email""":"""test@test.com""","""password""":"""123456"""}'"
Write-Host ""
Write-Host "Logs saved to ./logs/" -ForegroundColor Cyan
Write-Host "To stop all services, run: .\scripts\stop-local.ps1" -ForegroundColor Yellow
Write-Host ""
Write-Host "Press Enter to stop all services..." -ForegroundColor Yellow
Read-Host

# Stop services
Write-Host ""
Write-Host "Stopping all services..." -ForegroundColor Yellow

foreach ($proc in $processes.GetEnumerator()) {
    try {
        Stop-Process -Id $proc.Value -Force -ErrorAction SilentlyContinue
        Write-Host "  * Stopped: $($proc.Key) (PID: $($proc.Value))"
    } catch {
        Write-Host "  * Failed to stop: $($proc.Key)" -ForegroundColor Yellow
    }
}

Write-Host ""
Write-Host "Stopping Docker containers..." -ForegroundColor Yellow
docker-compose -f deployments/docker-compose.yml down 2>&1 | Tee-Object -FilePath "logs/docker-stop.log"

# Cleanup
Remove-Item "scripts/.pids.json" -Force -ErrorAction SilentlyContinue

Write-Host ""
Write-Host "All services stopped!" -ForegroundColor Green