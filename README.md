# Study In

Учебная платформа для репетиторов: домашние задания, тесты, оценки, посещаемость, родительский модуль.

## Стек

- **Backend:** Go 1.22+ (Chi, sqlx, PostgreSQL 16, Redis 7, MinIO, Asynq)
- **Frontend:** Flutter 3.24+ (Riverpod, go_router, dio, Hive)
- **Инфраструктура:** Docker + Caddy, Hetzner Cloud, GitHub Actions

## Клонирование

```bash
git clone https://github.com/mahmudovbahrom555-lab/study_in.git
cd study_in
```

## Структура репозитория

```
study_in/
├── backend/          # Go API
├── flutter_app/      # Flutter мобильное приложение
├── docs/             # Документация проекта
├── .github/          # GitHub Actions CI/CD
└── docker-compose.yml
```

## Быстрый старт

### Требования

- Docker и Docker Compose
- Go 1.22+ (для разработки бэкенда)
- Flutter 3.24+ (для разработки мобилки)
- Make

### Запуск инфраструктуры

```bash
# Поднять PostgreSQL + Redis + MinIO
docker-compose up -d

# Дождаться запуска и применить миграции
cd backend
cp .env.example .env
make deps
make migrate-up

# Запустить бэкенд
make dev
```

### Запуск Flutter

```bash
cd flutter_app
flutter pub get
flutter run --dart-define=API_URL=http://localhost:8080/api/v1
```

### Проверка работы

```bash
curl http://localhost:8080/api/v1/health
```

## Документация

- [Архитектура](docs/architecture.md)
- [Соглашения по коду](docs/conventions.md)
- [Backend](backend/README.md)
- [Frontend](flutter_app/README.md)

## Этапы разработки

- [x] **Этап 0:** Инфраструктура и каркас
- [ ] Этап 1: Авторизация по SMS
- [ ] Этап 2: Группы
- [ ] Этап 3: Лента и объявления
- [ ] Этап 4: Домашние задания
- [ ] Этап 5: Тесты
- [ ] Этап 6: Журнал оценок
- [ ] Этап 7: Посещаемость
- [ ] Этап 8: Родительский модуль
- [ ] Этап 9: Полировка и QA
- [ ] Этап 10: Релиз
