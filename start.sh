#!/bin/bash

# Start Docker Compose services
echo "Starting application..."
docker compose up --build -d

# Check if docker compose command succeeded
if [ $? -eq 0 ]; then
    echo "✅ Application is running"
else
    echo "❌ Failed to start application"
    exit 1
fi