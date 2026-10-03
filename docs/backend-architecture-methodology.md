# Методология и целевая архитектура бэкенда

> **Статус:** v0.1 (2026-10-02). Дополняет `docs/architecture.md` (текущее
> устройство) и `docs/conventions.md` (стиль кода). Этот документ отвечает на
> вопрос: **как развивать бэкенд, чтобы новые модули добавлялись без роста
> стоимости изменений.**
>
> Решения, принятые по этому документу, фиксируются в `DECISIONS.md`.

---

## 1. Научная основа: что делает архитектуру расширяемой

Расширяемость — не свойство технологии (микросервисы, Kafka), а свойство
**границ между частями системы**. Опираемся на результаты, которые подтверждены
исследованиями и многолетней практикой:

| Принцип | Источник | Что из этого следует для нас |
|---|---|---|
| **Скрывать решения, которые могут измениться**, а не делить систему по шагам обработки | Parnas (1972) | Модуль = одна область знаний (оплаты, тесты), а не слой (все репозитории в одной папке) |
| **Структура системы повторяет структуру коммуникаций команды** | Conway (1968); эмпирически — MacCormack, Rusnak & Baldwin (2012) | Границы модулей = будущие границы ответственности людей/команд |
| **Слабая связанность предсказывает скорость поставки** | Forsgren, Humble & Kim (2018), исследование DORA | Цель: модуль можно изменить, протестировать и выкатить, не трогая другие |
| **Сложность растёт, если с ней не бороться целенаправленно** | Законы эволюции ПО, Lehman (1980) | Нужны автоматические проверки архитектуры («fitness functions»), а не только договорённости |
| **Ограниченные контексты** (bounded contexts) с явными контрактами | Evans (2003) | Одно слово («группа», «оценка») значит одно внутри модуля; между модулями — явный API |
| **Зависимости направлены к стабильному ядру** | Martin (2017); Cockburn (2005), hexagonal | Бизнес-логика не знает о Postgres, HTTP, OpenAI |
| **Сначала монолит, делить — когда границы проверены жизнью** | Fowler (2015), «MonolithFirst» | Целевая форма — **модульный монолит**, не микросервисы |
| **Архитектура проверяется автоматически** | Ford, Parsons & Kua (2017), fitness functions | Правила зависимостей — в CI, а не в ревью |
| **Горячие точки = частота изменений × сложность** | Tornhill (2015) | Рефакторим там, где часто меняем, а не там, где «некрасиво» |

---

## 2. Диагностика текущего состояния (2026-10-02)

Текущая архитектура (Clean Architecture + feature-папки) — хороший старт.
Проблемы ниже найдены в коде и **уже проявились как баги** в этой же сессии.

| # | Проблема | Где | Проявление |
|---|---|---|---|
| P1 | **`SELECT *` + строгий маппинг sqlx** — 48 запросов. Новая колонка в таблице без поля в структуре ломает все запросы к таблице | `infrastructure/postgres/*` | Миграция 000016 добавила `groups.is_demo` и `cefr_level` → **500 на всех экранах групп** |
| P2 | **Ошибки проглатываются**: 7 копий `writeError`; ошибка 500 не логируется вместе с причиной | `features/*/handler.go` | Причину 500 пришлось искать вручную в коде |
| P3 | **Роль зашита в JWT без перевыпуска** при изменении | `features/auth`, `middleware/auth.go` | После выбора роли — 403 до перевхода (исправлено на клиенте) |
| P4 | **Ручная сборка всех зависимостей в одном файле** (441 строка); каждая фича правит `server.go` | `server/server.go` | Конфликты при параллельной разработке; бизнес-логика попадает в адаптеры |
| P5 | **Бизнес-логика в слое сборки**: `aiQuizCreatorAdapter` создаёт тест/вопросы/варианты; ошибки привязки тем игнорируются (`_ =`) | `server/server.go` | Логику невозможно протестировать в модуле |
| P6 | **Нет транзакций** для многошаговых операций (транзакции используются в 1 файле репозиториев из 15) | AI-генерация теста, сдача теста и т.д. | Сбой посередине оставляет «полутест» в базе |
| P7 | **«Божественный» сервис AI**: `ai/service.go` — 895 строк, конструктор на 15 зависимостей | `features/ai` | Любое изменение AI задевает несвязанные части |
| P8 | **Межмодульные вызовы через ручные адаптеры и `SetObserver`** (обход циклической инициализации) | quizzes ↔ ai, parents → groups | Скрытые связи, порядок инициализации важен |
| P9 | **Все репозитории в одном пакете** `infrastructure/postgres` | — | Модуль нельзя удалить или вынести целиком |
| P10 | **Нет интеграционных тестов против реальной БД** (`tests/integration` упомянут в Makefile, но отсутствует) | — | P1 не был пойман тестами |
| P11 | **Дрейф окружений**: Dockerfile собирал Go 1.22 при `go 1.25` в `go.mod`; образы MinIO удалены из Docker Hub | `backend/Dockerfile`, compose | Сборка не проходила |

