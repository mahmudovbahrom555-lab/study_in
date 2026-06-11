# Прогресс разработки repetapp

## Статус: AI Phase 1 завершена ✅

### AI Phase 1 — PDF→Quiz + AI Mentor + Writing Checker + Gamification ✅ (commit abf3a2d)

**Backend (28 новых файлов):**

- Миграция 000011: pgvector расширение + 10 новых таблиц
  - ai_documents, ai_document_chunks (vector(1536) + HNSW индекс)
  - ai_jobs (async статусы задач)
  - ai_sessions, ai_messages (история чатов)
  - topic_tags, question_topics (теги тем для SM-2)
  - topic_mastery (SM-2 алгоритм: interval_days, ease_factor, next_review)
  - student_gamification (XP, streak, longest_streak)
  - xp_events, skill_assessments, ai_token_usage (биллинг)

- infrastructure/openai: Chat (JSONMode), Embed, EmbedBatch, CalcCostUSD
- infrastructure/pdf: ExtractText + ChunkText (sliding window, overlap)
- infrastructure/spaced_repetition: SM-2 алгоритм (Update + IsDue)
- infrastructure/minio: добавлен GetObject
- queue: TypeAIProcessDoc + TypeAIGenerateQuiz + AIProcessor interface

- features/ai — 13 эндпоинтов:
  - POST   /ai/documents                    — загрузка PDF (async chunking)
  - GET    /ai/documents                    — список документов
  - DELETE /ai/documents/{docID}            — удалить документ
  - POST   /ai/documents/{docID}/generate-quiz — запустить генерацию (async)
  - GET    /ai/jobs/{jobID}                 — статус задачи
  - POST   /ai/sessions                     — создать сессию ментора
  - GET    /ai/sessions/{id}                — история сессии
  - POST   /ai/sessions/{id}/messages       — отправить сообщение ментору
  - POST   /ai/writing/check                — проверить эссе (без переписывания)
  - GET    /ai/me/gamification              — XP, стрик, уровень
  - GET    /ai/me/weaknesses                — слабые темы
  - GET    /ai/me/review-queue              — очередь повторения (SM-2)

- server.go: 4 новых адаптера (aiQuizCreator, aiQueueEnqueuer, aiProcessor, noopAIStore)
- config: OpenAI секция (APIKey, Model, EmbeddingModel, MonthlyTokensMax)
- docker-compose: postgres → pgvector/pgvector:pg16 (dev + prod)
- Тесты: 5 unit (CreateDocument, forbidden delete, CreateJob, AwardXP, Gamification) — race clean

---

## Завершённые этапы (1-10)

### Этап 10 — Релиз ✅ (commit e2797bc)
- auto-migrate при AUTO_MIGRATE=true
- docker-compose.prod.yml полный production-стек

### Этап 9 — Домашние задания Flutter ✅ (commit a69322f)
- AssignmentsPage, _CreateAssignmentSheet, submit dialog
- HomeShell role-aware NavigationBar

### Этап 8 — Модуль родителей ✅ (commit c332c40)
- parent_links, ParentsPage, drill-in к оценкам и посещаемости

### Этап 7 — Посещаемость ✅ (commit 883e18c)
- ENUM (present/absent/late/excused), Upsert ON CONFLICT

### Этап 6 — Журнал оценок ✅ (commit e8a3948)
- grades, CRUD + ListByGroupStudent, Flutter SummaryBar

### Этап 5 — Тесты (quizzes) ✅ (commit d14a7f7)
- авто-оценка, MaxAttempts, студенты не видят IsCorrect

### Этап 4 — Домашние задания + MinIO ✅
- assignments, submissions, attachments, signed URLs

### Этап 3 — Лента ✅
- posts, post_attachments, pinned

### Этап 2 — Группы ✅
- groups, group_members, invite code (crypto/rand)

### Этап 1 — Auth ✅
- SMS OTP, JWT access+refresh, Eskiz, rate-limit, 10 тестов

### Этап 0 — Инфраструктура ✅
- Go+Flutter skeleton, Docker, CI/CD

## Мониторинг (commit 28c25fc)
- Prometheus /metrics, Grafana, Loki, Sentry, FCM stub
- docker-compose.monitoring.yml

## Следующие фазы AI

### AI Phase 2 (не начата)
- Speaking Assessment (Whisper API)
- История проверок writing
- DELETE /ai/sessions/{id}

### AI Phase 3 (не начата)
- skill_assessments: speaking + reading CEFR оценки
- GET /students/{id}/skill-history

### AI Phase 4 (не начата)
- Дашборд учителя: GET /groups/{id}/ai-insights
- POST /me/daily-goal

### AI Phase 5 (не начата)
- AI Progress Coach для родителей/учителей
- GET /groups/{id}/students/{sid}/progress
