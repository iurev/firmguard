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

test:
	@go test -v ./internal/api/... ./internal/service/... ./internal/worker/... ./internal/repository/...

race:
	@go test -race -v ./internal/api/... ./internal/service/... ./internal/worker/... ./internal/repository/...

coverage:
	@go test -coverprofile=coverage.out ./internal/api/... ./internal/service/... ./internal/worker/... ./internal/repository/...
	@go tool cover -func=coverage.out

fmt:
	@go fmt ./...

clean:
	@rm -f main
