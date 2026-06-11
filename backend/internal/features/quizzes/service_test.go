package quizzes_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/features/quizzes"
)

// --- in-memory repo ---

type memRepo struct {
	quizzes   map[uuid.UUID]*domain.Quiz
	questions map[uuid.UUID]*domain.Question
	options   map[uuid.UUID]*domain.Option
	attempts  map[uuid.UUID]*domain.QuizAttempt
	answers   map[uuid.UUID][]*domain.StudentAnswer
}

func newMemRepo() *memRepo {
	return &memRepo{
		quizzes:   make(map[uuid.UUID]*domain.Quiz),
		questions: make(map[uuid.UUID]*domain.Question),
		options:   make(map[uuid.UUID]*domain.Option),
		attempts:  make(map[uuid.UUID]*domain.QuizAttempt),
		answers:   make(map[uuid.UUID][]*domain.StudentAnswer),
	}
}

func (r *memRepo) CreateQuiz(_ context.Context, q *domain.Quiz) error {
	cp := *q; r.quizzes[q.ID] = &cp; return nil
}
func (r *memRepo) GetQuizByID(_ context.Context, id uuid.UUID) (*domain.Quiz, error) {
	q, ok := r.quizzes[id]; if !ok || q.DeletedAt != nil { return nil, nil }; cp := *q; return &cp, nil
}
func (r *memRepo) ListGroupQuizzes(_ context.Context, groupID uuid.UUID) ([]*domain.Quiz, error) {
	var out []*domain.Quiz
	for _, q := range r.quizzes {
		if q.GroupID == groupID && q.DeletedAt == nil { cp := *q; out = append(out, &cp) }
	}
	return out, nil
}
func (r *memRepo) UpdateQuiz(_ context.Context, q *domain.Quiz) error { cp := *q; r.quizzes[q.ID] = &cp; return nil }
func (r *memRepo) SoftDeleteQuiz(_ context.Context, id uuid.UUID) error {
	if q, ok := r.quizzes[id]; ok { now := time.Now(); q.DeletedAt = &now }; return nil
}
func (r *memRepo) PublishQuiz(_ context.Context, id uuid.UUID, published bool) error {
	if q, ok := r.quizzes[id]; ok { q.IsPublished = published }; return nil
}
func (r *memRepo) CreateQuestion(_ context.Context, q *domain.Question) error {
	cp := *q; r.questions[q.ID] = &cp; return nil
}
func (r *memRepo) GetQuestion(_ context.Context, id uuid.UUID) (*domain.Question, error) {
	q, ok := r.questions[id]; if !ok { return nil, nil }; cp := *q; return &cp, nil
}
func (r *memRepo) ListQuestions(_ context.Context, quizID uuid.UUID) ([]*domain.Question, error) {
	var out []*domain.Question
	for _, q := range r.questions { if q.QuizID == quizID { cp := *q; out = append(out, &cp) } }
	return out, nil
}
func (r *memRepo) UpdateQuestion(_ context.Context, q *domain.Question) error { cp := *q; r.questions[q.ID] = &cp; return nil }
func (r *memRepo) DeleteQuestion(_ context.Context, id uuid.UUID) error { delete(r.questions, id); return nil }

func (r *memRepo) CreateOption(_ context.Context, o *domain.Option) error {
	cp := *o; r.options[o.ID] = &cp; return nil
}
func (r *memRepo) ListOptions(_ context.Context, questionID uuid.UUID) ([]*domain.Option, error) {
	var out []*domain.Option
	for _, o := range r.options { if o.QuestionID == questionID { cp := *o; out = append(out, &cp) } }
	return out, nil
}
func (r *memRepo) DeleteOption(_ context.Context, id uuid.UUID) error { delete(r.options, id); return nil }

func (r *memRepo) CreateAttempt(_ context.Context, a *domain.QuizAttempt) error {
	cp := *a; r.attempts[a.ID] = &cp; return nil
}
func (r *memRepo) GetAttempt(_ context.Context, id uuid.UUID) (*domain.QuizAttempt, error) {
	a, ok := r.attempts[id]; if !ok { return nil, nil }; cp := *a; return &cp, nil
}
func (r *memRepo) CountAttempts(_ context.Context, quizID, studentID uuid.UUID) (int, error) {
	count := 0
	for _, a := range r.attempts {
		if a.QuizID == quizID && a.StudentID == studentID { count++ }
	}
	return count, nil
}
func (r *memRepo) FinishAttempt(_ context.Context, id uuid.UUID, score, maxScore int16) error {
	if a, ok := r.attempts[id]; ok {
		now := time.Now(); a.FinishedAt = &now; a.Score = &score; a.MaxScore = maxScore
	}
	return nil
}
func (r *memRepo) SaveAnswers(_ context.Context, ans []*domain.StudentAnswer) error {
	for _, a := range ans {
		cp := *a; r.answers[a.AttemptID] = append(r.answers[a.AttemptID], &cp)
	}
	return nil
}
func (r *memRepo) GetAnswers(_ context.Context, attemptID uuid.UUID) ([]*domain.StudentAnswer, error) {
	return r.answers[attemptID], nil
}

