# Соглашения по коду

## Git

### Ветки

- `main` — продакшен. Защищена, только PR из dev.
- `dev` — staging. Авто-деплой на staging-сервер.
- `feature/<name>` — разработка новой фичи.
- `hotfix/<name>` — срочные правки прода.
- `fix/<name>` — исправление багов.

### Коммиты

Conventional Commits:

```
feat: добавлен экран регистрации
fix: исправлено падение при пустом списке групп
refactor: вынесена логика подсчёта баллов в отдельный сервис
docs: обновлён README по запуску
test: добавлены тесты для AuthService
chore: обновлены зависимости
```

### Pull Requests

- Название по Conventional Commits
- Описание: что меняется и зачем
- Ссылка на задачу (если есть)
- Скриншоты для UI изменений
- Минимум 1 ревью перед merge

## Go

### Naming

- **Файлы:** `snake_case.go`
- **Типы:** `PascalCase`
- **Функции экспортируемые:** `PascalCase`
- **Функции приватные:** `camelCase`
- **Константы:** `PascalCase` или `UPPER_SNAKE` для групп
- **Интерфейсы:** часто заканчиваются на `-er` (`Repository`, `Sender`)

### Структура файла

```go
package mypackage

import (
    // stdlib
    "context"
    "errors"

    // external
    "github.com/google/uuid"

    // internal
    "github.com/repetapp/backend/internal/domain"
)

// Константы
const defaultLimit = 20

// Типы
type Service struct {
    repo Repository
}

// Конструкторы
func NewService(repo Repository) *Service {
    return &Service{repo: repo}
}

// Методы
func (s *Service) DoSomething(ctx context.Context) error {
    // ...
}
```

### Обработка ошибок

Всегда оборачивайте ошибки с контекстом:

```go
// ✗ Плохо
if err != nil {
    return err
}

// ✓ Хорошо
if err != nil {
    return fmt.Errorf("create user: %w", err)
}
```

Используйте доменные ошибки для бизнес-логики:

```go
if user == nil {
    return domain.ErrNotFound
}
```

### Тесты

Файлы тестов рядом с кодом: `service.go` → `service_test.go`.

```go
func TestService_DoSomething(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:  "успех",
            input: "valid",
            want:  "result",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // ...
        })
    }
}
```

## Dart / Flutter

### Naming

- **Файлы:** `snake_case.dart`
- **Классы:** `PascalCase`
- **Переменные/функции:** `camelCase`
- **Константы:** `lowerCamelCase` (Dart style)
- **Приватные:** с `_` префиксом

### Структура виджета

```dart
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

class MyPage extends ConsumerWidget {
  const MyPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return Scaffold(
      // ...
    );
  }
}
```

### Константы

```dart
// ✗ Плохо
const padding = EdgeInsets.all(16);

// ✓ Хорошо
static const _padding = EdgeInsets.all(16);

// ✓ Лучше — в shared/dimensions.dart
class AppSpacing {
  static const md = 16.0;
}
```

### Стейт

- **Локальное состояние виджета** — `StatefulWidget` или `useState`
- **Состояние фичи** — Riverpod провайдеры
- **Глобальное состояние** — Riverpod провайдеры верхнего уровня

## SQL

### Naming

- **Таблицы:** `snake_case`, множественное число (`users`, `assignments`)
- **Колонки:** `snake_case` (`created_at`, `phone_number`)
- **Индексы:** `idx_<table>_<columns>` (`idx_users_phone`)
- **Foreign keys:** `<table>_<column>_fkey` (создаётся автоматически)

### Миграции

- Каждая миграция в отдельной паре файлов: `NNNNNN_name.up.sql` + `NNNNNN_name.down.sql`
- Down миграции обязательны для отката
- Не редактировать применённые миграции — создавать новые

```sql
-- 000002_add_users.up.sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone VARCHAR(20) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_users_phone ON users(phone);
```

```sql
-- 000002_add_users.down.sql
DROP TABLE IF EXISTS users;
```

## API

### URL

- Версия в URL: `/api/v1/...`
- Множественное число для ресурсов: `/groups`, `/assignments`
- Глаголы только для действий: `/auth/verify`, `/groups/:id/archive`

### Методы

- `GET` — чтение
- `POST` — создание или действие
- `PATCH` — частичное обновление
- `PUT` — полная замена (редко)
- `DELETE` — удаление (soft delete на бэке)

### Формат ответа

Всегда обёртка `data` или `error`:

```json
// Успех
{ "data": { ... } }

// Список с пагинацией
{ "data": [...], "meta": { "page": 1, "total": 100 } }

// Ошибка
{ "error": { "code": "...", "message": "...", "details": {...} } }
```

### Коды ответов

- `200 OK` — успешный GET, PATCH
- `201 Created` — успешный POST с созданием
- `204 No Content` — успешный DELETE
- `400 Bad Request` — невалидные данные
- `401 Unauthorized` — нет/невалидный JWT
- `403 Forbidden` — JWT валидный, но нет прав
- `404 Not Found` — ресурс не найден
- `409 Conflict` — конфликт (дубликат)
- `429 Too Many Requests` — rate limit
- `500 Internal Server Error` — баг сервера

## Code Review

### Чек-лист

- [ ] Тесты добавлены/обновлены
- [ ] Линтер проходит
- [ ] Названия понятные
- [ ] Нет TODO без задачи
- [ ] Нет закомментированного кода
- [ ] Логи структурированные
- [ ] Ошибки обёрнуты
- [ ] Документация обновлена при необходимости
