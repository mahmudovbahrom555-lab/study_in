# Backend (Go)

## Структура

```
backend/
├── cmd/api/              # Точка входа
├── internal/
│   ├── config/           # Конфигурация
│   ├── domain/           # Бизнес-сущности (ядро, не зависит ни от чего)
│   ├── features/         # Use cases по фичам
│   ├── infrastructure/   # PostgreSQL, Redis, S3, FCM, SMS
│   ├── middleware/       # HTTP middleware
│   ├── pkg/              # Общие утилиты
│   └── server/           # HTTP сервер и роутинг
├── migrations/           # SQL миграции
├── tests/                # Интеграционные тесты
└── scripts/              # Утилиты
```

## Быстрый старт

```bash
# 1. Скопировать пример конфигурации
cp .env.example .env

# 2. Поднять зависимости
make docker-up

# 3. Установить goose или migrate (один раз)
brew install golang-migrate

# 4. Применить миграции
make migrate-up

# 5. Запустить сервер
make dev
```

Проверка работы:

```bash
curl http://localhost:8080/api/v1/health
```

## Команды

```bash
make help              # Показать все команды
make dev               # Запуск локально
make test              # Все тесты
make lint              # Линтер
make migrate-up        # Применить миграции
make migrate-create name=add_users   # Создать миграцию
make build             # Собрать бинарь
```

## Принципы

- **Clean Architecture**: domain не зависит от infrastructure
- **Feature-First**: код группируется по фичам, не по типам
- **Зависимости через интерфейсы**: сервисы зависят от интерфейсов репозиториев
- **UUID везде**: никаких автоинкрементов в публичных API
- **Stateless API**: JWT, без сессий в памяти
- **Soft delete**: через `deleted_at`

## Добавление новой фичи

1. Создать папку `internal/features/<name>/`
2. Создать файлы:
   - `service.go` — бизнес-логика
   - `handler.go` — HTTP обработчики
   - `repository.go` — интерфейс репозитория
   - `dto.go` — Request/Response структуры
3. Реализовать репозиторий в `internal/infrastructure/postgres/<name>_repository.go`
4. Зарегистрировать роуты в `internal/server/server.go`
5. Написать тесты

## Тестирование

```bash
make test-unit         # Юнит-тесты (моки репозиториев)
make test-integration  # С реальной БД через testcontainers
```
