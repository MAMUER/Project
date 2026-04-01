#!/bin/bash
# scripts/migrate.sh
# Fitness Platform - Database Migration Script for Linux

echo "========================================"
echo "   DATABASE MIGRATIONS"
echo "========================================"
echo ""

# Configuration
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-postgres}
DB_NAME=${DB_NAME:-fitness}

echo "Configuration:"
echo "  Host:     $DB_HOST"
echo "  Port:     $DB_PORT"
echo "  Database: $DB_NAME"
echo "  User:     $DB_USER"
echo ""

# Check if psql is available
echo "[1/3] Checking psql installation..."
if command -v psql &> /dev/null; then
    echo "  ✓ psql is available"
else
    echo "  ⚠ psql not found. Using Docker..."
    
    # Run migrations via Docker
    echo ""
    echo "[2/3] Running migrations via Docker..."
    docker-compose -f deployments/docker-compose.yml exec -T postgres psql -U $DB_USER -d $DB_NAME -f /docker-entrypoint-initdb.d/init.sql
    
    if [ $? -eq 0 ]; then
        echo "  ✓ Migrations completed via Docker"
    else
        echo "  ✗ Migrations failed!"
        exit 1
    fi
    
    echo ""
    echo "========================================"
    echo "   MIGRATIONS COMPLETED!"
    echo "========================================"
    exit 0
fi

# Set password for psql
export PGPASSWORD=$DB_PASSWORD

# Run migrations
echo ""
echo "[2/3] Running migrations..."
MIGRATION_FILE="scripts/init-db.sql"

if [ -f "$MIGRATION_FILE" ]; then
    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f $MIGRATION_FILE
    
    if [ $? -eq 0 ]; then
        echo "  ✓ Migrations completed successfully"
    else
        echo "  ✗ Migrations failed!"
        exit 1
    fi
else
    echo "  ✗ Migration file not found: $MIGRATION_FILE"
    exit 1
fi

# Verify
echo ""
echo "[3/3] Verifying migrations..."
TABLES=$(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';")
echo "  ✓ Tables created: $TABLES"

echo ""
echo "========================================"
echo "   MIGRATIONS COMPLETED!"
echo "========================================"
echo ""