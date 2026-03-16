.PHONY: dev build run migrate-up migrate-down test race db-up db-down db-reset

DB_CONTAINER=ustatus-postgres
DB_USER=ustatus
DB_PASSWORD=ustatus
DB_NAME=ustatus
DB_PORT=5432

## Start Postgres container
db-up:
	@docker run -d --name $(DB_CONTAINER) \
		-e POSTGRES_USER=$(DB_USER) \
		-e POSTGRES_PASSWORD=$(DB_PASSWORD) \
		-e POSTGRES_DB=$(DB_NAME) \
		-p $(DB_PORT):5432 \
		postgres:16-alpine
	@echo "Waiting for Postgres to be ready..."
	@until docker exec $(DB_CONTAINER) pg_isready -U $(DB_USER) > /dev/null 2>&1; do sleep 1; done
	@echo "Postgres is ready on port $(DB_PORT)"

## Stop and remove Postgres container
db-down:
	@docker stop $(DB_CONTAINER) 2>/dev/null || true
	@docker rm $(DB_CONTAINER) 2>/dev/null || true
	@echo "Postgres container removed"

## Reset: remove container and start fresh
db-reset: db-down db-up

## Run the Go backend (requires local Postgres via make db-up)
dev:
	@DATABASE_URL="postgres://$(DB_USER):$(DB_PASSWORD)@localhost:$(DB_PORT)/$(DB_NAME)?sslmode=disable" \
	JWT_SECRET=change-me-in-dev \
	PORT=8080 \
	ALLOWED_ORIGINS=http://localhost:5173 \
	go run ./cmd/server

## Build the binary
build:
	@mkdir -p bin
	go build -o bin/server ./cmd/server

## Apply migrations (standalone, for local dev without running server)
migrate-up:
	@mkdir -p data
	@go run -tags migrator ./cmd/migrate up

## Roll back last migration
migrate-down:
	@mkdir -p data
	@go run -tags migrator ./cmd/migrate down

## Run tests
test:
	go test ./...

## Race detector
race:
	go test -race ./...
