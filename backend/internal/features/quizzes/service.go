package quizzes

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

type Service struct {
	repo   Repository
	groups GroupChecker
}

func NewService(repo Repository, groups GroupChecker) *Service {
	return &Service{repo: repo, groups: groups}
}

// CreateQuiz creates a new quiz (teacher only, unpublished by default).
func (s *Service) CreateQuiz(ctx context.Context, teacherID, groupID uuid.UUID, req CreateQuizRequest) (*domain.Quiz, error) {
	g, err := s.groups.GetGroupByID(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("quizzes.CreateQuiz fetch group: %w", err)
	}
	if g == nil {
		return nil, domain.ErrNotFound
	}
	if g.TeacherID != teacherID {
		return nil, domain.ErrForbidden
	}

	maxAttempts := req.MaxAttempts
	if maxAttempts == 0 {
		maxAttempts = 1
	}

	now := time.Now()
	q := &domain.Quiz{
		ID:          uuid.New(),
		GroupID:     groupID,
		TeacherID:   teacherID,
		Title:       strings.TrimSpace(req.Title),
		Description: req.Description,
		TimeLimit:   req.TimeLimit,
		MaxAttempts: maxAttempts,
		OpenAt:      req.OpenAt,
		CloseAt:     req.CloseAt,
		IsPublished: false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.CreateQuiz(ctx, q); err != nil {
		return nil, fmt.Errorf("quizzes.CreateQuiz: %w", err)
	}
	return q, nil
}

// ListGroupQuizzes returns all quizzes for a group.
// Students only see published quizzes that are currently open.
func (s *Service) ListGroupQuizzes(ctx context.Context, groupID, callerID uuid.UUID, role domain.Role) ([]*domain.Quiz, error) {
	if err := s.assertVisible(ctx, groupID, callerID, role); err != nil {
		return nil, err
	}

	qs, err := s.repo.ListGroupQuizzes(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("quizzes.ListGroupQuizzes: %w", err)
	}

	if role == domain.RoleTeacher {
		return qs, nil
	}

	// Students see only published quizzes.
	now := time.Now()
	var visible []*domain.Quiz
	for _, q := range qs {
		if !q.IsPublished {
			continue
		}
		if q.OpenAt != nil && now.Before(*q.OpenAt) {
			continue
		}
		if q.CloseAt != nil && now.After(*q.CloseAt) {
			continue
		}
		visible = append(visible, q)
	}
	return visible, nil
}

// GetQuiz returns quiz detail including questions.
// For students: options are returned without IsCorrect flag.
func (s *Service) GetQuiz(ctx context.Context, quizID, callerID uuid.UUID, role domain.Role) (*domain.Quiz, []*domain.Question, map[uuid.UUID][]*domain.Option, error) {
	q, err := s.repo.GetQuizByID(ctx, quizID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("quizzes.GetQuiz: %w", err)
	}
	if q == nil {
		return nil, nil, nil, domain.ErrNotFound
	}

	if err := s.assertVisible(ctx, q.GroupID, callerID, role); err != nil {
		return nil, nil, nil, err
	}
	if role != domain.RoleTeacher && !q.IsPublished {
		return nil, nil, nil, domain.ErrForbidden
	}

	questions, err := s.repo.ListQuestions(ctx, quizID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("quizzes.GetQuiz questions: %w", err)
	}

	optMap := make(map[uuid.UUID][]*domain.Option, len(questions))
	for _, qst := range questions {
		opts, err := s.repo.ListOptions(ctx, qst.ID)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("quizzes.GetQuiz options: %w", err)
		}
		optMap[qst.ID] = opts
	}
	return q, questions, optMap, nil
}

