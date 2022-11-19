.PHONY: run
run:
	go run cmd/runtime-metrics/main.go

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: up
up:
	docker compose up -d --remove-orphans

.PHONY: down
down:
	docker compose down --remove-orphans

.PHONY: build
build:
	docker compose build