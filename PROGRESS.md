# Прогресс разработки repetapp

## Текущий этап: 9 — Полировка и QA

## Завершённые этапы

### Этап 8 — Модуль родителей ✅ (commit c332c40)

**Backend:**
- Миграция 000009: parent_links (UNIQUE parent_id+student_id)
- domain/parent.go: ParentLink
- features/parents: LinkChild, UnlinkChild, ListChildren, ChildGroups, ChildGrades, ChildAttendance
- service_test.go: 8 тестов, -race clean
- handler: 6 endpoints /parent/children/**
- postgres/parent_repository.go: полная реализация
- server.go: groupMemberAdapter + wiring

**Flutter:**
- domain: ParentLink entity + ParentsRepository
- data: ParentLinkDto, ParentsApi, ParentsRepositoryImpl
- providers: ChildrenNotifier (StateNotifier) + family providers для групп/оценок/посещаемости
- pages: ParentsPage (список детей, link/unlink dialog, drill-in к оценкам и посещаемости)
- router: /parent/children → ParentsPage

### Этап 7 — Посещаемость ✅ (commit 883e18c)

**Backend:**
- Миграция 000008: attendance (ENUM: present/absent/late/excused), UNIQUE constraint
- features/attendance: Upsert, ListByGroup, ListByGroupStudent (9 тестов, -race)
- Upsert с ON CONFLICT DO UPDATE для идемпотентной записи

**Flutter:**
- teacher view: DatePicker → отметка присутствия для каждого студента
- student view: история посещаемости, цвет по статусу

### Этап 6 — Журнал оценок ✅ (commit e8a3948)

**Backend:**
- Миграция 000007: grades (value NUMERIC, graded_at, comment)
- features/grades: CRUD + ListByGroupStudent (9 тестов, -race)

**Flutter:**
- teacher view: Add/Edit/Delete оценок с sheet
- student view: история оценок с SummaryBar (средний балл)

### Этап 5 — Тесты (quizzes) ✅ (commit d14a7f7)

**Backend:**
- Миграция 000006: quizzes, questions, options, quiz_attempts, student_answers
- features/quizzes: авто-оценка, MaxAttempts, студенты не видят IsCorrect (10 тестов)

**Flutter:**
- pages: QuizzesPage, QuizDetailPage, QuizAttemptPage (start + attempt + result)

### Этап 4 — Домашние задания + MinIO ✅

**Backend:**
- Миграция 000005: assignments, submissions, attachments
- infrastructure/minio: PresignedGetURL, PutObject, noopSigner fallback

### Этап 3 — Лента и объявления ✅

**Backend:**
- Миграция 000004: posts (pinned, soft delete), post_attachments

**Flutter:**
- FeedPage, PostCard, CreatePostSheet

### Этап 2 — Группы и приглашения ✅

**Backend:**
- Миграция 000003: groups, group_members (payment_status)
- invite code: uppercase safe alphabet, crypto/rand, 11 тестов

**Flutter:**
- GroupsPage, GroupDetailPage, GroupCard, JoinGroupSheet

### Этап 1 — Авторизация по SMS ✅
- 10 endpoints, 10 unit-тестов, Eskiz SMS, rate-limit
- Flutter: SplashPage → PhonePage → VerifyPage → RoleSelectPage → ProfileSetupPage

### Этап 0 — Инфраструктура и каркас ✅

## Следующие этапы
- Этап 9: Полировка и QA ← ТЕКУЩИЙ
- Этап 10: Релиз