**Горячие точки для рефакторинга** (оценить по `git log` — частота × размер):
`server/server.go`, `features/ai/service.go`, `features/ai/handler.go`,
`infrastructure/postgres/insights_repository.go`.

---

## 3. Целевая архитектура: модульный монолит

### 3.1. Модули (bounded contexts)

Границы выводятся из карты работы репетитора (`research/01-jtbd-tutors.md`):
каждая работа меняется по своим причинам, значит — свой модуль.

```
┌──────────────────────────────────────────────────────────────────────┐
│                         HTTP API (chi) / Worker (asynq)              │
├──────────┬───────────┬────────────┬────────────┬──────────┬──────────┤
│ identity │ classroom │  learning  │ assessment │ billing  │ family   │
│ auth,    │ groups,   │ feed,      │ grades,    │ тарифы,  │ parents, │
│ роли,    │ members,  │ assign-    │ attendance │ периоды, │ отчёты   │
│ орг-ции  │ schedule* │ ments,     │            │ платежи* │ родителю │
│          │           │ quizzes    │            │          │          │
├──────────┴───────────┴────────────┴────────────┴──────────┴──────────┤
│ intelligence (AI: documents, generation, mastery/SM-2, rules, coach) │
├──────────────────────────────────────────────────────────────────────┤
│ notifications   │   reporting (read-models: Parent ROI, Owner Risk)  │
├──────────────────────────────────────────────────────────────────────┤
│ platform: db/tx, events (outbox), errors, authz, config, observability│
└──────────────────────────────────────────────────────────────────────┘
 * — модули, которых пока нет; появятся по итогам JTBD (оплаты, расписание)
```

| Модуль | Владеет таблицами | Публикует события (примеры) |
|---|---|---|
| identity | users, refresh_tokens, verification_codes, organizations* | `UserRoleAssigned` |
| classroom | groups, group_members, lessons* | `StudentJoinedGroup`, `LessonScheduled`* |
| learning | posts, assignments, submissions, quizzes, questions, options, quiz_attempts, student_answers | `QuizAnswerSubmitted`, `SubmissionGraded` |
| assessment | grades, attendance | `GradeRecorded`, `AttendanceMarked` |
| billing* | tariffs, invoices, payments | `PaymentOverdue`, `PaymentReceived` |
| family | parent_links, parent_reports | `ParentLinked` |
| intelligence | ai_*, topic_*, question_topics, quiz_question_feedback, recommendation_outcomes, student_gamification, xp_events, skill_assessments, skill_logs, consent_events | `RecommendationCreated` |
| notifications | notifications, device_tokens | — (только потребитель) |
| reporting | owner_risk_alerts + read-only представления | — |

### 3.2. Устройство модуля

Каждый модуль — **вертикальный срез** со всем, что ему нужно. Удаление папки
модуля удаляет фичу целиком.

```
internal/modules/learning/
├── module.go          ← ЕДИНСТВЕННАЯ точка сборки модуля: New(deps) + Routes + Subscriptions
├── api.go             ← публичный контракт для других модулей (интерфейсы + DTO)
├── events.go          ← события, которые модуль публикует
├── domain/            ← сущности и правила, без зависимостей от инфраструктуры
├── app/               ← сервисы (use cases)
├── store/             ← репозитории (Postgres), только таблицы этого модуля
├── http/              ← handlers + DTO запросов/ответов
└── store_test.go      ← интеграционные тесты против реальной БД
```

### 3.3. Правила зависимостей (проверяются в CI)

1. Модуль импортирует из другого модуля **только его `api.go`/`events.go`**,
   никогда `domain/`, `app/`, `store/`.
