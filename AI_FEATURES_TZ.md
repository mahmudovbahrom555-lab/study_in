# ТЗ: AI-слой для RepetApp

**Версия:** 1.0  
**Дата:** 2026-06-12  
**Статус:** Планирование

---

## Стратегия

RepetApp сейчас — электронный журнал. С AI он становится **платформой адаптивного обучения**.

Конкурентная формула:
```
Репетитор + RepetApp AI  ≈  Duolingo + живой учитель + CEFR-тренажёр
```

Монетизационная ставка: преподаватель платит за экономию времени, ученик/родитель — за результат (CEFR, IELTS, олимпиада).

---

## Этапы внедрения

```
Фаза 1 (MVP AI)     → Генерация тестов по PDF           ← НАЧИНАЕМ ЗДЕСЬ
Фаза 2              → AI Mentor + Writing Checker
Фаза 3              → Speaking Assessment (Whisper)
Фаза 4              → Duolingo-механика + XP + Streak
Фаза 5              → Progress Coach + уведомления родителям
```

---

## Архитектурные принципы

1. **Не ломаем существующую архитектуру.** AI — отдельный feature-пакет `features/ai/`.
2. **Все тяжёлые задачи через Asynq.** PDF-обработка, генерация изображений, TTS — в очередь.
3. **RAG через pgvector.** Никакой отдельной векторной БД. Расширение pgvector в PostgreSQL.
4. **OpenAI с fallback.** Если API недоступен — graceful degradation, не 500.
5. **Промпты — код.** Промпты хранятся в `internal/features/ai/prompts/` как константы Go, версионируются вместе с кодом.
6. **Токены — деньги.** Каждый запрос к GPT логируется с количеством токенов. Лимиты на уровне тарифа.

---

## Модель данных (новые таблицы)

### `ai_documents` — загруженные учебники/материалы
```sql
CREATE TABLE ai_documents (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id    UUID REFERENCES groups(id) ON DELETE CASCADE,
    teacher_id  UUID REFERENCES users(id) ON DELETE CASCADE,
    title       VARCHAR(300) NOT NULL,
    object_key  TEXT NOT NULL,              -- MinIO key для PDF
    mime_type   VARCHAR(100) NOT NULL,
    size_bytes  BIGINT NOT NULL,
    status      VARCHAR(30) NOT NULL DEFAULT 'pending'  -- pending/processing/ready/failed
                CHECK (status IN ('pending','processing','ready','failed')),
    chunk_count INT NOT NULL DEFAULT 0,
    error_msg   TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_ai_documents_group ON ai_documents(group_id);
CREATE INDEX idx_ai_documents_teacher ON ai_documents(teacher_id);
```

### `ai_document_chunks` — чанки + векторные эмбеддинги (pgvector)
```sql
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE ai_document_chunks (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES ai_documents(id) ON DELETE CASCADE,
    chunk_index INT NOT NULL,
    content     TEXT NOT NULL,
    embedding   vector(1536),   -- text-embedding-3-small
    token_count INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_chunks_document ON ai_document_chunks(document_id);
-- HNSW индекс для быстрого similarity search
CREATE INDEX idx_chunks_embedding ON ai_document_chunks
    USING hnsw (embedding vector_cosine_ops)
    WITH (m = 16, ef_construction = 64);
```

### `ai_sessions` — сессии разговора с AI
```sql
CREATE TABLE ai_sessions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_type VARCHAR(30) NOT NULL  -- mentor/writing/speaking/reading/quiz_gen
                 CHECK (session_type IN ('mentor','writing','speaking','reading','quiz_gen')),
    document_id UUID REFERENCES ai_documents(id) ON DELETE SET NULL,
    subject     VARCHAR(100),
    cefr_level  VARCHAR(5),             -- A1/A2/B1/B2/C1/C2
    metadata    JSONB NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_ai_sessions_user ON ai_sessions(user_id, created_at DESC);
```

### `ai_messages` — история сообщений
```sql
CREATE TABLE ai_messages (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id  UUID NOT NULL REFERENCES ai_sessions(id) ON DELETE CASCADE,
    role        VARCHAR(15) NOT NULL CHECK (role IN ('user','assistant','system')),
    content     TEXT NOT NULL,
    tokens_in   INT NOT NULL DEFAULT 0,
    tokens_out  INT NOT NULL DEFAULT 0,
    model       VARCHAR(50) NOT NULL DEFAULT 'gpt-4o-mini',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_ai_messages_session ON ai_messages(session_id, created_at);
```

