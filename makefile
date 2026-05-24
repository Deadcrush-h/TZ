.PHONY: help build run clean migrate-up migrate-down migrate-create migrate-status docker-build docker-up docker-down

# Variables
BINARY_NAME=org-structure-api
MIGRATION_DIR=migrations

# Colors
GREEN := \033[0;32m
RED := \033[0;31m
RESET := \033[0m

help:
	@echo "${GREEN}Available commands:${RESET}"
	@echo "  make build         - Build the application"
	@echo "  make run           - Run the application"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make migrate-up    - Run database migrations"
	@echo "  make migrate-down  - Rollback last migration"
	@echo "  make migrate-create NAME=migration_name - Create new migration"
	@echo "  make migrate-status- Check migration status"
	@echo "  make docker-build  - Build Docker image"
	@echo "  make docker-up     - Start Docker containers"
	@echo "  make docker-down   - Stop Docker containers"
	@echo "  make docker-logs   - View Docker logs"

build:
	@echo "${GREEN}Building application...${RESET}"
	go build -o $(BINARY_NAME) ./cmd/main.go
	@echo "${GREEN}Build complete!${RESET}"

run:
	@echo "${GREEN}Running application...${RESET}"
	go run ./cmd/main.go

clean:
	@echo "${GREEN}Cleaning...${RESET}"
	rm -f $(BINARY_NAME)
	go clean
	@echo "${GREEN}Clean complete!${RESET}"

migrate-up:
	@echo "${GREEN}Running migrations up...${RESET}"
	@goose -dir $(MIGRATION_DIR) postgres "host=localhost user=postgres password=postgres dbname=org_structure port=5432 sslmode=disable" up
	@echo "${GREEN}Migrations completed!${RESET}"

migrate-down:
	@echo "${RED}Rolling back migration...${RESET}"
	@goose -dir $(MIGRATION_DIR) postgres "host=localhost user=postgres password=postgres dbname=org_structure port=5432 sslmode=disable" down
	@echo "${GREEN}Rollback completed!${RESET}"

migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "${RED}Error: NAME is required. Usage: make migrate-create NAME=migration_name${RESET}"; \
		exit 1; \
	fi
	@echo "${GREEN}Creating migration: $(NAME)${RESET}"
	@goose -dir $(MIGRATION_DIR) create $(NAME) sql
	@echo "${GREEN}Migration created!${RESET}"

migrate-status:
	@echo "${GREEN}Migration status:${RESET}"
	@goose -dir $(MIGRATION_DIR) postgres "host=localhost user=postgres password=postgres dbname=org_structure port=5432 sslmode=disable" status

docker-build:
	@echo "${GREEN}Building Docker image...${RESET}"
	docker-compose build
	@echo "${GREEN}Docker image built!${RESET}"

docker-up:
	@echo "${GREEN}Starting Docker containers...${RESET}"
	docker-compose up -d
	@echo "${GREEN}Waiting for services to be ready...${RESET}"
	sleep 5
	@echo "${GREEN}Docker containers started!${RESET}"

docker-down:
	@echo "${RED}Stopping Docker containers...${RESET}"
	docker-compose down
	@echo "${GREEN}Docker containers stopped!${RESET}"

docker-logs:
	docker-compose logs -f

docker-clean:
	@echo "${RED}Cleaning Docker containers and volumes...${RESET}"
	docker-compose down -v
	@echo "${GREEN}Docker cleaned!${RESET}"

dev: docker-up migrate-up
	@echo "${GREEN}Development environment ready!${RESET}"