2. `domain/` не импортирует ничего, кроме stdlib и `platform/errors`.
3. Модуль читает и пишет **только свои таблицы**. Исключение — `reporting`:
   ему разрешены read-only SQL-представления поверх чужих таблиц, и только ему.
4. Циклические зависимости между модулями запрещены. Если A нужен B, а B — A,
   одна из связей становится событием.

**Проверка:** `depguard` в `.golangci.yml` (линтер уже подключён к CI) +
архитектурный тест на `go/packages`, который падает при нарушении правил 1–2.

### 3.4. Сборка: модули регистрируют себя сами

Вместо 441 строки ручной сборки в `server.go`:

```go
// internal/platform/module.go
type Module interface {
    Name() string
    Routes(r chi.Router)                 // HTTP-маршруты модуля
    Subscribe(bus events.Subscriber)     // подписки на события других модулей
    Jobs() []queue.Handler               // фоновые задачи модуля
}

// cmd/api/main.go
mods := []platform.Module{
    identity.New(deps),
    classroom.New(deps),
    learning.New(deps, classroom.API(...)),
    intelligence.New(deps),
    // новый модуль = одна строка здесь
}
```

`deps` — общие платформенные зависимости (БД, транзакции, логгер, шина событий,
конфиг). Межмодульные зависимости передаются **через `api.go`** другого модуля.

---

## 4. Взаимодействие модулей

### 4.1. Синхронный вызов vs событие

| Ситуация | Механизм | Пример |
|---|---|---|
| Нужен ответ, чтобы продолжить (проверка прав, чтение) | Синхронный вызов через `api.go` | learning спрашивает classroom: «состоит ли ученик в группе?» |
| Другой модуль должен **отреагировать** на факт | Доменное событие | `QuizAnswerSubmitted` → intelligence обновляет SM-2 (сейчас — `SetObserver`) |
| Реакция медленная или внешняя (push, SMS, OpenAI) | Событие → очередь (asynq) | `PaymentOverdue` → notifications → push |

### 4.2. Transactional outbox

Проблема: если сохранить данные и отправить событие отдельными шагами,
при сбое между ними событие теряется или уходит без данных («двойная запись»;
Kleppmann, 2017; Richardson, microservices.io).

Решение: событие записывается **в ту же транзакцию**, что и данные, в таблицу
`outbox`; отдельный воркер читает `outbox` и доставляет события подписчикам /
в asynq. Подписчики **идемпотентны** (повторная доставка не ломает данные).

```
BEGIN
  INSERT INTO student_answers ...
  INSERT INTO outbox (event_type, payload) VALUES ('QuizAnswerSubmitted', ...)
COMMIT
        │
        ▼  outbox-воркер (asynq)
  intelligence.OnQuizAnswerSubmitted(...)  ← идемпотентно по event_id
```

Это убирает `SetObserver` и порядок инициализации (P8) и делает AI-обновления
устойчивыми к сбоям OpenAI.

### 4.3. Транзакции

Платформа даёт `tx.Run(ctx, func(ctx) error)`; репозитории берут транзакцию из
`ctx`. **Любая операция, которая пишет в несколько таблиц, — в транзакции**
(закрывает P6; первый кандидат — генерация теста из `aiQuizCreatorAdapter`,
которая переезжает в модуль learning: P5).

---

## 5. Сквозные стандарты

### 5.1. Доступ к данным (закрывает P1)

- **Запрещено** `SELECT *` в коде приложения. Явный список колонок.
- Интеграционный тест на каждую таблицу: структура ↔ колонки (`information_schema`).
- Миграции — по схеме **expand → migrate → contract** (parallel change;
  Sato, 2014): добавили колонку (nullable / с default) → выкатили код → удалили
  старое отдельной миграцией. Каждая миграция обратно совместима с предыдущей
  версией кода, иначе rolling deploy невозможен.

### 5.2. Ошибки (закрывает P2)

- Один пакет `platform/errors`: доменные ошибки с кодом
  (`NOT_FOUND`, `FORBIDDEN`, `VALIDATION`, `CONFLICT`, `INTERNAL`).
- Один HTTP-маппер ошибок для всех модулей (копии `writeError` удаляются).
- **Каждая 5xx логируется** с `request_id`, user_id, маршрутом и цепочкой
  `%w`, и отправляется в Sentry. Клиент получает только код и `request_id`.
- Клиент показывает пользователю текст по **коду**, а не текст исключения
  (см. находку F1 в `research/02-ux-evaluation-methodology.md`).

