.PHONY: help dev build test test-unit test-integration lint migrate-up migrate-down migrate-create docker-up docker-down clean

# Загрузить .env если есть
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

DATABASE_URL ?= postgres://repetapp:repetapp_dev_password@localhost:5432/repetapp?sslmode=disable
MIGRATE_BIN ?= migrate

help: ## Показать справку
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-25s\033[0m %s\n", $$1, $$2}'

dev: ## Запустить сервер локально
	go run ./cmd/api

build: ## Собрать бинарь в bin/api
	mkdir -p bin
	go build -o bin/api ./cmd/api

test: ## Все тесты
	go test ./... -race -cover

test-unit: ## Только юнит-тесты
	go test ./internal/features/... -race -cover

test-integration: ## Интеграционные тесты
	go test ./tests/integration/... -race

lint: ## Запустить линтер
	@which golangci-lint > /dev/null || (echo "Install golangci-lint: https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run ./...

migrate-up: ## Применить все миграции
	@which $(MIGRATE_BIN) > /dev/null || (echo "Install migrate: brew install golang-migrate" && exit 1)
	$(MIGRATE_BIN) -path migrations -database "$(DATABASE_URL)" up

migrate-down: ## Откатить последнюю миграцию
	$(MIGRATE_BIN) -path migrations -database "$(DATABASE_URL)" down 1

migrate-create: ## Создать миграцию: make migrate-create name=add_users
	@if [ -z "$(name)" ]; then echo "Usage: make migrate-create name=<migration_name>"; exit 1; fi
	$(MIGRATE_BIN) create -ext sql -dir migrations -seq $(name)

migrate-status: ## Текущая версия миграции
	$(MIGRATE_BIN) -path migrations -database "$(DATABASE_URL)" version

docker-up: ## Поднять зависимости в Docker
	docker-compose -f ../docker-compose.yml up -d

docker-down: ## Остановить Docker
	docker-compose -f ../docker-compose.yml down

deps: ## Установить зависимости
	go mod download
	go mod tidy

clean: ## Очистить артефакты сборки
	rm -rf bin/ tmp/ coverage.out coverage.html
