# RepetApp — учебная платформа для репетиторов

Go backend + Flutter мобильное приложение: домашние задания, тесты, оценки, посещаемость, родительский модуль.

## Стек

| Слой | Технологии |
|------|-----------|
| Backend | Go 1.22, Chi v5, sqlx + pgx/v5, golang-migrate |
| Хранилища | PostgreSQL 16, Redis 7, MinIO (private bucket, signed URLs) |
| Async | **Asynq** (SMS + push через очередь, retry ×3) |
| Push | Firebase Cloud Messaging (FCM) |
| Мониторинг | Prometheus `/metrics`, Grafana, Loki, Sentry |
| Mobile | Flutter 3.24+, Riverpod (StateNotifier), go_router, Dio |
| Инфраструктура | Docker Compose, GitHub Actions CI |

## Быстрый старт (dev)

```bash
git clone https://github.com/mahmudovbahrom555-lab/study_in.git
cd study_in

# Поднять зависимости
docker compose up -d

# Применить миграции
cd backend && make migrate-up

# Запустить сервер
make dev
```

## Переменные окружения

Скопируй и заполни:
```bash
cp backend/.env.example backend/.env
```

Обязательные:
- `DATABASE_URL` — postgres connection string
- `REDIS_ADDR` — e.g. `localhost:6379`
- `JWT_ACCESS_SECRET` — минимум 32 символа
- `JWT_REFRESH_SECRET` — минимум 32 символа

## Production deploy

```bash
# Полный стек (API + Postgres + Redis + MinIO)
docker compose -f docker-compose.prod.yml up -d

# + Мониторинг (Prometheus + Grafana + Loki)
docker compose -f docker-compose.prod.yml -f docker-compose.monitoring.yml up -d
```

Grafana доступна на `:3000`, Prometheus на `:9090`.

## API — эндпоинты (62 шт.)

| Группа | Эндпоинты |
|--------|-----------|
| Auth | `/auth/send-code`, `/auth/verify`, `/auth/refresh`, `/auth/logout`, `/auth/me`, `/auth/profile` + device |
| Groups | CRUD + archive + join + members |
| Feed | CRUD posts + attachments |
| Assignments | CRUD + file upload + submit + grade submissions |
| Quizzes | CRUD quiz/questions/options + attempt + submit (авто-оценка) |
| Grades | CRUD + list by student |
| Attendance | Upsert + list by group/student |
| Parents | link/unlink children + read child grades/attendance/groups |
| Notifications | list + mark read + device registration |
| System | `/health`, `/version`, `/metrics` |

## Архитектура backend

```
cmd/api/main.go          ← точка входа, Sentry, Asynq worker, HTTP server
internal/
  config/                ← Viper, валидация при старте
  domain/                ← чистые сущности (User, Group, Grade, …)
  features/
    auth/                ← SMS OTP, JWT access+refresh
    groups/              ← invite code (crypto/rand), payment_status
    feed/                ← посты с вложениями
    assignments/         ← ДЗ + сабмиты + оценка
    quizzes/             ← тесты, авто-оценка, MaxAttempts
    grades/              ← журнал оценок
    attendance/          ← upsert, ENUM статусы
    parents/             ← parent_links, drill-in к данным ребёнка
    notifications/       ← in-app + FCM push через Asynq
  infrastructure/
    postgres/            ← sqlx репозитории
    redis/               ← rate limiter
    minio/               ← S3-совместимое хранилище
    queue/               ← Asynq client + worker (SMS, push)
    fcm/                 ← Firebase Cloud Messaging pusher
    monitoring/          ← Sentry init + SentryRecovery middleware
    sms/                 ← Eskiz.uz + MockSender
  middleware/            ← Auth (JWT), Recovery, Logging, Metrics (Prometheus)
  server/                ← Chi router, wiring всех зависимостей
  pkg/
    jwt/                 ← access + refresh токены
    response/            ← стандартные JSON-ответы
    logger/              ← slog JSON/text
migrations/              ← 000001…000010, up + down
```

## Тесты

```bash
cd backend && go test -race ./...
# 9 пакетов: auth, groups, feed, assignments, quizzes, grades, attendance, parents, notifications
```

## Структура Flutter

```
flutter_app/lib/
  core/
    network/    ← Dio + AuthInterceptor (auto-refresh JWT)
    router/     ← go_router, role-aware redirect
    theme/      ← Material 3
  features/
    auth/       ← SMS OTP flow
    groups/     ← список + detail + QuickAction chips
    feed/
    assignments/ ← список ДЗ, сдача, оценка
    quizzes/    ← прохождение теста, результат
    grades/     ← журнал, SummaryBar
    attendance/ ← teacher + student view, DatePicker
    parents/    ← список детей, drill-in к данным
    home/       ← HomeShell с role-aware NavigationBar
```
