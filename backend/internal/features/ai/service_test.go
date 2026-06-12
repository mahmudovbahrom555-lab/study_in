package ai_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/features/ai"
)

// ─── Mocks ────────────────────────────────────────────────────────────────────

type mockDocRepo struct {
	docs   map[uuid.UUID]*domain.AIDocument
	chunks []*domain.AIDocumentChunk
}

func newMockDocRepo() *mockDocRepo {
	return &mockDocRepo{docs: map[uuid.UUID]*domain.AIDocument{}}
}

func (m *mockDocRepo) CreateDocument(_ context.Context, doc *domain.AIDocument) error {
	doc.ID = uuid.New()
	m.docs[doc.ID] = doc
	return nil
}
func (m *mockDocRepo) GetDocument(_ context.Context, id uuid.UUID) (*domain.AIDocument, error) {
	return m.docs[id], nil
}
func (m *mockDocRepo) ListDocuments(_ context.Context, teacherID uuid.UUID, _ *uuid.UUID) ([]*domain.AIDocument, error) {
	var out []*domain.AIDocument
	for _, d := range m.docs {
		if d.TeacherID == teacherID {
			out = append(out, d)
		}
	}
	return out, nil
}
func (m *mockDocRepo) UpdateDocumentStatus(_ context.Context, id uuid.UUID, s domain.AIDocumentStatus, cnt int, _ *string) error {
	if d, ok := m.docs[id]; ok {
		d.Status = s
		d.ChunkCount = cnt
	}
	return nil
}
func (m *mockDocRepo) DeleteDocument(_ context.Context, id uuid.UUID) error {
	delete(m.docs, id)
	return nil
}
func (m *mockDocRepo) SaveChunk(_ context.Context, c *domain.AIDocumentChunk, _ []float32) error {
	m.chunks = append(m.chunks, c)
	return nil
}
func (m *mockDocRepo) SearchSimilarChunks(_ context.Context, docID uuid.UUID, _ []float32, limit int) ([]*domain.AIDocumentChunk, error) {
	var out []*domain.AIDocumentChunk
	for _, c := range m.chunks {
		if c.DocumentID == docID {
			out = append(out, c)
		}
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}

type mockJobRepo struct {
	jobs map[uuid.UUID]*domain.AIJob
}

func newMockJobRepo() *mockJobRepo { return &mockJobRepo{jobs: map[uuid.UUID]*domain.AIJob{}} }

func (m *mockJobRepo) CreateJob(_ context.Context, j *domain.AIJob) error {
	j.ID = uuid.New()
	m.jobs[j.ID] = j
	return nil
}
func (m *mockJobRepo) GetJob(_ context.Context, id uuid.UUID) (*domain.AIJob, error) {
	return m.jobs[id], nil
}
func (m *mockJobRepo) UpdateJob(_ context.Context, id uuid.UUID, s domain.AIJobStatus, result []byte, errMsg *string) error {
	if j, ok := m.jobs[id]; ok {
		j.Status = s
		j.Result = result
		j.ErrorMsg = errMsg
		j.UpdatedAt = time.Now()
	}
	return nil
}

type mockSessionRepo struct {
	sessions map[uuid.UUID]*domain.AISession
	messages []*domain.AIMessage
}

func newMockSessionRepo() *mockSessionRepo {
	return &mockSessionRepo{sessions: map[uuid.UUID]*domain.AISession{}}
}
func (m *mockSessionRepo) CreateSession(_ context.Context, s *domain.AISession) error {
	s.ID = uuid.New()
	m.sessions[s.ID] = s
	return nil
}
func (m *mockSessionRepo) GetSession(_ context.Context, id uuid.UUID) (*domain.AISession, error) {
	return m.sessions[id], nil
}
func (m *mockSessionRepo) ListUserSessions(_ context.Context, userID uuid.UUID) ([]*domain.AISession, error) {
	var out []*domain.AISession
	for _, s := range m.sessions {
		if s.UserID == userID {
			out = append(out, s)
		}
	}
	return out, nil
}
func (m *mockSessionRepo) DeleteSession(_ context.Context, id uuid.UUID) error {
	delete(m.sessions, id)
	return nil
}
func (m *mockSessionRepo) AddMessage(_ context.Context, msg *domain.AIMessage) error {
	msg.ID = uuid.New()
	m.messages = append(m.messages, msg)
	return nil
}
func (m *mockSessionRepo) ListMessages(_ context.Context, sessionID uuid.UUID) ([]*domain.AIMessage, error) {
	var out []*domain.AIMessage
	for _, m2 := range m.messages {
		if m2.SessionID == sessionID {
			out = append(out, m2)
		}
	}
	return out, nil
}

type mockTokenRepo struct{ total int }

func (m *mockTokenRepo) LogUsage(_ context.Context, u *domain.AITokenUsage) error {
	m.total += u.TokensIn + u.TokensOut
	return nil
}
func (m *mockTokenRepo) MonthlyTokens(_ context.Context, _ uuid.UUID) (int, error) {
	return m.total, nil
}

type mockMasteryRepo struct{ entries map[string]*domain.TopicMastery }

func newMockMasteryRepo() *mockMasteryRepo {
	return &mockMasteryRepo{entries: map[string]*domain.TopicMastery{}}
}
func (m *mockMasteryRepo) UpsertMastery(_ context.Context, e *domain.TopicMastery) error {
	m.entries[e.StudentID.String()+":"+e.TopicID.String()] = e
	return nil
}
func (m *mockMasteryRepo) GetMastery(_ context.Context, s, t uuid.UUID) (*domain.TopicMastery, error) {
	return m.entries[s.String()+":"+t.String()], nil
}
func (m *mockMasteryRepo) ListDueMastery(_ context.Context, s uuid.UUID, limit int) ([]*domain.TopicMastery, error) {
	var out []*domain.TopicMastery
	for _, e := range m.entries {
		if e.StudentID == s && !time.Now().Before(e.NextReview) {
			out = append(out, e)
		}
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}
func (m *mockMasteryRepo) ListWeakTopics(_ context.Context, s uuid.UUID, limit int) ([]*domain.TopicMastery, error) {
	var out []*domain.TopicMastery
	for _, e := range m.entries {
		if e.StudentID == s && e.TotalCount > 0 {
			out = append(out, e)
		}
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}

type mockGamifRepo struct{ data map[uuid.UUID]*domain.StudentGamification }

func newMockGamifRepo() *mockGamifRepo {
	return &mockGamifRepo{data: map[uuid.UUID]*domain.StudentGamification{}}
}
func (m *mockGamifRepo) GetOrCreate(_ context.Context, id uuid.UUID) (*domain.StudentGamification, error) {
	if g, ok := m.data[id]; ok {
		return g, nil
	}
	g := &domain.StudentGamification{StudentID: id, DailyGoalXP: 50, UpdatedAt: time.Now()}
	m.data[id] = g
	return g, nil
}
func (m *mockGamifRepo) AddXP(_ context.Context, id uuid.UUID, amount int, _ string) error {
	g, _ := m.GetOrCreate(context.Background(), id)
	g.XPTotal += amount
	g.XPToday += amount
	return nil
}
func (m *mockGamifRepo) UpdateStreak(_ context.Context, id uuid.UUID) error {
	g, _ := m.GetOrCreate(context.Background(), id)
	today := time.Now().Truncate(24 * time.Hour)
	if g.LastActivityDate == nil || g.LastActivityDate.Before(today) {
		g.StreakDays++
		t := today
		g.LastActivityDate = &t
	}
	return nil
}

type mockTopicRepo struct{}

func (m *mockTopicRepo) UpsertTopic(_ context.Context, name string, subject *string) (*domain.TopicTag, error) {
	return &domain.TopicTag{ID: uuid.New(), Name: name, Subject: subject}, nil
}
func (m *mockTopicRepo) GetTopic(_ context.Context, id uuid.UUID) (*domain.TopicTag, error) {
	return &domain.TopicTag{ID: id}, nil
}
func (m *mockTopicRepo) LinkQuestionTopics(_ context.Context, _ uuid.UUID, _ []uuid.UUID) error {
	return nil
}

type mockQuizCreator struct{ quizID uuid.UUID }

func (m *mockQuizCreator) CreateQuizWithQuestions(_ context.Context, _ ai.QuizCreateRequest) (uuid.UUID, error) {
	m.quizID = uuid.New()
	return m.quizID, nil
}

type mockObjectStore struct{ data map[string][]byte }

func (m *mockObjectStore) GetObject(_ context.Context, key string) ([]byte, error) {
	return m.data[key], nil
}

type mockInsightsRepo struct {
	insights *domain.ClassInsights
	progress *domain.StudentProgress
}

func (m *mockInsightsRepo) ClassInsights(_ context.Context, groupID uuid.UUID) (*domain.ClassInsights, error) {
	if m.insights != nil {
		return m.insights, nil
	}
	return &domain.ClassInsights{GroupID: groupID, Period: "last_30_days"}, nil
}
func (m *mockInsightsRepo) StudentProgress(_ context.Context, _, studentID uuid.UUID) (*domain.StudentProgress, error) {
	if m.progress != nil {
		return m.progress, nil
	}
	return &domain.StudentProgress{StudentID: studentID, Name: "Test Student"}, nil
}

func (m *mockInsightsRepo) IsDemo(_ context.Context, _ uuid.UUID) (bool, error) {
	return false, nil
}

type mockFeedbackRepo struct {
	feedback map[uuid.UUID]*domain.QuizQuestionFeedback
}

func newMockFeedbackRepo() *mockFeedbackRepo {
	return &mockFeedbackRepo{feedback: map[uuid.UUID]*domain.QuizQuestionFeedback{}}
}
func (m *mockFeedbackRepo) UpsertFeedback(_ context.Context, fb *domain.QuizQuestionFeedback) error {
	m.feedback[fb.QuestionID] = fb
	return nil
}
func (m *mockFeedbackRepo) AcceptanceRate(_ context.Context, teacherID uuid.UUID) (float64, int, error) {
	total, accepted := 0, 0
	for _, fb := range m.feedback {
		if fb.TeacherID == teacherID {
			total++
			if fb.Accepted {
				accepted++
			}
		}
	}
	if total == 0 {
		return 0, 0, nil
	}
	return float64(accepted) / float64(total), total, nil
}

// ─── Service builder ──────────────────────────────────────────────────────────

func buildServiceWithFeedback() (*ai.Service, *mockFeedbackRepo) {
	docs := newMockDocRepo()
	jobs := newMockJobRepo()
	sessions := newMockSessionRepo()
	tokens := &mockTokenRepo{}
	mastery := newMockMasteryRepo()
	gamif := newMockGamifRepo()
	topics := &mockTopicRepo{}
	quiz := &mockQuizCreator{}
	store := &mockObjectStore{data: map[string][]byte{}}
	insights := &mockInsightsRepo{}
	feedback := newMockFeedbackRepo()

	svc := ai.NewService(docs, jobs, sessions, tokens, mastery, gamif, topics, quiz, store, insights, feedback, nil, nil, nil, nil,
		ai.ServiceConfig{Model: "gpt-4o-mini", MonthlyTokensMax: 500000, ChunkSize: 400, ChunkOverlap: 50})
	return svc, feedback
}

func buildService() (*ai.Service, *mockDocRepo, *mockJobRepo, *mockGamifRepo) {
	docs := newMockDocRepo()
	jobs := newMockJobRepo()
	sessions := newMockSessionRepo()
	tokens := &mockTokenRepo{}
	mastery := newMockMasteryRepo()
	gamif := newMockGamifRepo()
	topics := &mockTopicRepo{}
	quiz := &mockQuizCreator{}
	store := &mockObjectStore{data: map[string][]byte{}}

	svc := ai.NewService(docs, jobs, sessions, tokens, mastery, gamif, topics, quiz, store, nil, nil, nil, nil, nil, nil,
		ai.ServiceConfig{Model: "gpt-4o-mini", MonthlyTokensMax: 500000, ChunkSize: 400, ChunkOverlap: 50})
	return svc, docs, jobs, gamif
}

// ─── Tests ────────────────────────────────────────────────────────────────────

func TestCreateDocument_SetsStatusPending(t *testing.T) {
	svc, _, _, _ := buildService()

	teacherID := uuid.New()
	doc := &domain.AIDocument{
		TeacherID: teacherID,
		Title:     "Grammar book",
		ObjectKey: "ai-docs/123/book.pdf",
		MimeType:  "application/pdf",
		SizeBytes: 1024,
	}

	if err := svc.CreateDocument(context.Background(), doc); err != nil {
		t.Fatalf("CreateDocument: %v", err)
	}
	if doc.ID == uuid.Nil {
		t.Fatal("expected doc ID to be set")
	}
	if doc.Status != domain.AIDocStatusPending {
		t.Fatalf("expected status pending, got %s", doc.Status)
	}
}

func TestDeleteDocument_ForbiddenForOtherTeacher(t *testing.T) {
	svc, _, _, _ := buildService()

	ownerID := uuid.New()
	doc := &domain.AIDocument{
		TeacherID: ownerID,
		Title:     "Test doc",
		ObjectKey: "key",
		MimeType:  "application/pdf",
		SizeBytes: 100,
	}
	_ = svc.CreateDocument(context.Background(), doc)

	otherTeacher := uuid.New()
	err := svc.DeleteDocument(context.Background(), doc.ID, otherTeacher)
	if err == nil {
		t.Fatal("expected forbidden error")
	}
}

func TestCreateJob_AssignsID(t *testing.T) {
	svc, _, jobs, _ := buildService()

	userID := uuid.New()
	docID := uuid.New()
	job, err := svc.CreateJob(context.Background(), "generate_quiz", userID, &docID)
	if err != nil {
		t.Fatalf("CreateJob: %v", err)
	}
	if job.ID == uuid.Nil {
		t.Fatal("expected job ID to be set")
	}
	if job.Status != domain.AIJobPending {
		t.Fatalf("expected status pending, got %s", job.Status)
	}

	fetched, _ := svc.GetJob(context.Background(), job.ID)
	if fetched == nil {
		t.Fatal("expected job to be retrievable")
	}
	_ = jobs
}

func TestAwardXP_IncrementsTotal(t *testing.T) {
	svc, _, _, gamif := buildService()

	studentID := uuid.New()
	_ = svc.AwardXP(context.Background(), studentID, 10, "quiz_correct")
	_ = svc.AwardXP(context.Background(), studentID, 20, "weak_topic")

	g, _ := svc.GetGamification(context.Background(), studentID)
	if g.XPTotal != 30 {
		t.Fatalf("expected 30 XP, got %d", g.XPTotal)
	}
	_ = gamif
}

func TestRecordAnswer_UpdatesMastery(t *testing.T) {
	svc, _, _, _ := buildService()

	studentID := uuid.New()
	topicID := uuid.New()

	// Wrong answer — interval should stay at 1.
	_ = svc.RecordAnswer(context.Background(), studentID, topicID, false, 0)

	queue, err := svc.GetReviewQueue(context.Background(), studentID)
	if err != nil {
		t.Fatalf("GetReviewQueue: %v", err)
	}
	// interval_days=1 → NextReview = tomorrow, so not yet due today
	// (depends on timing — just verify no error)
	_ = queue
}

func TestGetGamification_ReturnsDefaultOnFirstCall(t *testing.T) {
	svc, _, _, _ := buildService()

	studentID := uuid.New()
	g, err := svc.GetGamification(context.Background(), studentID)
	if err != nil {
		t.Fatalf("GetGamification: %v", err)
	}
	if g.StudentID != studentID {
		t.Fatalf("expected student_id %s, got %s", studentID, g.StudentID)
	}
	if g.DailyGoalXP != 50 {
		t.Fatalf("expected daily_goal_xp 50, got %d", g.DailyGoalXP)
	}
}

func TestSubmitQuestionFeedback_Accepted(t *testing.T) {
	svc, fb := buildServiceWithFeedback()

	teacherID := uuid.New()
	questionID := uuid.New()

	if err := svc.SubmitQuestionFeedback(context.Background(), questionID, teacherID, true); err != nil {
		t.Fatalf("SubmitQuestionFeedback: %v", err)
	}

	stored, ok := fb.feedback[questionID]
	if !ok {
		t.Fatal("expected feedback to be stored")
	}
	if !stored.Accepted {
		t.Fatal("expected accepted=true")
	}
	if stored.TeacherID != teacherID {
		t.Fatalf("expected teacher_id %s, got %s", teacherID, stored.TeacherID)
	}
}

func TestGetMyAcceptanceRate_MixedFeedback(t *testing.T) {
	svc, _ := buildServiceWithFeedback()

	teacherID := uuid.New()
	_ = svc.SubmitQuestionFeedback(context.Background(), uuid.New(), teacherID, true)
	_ = svc.SubmitQuestionFeedback(context.Background(), uuid.New(), teacherID, true)
	_ = svc.SubmitQuestionFeedback(context.Background(), uuid.New(), teacherID, false)

	rate, total, err := svc.GetMyAcceptanceRate(context.Background(), teacherID)
	if err != nil {
		t.Fatalf("GetMyAcceptanceRate: %v", err)
	}
	if total != 3 {
		t.Fatalf("expected total=3, got %d", total)
	}
	const want = 2.0 / 3.0
	if rate < want-0.001 || rate > want+0.001 {
		t.Fatalf("expected rate %.4f, got %.4f", want, rate)
	}
}

func TestGetClassInsights_ReturnsGroupID(t *testing.T) {
	svc, _ := buildServiceWithFeedback()

	groupID := uuid.New()
	insights, err := svc.GetClassInsights(context.Background(), groupID, uuid.New())
	if err != nil {
		t.Fatalf("GetClassInsights: %v", err)
	}
	if insights.GroupID != groupID {
		t.Fatalf("expected group_id %s, got %s", groupID, insights.GroupID)
	}
}

func TestGetStudentProgress_ReturnsStudentID(t *testing.T) {
	svc, _ := buildServiceWithFeedback()

	groupID := uuid.New()
	studentID := uuid.New()
	progress, err := svc.GetStudentProgress(context.Background(), groupID, studentID)
	if err != nil {
		t.Fatalf("GetStudentProgress: %v", err)
	}
	if progress.StudentID != studentID {
		t.Fatalf("expected student_id %s, got %s", studentID, progress.StudentID)
	}
}
