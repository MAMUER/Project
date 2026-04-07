# api-test.ps1 - Fitness Platform Full API Test Suite
# Uses curl.exe for reliable HTTPS with self-signed certs

param(
    [string]$BaseUrl = "https://localhost:8443"
)

$ErrorActionPreference = "Continue"

# Fix: PowerShell 7 may not have Docker in PATH
$dockerPath = "C:\Program Files\Docker\Docker\resources\bin"
if (-not ($env:PATH -split ";" | Where-Object { $_ -eq $dockerPath })) {
    $env:PATH = "$dockerPath;$env:PATH"
}

# Ignore self-signed SSL certificate errors for local dev
[System.Net.ServicePointManager]::SecurityProtocol = [System.Net.SecurityProtocolType]::Tls12 -bor [System.Net.SecurityProtocolType]::Tls13

function Invoke-Curl {
    param(
        [string]$Method,
        [string]$Path,
        [string]$Body = $null,
        [string]$Token = $null
    )
    
    $url = "$BaseUrl$Path"
    
    # Build arguments for curl.exe
    $curlArgs = @("-sk", "-w", "\n%{http_code}", "-m", "30")
    if ($Method -eq "POST") { $curlArgs += "-X", "POST" }
    elseif ($Method -eq "PUT") { $curlArgs += "-X", "PUT" }
    if ($Token) { $curlArgs += "-H", "Authorization: Bearer $Token" }
    $curlArgs += "-H", "Content-Type: application/json"

    # Write body to temp file (curl on Windows can't reliably receive JSON via -d from PowerShell)
    $tmpFile = $null
    if ($Body) {
        $tmpFile = Join-Path $env:TEMP "curl_$PID.json"
        $utf8NoBom = New-Object System.Text.UTF8Encoding $false
        [System.IO.File]::WriteAllText($tmpFile, $Body, $utf8NoBom)
        $curlArgs += "--data-binary", "@$tmpFile"
    }
    
    $curlArgs += $url

    # Call curl.exe
    $output = & "curl.exe" $curlArgs 2>$null

    # Cleanup temp file
    if ($tmpFile) { Remove-Item $tmpFile -Force -ErrorAction SilentlyContinue }
    
    # Parse output: HTTP code is appended by -w, body is everything else
    # curl.exe returns body + newline + http_code, PowerShell captures as array of lines
    if ($output -is [array]) {
        # Find the last line that is a 3-digit number
        for ($i = $output.Length - 1; $i -ge 0; $i--) {
            $line = $output[$i].Trim()
            if ($line -match '^\d{3}$') {
                $httpCode = $line
                $bodyLines = $output[0..($i-1)]
                $bodyStr = ($bodyLines -join "`n").Trim()
                break
            }
        }
    } else {
        # Single string output
        $outputStr = $output.Trim()
        $lastNewlineIdx = $outputStr.LastIndexOf("`n")
        if ($lastNewlineIdx -gt 0) {
            $httpCode = $outputStr.Substring($lastNewlineIdx + 1).Trim()
            $bodyStr = $outputStr.Substring(0, $lastNewlineIdx).Trim()
        } else {
            $httpCode = "000"
            $bodyStr = $outputStr
        }
    }

    if (-not $httpCode) { $httpCode = "000" }
    if (-not $bodyStr) { $bodyStr = "" }

    # Validate httpCode
    if ($httpCode -notmatch '^\d{3}$') {
        $httpCode = "000"
    }
    
    return @{
        StatusCode = [int]$httpCode
        Content    = $bodyStr.Trim()
    }
}

function Test-EP {
    param(
        [string]$Name,
        [string]$Method = "GET",
        [string]$Path,
        [string]$Body = $null,
        [int]$ExpSt = 200
    )
    
    $num = $Script:Passed + $Script:Failed + $Script:Skipped + 1
    Write-Host "[$num] $Name ($Method $Path) " -NoNewline -ForegroundColor Yellow
    
    $result = Invoke-Curl -Method $Method -Path $Path -Body $Body -Token $Script:Token
    $sc = $result.StatusCode
    $ok = $sc -eq $ExpSt
    
    $preview = ""
    if ($result.Content -and $result.Content.Length -gt 0) {
        $preview = $result.Content.Substring(0, [Math]::Min(150, $result.Content.Length))
    }
    $script:Results += "$Method $Path | Exp:$ExpSt | Act:$sc | $(if($ok){'PASS'}else{'FAIL'}) | $preview"
    
    if ($ok) { $Script:Passed++ } else { $Script:Failed++ }
    
    if ($ok) { Write-Host "PASS ($sc)" -ForegroundColor Green }
    else {
        Write-Host "FAIL (exp:$ExpSt got:$sc)" -ForegroundColor Red
        if ($result.Content -and $result.Content.Length -lt 200) {
            Write-Host "  $($result.Content)" -ForegroundColor DarkGray
        }
    }
    return $result
}

