# Morning QA Report — 2026-06-14

Автономный QA-прогон выполнен пока ты спал.

---

## Итог одной строкой

**2 коммита, 0 оставшихся ошибок, все Go-тесты зелёные.**

---

## Что было сделано

### Коммит 1 — `9f882df`
`fix(reports): 6 code-review bugs — auth, transaction, analytics, cache, sort`

Все баги из code-review исправлены в backend:

| # | Баг | Статус |
|---|-----|--------|
| P0 | Risk Alert IDs = `00000000...` | ✅ FIXED — `RETURNING id` + scan в цикл |
| P0 | DELETE + INSERT без транзакции | ✅ FIXED — `BeginTxx`/`Commit`/`Rollback` |
| P1 | MasteryDelta всегда = 100% | ✅ FIXED — заменил `topic_mastery` на `skill_logs` |
| P1 | Кеш ломался через `created_at` | ✅ FIXED — кешируем по `PeriodStart.Equal()` |
| P2 | `in_progress` дубликаты при refresh | ✅ FIXED — DELETE `IN ('new','in_progress')` |
| P3 | ORDER BY строковый: `h < l < m` | ✅ FIXED — CASE expression в обоих запросах |

### Коммит 2 — `fc52350`
`fix(flutter): resolve all analyze errors found by QA pass`

`flutter analyze` выявил ошибки — все исправлены:

| Файл | Ошибка | Фикс |
|------|--------|------|
| `app_theme.dart:25` | `CardTheme` → требует `CardThemeData` | Тип переименован (Flutter 3.32 breaking change) |
| `parents_api.dart:3-5` | `../../../` указывал в `lib/` вместо `lib/features/` | `../../` (2 уровня вверх, не 3) |
| `parents_repository_impl.dart:1-3` | То же самое | Аналогично |
| `quizzes_page.dart:6` | Unused import `quiz.dart` | Удалён |
| `quizzes_provider.dart:113` | Unused field warning на `_groupId` | `// ignore:` — поле нужно, приходит от провайдера |

---

## Результаты тестирования

### Go backend tests
```
ok  internal/features/ai          1.539s
ok  internal/features/assignments (cached)
ok  internal/features/attendance  (cached)
ok  internal/features/auth        2.087s
ok  internal/features/feed        (cached)
ok  internal/features/grades      (cached)
ok  internal/features/groups      (cached)
ok  internal/features/notifications (cached)
ok  internal/features/parents     (cached)
ok  internal/features/quizzes     2.048s
ok  internal/infrastructure/spaced_repetition (cached)
```
**Все 11 пакетов с тестами — зелёные. Пакет `reports` тестов не имеет (см. ниже).**

### Flutter analyze
```
Before: 64 issues (10 errors, 1 warning, 53 info)
After:  41 issues (0 errors, 0 warnings, 41 info)
```
Оставшиеся 41 — только `info` стиля (`require_trailing_commas`, `deprecated_member_use`). Не блокируют сборку.

### E2E через curl
❌ **Не выполнено** — Docker не установлен в среде выполнения агента.
Тест нужно запустить вручную после `docker compose up -d`.

---

## Что НЕ покрыто тестами (технический долг)

| Пакет | Что нужно протестировать |
|-------|------------------------|
| `reports` | `GenerateParentReport` cache logic, `SaveRiskAlerts` транзакция, role auth в handler |
| `postgres/report_repository` | `ComputeMetrics` с реальной БД (mastery delta, group_topics CTE) |
| Flutter widget tests | `ParentRoiPage`, `OwnerRiskPage`, `QuizResultPage` (gamification banner) |

---

## Оставшиеся style warnings (не блокируют)

- `withOpacity()` deprecated → нужно заменить на `.withValues(alpha: ...)` в ~8 местах
- `surfaceVariant` deprecated → `surfaceContainerHighest` в 2 местах
- `require_trailing_commas` — ~30 мест

Всё это можно сделать одной командой `dart fix --apply` когда будет время.

---

## Следующие шаги (предлагаю)

1. `docker compose up -d` → запустить E2E curl тест вручную
2. Написать тесты для `reports` пакета (service + mock repository)
3. `dart fix --apply` для автоматического исправления deprecated API

---

*Отчёт сгенерирован автономным QA-агентом 2026-06-14*
