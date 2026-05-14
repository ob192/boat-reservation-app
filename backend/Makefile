.PHONY: run build test tidy fmt vet migrate seed lint

# Default target shows what's available.
help:
	@echo "Targets:"
	@echo "  run      - go run ./cmd/api"
	@echo "  build    - build static binary into bin/api"
	@echo "  test     - go test ./..."
	@echo "  tidy     - go mod tidy"
	@echo "  fmt      - gofmt the tree"
	@echo "  vet      - go vet ./..."
	@echo "  migrate  - psql \$$DATABASE_URL -f migrations/0001_init.sql"
	@echo "  seed     - psql \$$DATABASE_URL -f migrations/0002_seed_slots.sql"

run:
	go run ./cmd/api

build:
	mkdir -p bin
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/api ./cmd/api

test:
	go test ./...

tidy:
	go mod tidy

fmt:
	gofmt -w .

vet:
	go vet ./...

migrate:
	@if [ -z "$$DATABASE_URL" ]; then echo "DATABASE_URL is not set" && exit 1; fi
	psql "$$DATABASE_URL" -f migrations/0001_init.sql

seed:
	@if [ -z "$$DATABASE_URL" ]; then echo "DATABASE_URL is not set" && exit 1; fi
	psql "$$DATABASE_URL" -f migrations/0002_seed_slots.sql
