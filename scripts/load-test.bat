@echo off
echo ========================================
echo   FITNESS PLATFORM - LOAD TEST
echo ========================================
echo.
powershell -ExecutionPolicy Bypass -File "%~dp0load-test.ps1"
echo.
echo ========================================
pause
