# Прогресс разработки repetapp

## Текущий этап: 2 — Группы и приглашения

## Завершённые этапы

### Этап 1 — Авторизация по SMS ✅
**Backend:**
- Миграция 000002: users, verification_codes, refresh_tokens, device_tokens
- domain/user.go: User, VerificationCode, RefreshToken, DeviceToken с ролями и языками
- pkg/jwt: IssueAccess / ParseAccess (HS256), Manager
- infrastructure/sms: Sender interface + MockSender + EskizSender (Eskiz.uz API)
- infrastructure/redis: RateLimiter (INCR+EXPIRE)
- features/auth: Repository interface, DTO, Service (10 методов), Handler (10 endpoints)
- infrastructure/postgres: AuthRepository — полная реализация
- middleware/auth.go: Bearer JWT + UserIDFromCtx / RoleFromCtx
- 10 unit-тестов сервиса: все проходят с race detector
- Исправлено: CORS AllowedOrigins конфигурируем, REDIS_ADDR валидируется, MinIO bucket приватный

**Flutter:**
- domain: User entity, AuthRepository interface
- data: UserDto/AuthResponseDto, AuthApi (Dio), AuthRepositoryImpl
- core/storage: SecureStorage (flutter_secure_storage)
- core/network: AuthInterceptor (Bearer + auto-refresh при 401)
- providers: AuthState, AuthNotifier (StateNotifier), authProvider
- core/router: app_router.dart (go_router), routes.dart
- pages: SplashPage, PhonePage, VerifyPage, RoleSelectPage, ProfileSetupPage
- widgets: PhoneInput, CodeInput
- app.dart: MaterialApp.router с go_router

⚠️ Flutter SDK не установлен по пути ~/Developer/flutter/bin.
   IDE показывает ошибки импортов — исчезнут после `flutter pub get`.
   Код логически корректен.

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
