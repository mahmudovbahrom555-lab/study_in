# Прогресс разработки repetapp

## Текущий этап: 1 — Авторизация по SMS (в процессе)

## Завершённые этапы

### Этап 0 — Инфраструктура и каркас ✅
- Go: Chi, config (Viper), middleware (logging, recovery, CORS), health/version endpoints, graceful shutdown
- Flutter: Riverpod, go_router, Dio, тема (light/dark), локализация (uz/uz-cyrl/ru/en)
- Docker Compose: PostgreSQL 16 + Redis 7 + MinIO
- GitHub Actions CI: lint → test → build для Go и Flutter

## Текущая итерация (Этап 1)

### Выполнено в этой итерации
- [ ] Исправлены известные проблемы (MinIO анонимный доступ, CORS, Redis валидация)
- [ ] Миграция 000002_auth (users, verification_codes, refresh_tokens, device_tokens)
- [ ] domain/user.go — сущности User, VerificationCode, RefreshToken
- [ ] pkg/jwt/jwt.go — генерация/валидация JWT
- [ ] infrastructure/sms — интерфейс + mock + Eskiz
- [ ] infrastructure/redis/rate_limit.go
- [ ] features/auth — repository interface, dto, service, handler
- [ ] infrastructure/postgres/auth_repository.go
- [ ] middleware/auth.go — JWT middleware
- [ ] Обновлён server.go — регистрация auth routes
- [ ] Тесты service_test.go

## Следующие этапы
- Этап 2: Группы и приглашения
- Этап 3: Лента и объявления
- Этап 4: Домашние задания
- Этап 5: Тесты
- Этап 6: Журнал оценок
- Этап 7: Посещаемость
- Этап 8: Родительский модуль
- Этап 9: Полировка и QA
- Этап 10: Релиз
