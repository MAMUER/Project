# scripts/stop-local.ps1
# Stop all services

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "   STOPPING ALL SERVICES" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Stop processes by saved PIDs
if (Test-Path "scripts/.pids.json") {
    $processes = Get-Content "scripts/.pids.json" | ConvertFrom-Json
    
    Write-Host "Stopping Go and Python services..." -ForegroundColor Yellow
    foreach ($proc in $processes.PSObject.Properties) {
        try {
            Stop-Process -Id $proc.Value -Force -ErrorAction SilentlyContinue
            Write-Host "  * Stopped: $($proc.Name) (PID: $($proc.Value))"
        } catch {
            Write-Host "  * Not found: $($proc.Name)" -ForegroundColor Yellow
        }
    }
    Remove-Item "scripts/.pids.json" -Force -ErrorAction SilentlyContinue
} else {
    Write-Host "PID file not found. Searching for processes..." -ForegroundColor Yellow
    
    # Find Go processes
    $goProcesses = Get-Process -Name "go" -ErrorAction SilentlyContinue
    foreach ($proc in $goProcesses) {
        try {
            Stop-Process -Id $proc.Id -Force -ErrorAction SilentlyContinue
            Write-Host "  * Stopped Go process: PID $($proc.Id)"
        } catch {}
    }
    
    # Find Python processes (ML services)
    $pythonProcesses = Get-Process -Name "python" -ErrorAction SilentlyContinue
    foreach ($proc in $pythonProcesses) {
        try {
            $cmdLine = (Get-WmiObject Win32_Process -Filter "ProcessId = $($proc.Id)" | Select-Object -ExpandProperty CommandLine)
            if ($cmdLine -like "*ml-*") {
                Stop-Process -Id $proc.Id -Force -ErrorAction SilentlyContinue
                Write-Host "  * Stopped ML process: PID $($proc.Id)"
            }
        } catch {}
    }
}

Write-Host ""
Write-Host "Stopping Docker containers..." -ForegroundColor Yellow
docker-compose -f deployments/docker-compose.yml down 2>&1

Write-Host ""
Write-Host "All services stopped!" -ForegroundColor Green