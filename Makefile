-include .env
export

GOOSE = go run github.com/pressly/goose/v3/cmd/goose@v3.22.1

.PHONY: up down migrate migrate-down test test-integration run-api fmt

up:
	docker compose up -d

down:
	docker compose down

migrate:
	$(GOOSE) -dir migrations postgres "$(DATABASE_URL)" up

migrate-down:
	$(GOOSE) -dir migrations postgres "$(DATABASE_URL)" down

test:
	go test ./...

test-integration:
	TEST_DATABASE_URL="$(DATABASE_URL)" go test ./internal/store/... -v

run-api:
	go run ./cmd/api

fmt:
	gofmt -w .