### `topic_tags` — предметные темы для отслеживания слабых мест
```sql
CREATE TABLE topic_tags (
    id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name    VARCHAR(200) NOT NULL,
    subject VARCHAR(100),
    UNIQUE (name, subject)
);

CREATE TABLE question_topics (
    question_id UUID REFERENCES questions(id) ON DELETE CASCADE,
    topic_id    UUID REFERENCES topic_tags(id) ON DELETE CASCADE,
    PRIMARY KEY (question_id, topic_id)
);
```

### `topic_mastery` — прогресс студента по темам (основа для Duolingo-механики)
```sql
CREATE TABLE topic_mastery (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    topic_id      UUID NOT NULL REFERENCES topic_tags(id) ON DELETE CASCADE,
    correct_count INT NOT NULL DEFAULT 0,
    total_count   INT NOT NULL DEFAULT 0,
    -- Spaced repetition: когда показывать снова
    next_review   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    interval_days INT NOT NULL DEFAULT 1,   -- SM-2 интервал
    ease_factor   NUMERIC(4,2) NOT NULL DEFAULT 2.5,  -- SM-2 ease
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (student_id, topic_id)
);
CREATE INDEX idx_mastery_student ON topic_mastery(student_id);
CREATE INDEX idx_mastery_review ON topic_mastery(student_id, next_review);
```

### `student_gamification` — XP, стрики, уровни (Duolingo)
```sql
CREATE TABLE student_gamification (
    student_id        UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    xp_total          INT NOT NULL DEFAULT 0,
    xp_today          INT NOT NULL DEFAULT 0,
    daily_goal_xp     INT NOT NULL DEFAULT 50,
    streak_days       INT NOT NULL DEFAULT 0,
    longest_streak    INT NOT NULL DEFAULT 0,
    last_activity_date DATE,
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE xp_events (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount     INT NOT NULL,
    reason     VARCHAR(100) NOT NULL,  -- 'quiz_correct','writing_submitted','streak_bonus'
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_xp_events_student ON xp_events(student_id, created_at DESC);
```

### `skill_assessments` — результаты AI-оценки навыков
```sql
CREATE TABLE skill_assessments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    skill       VARCHAR(20) NOT NULL CHECK (skill IN ('reading','writing','speaking','listening')),
    cefr_level  VARCHAR(5),
    score       NUMERIC(5,2),
    feedback    TEXT,
    audio_key   TEXT,               -- MinIO key для speaking записей
    transcript  TEXT,               -- Whisper транскрипция
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_assessments_student ON skill_assessments(student_id, skill, created_at DESC);
```

### `ai_token_usage` — биллинг и лимиты
```sql
CREATE TABLE ai_token_usage (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    model       VARCHAR(50) NOT NULL,
    tokens_in   INT NOT NULL,
    tokens_out  INT NOT NULL,
    feature     VARCHAR(50) NOT NULL,  -- 'mentor','quiz_gen','writing','speaking'
    cost_usd    NUMERIC(10,6) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_token_usage_user_month ON ai_token_usage(user_id, created_at);
```

---

## API Design

### Фаза 1: Документы и генерация тестов

```
POST   /ai/documents                          — загрузить PDF/текст (multipart)
GET    /ai/documents                          — список документов в группе
DELETE /ai/documents/{docID}                  — удалить документ

POST   /ai/documents/{docID}/generate-quiz    — запустить генерацию (async, возвращает job_id)
GET    /ai/jobs/{jobID}                        — статус задачи (pending/done/failed)

POST   /ai/documents/{docID}/ask              — задать вопрос по документу (RAG)
```

### Фаза 2: AI Mentor и Writing

```
POST   /ai/sessions                           — создать сессию (type: mentor/writing/reading)
GET    /ai/sessions/{id}                      — получить историю
POST   /ai/sessions/{id}/messages             — отправить сообщение
DELETE /ai/sessions/{id}                      — удалить сессию

POST   /ai/writing/check                      — проверить эссе без переписывания
GET    /ai/writing/history                    — история проверок
```

### Фаза 3: Speaking

```
POST   /ai/speaking/sessions                  — создать speaking-сессию
POST   /ai/speaking/{id}/submit               — загрузить аудио (multipart, ≤5 MB)
GET    /ai/speaking/{id}/result               — получить оценку (async)
GET    /students/{id}/skill-history           — история оценок навыков
```

### Фаза 4: Gamification + Progress

