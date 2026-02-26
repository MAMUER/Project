#!/bin/bash

# Проверка свободного места перед деплоем
FREE_SPACE=$(df / | awk 'NR==2 {print $4}')
FREE_SPACE_GB=$((FREE_SPACE / 1024 / 1024))

echo "Free space: ${FREE_SPACE_GB}GB"

if [ $FREE_SPACE_GB -lt 5 ]; then
    echo "WARNING: Less than 5GB free!"
    echo "Cleaning cache..."
    docker builder prune -f
    docker system prune -f
    
    FREE_SPACE_AFTER=$(df / | awk 'NR==2 {print $4}')
    FREE_SPACE_AFTER_GB=$((FREE_SPACE_AFTER / 1024 / 1024))
    echo "Free space after cleanup: ${FREE_SPACE_AFTER_GB}GB"
fi

if [ $FREE_SPACE_GB -lt 2 ]; then
    echo "ERROR: Less than 2GB free! Aborting deployment."
    exit 1
fi

echo "Proceeding with deployment..."