# scripts/run-local.ps1
# Fitness Platform - Local Startup Script for Windows

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "   FITNESS PLATFORM - LOCAL STARTUP" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Cleanup function
function Cleanup {
    Write-Host "`n[Cleanup] Stopping all services..." -ForegroundColor Yellow
    if (Test-Path "scripts/.pids.json") {
        $processes = Get-Content "scripts/.pids.json" | ConvertFrom-Json
        foreach ($proc in $processes.PSObject.Properties) {
            Stop-Process -Id $proc.Value -Force -ErrorAction SilentlyContinue
        }
        Remove-Item "scripts/.pids.json" -Force -ErrorAction SilentlyContinue
    }
    docker-compose -f deployments/docker-compose.yml down 2>$null
    Write-Host "[Cleanup] Done!" -ForegroundColor Green
}

# Handle Ctrl+C
$ctrlC = Register-EngineEvent -SourceIdentifier PowerShell.Exiting -Action { Cleanup }

# Check Docker
Write-Host "[1/8] Checking Docker..." -ForegroundColor Yellow
try {
    $null = docker ps 2>&1
    Write-Host "  ✓ Docker is running" -ForegroundColor Green
} catch {
    Write-Host "  ✗ Docker is not running!" -ForegroundColor Red
    Write-Host "     Please start Docker Desktop and try again." -ForegroundColor Yellow
    exit 1
}

# Check ports
function Test-Port {
    param([int]$Port)
    $connection = Test-NetConnection -ComputerName localhost -Port $Port -WarningAction SilentlyContinue -ErrorAction SilentlyContinue
    return $connection.TcpTestSucceeded
}

$ports = @{
    "PostgreSQL" = 5432
    "Redis" = 6379
    "RabbitMQ" = 5672
    "Gateway" = 8080
    "ML Classifier" = 8001
    "ML Generator" = 8002
}

$busyPorts = @()
foreach ($name in $ports.Keys) {
    if (Test-Port -Port $ports[$name]) {
        $busyPorts += "$name (:$( $ports[$name] ))"
    }
}

if ($busyPorts.Count -gt 0) {
    Write-Host "[WARNING] Busy ports: $($busyPorts -join ', ')" -ForegroundColor Yellow
    $response = Read-Host "Continue anyway? (Y/N)"
    if ($response -ne 'Y' -and $response -ne 'y') {
        Write-Host "Aborted." -ForegroundColor Red
        exit 1
    }
}

# Set environment variables
$env:DB_HOST = "localhost"
$env:DB_PORT = "5432"
$env:DB_USER = "postgres"
$env:DB_PASSWORD = "postgres"
$env:DB_NAME = "fitness"
$env:JWT_SECRET = "my-super-secret-key-change-in-production-2024"
$env:JWT_EXPIRATION_HOURS = "24"
$env:ML_CLASSIFIER_URL = "http://localhost:8001"
$env:ML_GENERATOR_URL = "http://localhost:8002"

# Create directories
Write-Host "[2/8] Creating directories..." -ForegroundColor Yellow
New-Item -ItemType Directory -Force -Path "logs" | Out-Null
New-Item -ItemType Directory -Force -Path "datasets/processed" | Out-Null
New-Item -ItemType Directory -Force -Path "models" | Out-Null
Write-Host "  ✓ Directories created" -ForegroundColor Green

# Start Docker containers
Write-Host "`n[3/8] Starting Docker containers (PostgreSQL, Redis, RabbitMQ)..." -ForegroundColor Yellow
docker-compose -f deployments/docker-compose.yml up -d 2>&1 | Tee-Object -FilePath "logs/docker-start.log"
Write-Host "  ✓ Containers started" -ForegroundColor Green

# Health check for PostgreSQL
Write-Host "`n[4/8] Waiting for PostgreSQL..." -ForegroundColor Yellow
for ($i = 1; $i -le 30; $i++) {
    try {
        $null = docker-compose -f deployments/docker-compose.yml exec -T postgres pg_isready -U postgres 2>$null
        Write-Host "  ✓ PostgreSQL is ready (attempt $i)" -ForegroundColor Green
        break
    } catch {
        if ($i -eq 30) {
            Write-Host "  ✗ PostgreSQL timeout after 30 attempts!" -ForegroundColor Red
            Cleanup
            exit 1
        }
        Start-Sleep -Seconds 2
    }
}

# Run migrations
Write-Host "`n[5/8] Running database migrations..." -ForegroundColor Yellow
if (Test-Path "scripts\migrate.ps1") {
    & .\scripts\migrate.ps1
} else {
    docker-compose -f deployments/docker-compose.yml exec -T postgres psql -U postgres -d fitness -f /docker-entrypoint-initdb.d/init.sql 2>$null
}
Write-Host "  ✓ Migrations completed" -ForegroundColor Green

