# scripts/run-local.ps1
# Fitness Platform - Local Startup Script for Windows

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "   FITNESS PLATFORM - LOCAL STARTUP" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Check Docker
Write-Host "[1/6] Checking Docker..." -ForegroundColor Yellow
try {
    $null = docker ps 2>&1
    Write-Host "  OK - Docker is running" -ForegroundColor Green
} catch {
    Write-Host "  ERROR - Docker is not running!" -ForegroundColor Red
    Write-Host "     Please start Docker Desktop and try again." -ForegroundColor Yellow
    exit 1
}

# Create directories
Write-Host ""
Write-Host "[2/6] Creating directories..." -ForegroundColor Yellow
New-Item -ItemType Directory -Force -Path "logs" | Out-Null
New-Item -ItemType Directory -Force -Path "datasets/processed" | Out-Null
New-Item -ItemType Directory -Force -Path "models" | Out-Null
Write-Host "  OK - Directories created" -ForegroundColor Green

# Check ML models - ИСПРАВЛЕНО!
Write-Host ""
Write-Host "[3/6] Checking ML models..." -ForegroundColor Yellow
$classifierExists = Test-Path "models/classifier.keras"
$generatorExists = Test-Path "models/generator.keras"

if ($classifierExists -and $generatorExists) {
    Write-Host "  OK - ML models found" -ForegroundColor Green
} else {
    Write-Host "  WARNING - ML models not found!" -ForegroundColor Yellow
    Write-Host "  Run training first:" -ForegroundColor Gray
    Write-Host "    python cmd/ml-classifier/train.py" -ForegroundColor Gray
    Write-Host "    python cmd/ml-generator/train_gan.py" -ForegroundColor Gray
}

# Set environment variables
Write-Host ""
Write-Host "[4/6] Setting environment variables..." -ForegroundColor Yellow
$env:DB_HOST = "localhost"
$env:DB_PORT = "5432"
$env:DB_USER = "postgres"
$env:DB_PASSWORD = "postgres"
$env:DB_NAME = "fitness"
$env:JWT_SECRET = "my-super-secret-key-change-in-production-2024"
$env:ML_CLASSIFIER_URL = "http://localhost:8001"
$env:ML_GENERATOR_URL = "http://localhost:8002"
Write-Host "  OK - Variables set" -ForegroundColor Green

# Start Docker containers
Write-Host ""
Write-Host "[5/6] Starting Docker containers..." -ForegroundColor Yellow
docker-compose -f deployments/docker-compose.yml up -d 2>&1 | Tee-Object -FilePath "logs/docker-start.log"
Write-Host "  Waiting for containers to start..." -ForegroundColor Gray
Start-Sleep -Seconds 15

# Check container status
$containerStatus = docker-compose -f deployments/docker-compose.yml ps 2>&1
if ($containerStatus -match "Up") {
    Write-Host "  OK - Containers running" -ForegroundColor Green
} else {
    Write-Host "  WARNING - Some containers may not be running" -ForegroundColor Yellow
}

# Start Go services
Write-Host ""
Write-Host "[6/6] Starting Go services..." -ForegroundColor Yellow

$processes = @{}

# User Service
Write-Host "  Starting User Service (gRPC :50051)..." -ForegroundColor Gray
$userService = Start-Process -NoNewWindow -PassThru -FilePath "go" `
    -ArgumentList "run", "cmd/user-service/main.go" `
    -RedirectStandardOutput "logs/user-service.log" `
    -RedirectStandardError "logs/user-service-error.log"
$processes["user-service"] = $userService.Id
Start-Sleep -Seconds 2

# Biometric Service
Write-Host "  Starting Biometric Service (gRPC :50052)..." -ForegroundColor Gray
$biometricService = Start-Process -NoNewWindow -PassThru -FilePath "go" `
    -ArgumentList "run", "cmd/biometric-service/main.go" `
    -RedirectStandardOutput "logs/biometric-service.log" `
    -RedirectStandardError "logs/biometric-service-error.log"
$processes["biometric-service"] = $biometricService.Id
Start-Sleep -Seconds 2