```
GET    /me/gamification                       — XP, стрик, уровень
GET    /me/weaknesses                         — темы со слабыми результатами
GET    /me/review-queue                       — вопросы для повторения (spaced repetition)
POST   /me/daily-goal                         — установить дневную цель XP

GET    /groups/{id}/ai-insights               — дашборд учителя: слабые места класса
GET    /groups/{id}/students/{sid}/progress   — детальный прогресс ученика (для учителя/родителя)
```

---

## Стратегия промптов

Промпты — это бизнес-логика. Хранить в `internal/features/ai/prompts/`.

### Mentor (никогда не даёт готовый ответ)

```go
const MentorSystem = `
Ты — AI-наставник для ученика. Твоя роль:
ДЕЛАЙ:
- Задавай наводящие вопросы
- Давай подсказки, но не ответы
- Разбивай сложное на шаги
- Хвали за правильные рассуждения
- Говори на языке ученика (рус/узб/англ — как он пишет)

НЕ ДЕЛАЙ:
- Не давай готовое решение домашнего задания
- Не пиши текст вместо ученика
- Если ученик настаивает на ответе — объясни, что понимание важнее ответа

Уровень ученика: {{.CEFRLevel}}
Предмет: {{.Subject}}
Максимум 3-4 предложения на каждый ответ.
`
```

### Генератор тестов (детерминированный JSON)

```go
const QuizGeneratorSystem = `
Ты генерируешь тесты по учебным материалам.
Верни ТОЛЬКО валидный JSON без markdown, без объяснений.

Формат:
{
  "title": "...",
  "questions": [
    {
      "text": "Вопрос",
      "topic_tags": ["тема1", "тема2"],
      "difficulty": "A1|A2|B1|B2|C1|C2",
      "options": [
        {"text": "Вариант", "is_correct": true},
        {"text": "Вариант", "is_correct": false}
      ],
      "explanation": "Краткое объяснение правильного ответа"
    }
  ]
}

Правила:
- Ровно 4 варианта ответа на каждый вопрос
- Ровно 1 правильный ответ
- topic_tags — конкретные темы из материала (2-3 слова)
- difficulty — CEFR-уровень вопроса
`
```

### Writing Evaluator (не исправляет, направляет)

```go
const WritingEvaluatorSystem = `
Оцени письменную работу студента. НЕ переписывай текст.

Дай оценку в формате JSON:
{
  "cefr_estimate": "B1",
  "scores": {
    "grammar": 7,
    "vocabulary": 6,
    "coherence": 8,
    "task_completion": 7
  },
  "errors": [
    {"fragment": "I goed to", "hint": "Какая форма Past Simple у неправильных глаголов?"}
  ],
  "strengths": ["..."],
  "one_improvement": "Одна конкретная рекомендация"
}

Подсказки должны направлять к самостоятельному исправлению.
Максимум 3 ошибки в errors (самые важные).
`
```

### Speaking Evaluator

```go
const SpeakingEvaluatorSystem = `
Оцени транскрипцию устной речи студента.

JSON:
{
  "cefr_estimate": "B1",
  "scores": {
    "fluency": 6,
    "grammar_range": 7,
    "vocabulary": 6,
    "coherence": 8
  },
  "notable_errors": ["глагол to be опущен в предложении №2"],
  "strengths": ["хорошая связность"],
  "feedback": "2-3 предложения мотивирующего фидбека"
}
`
```

---

## Spaced Repetition (SM-2 алгоритм)

После каждого ответа на вопрос обновляем `topic_mastery`:

```go
// SM-2: при правильном ответе увеличиваем интервал
func updateSM2(mastery *TopicMastery, correct bool, quality int) {
    // quality: 0-5 (0=неправильно, 3=с трудом, 5=легко)
    if !correct {
        mastery.IntervalDays = 1
        mastery.EaseFactor = max(1.3, mastery.EaseFactor - 0.2)
    } else {
        switch mastery.IntervalDays {
        case 1: mastery.IntervalDays = 6
        case 6: mastery.IntervalDays = 15
        default:
            mastery.IntervalDays = int(float64(mastery.IntervalDays) * mastery.EaseFactor)
        }
        mastery.EaseFactor = mastery.EaseFactor + 0.1 - (5-quality)*0.08
        mastery.EaseFactor = max(1.3, mastery.EaseFactor)
    }
    mastery.NextReview = time.Now().AddDate(0, 0, mastery.IntervalDays)
}
```

Это значит: ученик ответил неправильно на тему "Present Perfect" → вопросы по этой теме появятся завтра. Ответил правильно → через 6 дней. Ответил легко несколько раз → через 2-3 недели.

