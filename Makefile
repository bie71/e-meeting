# Simple Makefile for a Go project

SHELL := /bin/bash
include .env
export

DB_HOST=aws-0-ap-southeast-1.pooler.supabase.com
DB_PORT=5432
DB_NAME=postgres
DB_PARAMS=sslmode=require

# Build the application
all: build test

build:
	@echo "Building..."
	
	
	@go build -o main cmd/api/main.go

# Run the application
run:
	@go run cmd/api/main.go
# Create DB container
docker-run:
	@if docker compose up --build 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose up --build; \
	fi

# Shutdown DB container
docker-down:
	@if docker compose down 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose down; \
	fi

# Test the application
test:
	@echo "Testing..."
	@go test ./... -v
# Integrations Tests for the application
itest:
	@echo "Running integration tests..."
	@go test ./internal/database -v

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main

# Live Reload
watch:
	@if command -v air > /dev/null; then \
            air; \
            echo "Watching...";\
        else \
            read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
            if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
                go install github.com/air-verse/air@latest; \
                air; \
                echo "Watching...";\
            else \
                echo "You chose not to install air. Exiting..."; \
                exit 1; \
            fi; \
        fi

# Database migration commands
migrate-up:
	@echo "Running migrations up..."
	migrate -path migrations -database "postgres://$(DATABASE_USER):$(DATABASE_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?$(DB_PARAMS)" up

migrate-down:
	@echo "Running migrations down..."
	migrate -path migrations -database "postgres://$(DATABASE_USER):$(DATABASE_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?$(DB_PARAMS)" down

# Force a specific migration version
migrate-force:
	@read -p "Enter version to force (e.g., 1): " version; \
	migrate -path migrations -database "postgres://$(DATABASE_USER):$(DATABASE_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?$(DB_PARAMS)" force $$version

migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir migrations -seq $$name

# Show current migration version
migrate-version:
	@echo "Current migration version:"
	migrate -path migrations -database "postgres://$(DATABASE_USER):$(DATABASE_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?$(DB_PARAMS)" version

# Database seeder
db-seed:
	@echo "Running database seeder..."
	@if [ ! -f main ]; then \
		echo "Building application first..."; \
		go build -o main cmd/api/main.go; \
	fi
	@if [ -f .env ]; then \
		source .env && ./main --seed-only; \
	else \
		echo "Error: .env file not found"; \
		exit 1; \
	fi

.PHONY: all build run test clean watch docker-run docker-down itest migrate-up migrate-down migrate-force migrate-create migrate-version