# Training Service
Write-Host "  Starting Training Service (gRPC :50053)..." -ForegroundColor Gray
$trainingService = Start-Process -NoNewWindow -PassThru -FilePath "go" `
    -ArgumentList "run", "cmd/training-service/main.go" `
    -RedirectStandardOutput "logs/training-service.log" `
    -RedirectStandardError "logs/training-service-error.log"
$processes["training-service"] = $trainingService.Id
Start-Sleep -Seconds 2

# Gateway
Write-Host "  Starting Gateway (HTTP :8080)..." -ForegroundColor Gray
$gateway = Start-Process -NoNewWindow -PassThru -FilePath "go" `
    -ArgumentList "run", "cmd/gateway/main.go" `
    -RedirectStandardOutput "logs/gateway.log" `
    -RedirectStandardError "logs/gateway-error.log"
$processes["gateway"] = $gateway.Id
Start-Sleep -Seconds 3

# ML Classifier
Write-Host "  Starting ML Classifier (HTTP :8001)..." -ForegroundColor Gray
$mlClassifier = Start-Process -NoNewWindow -PassThru -FilePath "python" `
    -ArgumentList "cmd/ml-classifier/main.py" `
    -RedirectStandardOutput "logs/ml-classifier.log" `
    -RedirectStandardError "logs/ml-classifier-error.log"
$processes["ml-classifier"] = $mlClassifier.Id
Start-Sleep -Seconds 3

# ML Generator
Write-Host "  Starting ML Generator (HTTP :8002)..." -ForegroundColor Gray
$mlGenerator = Start-Process -NoNewWindow -PassThru -FilePath "python" `
    -ArgumentList "cmd/ml-generator/main.py" `
    -RedirectStandardOutput "logs/ml-generator.log" `
    -RedirectStandardError "logs/ml-generator-error.log"
$processes["ml-generator"] = $mlGenerator.Id
Start-Sleep -Seconds 3

# Save PIDs
$processes | ConvertTo-Json | Out-File -FilePath "scripts/.pids.json" -Encoding UTF8

# Wait for services to start
Write-Host ""
Write-Host "  Waiting for services to initialize..." -ForegroundColor Gray
Start-Sleep -Seconds 10

# Health checks
Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "   HEALTH CHECKS" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green

$services = @(
    @{Name="Gateway"; URL="http://localhost:8080/health"},
    @{Name="ML Classifier"; URL="http://localhost:8001/health"},
    @{Name="ML Generator"; URL="http://localhost:8002/health"}
)

foreach ($svc in $services) {
    try {
        $response = Invoke-WebRequest -Uri $svc.URL -TimeoutSec 5 -UseBasicParsing -ErrorAction Stop
        if ($response.StatusCode -eq 200) {
            Write-Host "  ✓ $($svc.Name)" -ForegroundColor Green
        } else {
            Write-Host "  ⚠ $($svc.Name) (Status: $($response.StatusCode))" -ForegroundColor Yellow
        }
    } catch {
        Write-Host "  ⚠ $($svc.Name) (Not responding)" -ForegroundColor Yellow
    }
}

# Summary
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "   ALL SERVICES STARTED!" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Gateway:         http://localhost:8080"
Write-Host "  ML Classifier:   http://localhost:8001"
Write-Host "  ML Generator:    http://localhost:8002"
Write-Host "  RabbitMQ UI:     http://localhost:15672 (guest/guest)"
Write-Host "  PostgreSQL:      localhost:5432"
Write-Host "  Redis:           localhost:6379"
Write-Host ""
Write-Host "  Logs:            ./logs/"
Write-Host "  Stop command:    .\scripts\stop-local.ps1"
Write-Host ""
Write-Host "Press Ctrl+C to stop all services..." -ForegroundColor Yellow

# Wait for interrupt
try {
    while ($true) {
        Start-Sleep -Seconds 1
    }
} finally {
    Write-Host ""
    Write-Host "Stopping all services..." -ForegroundColor Yellow
    
    # Stop Go and Python processes
    foreach ($proc in $processes.GetEnumerator()) {
        try {
            Stop-Process -Id $proc.Value -Force -ErrorAction SilentlyContinue
            Write-Host "  ✓ Stopped: $($proc.Key)" -ForegroundColor Green
        } catch {
            Write-Host "  ⚠ Failed to stop: $($proc.Key)" -ForegroundColor Yellow
        }
    }
    
    # Stop Docker containers
    Write-Host ""
    Write-Host "Stopping Docker containers..." -ForegroundColor Yellow
    docker-compose -f deployments/docker-compose.yml down 2>$null
    
    # Cleanup
    Remove-Item "scripts/.pids.json" -Force -ErrorAction SilentlyContinue
    
    Write-Host ""
    Write-Host "All services stopped!" -ForegroundColor Green
}