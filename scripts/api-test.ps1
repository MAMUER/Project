# scripts/load-test.ps1
# Load Test Runner for Windows

param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$Duration = "2m",
    [int]$VUs = 50,
    [switch]$Help
)

if ($Help) {
    Write-Host @"
========================================
   FITNESS PLATFORM - LOAD TEST
========================================

Usage: .\scripts\load-test.ps1 [options]

Options:
  -BaseUrl    API base URL (default: http://localhost:8080)
  -Duration   Test duration (default: 2m)
  -VUs        Virtual users (default: 50)
  -Help       Show this help

Examples:
  .\scripts\load-test.ps1
  .\scripts\load-test.ps1 -Duration 5m -VUs 100
  .\scripts\load-test.ps1 -BaseUrl http://api.example.com

========================================
"@
    exit 0
}

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "   FITNESS PLATFORM - LOAD TEST" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Check k6
Write-Host "[1/3] Checking k6 installation..." -ForegroundColor Yellow
try {
    $k6Version = k6 version 2>&1
    Write-Host "  ✓ k6 is installed: $k6Version" -ForegroundColor Green
} catch {
    Write-Host "  ✗ k6 is not installed!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Install k6:" -ForegroundColor Yellow
    Write-Host "  Windows: winget install k6" -ForegroundColor Gray
    Write-Host "  Or download: https://k6.io/docs/getting-started/installation/" -ForegroundColor Gray
    exit 1
}

# Check if service is running
Write-Host ""
Write-Host "[2/3] Checking API availability..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "$BaseUrl/health" -TimeoutSec 5 -UseBasicParsing
    if ($response.StatusCode -eq 200) {
        Write-Host "  ✓ API is running at $BaseUrl" -ForegroundColor Green
    } else {
        Write-Host "  ✗ API returned status: $($response.StatusCode)" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "  ✗ Cannot connect to $BaseUrl" -ForegroundColor Red
    Write-Host "  Make sure services are running: .\scripts\run-local.ps1" -ForegroundColor Yellow
    exit 1
}

# Run load test
Write-Host ""
Write-Host "[3/3] Starting load test..." -ForegroundColor Yellow
Write-Host "  Base URL: $BaseUrl" -ForegroundColor Gray
Write-Host "  Duration: $Duration" -ForegroundColor Gray
Write-Host "  VUs:      $VUs" -ForegroundColor Gray
Write-Host ""

$env:BASE_URL = $BaseUrl

k6 run `
    --duration $Duration `
    --vus $VUs `
    --out json=logs/load-test-results.json `
    scripts/load-test/load-test.js

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Green
    Write-Host "   LOAD TEST COMPLETED SUCCESSFULLY!" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Green
    Write-Host ""
    Write-Host "Results saved to: logs/load-test-results.json" -ForegroundColor Cyan
    Write-Host ""
} else {
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Red
    Write-Host "   LOAD TEST COMPLETED WITH ERRORS!" -ForegroundColor Red
    Write-Host "========================================" -ForegroundColor Red
    Write-Host ""
}

exit $LASTEXITCODE