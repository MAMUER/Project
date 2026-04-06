@echo off
REM Fitness Platform - API Test Suite
REM Auto-fixes Docker PATH for PowerShell 7

set SCRIPT_DIR=%~dp0
set "PATH=C:\Program Files\Docker\Docker\resources\bin;%PATH%"

echo ========================================
echo   FITNESS PLATFORM - API TEST
echo ========================================
echo.

if "%~1"=="" (
    powershell.exe -ExecutionPolicy Bypass -NoProfile -File "%SCRIPT_DIR%api-test.ps1"
) else (
    powershell.exe -ExecutionPolicy Bypass -NoProfile -File "%SCRIPT_DIR%api-test.ps1" -BaseUrl %1
)

pause
