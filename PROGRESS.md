# Прогресс разработки repetapp

## Текущий этап: 6 — Журнал оценок

## Завершённые этапы

### Этап 5 — Тесты (quizzes) ✅

**Backend:**
- Миграция 000006: quizzes, questions, options, quiz_attempts, student_answers
- domain/quiz.go: Quiz, Question, Option, QuizAttempt, StudentAnswer
- features/quizzes: Repository interface, DTO, Service (10 тестов, -race), Handler (8 endpoints)
- postgres/quiz_repository.go: полная реализация
- server.go: wiring quizzes handler
- Авто-оценка при SubmitAttempt, MaxAttempts, студенты не видят IsCorrect

**Flutter:**
- domain: Quiz, Question, Option, QuizAttempt entities + QuizzesRepository
- data: Quiz/Question/Option/Attempt DTOs, QuizzesApi, QuizzesRepositoryImpl
- providers: QuizzesNotifier, QuizDetailNotifier, AttemptNotifier
- pages: QuizzesPage, QuizDetailPage, QuizAttemptPage (start screen + attempt + result dialog)
- widgets: QuizCard, CreateQuizSheet
- router: /groups/:id/quizzes, /groups/:groupId/quizzes/:quizId, .../attempt

### Этап 4 — Домашние задания + MinIO ✅

**Backend:**
- Миграция 000005: assignments, assignment_attachments, submissions, submission_attachments
- infrastructure/minio: Client (PresignedGetURL, PutObject, RemoveObject)
- features/assignments: service + handler + postgres repo
- noopSigner/noopStore — graceful degradation при отсутствии MinIO

### Этап 3 — Лента и объявления ✅

**Backend:**
- Миграция 000004: posts (pinned, soft delete), post_attachments (MinIO object_key)
- features/feed: GroupChecker interface, Service, Handler (URL signer), Repository

**Flutter:**
- pages: FeedPage, widgets: PostCard, CreatePostSheet
- router: /groups/:id/feed

### Этап 2 — Группы и приглашения ✅

**Backend:**
- Миграция 000003: groups, group_members (payment_status)
- features/groups: invite code (uppercase safe alphabet, crypto/rand), 11 тестов

**Flutter:**
- pages: GroupsPage, GroupDetailPage
- widgets: GroupCard, MemberTile, CreateGroupSheet, JoinGroupSheet

### Этап 1 — Авторизация по SMS ✅
- Миграция 000002: users, verification_codes, refresh_tokens, device_tokens
- features/auth: 10 endpoints, 10 unit-тестов
- Flutter: SplashPage, PhonePage, VerifyPage, RoleSelectPage, ProfileSetupPage

### Этап 0 — Инфраструктура и каркас ✅

## Следующие этапы
- Этап 6: Журнал оценок (grades) ← ТЕКУЩИЙ
- Этап 7: Посещаемость (attendance)
- Этап 8: Модуль родителей
- Этап 9: Полировка и QA
- Этап 10: Релиз
