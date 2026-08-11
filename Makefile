.PHONY: dev run build test vet up down tidy

# Sobe o MongoDB local e a API em foreground.
dev: up run

run:
	go run ./cmd/api

build:
	go build -o bin/gestorbuy-api ./cmd/api

test:
	go test ./...

vet:
	go vet ./...

up:
	docker compose up -d mongodb

down:
	docker compose down

tidy:
	go mod tidy
