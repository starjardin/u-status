.PHONY: dev build run migrate-up migrate-down test race

## Run the Go backend (requires .env)
dev:
	@mkdir -p data
	@export $$(grep -v '^#' .env | xargs) && go run ./cmd/server

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
