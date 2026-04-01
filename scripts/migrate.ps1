# scripts/migrate.ps1
# Fitness Platform - Database Migration Script for Windows

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "   DATABASE MIGRATIONS" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Configuration
$DB_HOST = $env:DB_HOST ?? "localhost"
$DB_PORT = $env:DB_PORT ?? "5432"
$DB_USER = $env:DB_USER ?? "postgres"
$DB_PASSWORD = $env:DB_PASSWORD ?? "postgres"
$DB_NAME = $env:DB_NAME ?? "fitness"

Write-Host "Configuration:" -ForegroundColor Yellow
Write-Host "  Host:     $DB_HOST"
Write-Host "  Port:     $DB_PORT"
Write-Host "  Database: $DB_NAME"
Write-Host "  User:     $DB_USER"
Write-Host ""

# Check if psql is available
Write-Host "[1/3] Checking psql installation..." -ForegroundColor Yellow
try {
    $null = Get-Command psql -ErrorAction Stop
    Write-Host "  ✓ psql is available" -ForegroundColor Green
} catch {
    Write-Host "  ⚠ psql not found. Using Docker..." -ForegroundColor Yellow
    
    # Run migrations via Docker
    Write-Host "`n[2/3] Running migrations via Docker..." -ForegroundColor Yellow
    docker-compose -f deployments/docker-compose.yml exec -T postgres psql -U $DB_USER -d $DB_NAME -f /docker-entrypoint-initdb.d/init.sql 2>&1
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "  ✓ Migrations completed via Docker" -ForegroundColor Green
    } else {
        Write-Host "  ✗ Migrations failed!" -ForegroundColor Red
        exit 1
    }
    
    Write-Host "`n========================================" -ForegroundColor Green
    Write-Host "   MIGRATIONS COMPLETED!" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Green
    exit 0
}

# Set password for psql
$env:PGPASSWORD = $DB_PASSWORD

# Run migrations
Write-Host "`n[2/3] Running migrations..." -ForegroundColor Yellow
$migrationFile = "scripts/init-db.sql"

if (Test-Path $migrationFile) {
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f $migrationFile
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "  ✓ Migrations completed successfully" -ForegroundColor Green
    } else {
        Write-Host "  ✗ Migrations failed!" -ForegroundColor Red
        exit 1
    }
} else {
    Write-Host "  ✗ Migration file not found: $migrationFile" -ForegroundColor Red
    exit 1
}

# Verify
Write-Host "`n[3/3] Verifying migrations..." -ForegroundColor Yellow
$tables = psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';"
Write-Host "  ✓ Tables created: $tables" -ForegroundColor Green

Write-Host "`n========================================" -ForegroundColor Green
Write-Host "   MIGRATIONS COMPLETED!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""