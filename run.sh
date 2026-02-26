#!/bin/bash
# run.sh

case "$1" in
  "build")
    echo "Building the application..."
    docker-compose build
    ;;
  "start")
    echo "Starting containers..."
    docker-compose up -d
    echo "Application is running at http://localhost:8080"
    ;;
  "stop")
    echo "Stopping containers..."
    docker-compose down
    ;;
  "restart")
    echo "Restarting containers..."
    docker-compose restart
    ;;
  "logs")
    echo "Showing logs..."
    docker-compose logs -f
    ;;
  "clean")
    echo "Cleaning up containers and volumes..."
    docker-compose down -v
    ;;
  "rebuild")
    echo "Rebuilding and starting..."
    docker-compose down
    docker-compose build --no-cache
    docker-compose up -d
    ;;
  *)
    echo "Usage: ./run.sh {build|start|stop|restart|logs|clean|rebuild}"
    exit 1
    ;;
esac