// AddQuestion adds a question to a quiz (teacher, quiz must be unpublished).
func (s *Service) AddQuestion(ctx context.Context, quizID, teacherID uuid.UUID, req CreateQuestionRequest) (*domain.Question, error) {
	q, err := s.repo.GetQuizByID(ctx, quizID)
	if err != nil {
		return nil, fmt.Errorf("quizzes.AddQuestion fetch: %w", err)
	}
	if q == nil {
		return nil, domain.ErrNotFound
	}
	if q.TeacherID != teacherID {
		return nil, domain.ErrForbidden
	}
	if q.IsPublished {
		return nil, domain.ErrConflict // can't edit published quiz
	}

	points := req.Points
	if points == 0 {
		points = 1
	}

	qst := &domain.Question{
		ID:          uuid.New(),
		QuizID:      quizID,
		Body:        strings.TrimSpace(req.Body),
		Explanation: req.Explanation,
		Position:    req.Position,
		Points:      points,
	}
	if err := s.repo.CreateQuestion(ctx, qst); err != nil {
		return nil, fmt.Errorf("quizzes.AddQuestion: %w", err)
	}
	return qst, nil
}

// AddOption adds an answer option to a question (teacher, quiz must be unpublished).
func (s *Service) AddOption(ctx context.Context, questionID, teacherID uuid.UUID, req CreateOptionRequest) (*domain.Option, error) {
	qst, err := s.repo.GetQuestion(ctx, questionID)
	if err != nil {
		return nil, fmt.Errorf("quizzes.AddOption fetch question: %w", err)
	}
	if qst == nil {
		return nil, domain.ErrNotFound
	}

	quiz, err := s.repo.GetQuizByID(ctx, qst.QuizID)
	if err != nil {
		return nil, fmt.Errorf("quizzes.AddOption fetch quiz: %w", err)
	}
	if quiz == nil || quiz.TeacherID != teacherID {
		return nil, domain.ErrForbidden
	}
	if quiz.IsPublished {
		return nil, domain.ErrConflict
	}

	o := &domain.Option{
		ID:         uuid.New(),
		QuestionID: questionID,
		Body:       strings.TrimSpace(req.Body),
		IsCorrect:  req.IsCorrect,
		Position:   req.Position,
	}
	if err := s.repo.CreateOption(ctx, o); err != nil {
		return nil, fmt.Errorf("quizzes.AddOption: %w", err)
	}
	return o, nil
}

// Publish toggles published state (teacher only).
func (s *Service) Publish(ctx context.Context, quizID, teacherID uuid.UUID, published bool) (*domain.Quiz, error) {
	q, err := s.repo.GetQuizByID(ctx, quizID)
	if err != nil {
		return nil, fmt.Errorf("quizzes.Publish fetch: %w", err)
	}
	if q == nil {
		return nil, domain.ErrNotFound
	}
	if q.TeacherID != teacherID {
		return nil, domain.ErrForbidden
	}

	if err := s.repo.PublishQuiz(ctx, quizID, published); err != nil {
		return nil, fmt.Errorf("quizzes.Publish: %w", err)
	}
	q.IsPublished = published
	return q, nil
}

// StartAttempt begins a new quiz attempt for a student.
func (s *Service) StartAttempt(ctx context.Context, quizID, studentID uuid.UUID) (*domain.QuizAttempt, error) {
	q, err := s.repo.GetQuizByID(ctx, quizID)
	if err != nil {
		return nil, fmt.Errorf("quizzes.StartAttempt fetch: %w", err)
	}
	if q == nil || !q.IsPublished {
		return nil, domain.ErrNotFound
	}

	m, err := s.groups.GetMember(ctx, q.GroupID, studentID)
	if err != nil {
		return nil, fmt.Errorf("quizzes.StartAttempt member: %w", err)
	}
	if m == nil {
		return nil, domain.ErrForbidden
	}

	count, err := s.repo.CountAttempts(ctx, quizID, studentID)
	if err != nil {
		return nil, fmt.Errorf("quizzes.StartAttempt count: %w", err)
	}
	if int16(count) >= q.MaxAttempts {
		return nil, domain.ErrConflict // attempts exhausted
	}

	// Compute max score from questions.
	questions, err := s.repo.ListQuestions(ctx, quizID)
	if err != nil {
		return nil, fmt.Errorf("quizzes.StartAttempt questions: %w", err)
	}
	var maxScore int16
	for _, qst := range questions {
		maxScore += qst.Points
	}

	a := &domain.QuizAttempt{
		ID:        uuid.New(),
		QuizID:    quizID,
		StudentID: studentID,
		StartedAt: time.Now(),
		MaxScore:  maxScore,
	}
	if err := s.repo.CreateAttempt(ctx, a); err != nil {
		return nil, fmt.Errorf("quizzes.StartAttempt: %w", err)
	}
	return a, nil
}