// --- group checker ---

type memGroups struct {
	groups  map[uuid.UUID]*domain.Group
	members map[string]struct{}
}

func newGroups(teacherID uuid.UUID) (*memGroups, uuid.UUID) {
	gid := uuid.New()
	return &memGroups{
		groups:  map[uuid.UUID]*domain.Group{gid: {ID: gid, TeacherID: teacherID}},
		members: make(map[string]struct{}),
	}, gid
}
func (mg *memGroups) addMember(gid, sid uuid.UUID) { mg.members[gid.String()+":"+sid.String()] = struct{}{} }
func (mg *memGroups) GetGroupByID(_ context.Context, id uuid.UUID) (*domain.Group, error) {
	g, ok := mg.groups[id]; if !ok { return nil, nil }; cp := *g; return &cp, nil
}
func (mg *memGroups) GetMember(_ context.Context, gid, uid uuid.UUID) (*domain.GroupMember, error) {
	_, ok := mg.members[gid.String()+":"+uid.String()]
	if !ok { return nil, nil }
	return &domain.GroupMember{GroupID: gid, StudentID: uid}, nil
}

// --- helpers ---

var (
	teacherID = uuid.New()
	studentID = uuid.New()
)

func buildSvc() (*quizzes.Service, *memGroups, uuid.UUID) {
	mg, gid := newGroups(teacherID)
	mg.addMember(gid, studentID)
	return quizzes.NewService(newMemRepo(), mg), mg, gid
}

func addQuestionWithCorrectOption(t *testing.T, svc *quizzes.Service, quizID uuid.UUID) (uuid.UUID, uuid.UUID) {
	t.Helper()
	qst, err := svc.AddQuestion(context.Background(), quizID, teacherID, quizzes.CreateQuestionRequest{
		Body: "2+2=?", Points: 1,
	})
	require.NoError(t, err)
	opt, err := svc.AddOption(context.Background(), qst.ID, teacherID, quizzes.CreateOptionRequest{
		Body: "4", IsCorrect: true,
	})
	require.NoError(t, err)
	_, _ = svc.AddOption(context.Background(), qst.ID, teacherID, quizzes.CreateOptionRequest{
		Body: "3", IsCorrect: false,
	})
	return qst.ID, opt.ID
}

// --- tests ---

func TestCreateQuiz_Success(t *testing.T) {
	svc, _, gid := buildSvc()
	q, err := svc.CreateQuiz(context.Background(), teacherID, gid,
		quizzes.CreateQuizRequest{Title: "Тест 1", MaxAttempts: 1})
	require.NoError(t, err)
	assert.Equal(t, "Тест 1", q.Title)
	assert.False(t, q.IsPublished)
}

