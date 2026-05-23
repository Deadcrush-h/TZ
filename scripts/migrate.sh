#!/bin/bash

# Скрипт для запуска миграций
DB_URL="postgres://postgres:postgres@localhost:5432/org_structure?sslmode=disable"

case "$1" in
    up)
        echo "Running migrations up..."
        goose -dir migrations postgres "$DB_URL" up
        ;;
    down)
        echo "Rolling back migration..."
        goose -dir migrations postgres "$DB_URL" down
        ;;
    status)
        echo "Migration status..."
        goose -dir migrations postgres "$DB_URL" status
        ;;
    create)
        if [ -z "$2" ]; then
            echo "Error: migration name is required"
            echo "Usage: ./migrate.sh create <migration_name>"
            exit 1
        fi
        echo "Creating migration: $2"
        goose -dir migrations create "$2" sql
        ;;
    *)
        echo "Usage: ./migrate.sh {up|down|status|create}"
        exit 1
        ;;
esac