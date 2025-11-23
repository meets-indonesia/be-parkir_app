#!/bin/bash
# Script to run migration 07 to drop hourly_rate column

echo "Running migration to drop hourly_rate column..."

# Get postgres container name
CONTAINER_NAME=$(docker compose ps -q postgres)

if [ -z "$CONTAINER_NAME" ]; then
    echo "Error: Postgres container is not running. Please run 'docker-compose up -d' first."
    exit 1
fi

# Run migration
docker exec -i "$CONTAINER_NAME" psql -U parking_user -d parking_app < migrations/07_drop_old_hourly_rate_column.sql

if [ $? -eq 0 ]; then
    echo "✅ Migration completed successfully!"
    echo "The hourly_rate column has been dropped."
else
    echo "❌ Migration failed. Please check the error above."
    exit 1
fi

