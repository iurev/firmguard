ifneq (,$(wildcard ./.env))
    include .env
    export
endif

.PHONY: all build run docker-run docker-down itest watch test coverage fmt race clean migrate-up migrate-down

all: build

build:
	@go build -o main cmd/api/main.go

run:
	@go run cmd/api/main.go

docker-run:
	@docker compose up -d
	@echo "Waiting for database to be ready..."
	@sleep 2
	@$(MAKE) migrate-up

docker-down:
	@docker compose down

migrate-up:
	@goose -dir migrations postgres "postgres://$(BLUEPRINT_DB_USERNAME):$(BLUEPRINT_DB_PASSWORD)@$(BLUEPRINT_DB_HOST):$(BLUEPRINT_DB_PORT)/$(BLUEPRINT_DB_DATABASE)?sslmode=disable&search_path=$(BLUEPRINT_DB_SCHEMA)" up

migrate-down:
	@goose -dir migrations postgres "postgres://$(BLUEPRINT_DB_USERNAME):$(BLUEPRINT_DB_PASSWORD)@$(BLUEPRINT_DB_HOST):$(BLUEPRINT_DB_PORT)/$(BLUEPRINT_DB_DATABASE)?sslmode=disable&search_path=$(BLUEPRINT_DB_SCHEMA)" down

itest:
	@go test -v ./internal/database/...

watch:
	@if command -v air > /dev/null; then \
	    air; \
	else \
	    read -p "Go's 'air' is not installed on your machine. Do you want to install it? [y/N] " choice; \
	    if [ "$$choice" = "y" ] || [ "$$choice" = "Y" ]; then \
	        go install github.com/air-verse/air@latest; \
	        air; \
	    else \
	        echo "You can install it with 'go install github.com/air-verse/air@latest'"; \
	        exit 1; \
	    fi; \
	fi

test:
	@go test -v ./internal/api/... ./internal/service/...

race:
	@go test -race -v ./internal/api/... ./internal/service/...

coverage:
	@go test -coverprofile=coverage.out ./internal/api/... ./internal/service/...
	@go tool cover -func=coverage.out

fmt:
	@go fmt ./...

clean:
	@rm -f main