func TestCreateQuiz_ForbiddenForNonOwner(t *testing.T) {
	svc, _, gid := buildSvc()
	_, err := svc.CreateQuiz(context.Background(), uuid.New(), gid,
		quizzes.CreateQuizRequest{Title: "hack", MaxAttempts: 1})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestPublish_TeacherOnly(t *testing.T) {
	svc, _, gid := buildSvc()
	q, _ := svc.CreateQuiz(context.Background(), teacherID, gid,
		quizzes.CreateQuizRequest{Title: "Q", MaxAttempts: 1})
	updated, err := svc.Publish(context.Background(), q.ID, teacherID, true)
	require.NoError(t, err)
	assert.True(t, updated.IsPublished)
}

func TestStartAttempt_StudentSuccess(t *testing.T) {
	svc, _, gid := buildSvc()
	q, _ := svc.CreateQuiz(context.Background(), teacherID, gid,
		quizzes.CreateQuizRequest{Title: "Q", MaxAttempts: 2})
	addQuestionWithCorrectOption(t, svc, q.ID)
	_, _ = svc.Publish(context.Background(), q.ID, teacherID, true)

	attempt, err := svc.StartAttempt(context.Background(), q.ID, studentID)
	require.NoError(t, err)
	assert.Equal(t, q.ID, attempt.QuizID)
	assert.Equal(t, int16(1), attempt.MaxScore)
}

func TestStartAttempt_NonMemberForbidden(t *testing.T) {
	svc, _, gid := buildSvc()
	q, _ := svc.CreateQuiz(context.Background(), teacherID, gid,
		quizzes.CreateQuizRequest{Title: "Q", MaxAttempts: 1})
	_, _ = svc.Publish(context.Background(), q.ID, teacherID, true)

	_, err := svc.StartAttempt(context.Background(), q.ID, uuid.New())
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestStartAttempt_AttemptsExhausted(t *testing.T) {
	svc, _, gid := buildSvc()
	q, _ := svc.CreateQuiz(context.Background(), teacherID, gid,
		quizzes.CreateQuizRequest{Title: "Q", MaxAttempts: 1})
	_, _ = svc.Publish(context.Background(), q.ID, teacherID, true)

	_, _ = svc.StartAttempt(context.Background(), q.ID, studentID)
	_, err := svc.StartAttempt(context.Background(), q.ID, studentID)
	assert.ErrorIs(t, err, domain.ErrConflict)
}

func TestSubmitAttempt_CorrectAnswer(t *testing.T) {
	svc, _, gid := buildSvc()
	q, _ := svc.CreateQuiz(context.Background(), teacherID, gid,
		quizzes.CreateQuizRequest{Title: "Math", MaxAttempts: 1})
	qstID, optID := addQuestionWithCorrectOption(t, svc, q.ID)
	_, _ = svc.Publish(context.Background(), q.ID, teacherID, true)

	attempt, _ := svc.StartAttempt(context.Background(), q.ID, studentID)
	result, err := svc.SubmitAttempt(context.Background(), attempt.ID, studentID, quizzes.SubmitAnswersRequest{
		Answers: map[string]string{qstID.String(): optID.String()},
	})
	require.NoError(t, err)
	require.NotNil(t, result.Score)
	assert.Equal(t, int16(1), *result.Score)
	assert.NotNil(t, result.FinishedAt)
}

func TestSubmitAttempt_WrongAnswer(t *testing.T) {
	svc, _, gid := buildSvc()
	q, _ := svc.CreateQuiz(context.Background(), teacherID, gid,
		quizzes.CreateQuizRequest{Title: "Math", MaxAttempts: 1})
	qstID, _ := addQuestionWithCorrectOption(t, svc, q.ID)
	_, _ = svc.Publish(context.Background(), q.ID, teacherID, true)

	attempt, _ := svc.StartAttempt(context.Background(), q.ID, studentID)

	// Submit wrong option (random UUID that won't match any option).
	result, err := svc.SubmitAttempt(context.Background(), attempt.ID, studentID, quizzes.SubmitAnswersRequest{
		Answers: map[string]string{qstID.String(): uuid.New().String()},
	})
	require.NoError(t, err)
	require.NotNil(t, result.Score)
	assert.Equal(t, int16(0), *result.Score)
}

func TestAddQuestion_PublishedQuizForbidden(t *testing.T) {
	svc, _, gid := buildSvc()
	q, _ := svc.CreateQuiz(context.Background(), teacherID, gid,
		quizzes.CreateQuizRequest{Title: "Q", MaxAttempts: 1})
	_, _ = svc.Publish(context.Background(), q.ID, teacherID, true)

	_, err := svc.AddQuestion(context.Background(), q.ID, teacherID,
		quizzes.CreateQuestionRequest{Body: "X", Points: 1})
	assert.ErrorIs(t, err, domain.ErrConflict)
}

func TestListGroupQuizzes_StudentOnlySeesPublished(t *testing.T) {
	svc, _, gid := buildSvc()
	q1, _ := svc.CreateQuiz(context.Background(), teacherID, gid,
		quizzes.CreateQuizRequest{Title: "Published", MaxAttempts: 1})
	_, _ = svc.CreateQuiz(context.Background(), teacherID, gid,
		quizzes.CreateQuizRequest{Title: "Draft", MaxAttempts: 1})
	_, _ = svc.Publish(context.Background(), q1.ID, teacherID, true)

	visible, err := svc.ListGroupQuizzes(context.Background(), gid, studentID, domain.RoleStudent)
	require.NoError(t, err)
	assert.Len(t, visible, 1)
	assert.Equal(t, "Published", visible[0].Title)
}
