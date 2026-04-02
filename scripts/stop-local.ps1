# scripts/stop-local.ps1
# Fitness Platform - Stop All Services

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "   STOPPING ALL SERVICES" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Stop processes by saved PIDs
if (Test-Path "scripts/.pids.json") {
    Write-Host "[1/3] Stopping Go and Python services..." -ForegroundColor Yellow
    $processes = Get-Content "scripts/.pids.json" | ConvertFrom-Json
    
    foreach ($proc in $processes.PSObject.Properties) {
        try {
            Stop-Process -Id $proc.Value -Force -ErrorAction SilentlyContinue
            Write-Host "  OK - Stopped: $($proc.Name) (PID: $($proc.Value))" -ForegroundColor Green
        } catch {
            Write-Host "  WARN - Not found: $($proc.Name)" -ForegroundColor Yellow
        }
    }
    Remove-Item "scripts/.pids.json" -Force -ErrorAction SilentlyContinue
} else {
    Write-Host "[1/3] PID file not found. Searching for processes..." -ForegroundColor Yellow
    
    # Find Go processes
    $goProcesses = Get-Process -Name "go" -ErrorAction SilentlyContinue
    foreach ($proc in $goProcesses) {
        try {
            Stop-Process -Id $proc.Id -Force -ErrorAction SilentlyContinue
            Write-Host "  OK - Stopped Go process: PID $($proc.Id)" -ForegroundColor Green
        } catch {
            # Ignore
        }
    }
    
    # Find Python processes (ML services)
    $pythonProcesses = Get-Process -Name "python" -ErrorAction SilentlyContinue
    foreach ($proc in $pythonProcesses) {
        try {
            $cmdLine = (Get-WmiObject Win32_Process -Filter "ProcessId = $($proc.Id)" -ErrorAction SilentlyContinue | Select-Object -ExpandProperty CommandLine)
            if ($cmdLine -like "*ml-*") {
                Stop-Process -Id $proc.Id -Force -ErrorAction SilentlyContinue
                Write-Host "  OK - Stopped ML process: PID $($proc.Id)" -ForegroundColor Green
            }
        } catch {
            # Ignore
        }
    }
}

# Stop Docker containers
Write-Host "`n[2/3] Stopping Docker containers..." -ForegroundColor Yellow
docker-compose -f deployments/docker-compose.yml down 2>&1 | Tee-Object -FilePath "logs/docker-stop.log"
Write-Host "  OK - Containers stopped" -ForegroundColor Green

# Cleanup
Write-Host "`n[3/3] Cleaning up..." -ForegroundColor Yellow
Remove-Item "scripts/.pids.json" -Force -ErrorAction SilentlyContinue
Write-Host "  OK - Cleanup completed" -ForegroundColor Green

Write-Host "`n========================================" -ForegroundColor Green
Write-Host "   ALL SERVICES STOPPED!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""