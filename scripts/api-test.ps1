# scripts/api-test.ps1
# API Testing Script — PowerShell FIXED

param(
    [string]$BASE_URL = "http://localhost:8080"
)

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "   FITNESS PLATFORM — API TEST" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Base URL: $BASE_URL"
Write-Host ""

$testResults = @{
    Passed = 0
    Failed = 0
}
$token = $null
$registrationFailed = $false

function Test-Endpoint {
    param(
        [string]$Name,
        [string]$Method,
        [string]$Uri,
        [hashtable]$Headers = @{},
        [string]$Body = $null,
        [int]$ExpectedStatus = 200,
        [int]$TimeoutSec = 10
    )
    
    Write-Host "[$($testResults.Passed + $testResults.Failed + 1)] Testing: $Name..." -ForegroundColor Yellow
    
    try {
        $params = @{
            Uri = $Uri
            Method = $Method
            Headers = $Headers
            UseBasicParsing = $true
            TimeoutSec = $TimeoutSec
        }
        
        if ($Body) {
            $params.Body = $Body
            $params.ContentType = "application/json"
        }
        
        $response = Invoke-WebRequest @params
        
        if ($response.StatusCode -eq $ExpectedStatus) {
            Write-Host "  ✓ $Name (Status: $($response.StatusCode))" -ForegroundColor Green
            $testResults.Passed++
            return $response
        } else {
            Write-Host "  ✗ $Name (Expected: $ExpectedStatus, Got: $($response.StatusCode))" -ForegroundColor Red
            $testResults.Failed++
            return $null
        }
    } catch {
        Write-Host "  ✗ $Name (Error: $($_.Exception.Message))" -ForegroundColor Red
        $testResults.Failed++
        return $null
    }
}

# 1. Health check
Write-Host "`n--- Core Endpoints ---" -ForegroundColor Cyan
$healthResult = Test-Endpoint -Name "Health Check" -Method GET -Uri "$BASE_URL/health" -TimeoutSec 5
if ($healthResult) {
    Write-Host "  Response: $($healthResult.Content)" -ForegroundColor Gray
}

# 2. Register
Write-Host "`n[2] Testing: Register..." -ForegroundColor Yellow
$randomId = Get-Random -Maximum 10000
$email = "apitest$randomId@example.com"

$registerBody = @{
    email = $email
    password = "test123"
    full_name = "API Test User"
    role = "client"
} | ConvertTo-Json

Write-Host "  Email: $email" -ForegroundColor Gray

try {
    $registerResp = Invoke-WebRequest -Uri "$BASE_URL/api/v1/register" `
        -Method POST -ContentType "application/json" -Body $registerBody `
        -UseBasicParsing -TimeoutSec 30  # Увеличенный таймаут
    
    if ($registerResp.StatusCode -eq 200 -or $registerResp.StatusCode -eq 201) {
        Write-Host "  ✓ Register (Status: $($registerResp.StatusCode))" -ForegroundColor Green
        $testResults.Passed++
        $userData = $registerResp.Content | ConvertFrom-Json
        Write-Host "  User ID: $($userData.user_id)" -ForegroundColor Gray
    } else {
        Write-Host "  ✗ Register (Status: $($registerResp.StatusCode))" -ForegroundColor Red
        $testResults.Failed++
        $registrationFailed = $true
    }
} catch {
    Write-Host "  ✗ Register (Error: $($_.Exception.Message))" -ForegroundColor Red
    Write-Host "  Possible causes:" -ForegroundColor Yellow
    Write-Host "    - User service not running (check: docker-compose ps)" -ForegroundColor Gray
    Write-Host "    - Database connection failed" -ForegroundColor Gray
    Write-Host "    - Gateway timeout" -ForegroundColor Gray
    $testResults.Failed++
    $registrationFailed = $true
}

