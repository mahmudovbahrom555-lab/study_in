# Прогресс разработки repetapp

## Текущий этап: 5 — Тесты (quiz)

## Завершённые этапы

### Этап 2 — Группы и приглашения ✅

**Backend:**
- Миграция 000003: groups (invite_code VARCHAR(10) UNIQUE), group_members (payment_status CHECK)
- domain/group.go: Group, GroupMember, GroupMemberDetail, PaymentStatus (trial/pending/paid)
- domain/errors.go: добавлены ErrAlreadyMember, ErrNotMember, ErrRoleAlreadySet
- features/groups: Repository interface, DTO, Service (10 методов), Handler (13 endpoints)
- infrastructure/postgres/group_repository.go: полная реализация
- invite_code: 8 символов, uppercase-only safe alphabet (без 0/O/1/I), crypto/rand
- server.go: подключён groupHandler под Auth middleware
- 11 unit-тестов сервиса — все проходят с race detector

**Flutter:**
- domain: Group, GroupMember entities, PaymentStatus enum, GroupsRepository interface
- data: GroupDto/GroupMemberDto (fromJson/toDomain), GroupsApi, GroupsRepositoryImpl
- providers: GroupsState/GroupsNotifier (list+crud), MembersNotifier (per-group, family)
- pages: GroupsPage (список с FAB для учителя), GroupDetailPage (участники + оплата)
- widgets: GroupCard, MemberTile (с PaymentChip + picker), CreateGroupSheet, JoinGroupSheet
- router: добавлены /groups и /groups/:id, home → GroupsPage

⚠️ Flutter SDK не установлен. IDE-ошибки исчезнут после `flutter pub get`.

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

### Этап 0 — Инфраструктура и каркас ✅
- Go: Chi, config (Viper), middleware (logging, recovery, CORS), health/version endpoints, graceful shutdown
- Flutter: Riverpod, go_router, Dio, тема (light/dark), локализация (uz/uz-cyrl/ru/en)
- Docker Compose: PostgreSQL 16 + Redis 7 + MinIO
- GitHub Actions CI: lint → test → build для Go и Flutter

### Этап 3 — Лента и объявления ✅

**Backend:**
- Миграция 000004: posts (body, pinned, soft delete), post_attachments (object_key, MinIO)
- domain/post.go: Post, PostAttachment, PostWithMeta
- features/feed: Repository interface + GroupChecker interface, DTO, Service (8 методов), Handler (7 endpoints)
- Endpoints: GET/POST /groups/{id}/feed, GET/PATCH/DELETE /feed/{postID}, POST pin/unpin
- URLSigner interface: noopSigner (placeholder до Этапа 4 когда подключим MinIO)
- infrastructure/postgres/feed_repository.go: полная реализация
- 11 unit-тестов: все проходят с -race

**Flutter:**
- domain: Post, PostAttachment entities, FeedRepository interface
- data: PostDto/PostAttachmentDto, FeedApi, FeedRepositoryImpl
- providers: FeedState/FeedNotifier (family per groupId, infinite scroll)
- pages: FeedPage (infinite scroll, pull-to-refresh, FAB для учителя)
- widgets: PostCard (pinned highlight, popup menu, attachment chips), CreatePostSheet
- router: /groups/:id/feed → FeedPage

### Этап 4 — Домашние задания + MinIO ✅

**Backend:**
- Миграция 000005: assignments, assignment_attachments, submissions, submission_attachments
- domain/assignment.go: Assignment, AssignmentAttachment, Submission, SubmissionAttachment
- infrastructure/minio/client.go: PresignedGetURL, PutObject, RemoveObject (private bucket)
- features/assignments: Repository, GroupChecker, Signer, ObjectStore interfaces
- Service: CreateAssignment, UploadAssignmentFile, GetAssignment, ListAssignments,
  UpdateAssignment, DeleteAssignment, Submit, UploadSubmissionFile, Grade, ListSubmissions
- 50 MB file size limit, sanitizeFilename() против path traversal
- Handler: 10 endpoints включая multipart file upload
- Fallback: noopSigner/noopStore если MinIO недоступен (dev без S3_ENDPOINT)
- 10 unit-тестов, все проходят с -race

**Решение по ТЗ:** MinIO SDK добавлен (minio-go/v7). Avatar upload теперь тоже возможен
через тот же MinIO client (Этап 4 → endpoint POST /api/v1/me/avatar активируется).

## Следующие этапы
- **Этап 5:** Тесты/Quiz (quiz, questions, options, student_answers, scoring)
- Этап 4: Домашние задания + MinIO (загрузка файлов, signed URLs)
- Этап 5: Тесты
- Этап 6: Журнал оценок
- Этап 7: Посещаемость
- Этап 8: Родительский модуль
- Этап 9: Полировка и QA
- Этап 10: Релиз
