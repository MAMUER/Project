#!/bin/bash

# server-maintenance.sh - запускать по крону раз в неделю

echo "=== Server Maintenance $(date) ==="

# 1. Очистка Docker cache старше 24 часов
echo "Cleaning old build cache..."
docker builder prune -f --filter "until=24h"

# 2. Удаление неиспользуемых образов
echo "Removing unused images..."
docker image prune -f

# 3. Полная очистка (безопасная)
echo "System prune..."
docker system prune -f

# 4. Проверка занятого места
echo "=== Disk Usage ==="
df -h /
docker system df

echo "=== Maintenance Complete ==="