---

## XP и Gamification (Duolingo-механика)

| Действие | XP |
|----------|-----|
| Правильный ответ в тесте | +10 |
| Правильный ответ на слабую тему | +20 (бонус) |
| Сдал ДЗ вовремя | +15 |
| Письменная работа проверена AI | +25 |
| Speaking сессия завершена | +30 |
| Стрик 7 дней | +100 (бонус) |
| Достиг дневной цели | +50 |
| Все вопросы теста правильно | +50 (perfect score) |

Уровни (CEFR → XP маппинг):
```
0-500 XP      → Starter
500-2000 XP   → A1
2000-5000 XP  → A2
5000-12000 XP → B1
...
```

---

## Дашборд слабых мест (для учителя)

```json
{
  "group_id": "...",
  "period": "last_30_days",
  "class_weakness": [
    {
      "topic": "Present Perfect",
      "avg_accuracy": 0.42,
      "students_struggling": 8,
      "total_students": 12,
      "recommended_action": "Рекомендуется повторное объяснение"
    }
  ],
  "students": [
    {
      "student_id": "...",
      "name": "Алишер",
      "xp": 1240,
      "streak": 5,
      "top_weakness": "Conditionals",
      "skill_levels": {
        "reading": "B1",
        "writing": "A2",
        "speaking": "A2"
      }
    }
  ]
}
```

---

## Стоимость API

### Модели

| Задача | Модель | Стоимость |
|--------|--------|-----------|
| Генерация тестов | gpt-4o-mini | ~$0.001 за 20 вопросов |
| AI Mentor (чат) | gpt-4o-mini | ~$0.001 за сообщение |
| Writing check | gpt-4o-mini | ~$0.002 за эссе |
| Speaking (Whisper) | whisper-1 | $0.006/мин |
| TTS (читать текст) | tts-1 | $0.015/1k символов |
| Эмбеддинги (RAG) | text-embedding-3-small | $0.02/1M токенов |
| Изображения | dall-e-3 | $0.04/image |

### Расходы на активного пользователя

| Профиль | Расход/месяц |
|---------|-------------|
| Учитель (3 PDF, 100 тестов) | ~$0.50 |
| Ученик (mentor 5 msg/день) | ~$1.50 |
| Ученик (+ speaking 3x/нед) | ~$4.00 |
| Ученик (полный режим) | ~$7.00 |

### Маржинальность

| Тариф | Цена | Себестоимость AI | Маржа |
|-------|------|-----------------|-------|
| Teacher Pro | 100k сум (~$8) | $0.50-2 | 75-94% |
| Student AI | 50k сум (~$4) | $2-7 | 0-50% |
| Center (20 учеников) | 500k сум (~$40) | $15-30 | 25-63% |

**Риск:** Student AI при интенсивном использовании speaking убыточен. Решение: лимит 10 speaking-сессий в месяц, далее — pay-per-use.

---

## Структура кода (features/ai/)

```
internal/
  features/
    ai/
      repository.go         — интерфейсы: DocumentRepo, SessionRepo, MasteryRepo
      service.go            — оркестрация: GenerateQuiz, SendMentorMessage, CheckWriting
      handler.go            — HTTP endpoints
      prompts/
        mentor.go
        quiz_generator.go
        writing_evaluator.go
        speaking_evaluator.go
      service_test.go

  infrastructure/
    openai/
      client.go             — OpenAI API клиент (chat, embeddings, whisper, tts, dall-e)
      models.go             — константы моделей и лимиты
    spaced_repetition/
      sm2.go                — SM-2 алгоритм
    pdf/
      extractor.go          — PDF текст → чанки (pdfcpu или poppler)
```

---

## Порядок реализации (по фазам)

### Фаза 1 — Teacher AI (2-3 недели)

**Goal:** Учитель загружает PDF → получает готовые тесты за 30 секунд.

Задачи:
1. `go get github.com/sashabaranov/go-openai` — OpenAI SDK
2. `infrastructure/openai/client.go` — обёртка над API
3. `infrastructure/pdf/extractor.go` — парсинг PDF (pdfcpu, без OCR пока)
4. Migration: `ai_documents`, `ai_document_chunks` (pgvector)
5. Asynq task: `ai:process_document` — chunk + embed
6. Asynq task: `ai:generate_quiz` — RAG + GPT → JSON → populate quizzes tables
7. API: `POST /ai/documents`, `POST /ai/documents/{id}/generate-quiz`, `GET /ai/jobs/{id}`
8. Flutter: страница загрузки PDF с прогрессом, просмотр сгенерированного теста