### 5.3. Авторизация (закрывает P3)

- Политики доступа — в одном месте (`platform/authz`): `can(actor, action, resource)`.
  Повторяющиеся проверки «учитель этой группы?» (`GroupChecker` в 5 модулях)
  становятся одной политикой.
- В JWT — только идентичность и `token_version`; роль/членство либо
  перевыпускаются при изменении (`UserRoleAssigned` → новый токен), либо
  читаются из кэша (Redis, TTL ≤ access TTL).
- **Мультиарендность сразу в модели:** `organization_id` уже есть в `users`.
  Для учебных центров (сегмент S4) — `organization_id` в корневых сущностях
  classroom и проверка арендатора в authz. Добавить позже в 20 таблиц намного
  дороже, чем заложить в 2 корневые сейчас.

### 5.4. API

- Контракт — **OpenAPI 3** спецификация в репозитории; клиентские DTO Flutter
  генерируются или проверяются по ней в CI.
- Версия в пути (`/api/v1`); **ломающие** изменения — только в `/v2`.
  Добавление необязательных полей — не ломающее.
- Списки — с пагинацией по курсору с первого дня (offset ломается на растущих
  таблицах ленты/уведомлений).
- Идемпотентность для POST, которые клиент может повторить по сети
  (сдача ДЗ, оплата*): заголовок `Idempotency-Key`.

### 5.5. Устойчивость к внешним сервисам

Внешние вызовы (OpenAI, Eskiz SMS, FCM, MinIO) — по Nygard (2018):
**таймаут** на каждый вызов, **повтор с экспоненциальной задержкой** только для
идемпотентных операций, **circuit breaker** для OpenAI, **bulkhead**: AI-задачи
в отдельной очереди asynq с ограниченным параллелизмом, чтобы не съесть
воркеры SMS/push.

### 5.6. Наблюдаемость

- Четыре золотых сигнала на маршрут: latency, traffic, errors, saturation
  (Beyer et al., 2016, Google SRE). Prometheus уже подключён.
- **SLO** для ключевых работ, а не для всего API: например, «99 % запросов
  `POST /attendance` быстрее 300 мс; 99,5 % без 5xx за 28 дней».
- Логи — структурированные (slog), с `request_id` во всех строках запроса.

### 5.7. Feature flags

Для постепенного запуска (AI-фичи, оплаты) — флаги с владельцем и сроком
удаления (Hodgson, 2017). Флаг без срока удаления — технический долг.

---

## 6. Тестирование

| Уровень | Что | Где |
|---|---|---|
| Unit | Доменные правила, Rule Engine, SM-2 — без БД | `domain/`, `app/` (уже есть, 9+ пакетов) |
| Интеграционные | Репозитории против **реального Postgres** (testcontainers-go или compose в CI) | `store_test.go` каждого модуля — закрывает P10 и ловит P1 |
| Контрактные | Ответы API соответствуют OpenAPI | CI |
| Архитектурные | Правила зависимостей 3.3, запрет `SELECT *` | CI (fitness functions) |

---

## 7. Метрики архитектуры (fitness functions)

| Метрика | Как считать | Цель |
|---|---|---|
| Нарушения правил зависимостей | depguard / арх-тест | 0 (блокирует merge) |
| `SELECT *` в коде | grep в CI | 0 |
| Нестабильность модуля I = Ce / (Ca + Ce) (Martin, 2002) | По графу импортов | `platform` → ~0 (стабильна); фичи — выше; зависимости только в сторону меньшей I |
| Файлы, изменённые в одном PR **в разных модулях** | `git log` | Тренд вниз: большие значения = границы проведены неверно |
| Горячие точки | частота изменений × размер (Tornhill, 2015) | Топ-5 пересматривается раз в квартал |
| DORA: частота деплоев, lead time, change failure rate, MTTR | CI/CD | Тренд, а не абсолютная цель (Forsgren et al., 2018) |

---

## 8. Дорожная карта миграции

Переход постепенный, по паттерну **strangler fig** (Fowler, 2004): новый код
сразу пишется по целевой схеме, старый переносится, когда его всё равно трогают.
Большого переписывания нет.

