# scripts/load-test.ps1
# Load Testing Script — ТРЕБУЕТ k6!

param(
    [string]$BASE_URL = "http://localhost:8080",
    [string]$DURATION = "2m",
    [int]$VUS = 50
)

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "   FITNESS PLATFORM — LOAD TEST" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Base URL: $BASE_URL"
Write-Host "Duration: $DURATION"
Write-Host "VUs:      $VUS"
Write-Host ""

# Check k6
Write-Host "[1/3] Checking k6 installation..." -ForegroundColor Yellow
try {
    $k6Version = k6 version 2>&1
    Write-Host "  ✓ k6 is installed: $k6Version" -ForegroundColor Green
} catch {
    Write-Host "  ✗ k6 is not installed!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Install k6 (OPTIONAL for load testing):" -ForegroundColor Yellow
    Write-Host "  Windows: winget install k6" -ForegroundColor Gray
    Write-Host "  Or skip load tests and use api-test.ps1 instead" -ForegroundColor Gray
    exit 1
}

# Set environment variables
$env:BASE_URL = $BASE_URL

# Run load test
Write-Host "`n[2/3] Running load test..." -ForegroundColor Yellow
Write-Host "  Script: scripts/load-test/load-test.js" -ForegroundColor Gray
Write-Host "  Command: k6 run --duration $DURATION --vus $VUS scripts/load-test/load-test.js" -ForegroundColor Gray
Write-Host ""

k6 run --duration $DURATION --vus $VUS scripts/load-test/load-test.js

if ($LASTEXITCODE -eq 0) {
    Write-Host "`n[3/3] Load test completed successfully!" -ForegroundColor Green
} else {
    Write-Host "`n[3/3] Load test completed with errors!" -ForegroundColor Yellow
}

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "   LOAD TEST SUMMARY" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Check logs above for detailed metrics."
Write-Host "========================================" -ForegroundColor Cyan

exit $LASTEXITCODE