# Start Go services
$processes = @{}
$goServices = @(
    @{Name="user-service"; Port=50051; Path="cmd/user-service/main.go"},
    @{Name="biometric-service"; Port=50052; Path="cmd/biometric-service/main.go"},
    @{Name="training-service"; Port=50053; Path="cmd/training-service/main.go"},
    @{Name="gateway"; Port=8080; Path="cmd/gateway/main.go"}
)

Write-Host "`n[6/8] Starting Go services..." -ForegroundColor Yellow
foreach ($svc in $goServices) {
    Write-Host "  Starting $($svc.Name) (:$($svc.Port))..." -ForegroundColor Gray
    $proc = Start-Process -NoNewWindow -PassThru -FilePath "go" `
        -ArgumentList "run", $svc.Path `
        -RedirectStandardOutput "logs/$($svc.Name).log" `
        -RedirectStandardError "logs/$($svc.Name)-error.log"
    $processes[$svc.Name] = $proc.Id
    Start-Sleep -Seconds 2
}
Write-Host "  ✓ All Go services started" -ForegroundColor Green

# Start Python ML services
Write-Host "`n[7/8] Starting ML services..." -ForegroundColor Yellow
Write-Host "  Starting ML Classifier (:8001)..." -ForegroundColor Gray
$mlClassifier = Start-Process -NoNewWindow -PassThru -FilePath "python" `
    -ArgumentList "cmd/ml-classifier/main.py" `
    -RedirectStandardOutput "logs/ml-classifier.log" `
    -RedirectStandardError "logs/ml-classifier-error.log"
$processes["ml-classifier"] = $mlClassifier.Id
Start-Sleep -Seconds 2

Write-Host "  Starting ML Generator (:8002)..." -ForegroundColor Gray
$mlGenerator = Start-Process -NoNewWindow -PassThru -FilePath "python" `
    -ArgumentList "cmd/ml-generator/main.py" `
    -RedirectStandardOutput "logs/ml-generator.log" `
    -RedirectStandardError "logs/ml-generator-error.log"
$processes["ml-generator"] = $mlGenerator.Id
Start-Sleep -Seconds 2
Write-Host "  ✓ All ML services started" -ForegroundColor Green

# Save PIDs
$processes | ConvertTo-Json | Out-File -FilePath "scripts/.pids.json" -Encoding UTF8

# Health checks
Write-Host "`n[8/8] Running health checks..." -ForegroundColor Yellow
Start-Sleep -Seconds 5

$services = @(
    @{Name="Gateway"; URL="http://localhost:8080/health"},
    @{Name="ML Classifier"; URL="http://localhost:8001/health"},
    @{Name="ML Generator"; URL="http://localhost:8002/health"}
)

$allHealthy = $true
foreach ($svc in $services) {
    try {
        $response = Invoke-WebRequest -Uri $svc.URL -TimeoutSec 5 -UseBasicParsing
        if ($response.StatusCode -eq 200) {
            Write-Host "  ✓ $($svc.Name)" -ForegroundColor Green
        } else {
            Write-Host "  ✗ $($svc.Name) (Status: $($response.StatusCode))" -ForegroundColor Red
            $allHealthy = $false
        }
    } catch {
        Write-Host "  ✗ $($svc.Name) (Not responding)" -ForegroundColor Red
        $allHealthy = $false
    }
}

# Summary
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "   SERVICES RUNNING" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Gateway:        http://localhost:8080"
Write-Host "  ML Classifier:  http://localhost:8001"
Write-Host "  ML Generator:   http://localhost:8002"
Write-Host "  RabbitMQ UI:    http://localhost:15672 (guest/guest)"
Write-Host "  PostgreSQL:     localhost:5432"
Write-Host "  Redis:          localhost:6379"
Write-Host ""
Write-Host "  Logs:           ./logs/"
Write-Host "  Stop command:   .\scripts\stop-local.ps1"
Write-Host ""
if ($allHealthy) {
    Write-Host "  Status:         ALL SERVICES HEALTHY" -ForegroundColor Green
} else {
    Write-Host "  Status:         SOME SERVICES UNHEALTHY" -ForegroundColor Yellow
}
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "`nPress Ctrl+C to stop all services..." -ForegroundColor Yellow

# Wait for interrupt
try {
    while ($true) { Start-Sleep -Seconds 1 }
} finally {
    Unregister-Event $ctrlC.Name -ErrorAction SilentlyContinue
    Cleanup
}