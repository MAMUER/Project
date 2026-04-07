# scripts/load-test.ps1
# Load Testing Script — requires k6 (optional)

param(
    [string]$BASE_URL = "https://localhost:8443",
    [string]$DURATION = "2m",
    [int]$VUS = 50
)

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "   FITNESS PLATFORM - LOAD TEST" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Base URL: $BASE_URL"
Write-Host "Duration: $DURATION"
Write-Host "VUs:      $VUS"
Write-Host ""

# Check k6
Write-Host "[1/3] Checking k6 installation..." -ForegroundColor Yellow
$k6Path = Get-Command k6 -ErrorAction SilentlyContinue
if (-not $k6Path) {
    Write-Host "  [SKIP] k6 is not installed (optional tool)" -ForegroundColor Yellow
    Write-Host "  Install: winget install k6" -ForegroundColor Gray
    Write-Host "  Or use api-test.ps1 for functional tests" -ForegroundColor Gray
    exit 0
}

$k6Version = & k6 version 2>&1
Write-Host "  [OK] k6 is installed: $k6Version" -ForegroundColor Green

# Set environment variables
$env:BASE_URL = $BASE_URL

# Run load test
Write-Host "`n[2/3] Running load test..." -ForegroundColor Yellow
Write-Host "  Script: scripts/load-test/load-test.js" -ForegroundColor Gray
Write-Host ""

$scriptPath = Join-Path $PSScriptRoot "load-test\load-test.js"
& k6 run --duration $DURATION --vus $VUS $scriptPath

$exitCode = $LASTEXITCODE

if ($exitCode -eq 0) {
    Write-Host "`n[3/3] Load test completed successfully!" -ForegroundColor Green
} else {
    Write-Host "`n[3/3] Load test completed with errors (exit code: $exitCode)" -ForegroundColor Yellow
}

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "   LOAD TEST SUMMARY" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

exit $exitCode