# Counters
$Script:Passed = 0
$Script:Failed = 0
$Script:Skipped = 0
$Script:Token = $null
$Script:Results = @()
$TestEmail = "apitest-$(Get-Random -Min 1000 -Max 9999)@example.com"

function Sec { param([string]$T); Write-Host ""; Write-Host "=== $T ===" -ForegroundColor Cyan }

Write-Host ""
Write-Host "=== FITNESS PLATFORM - FULL API TEST ===" -ForegroundColor Cyan
Write-Host "Base URL : $BaseUrl" -ForegroundColor Gray
Write-Host "Test User: $TestEmail" -ForegroundColor Gray
Write-Host ""

# 0. Health
Sec "0. HEALTH"
$hl = Test-EP -Name "Health" -Method GET -Path "/health" -ExpSt 200
if ($hl.StatusCode -ne 200) {
    Write-Host "Gateway not responding." -ForegroundColor Red
    exit 1
}

# 1. Auth
Sec "1. AUTH"
Write-Host "Email: $TestEmail" -ForegroundColor DarkGray

$rb = @{email=$TestEmail;password="TestPass123!";full_name="API Test User";role="client"} | ConvertTo-Json
$rr = Test-EP -Name "Register" -Method POST -Path "/api/v1/register" -Body $rb -ExpSt 200

# Extract verification token from registration response (dev mode)
$verifyToken = ""
if ($rr.StatusCode -eq 200 -and $rr.Content) {
    try {
        $rj = $rr.Content | ConvertFrom-Json
        if ($rj.message) {
            if ($rj.message -match "token \(dev only\):\s*([a-f0-9]+)") {
                $verifyToken = $Matches[1]
                Write-Host "  Verification token captured" -ForegroundColor DarkGray
            }
        }
    } catch {}
}

# Confirm email before login
if ($verifyToken) {
    $confirmBody = @{token=$verifyToken} | ConvertTo-Json
    Test-EP -Name "Confirm Email" -Method POST -Path "/api/v1/auth/confirm" -Body $confirmBody -ExpSt 200
}

Test-EP -Name "Register (dup)" -Method POST -Path "/api/v1/register" -Body $rb -ExpSt 409
Test-EP -Name "Register (bad email)" -Method POST -Path "/api/v1/register" -Body (@{email="bad";password="TestPass123!";full_name="B";role="c"}|ConvertTo-Json) -ExpSt 400
Test-EP -Name "Register (short pw)" -Method POST -Path "/api/v1/register" -Body (@{email="s@e.com";password="123";full_name="S";role="c"}|ConvertTo-Json) -ExpSt 400
Test-EP -Name "Register (empty)" -Method POST -Path "/api/v1/register" -Body (@{}|ConvertTo-Json) -ExpSt 400

$lb = @{email=$TestEmail;password="TestPass123!"} | ConvertTo-Json
$lr = Test-EP -Name "Login" -Method POST -Path "/api/v1/login" -Body $lb -ExpSt 200
if ($lr.StatusCode -eq 200) {
    try { $Script:Token = ($lr.Content | ConvertFrom-Json).access_token } catch {}
}

Test-EP -Name "Login (wrong pw)" -Method POST -Path "/api/v1/login" -Body (@{email=$TestEmail;password="wrong"}|ConvertTo-Json) -ExpSt 401
Test-EP -Name "Login (empty email)" -Method POST -Path "/api/v1/login" -Body (@{email="";password="TestPass123!"}|ConvertTo-Json) -ExpSt 400

if (-not $Script:Token) {
    Write-Host "No token obtained. Skipping auth tests." -ForegroundColor Yellow
    exit 1
}

# 2. Profile
Sec "2. PROFILE"
Test-EP -Name "Get Profile" -Method GET -Path "/api/v1/profile" -ExpSt 200

$upb = @{age=28;gender="male";height_cm=180;weight_kg=75.5;fitness_level="intermediate";goals=@("weight_loss","endurance");contraindications=@("knee");nutrition="balanced";sleep_hours=7.5} | ConvertTo-Json
Test-EP -Name "Update Profile" -Method PUT -Path "/api/v1/profile" -Body $upb -ExpSt 200
Test-EP -Name "Get Profile (after)" -Method GET -Path "/api/v1/profile" -ExpSt 200