# 3. Login (только если регистрация успешна)
if (-not $registrationFailed) {
    Write-Host "`n[3] Testing: Login..." -ForegroundColor Yellow
    $loginBody = @{
        email = $email
        password = "test123"
    } | ConvertTo-Json

    try {
        $loginResp = Invoke-WebRequest -Uri "$BASE_URL/api/v1/login" `
            -Method POST -ContentType "application/json" -Body $loginBody `
            -UseBasicParsing -TimeoutSec 10
        
        if ($loginResp.StatusCode -eq 200) {
            Write-Host "  ✓ Login (Status: $($loginResp.StatusCode))" -ForegroundColor Green
            $testResults.Passed++
            $loginData = $loginResp.Content | ConvertFrom-Json
            $token = $loginData.access_token
            Write-Host "  Token: $($token.Substring(0, 50))..." -ForegroundColor Gray
        } else {
            Write-Host "  ✗ Login (Status: $($loginResp.StatusCode))" -ForegroundColor Red
            $testResults.Failed++
        }
    } catch {
        Write-Host "  ✗ Login (Error: $($_.Exception.Message))" -ForegroundColor Red
        $testResults.Failed++
    }
} else {
    Write-Host "`n  Skipping Login (registration failed)" -ForegroundColor Yellow
}

# Authenticated tests (только если есть токен)
if ($token) {
    $headers = @{
        "Authorization" = "Bearer $token"
        "Content-Type" = "application/json"
    }

    # 4. Get Profile
    Write-Host "`n--- User Endpoints ---" -ForegroundColor Cyan
    Test-Endpoint -Name "Get Profile" -Method GET -Uri "$BASE_URL/api/v1/profile" -Headers $headers -TimeoutSec 10

    # 5. Add Biometrics
    Write-Host "`n[5] Testing: Add Biometrics..." -ForegroundColor Yellow
    $bioBody = @{
        metric_type = "heart_rate"
        value = (70 + (Get-Random -Maximum 30))
        timestamp = (Get-Date -Format "o")
        device_type = "test_device"
    } | ConvertTo-Json

    Test-Endpoint -Name "Add Biometrics" -Method POST -Uri "$BASE_URL/api/v1/biometrics" `
        -Headers $headers -Body $bioBody -ExpectedStatus 201 -TimeoutSec 10

    # 6. ML Classify
    Write-Host "`n--- ML Endpoints ---" -ForegroundColor Cyan
    Test-Endpoint -Name "ML Classify" -Method GET -Uri "$BASE_URL/api/v1/ml/classify" -Headers $headers -TimeoutSec 30

    # 7. ML Generate Plan
    Write-Host "`n[7] Testing: Generate Plan..." -ForegroundColor Yellow
    $planBody = @{
        training_class = "endurance_e1e2"
        user_profile = @{
            gender = "male"
            age = 30
            fitness_level = "intermediate"
            goals = @("weight_loss")
        }
    } | ConvertTo-Json

    Test-Endpoint -Name "Generate Plan" -Method POST -Uri "$BASE_URL/api/v1/ml/generate-plan" `
        -Headers $headers -Body $planBody -TimeoutSec 30
} else {
    Write-Host "`n  Skipping authenticated tests (no token)" -ForegroundColor Yellow
}

# Summary
Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "   TEST SUMMARY" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Passed: $($testResults.Passed)" -ForegroundColor $(if ($testResults.Failed -eq 0) { "Green" } else { "Yellow" })
Write-Host "  Failed: $($testResults.Failed)" -ForegroundColor $(if ($testResults.Failed -eq 0) { "Green" } else { "Red" })
Write-Host "  Total:  $($testResults.Passed + $testResults.Failed)"
Write-Host "========================================" -ForegroundColor Cyan

if ($testResults.Failed -eq 0) {
    Write-Host "`n✓ ALL TESTS PASSED!" -ForegroundColor Green
    exit 0
} else {
    Write-Host "`n✗ SOME TESTS FAILED!" -ForegroundColor Red
    Write-Host "`nTroubleshooting:" -ForegroundColor Yellow
    Write-Host "  1. Check service status:" -ForegroundColor Gray
    Write-Host "     docker-compose -f deployments/docker-compose.yml ps" -ForegroundColor Gray
    Write-Host "  2. Check failing service logs:" -ForegroundColor Gray
    Write-Host "     docker-compose logs user-service" -ForegroundColor Gray
    Write-Host "  3. Restart services:" -ForegroundColor Gray
    Write-Host "     docker-compose restart" -ForegroundColor Gray
    exit 1
}