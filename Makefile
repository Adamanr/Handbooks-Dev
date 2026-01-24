# Makefile

up:
	podman compose up -d

down:
	podman compose down -v

build:
	podman compose build

logs:
	podman compose logs -f app

dev:
	@if [ -f .env ]; then export $$(grep -v '^#' .env | xargs); fi
	go run cmd/main.go

migrate_up:
	goose up

migrate_down:
	goose down

.PHONY: up down build logs dev