| Этап | Что | Закрывает | Оценка |
|---|---|---|---|
| **0. Защита** | Явные колонки вместо `SELECT *`; общий маппер ошибок с логированием 5xx; интеграционный тест «структура ↔ таблица»; CI собирает Docker-образ | P1, P2, P10, P11 | 3–5 дней |
| **1. Платформа** | `platform/tx`, `platform/errors`, `platform/authz`, интерфейс `Module`; перевыпуск токена при смене роли | P3, P6 | 1 неделя |
| **2. Первый модуль по шаблону** | `learning` (quizzes): перенос `aiQuizCreatorAdapter` внутрь модуля, транзакция, событие `QuizAnswerSubmitted` через outbox вместо `SetObserver` | P5, P8 | 1–1,5 недели |
| **3. Разделение AI** | `intelligence` → подмодули documents / generation / mastery / recommendations / coach | P7 | 1–2 недели |
| **4. Остальные модули** | По мере изменений (strangler): classroom, assessment, family, notifications | P4, P9 | Постепенно |
| **5. Новые модули** | `billing` и `schedule` — сразу по шаблону, если JTBD-исследование подтвердит O1–O3 / O6 | — | По приоритету продукта |

**Когда выносить модуль в отдельный сервис:** только при измеренной
необходимости — независимое масштабирование (например, AI-воркеры),
отдельная команда-владелец или иные требования к надёжности. Чистые границы
модульного монолита делают такой вынос механическим.

---

## 9. Чек-лист нового модуля / фичи

- [ ] Связь с работой из JTBD (job story / O#) указана в PR
- [ ] Модуль владеет своими таблицами; миграция по схеме expand/contract
- [ ] Нет `SELECT *`; есть интеграционный тест репозитория
- [ ] Межмодульное взаимодействие — через `api.go` или событие
- [ ] Многотабличная запись — в транзакции; события — через outbox
- [ ] Ошибки — доменные коды; 5xx логируются
- [ ] Права — через `authz`, учитывают `organization_id`
- [ ] Эндпоинты описаны в OpenAPI; списки с курсорной пагинацией
- [ ] Внешние вызовы с таймаутом / повтором / ограничением параллелизма
- [ ] Метрики и SLO для ключевого сценария
- [ ] Решение записано в `DECISIONS.md`, если отклоняется от этого документа

---

## Источники

- Beyer, B., Jones, C., Petoff, J., & Murphy, N. R. (Eds.). (2016). *Site Reliability Engineering: How Google Runs Production Systems.* O'Reilly.
- Cockburn, A. (2005). *Hexagonal Architecture (Ports and Adapters).* alistair.cockburn.us.
- Conway, M. E. (1968). How do committees invent? *Datamation, 14*(4), 28–31.
- Evans, E. (2003). *Domain-Driven Design: Tackling Complexity in the Heart of Software.* Addison-Wesley.
- Ford, N., Parsons, R., & Kua, P. (2017). *Building Evolutionary Architectures.* O'Reilly.
- Forsgren, N., Humble, J., & Kim, G. (2018). *Accelerate: The Science of Lean Software and DevOps.* IT Revolution.
- Fowler, M. (2004). *StranglerFigApplication.* martinfowler.com.
- Fowler, M. (2015). *MonolithFirst.* martinfowler.com.
- Hodgson, P. (2017). *Feature Toggles (aka Feature Flags).* martinfowler.com.
- Kleppmann, M. (2017). *Designing Data-Intensive Applications.* O'Reilly.
- Lehman, M. M. (1980). Programs, life cycles, and laws of software evolution. *Proceedings of the IEEE, 68*(9), 1060–1076.
- MacCormack, A., Baldwin, C., & Rusnak, J. (2012). Exploring the duality between product and organizational architectures: A test of the "mirroring" hypothesis. *Research Policy, 41*(8), 1309–1324.
- Martin, R. C. (2002). *Agile Software Development: Principles, Patterns, and Practices.* Prentice Hall.
- Martin, R. C. (2017). *Clean Architecture.* Prentice Hall.
- Nygard, M. T. (2018). *Release It!* (2nd ed.). Pragmatic Bookshelf.
- Parnas, D. L. (1972). On the criteria to be used in decomposing systems into modules. *Communications of the ACM, 15*(12), 1053–1058.
- Richardson, C. *Pattern: Transactional outbox.* microservices.io.
- Sato, D. (2014). *ParallelChange.* martinfowler.com.
- Tornhill, A. (2015). *Your Code as a Crime Scene.* Pragmatic Bookshelf.
