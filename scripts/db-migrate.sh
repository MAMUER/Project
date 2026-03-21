#!/bin/bash

set -e

DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-healthfit}
DB_PASSWORD=${DB_PASSWORD:-healthfit123}
DB_NAME=${DB_NAME:-healthfit}

MIGRATIONS_DIR="deploy/docker/postgres/migrations"

# Применение миграций
for migration in ${MIGRATIONS_DIR}/*.up.sql; do
    echo "Applying migration: $(basename $migration)"
    PGPASSWORD=${DB_PASSWORD} psql \
        -h ${DB_HOST} \
        -p ${DB_PORT} \
        -U ${DB_USER} \
        -d ${DB_NAME} \
        -f ${migration}
done

echo "Migrations completed"