# 3. Biometrics
Sec "3. BIOMETRICS"
Test-EP -Name "Add Biometric (HR)" -Method POST -Path "/api/v1/biometrics" -Body (@{metric_type="heart_rate";value=72.0;timestamp=(Get-Date -Format "o");device_type="test"}|ConvertTo-Json) -ExpSt 201
Test-EP -Name "Add Biometric (SpO2)" -Method POST -Path "/api/v1/biometrics" -Body (@{metric_type="spo2";value=98.0;timestamp=(Get-Date -Format "o");device_type="test"}|ConvertTo-Json) -ExpSt 201
Test-EP -Name "Add Biometric (neg)" -Method POST -Path "/api/v1/biometrics" -Body (@{metric_type="heart_rate";value=-10.0;timestamp=(Get-Date -Format "o");device_type="test"}|ConvertTo-Json) -ExpSt 400
Test-EP -Name "Get Biometrics" -Method GET -Path "/api/v1/biometrics?metric_type=heart_rate&limit=10" -ExpSt 200
Test-EP -Name "Logout" -Method POST -Path "/api/v1/logout" -ExpSt 200
$Script:Token = $null

# 4. Post-logout
Sec "4. POST-LOGOUT"
Test-EP -Name "Profile (no token)" -Method GET -Path "/api/v1/profile" -ExpSt 404
Test-EP -Name "Biometrics (no token)" -Method GET -Path "/api/v1/biometrics" -ExpSt 404

# Re-login
$lr2 = Test-EP -Name "Re-login" -Method POST -Path "/api/v1/login" -Body (@{email=$TestEmail;password="TestPass123!"}|ConvertTo-Json) -ExpSt 200
if ($lr2.StatusCode -eq 200) {
    try { $Script:Token = ($lr2.Content | ConvertFrom-Json).access_token } catch {}
}

# 5. Training
Sec "5. TRAINING"
$gpb = @{duration_weeks=4;available_days=@(1,3,5);classification_class="endurance_e1e2";confidence=0.85} | ConvertTo-Json
$gpr = Test-EP -Name "Generate Plan" -Method POST -Path "/api/v1/training/generate" -Body $gpb -ExpSt 200
if ($gpr.StatusCode -eq 200) {
    try { $pd = $gpr.Content | ConvertFrom-Json; if ($pd.plan_id) { $Script:PlanId = $pd.plan_id; Write-Host "  Plan ID: $($Script:PlanId)" -ForegroundColor DarkGray } } catch {}
}

Test-EP -Name "Get Plans" -Method GET -Path "/api/v1/training/plans" -ExpSt 200

if ($Script:PlanId) {
    Test-EP -Name "Complete Workout" -Method POST -Path "/api/v1/training/complete" -Body (@{plan_id=$Script:PlanId;workout_id="w1";feedback="Great!"}|ConvertTo-Json) -ExpSt 200
} else { Write-Host "[SKIP] Complete Workout" -ForegroundColor DarkGray; $Script:Skipped++ }

Test-EP -Name "Get Progress" -Method GET -Path "/api/v1/training/progress" -ExpSt 200

# 6. ML
Sec "6. ML"
Test-EP -Name "ML Classify" -Method POST -Path "/api/v1/ml/classify" -ExpSt 200
Test-EP -Name "ML Gen Plan" -Method POST -Path "/api/v1/ml/generate-plan" -Body (@{training_class="endurance_e1e2";duration_weeks=4;available_days=@(1,3,5);preferences=@{max_duration=60}}|ConvertTo-Json) -ExpSt 200

# 7. Security
Sec "7. SECURITY"
$Script:Token = $null
Test-EP -Name "Profile (no token)" -Method GET -Path "/api/v1/profile" -ExpSt 404
Test-EP -Name "Training (no token)" -Method GET -Path "/api/v1/training/plans" -ExpSt 404

# Summary
Write-Host ""
$total = $Script:Passed + $Script:Failed + $Script:Skipped
Write-Host "=== SUMMARY ===" -ForegroundColor Cyan
Write-Host "Passed : $($Script:Passed) / $total" -ForegroundColor $(if($Script:Failed -eq 0){"Green"}else{"Yellow"})
Write-Host "Failed : $($Script:Failed) / $total" -ForegroundColor $(if($Script:Failed -eq 0){"Green"}else{"Red"})
if ($Script:Skipped -gt 0) { Write-Host "Skipped: $($Script:Skipped) / $total" -ForegroundColor DarkGray }

if ($Script:Failed -gt 0) {
    Write-Host ""
    Write-Host "=== FAILURES ===" -ForegroundColor Red
    foreach ($l in $Script:Results) { if ($l -match "FAIL") { Write-Host "  $l" -ForegroundColor Red } }
}

Write-Host ""
if ($Script:Failed -eq 0) { Write-Host "ALL TESTS PASSED!" -ForegroundColor Green; exit 0 }
else { Write-Host "SOME TESTS FAILED!" -ForegroundColor Red; exit 1 }