### Фаза 2 — AI Mentor + Writing (2 недели)

1. Migration: `ai_sessions`, `ai_messages`, `topic_tags`, `question_topics`
2. `features/ai/service.go` — mentor chat с оконным контекстом (last 10 messages)
3. Writing checker — POST endpoint + оценка без переписывания
4. Обновление `student_answers` — добавить topic tagging при ответе
5. `topic_mastery` обновляется после каждого quiz attempt
6. Flutter: chat UI для ментора, writing submission, результат с подсветкой

### Фаза 3 — Speaking (1-2 недели)

1. Migration: `skill_assessments`
2. Whisper API интеграция (аудио → транскрипция)
3. Speaking evaluator промпт → CEFR score
4. MinIO: хранение аудио (teacher может прослушать)
5. Flutter: запись аудио, upload, polling результата

### Фаза 4 — Gamification (1 неделя)

1. Migration: `student_gamification`, `xp_events`, `topic_mastery`
2. SM-2 алгоритм в `infrastructure/spaced_repetition/`
3. XP начисляется в сервисах quiz/assignments/ai
4. Streak: проверяется при каждом начислении XP
5. API: `/me/gamification`, `/me/weaknesses`, `/me/review-queue`
6. Flutter: XP bar, streak flame icon, weakness карточки, review queue

### Фаза 5 — Progress Coach (1-2 недели)

1. Еженедельный отчёт учителю (Asynq cron task)
2. Notification родителю: "Алишер завершил 5 сессий, слабая тема: Past Perfect"
3. `/groups/{id}/ai-insights` — дашборд слабых мест класса
4. `/students/{id}/progress` — для родительского модуля
5. Flutter: progress screen с графиками, weekly report

---

## Монетизация (финальная схема)

### Tiers

```
FREE
├── Группы + ДЗ + оценки (без AI)
└── 5 AI-запросов/месяц (бесплатный пробник)

TEACHER PRO — 100,000 сум/мес
├── Неограниченная загрузка PDF
├── Генерация тестов по документу
├── Генерация ДЗ и планов уроков
├── Writing checker (проверка работ)
├── Дашборд слабых мест класса
└── 200 AI-сообщений ментора/мес (для демонстраций)

STUDENT LEARNER — 50,000 сум/мес
├── AI Mentor (100 сообщений/мес)
├── Writing (20 проверок/мес)
├── Reading comprehension (неограниченно)
├── XP + Streak + Review Queue (Duolingo)
└── Skill history (reading/writing уровень)

STUDENT INTENSIVE — 100,000 сум/мес
├── Всё из Learner
├── Speaking (15 сессий/мес)
├── AI Mentor (300 сообщений/мес)
└── CEFR progress tracking

CENTER — 500,000 сум/мес (до 30 учеников)
├── Все Teacher Pro фичи
├── Student Intensive для всех учеников центра
└── Аналитика центра (сравнение классов)
```

### Overage pricing

Если лимит закончился:
- +1000 сообщений ментора: 30,000 сум
- +5 speaking сессий: 20,000 сум

---

## Технические риски и митигация

| Риск | Вероятность | Митигация |
|------|-------------|-----------|
| OpenAI недоступен из Узбекистана | Средняя | Azure OpenAI как резерв; прокси через собственный сервер |
| PDF с узбекским текстом (кириллица/латиница) | Высокая | Few-shot примеры в промпте; явно указывать язык материала |
| Speaking: шум в записи | Средняя | Whisper устойчив к шуму; добавить инструкцию по записи |
| Высокие расходы на speaking | Средняя | Жёсткие лимиты на уровне middleware до вызова API |
| Ученик "взламывает" ментора ("всё равно дай ответ") | Высокая | Prompt hardening; логирование нарушений; rate limit на попытки |
| OCR для отсканированных PDF | Высокая | В Фазе 1 только текстовые PDF; OCR (Tesseract) — Фаза 2 |

---

## Оценка PMF с AI

| Продукт | PMF |
|---------|-----|
| Текущий LMS | 7.5/10 |
| + Teacher AI (тесты по PDF) | 8.5/10 |
| + Student Mentor + Writing | 9.0/10 |
| + Speaking + Gamification | 9.5/10 |
| + Progress Coach | 9.5-10/10 |

---

*Следующий шаг: Фаза 1 — `go get github.com/sashabaranov/go-openai` + infrastructure/openai/client.go + ai_documents migration*
