# Прогресс разработки repetapp

## Статус: Teacher Dashboard завершён ✅ | AI Phase 2 в работе

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

---

## Teacher Dashboard ✅ (commits 0ed9056, d3ee7b1, 5b54964)

### Backend:
- Migration 000012: quiz_question_feedback (TAR — Teacher Acceptance Rate)
- Migration 000013: ai_generation_sessions (golden dataset: generated/accepted/edited/rejected)
- InsightsRepository: ClassInsights + StudentProgress (SQL analytics)
- QuizFeedbackRepository: UpsertFeedback + AcceptanceRate
- AIGenerationSessionRepository: CreateSession → LinkQuiz → SyncCounts → TeacherStats
- GenerateQuiz: открывает session до GPT-вызова, линкует к quiz после
- SubmitQuestionFeedback: lazily SyncCountsByQuestion
- "Next Best Action" эвристики (P1/P2/P3) — без GPT, чистый Go
- Новые эндпоинты: /groups/{id}/ai-insights, /groups/{id}/students/{sid}/progress,
  /questions/{id}/feedback, /me/acceptance-rate, /me/generation-stats
- 9 unit тестов (race clean)

### Flutter:
- AiInsightsPage: статкарты (студенты/TAR/средний балл), "Next Best Action" панель
  с цветными карточками + приоритетными бейджами, слабые темы, таблица лидеров
- StudentProgressPage: геймификация (XP/стрик/рекорд), CEFR навыки, темы, история тестов
- Роуты: /groups/:id/ai-insights, /groups/:id/students/:id/progress
- Кнопка "AI Insights" в group_detail быстрых действиях (только для teacher)
- TeacherGenerationStats DTO + provider

---

---

## AI Phase 2 — Confidence + Consistency + SM-2 Extension ✅

### Backend:
- Migration 000014: topic_mastery ← confidence_score (NUMERIC 4,3), consistency_score (NUMERIC 4,3), correct_streak (INT)
- SM-2: Update() расширен — EMA confidence (α=0.3), streak-based consistency (÷2 на ошибке), correct_streak
- domain/ai.go: TopicMastery + StudentTopicDetail получили confidence_score, consistency_score, correct_streak
- ai_repository.go: UpsertMastery SQL обновлён с новыми колонками
- insights_repository.go: StudentProgress запрос читает confidence_score, consistency_score
- 7 unit тестов SM-2 (race clean)

### Flutter:
- StudentTopicDetail entity: confidenceScore, consistencyScore (default 0.5)
- StudentTopicDetailDto: парсит confidence_score, consistency_score из JSON
- StudentProgressPage: `_TopicRow` → Card + `_ScoreBar` (mini progress bars для Уверенности и Стабильности)

---

## AI Phase 3 — Rule Engine + Recommendation Engine ✅ (commits 9fea19c, e5a5b62)

### Backend:
- Migration 000015: ai_recommendations + recommendation_outcomes
- Rule Engine (pure Go, 0 GPT calls): 7 rules — mastery_critical, mastery_low,
  consistency_critical, review_needed, at_risk, tar_low, celebrate
- rule_data (JSONB) snapshot of metrics at trigger time — feeds outcome measurement
- GetClassInsights() runs Rule Engine → ReplaceForGroup → lazy outcome measurement (background goroutine, 7-day window)
- POST /ai/recommendations/{id}/action — accept | dismiss | snooze
- GET  /ai/recommendations/{id}/explain — GPT prose (lazy; falls back to rule reason)
- TopicWeakness extended: avg_confidence + avg_consistency
- 9 Rule Engine tests + 11 feature tests, all passing

### Flutter:
- TeacherRecommendation entity: id field added
- TopicWeakness entity: avgConfidence, avgConsistency
- _RecommendationCard → ConsumerStatefulWidget with Accept / Dismiss / Почему? actions
- GPT explanation shown in bottom sheet on tap

## Следующие фазы AI

### AI Phase 4 — AI Coach для студентов (следующее)
- Student-facing confidence/consistency dashboard
- Speaking Assessment (Whisper API)
- skill_assessments CEFR оценки
- POST /me/daily-goal