// SubmitAttempt grades the attempt and saves student answers.
func (s *Service) SubmitAttempt(ctx context.Context, attemptID, studentID uuid.UUID, req SubmitAnswersRequest) (*SubmitAttemptResult, error) {
	attempt, err := s.repo.GetAttempt(ctx, attemptID)
	if err != nil {
		return nil, fmt.Errorf("quizzes.SubmitAttempt fetch: %w", err)
	}
	if attempt == nil {
		return nil, domain.ErrNotFound
	}
	if attempt.StudentID != studentID {
		return nil, domain.ErrForbidden
	}
	if attempt.FinishedAt != nil {
		return nil, domain.ErrConflict // already submitted
	}

	questions, err := s.repo.ListQuestions(ctx, attempt.QuizID)
	if err != nil {
		return nil, fmt.Errorf("quizzes.SubmitAttempt questions: %w", err)
	}

	var score int16
	var answers []*domain.StudentAnswer
	var resultItems []QuizResultItem

	for _, qst := range questions {
		opts, err := s.repo.ListOptions(ctx, qst.ID)
		if err != nil {
			return nil, fmt.Errorf("quizzes.SubmitAttempt options: %w", err)
		}

		item := QuizResultItem{
			QuestionID:   qst.ID,
			QuestionBody: qst.Body,
			Points:       qst.Points,
			Explanation:  qst.Explanation,
		}
		// find correct option
		for _, o := range opts {
			if o.IsCorrect {
				cid := o.ID
				item.CorrectOptionID = &cid
				item.CorrectBody = o.Body
				break
			}
		}

		// apply student answer if present
		if chosenIDStr, ok := req.Answers[qst.ID.String()]; ok {
			chosenID, parseErr := uuid.Parse(chosenIDStr)
			if parseErr != nil {
				fmt.Printf("SubmitAttempt: invalid option UUID %q for question %s: %v\n", chosenIDStr, qst.ID, parseErr)
			}
			if parseErr == nil {
				for _, o := range opts {
					if o.ID == chosenID {
						sid := o.ID
						sbody := o.Body
						item.SelectedOptionID = &sid
						item.SelectedBody = &sbody
						item.IsCorrect = o.IsCorrect
						if o.IsCorrect {
							score += qst.Points
						}
						answers = append(answers, &domain.StudentAnswer{
							AttemptID:  attemptID,
							QuestionID: qst.ID,
							OptionID:   chosenID,
						})
						break
					}
				}
			}
		}

		resultItems = append(resultItems, item)
	}

	if len(answers) > 0 {
		if err := s.repo.SaveAnswers(ctx, answers); err != nil {
			return nil, fmt.Errorf("quizzes.SubmitAttempt save answers: %w", err)
		}
	}

	if err := s.repo.FinishAttempt(ctx, attemptID, score, attempt.MaxScore); err != nil {
		return nil, fmt.Errorf("quizzes.SubmitAttempt finish: %w", err)
	}

	attempt.Score = &score
	now := time.Now()
	attempt.FinishedAt = &now
	return &SubmitAttemptResult{Attempt: attempt, QuestionResults: resultItems}, nil
}

func (s *Service) assertVisible(ctx context.Context, groupID, callerID uuid.UUID, role domain.Role) error {
	g, err := s.groups.GetGroupByID(ctx, groupID)
	if err != nil {
		return fmt.Errorf("assertVisible: %w", err)
	}
	if g == nil {
		return domain.ErrNotFound
	}
	if role == domain.RoleTeacher {
		if g.TeacherID != callerID {
			return domain.ErrForbidden
		}
		return nil
	}
	m, err := s.groups.GetMember(ctx, groupID, callerID)
	if err != nil {
		return fmt.Errorf("assertVisible: %w", err)
	}
	if m == nil {
		return domain.ErrForbidden
	}
	return nil
}
