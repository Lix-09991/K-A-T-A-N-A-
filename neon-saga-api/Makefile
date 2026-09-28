ifneq (,$(wildcard .env))
    include .env
    export
endif

.PHONY: dev build run db-up db-down migrate-up migrate-down migrate-create test

dev:
	air

build:
	go build -o bin/api ./cmd/api

run: build
	./bin/api

db-up:
	docker compose up -d

db-down:
	docker compose down

migrate-up:
	migrate -path ./migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path ./migrations -database "$(DATABASE_URL)" down 1

migrate-create:
	migrate create -ext sql -dir ./migrations -seq $(name)

test:
	go test ./... -v
