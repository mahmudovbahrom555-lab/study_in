package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/features/ai/prompts"
	infpdf "github.com/mahmudovbahrom555-lab/study_in/backend/internal/infrastructure/pdf"
	oai "github.com/mahmudovbahrom555-lab/study_in/backend/internal/infrastructure/openai"
	sm2 "github.com/mahmudovbahrom555-lab/study_in/backend/internal/infrastructure/spaced_repetition"
)

// Service is the AI feature orchestrator.
type Service struct {
	docs        DocumentRepository
	jobs        JobRepository
	sessions    SessionRepository
	tokens      TokenRepository
	mastery     MasteryRepository
	gamif       GamificationRepository
	topics      TopicRepository
	quiz        QuizCreator
	store       ObjectStore
	insights    InsightsRepository
	feedback    QuizFeedbackRepository
	genSess     GenerationSessionRepository
	recRepo     RecommendationRepository
	demoCreator DemoGroupCreator
	ai          *oai.Client
	cfg         ServiceConfig
}

type ServiceConfig struct {
	Model            string
	MonthlyTokensMax int
	ChunkSize        int
	ChunkOverlap     int
}

func NewService(
	docs DocumentRepository,
	jobs JobRepository,
	sessions SessionRepository,
	tokens TokenRepository,
	mastery MasteryRepository,
	gamif GamificationRepository,
	topics TopicRepository,
	quiz QuizCreator,
	store ObjectStore,
	insights InsightsRepository,
	feedback QuizFeedbackRepository,
	genSess GenerationSessionRepository,
	recRepo RecommendationRepository,
	demoCreator DemoGroupCreator,
	ai *oai.Client,
	cfg ServiceConfig,
) *Service {
	return &Service{
		docs: docs, jobs: jobs, sessions: sessions,
		tokens: tokens, mastery: mastery, gamif: gamif,
		topics: topics, quiz: quiz, store: store,
		insights: insights, feedback: feedback, genSess: genSess, recRepo: recRepo,
		demoCreator: demoCreator,
		ai: ai, cfg: cfg,
	}
}

// ─── Document management ──────────────────────────────────────────────────────

// CreateDocument persists a newly uploaded document record (status=pending).
func (s *Service) CreateDocument(ctx context.Context, doc *domain.AIDocument) error {
	doc.ID = uuid.New()
	doc.Status = domain.AIDocStatusPending
	doc.CreatedAt = time.Now()
	if err := s.docs.CreateDocument(ctx, doc); err != nil {
		return fmt.Errorf("ai.CreateDocument: %w", err)
	}
	return nil
}

func (s *Service) GetDocument(ctx context.Context, id uuid.UUID) (*domain.AIDocument, error) {
	doc, err := s.docs.GetDocument(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("ai.GetDocument: %w", err)
	}
	return doc, nil
}

func (s *Service) ListDocuments(ctx context.Context, teacherID uuid.UUID, groupID *uuid.UUID) ([]*domain.AIDocument, error) {
	return s.docs.ListDocuments(ctx, teacherID, groupID)
}

func (s *Service) DeleteDocument(ctx context.Context, docID, teacherID uuid.UUID) error {
	doc, err := s.docs.GetDocument(ctx, docID)
	if err != nil {
		return fmt.Errorf("ai.DeleteDocument: %w", err)
	}
	if doc == nil {
		return domain.ErrNotFound
	}
	if doc.TeacherID != teacherID {
		return domain.ErrForbidden
	}
	return s.docs.DeleteDocument(ctx, docID)
}

// ─── Document processing (called by Asynq worker) ────────────────────────────

// ProcessDocument downloads the PDF from MinIO, extracts text, chunks it,
// embeds each chunk, and stores the result in ai_document_chunks.
// Updates document status throughout.
func (s *Service) ProcessDocument(ctx context.Context, docID uuid.UUID) error {
	doc, err := s.docs.GetDocument(ctx, docID)
	if err != nil || doc == nil {
		return fmt.Errorf("ai.ProcessDocument fetch: %w", err)
	}

	_ = s.docs.UpdateDocumentStatus(ctx, docID, domain.AIDocStatusProcessing, 0, nil)

	data, err := s.store.GetObject(ctx, doc.ObjectKey)
	if err != nil {
		msg := err.Error()
		_ = s.docs.UpdateDocumentStatus(ctx, docID, domain.AIDocStatusFailed, 0, &msg)
		return fmt.Errorf("ai.ProcessDocument download: %w", err)
	}

	text, err := infpdf.ExtractTextFromBytes(data)
	if err != nil || strings.TrimSpace(text) == "" {
		msg := "PDF text extraction failed"
		if err != nil {
			msg = err.Error()
		}
		_ = s.docs.UpdateDocumentStatus(ctx, docID, domain.AIDocStatusFailed, 0, &msg)
		return fmt.Errorf("ai.ProcessDocument extract: %w", err)
	}

	chunks := infpdf.ChunkText(text, s.cfg.ChunkSize, s.cfg.ChunkOverlap)
	if len(chunks) == 0 {
		msg := "no text chunks extracted"
		_ = s.docs.UpdateDocumentStatus(ctx, docID, domain.AIDocStatusFailed, 0, &msg)
		return fmt.Errorf("ai.ProcessDocument: no chunks")
	}

	// Embed in batches of 20 to stay within API limits.
	const batchSize = 20
	totalTokens := 0
	count := 0
	for i := 0; i < len(chunks); i += batchSize {
		end := i + batchSize
		if end > len(chunks) {
			end = len(chunks)
		}
		batch := chunks[i:end]

		if s.ai.IsConfigured() {
			embeddings, batchTokens, embErr := s.ai.EmbedBatch(ctx, batch)
			if embErr != nil {
				// Log but continue — store chunks without embeddings so text
				// search still works.
				fmt.Printf("embed batch %d: %v\n", i, embErr)
				embeddings = make([][]float32, len(batch))
			}
			totalTokens += batchTokens
			for j, chunk := range batch {
				c := &domain.AIDocumentChunk{
					ID:         uuid.New(),
					DocumentID: docID,
					ChunkIndex: i + j,
					Content:    chunk,
					TokenCount: len(strings.Fields(chunk)), // approx
					CreatedAt:  time.Now(),
				}
				var emb []float32
				if embeddings != nil && j < len(embeddings) {
					emb = embeddings[j]
				}
				if sErr := s.docs.SaveChunk(ctx, c, emb); sErr != nil {
					return fmt.Errorf("ai.ProcessDocument save chunk: %w", sErr)
				}
				count++
			}
		} else {
			// No API key — store chunks without embeddings (text-only search).
			for j, chunk := range batch {
				c := &domain.AIDocumentChunk{
					ID:         uuid.New(),
					DocumentID: docID,
					ChunkIndex: i + j,
					Content:    chunk,
					TokenCount: len(strings.Fields(chunk)),
					CreatedAt:  time.Now(),
				}
				if sErr := s.docs.SaveChunk(ctx, c, nil); sErr != nil {
					return fmt.Errorf("ai.ProcessDocument save chunk (no-embed): %w", sErr)
				}
				count++
			}
		}
	}

	if totalTokens > 0 {
		_ = s.logTokens(ctx, doc.TeacherID, oai.ModelEmbedSmall, totalTokens, 0, "doc_embed")
	}

	return s.docs.UpdateDocumentStatus(ctx, docID, domain.AIDocStatusReady, count, nil)
}

// ─── Quiz generation ──────────────────────────────────────────────────────────

type GenerateQuizRequest struct {
	DocumentID   uuid.UUID
	GroupID      uuid.UUID
	TeacherID    uuid.UUID
	Title        string
	NumQuestions int
	CEFRLevel    string
	Subject      string
}

// GenerateQuiz is called by the Asynq worker.
// It retrieves relevant document chunks, calls GPT, and persists the quiz.
func (s *Service) GenerateQuiz(ctx context.Context, req GenerateQuizRequest, jobID uuid.UUID) error {
	_ = s.jobs.UpdateJob(ctx, jobID, domain.AIJobProcessing, nil, nil)

	doc, err := s.docs.GetDocument(ctx, req.DocumentID)
	if err != nil || doc == nil {
		return s.failJob(ctx, jobID, fmt.Errorf("document not found: %w", err))
	}
	if doc.Status != domain.AIDocStatusReady {
		return s.failJob(ctx, jobID, fmt.Errorf("document not ready (status=%s)", doc.Status))
	}

	// Check monthly token limit.
	if s.cfg.MonthlyTokensMax > 0 {
		used, lErr := s.tokens.MonthlyTokens(ctx, req.TeacherID)
		if lErr == nil && used >= s.cfg.MonthlyTokensMax {
			return s.failJob(ctx, jobID, fmt.Errorf("monthly token limit reached"))
		}
	}

	// Build context from top-k similar chunks (or first N if no embeddings).
	contextText, err := s.buildContext(ctx, req.DocumentID, req.Subject, 12)
	if err != nil {
		return s.failJob(ctx, jobID, err)
	}

	if req.NumQuestions <= 0 {
		req.NumQuestions = 20
	}

	if !s.ai.IsConfigured() {
		return s.failJob(ctx, jobID, fmt.Errorf("OpenAI not configured"))
	}

	resp, err := s.ai.Chat(ctx, oai.ChatRequest{
		SystemPrompt: prompts.QuizGeneratorSystem,
		Messages: []oai.ChatMessage{
			{
				Role:    "user",
				Content: prompts.QuizGeneratorUserPrompt(contextText, req.NumQuestions, req.CEFRLevel, req.Subject),
			},
		},
		MaxTokens: 4000,
		JSONMode:  true,
	})
	if err != nil {
		return s.failJob(ctx, jobID, fmt.Errorf("openai quiz gen: %w", err))
	}

	_ = s.logTokens(ctx, req.TeacherID, resp.Model, resp.TokensIn, resp.TokensOut, "quiz_gen")

	// Open a generation session to track TAR for this quiz.
	var sessionID uuid.UUID
	if s.genSess != nil {
		var cefrPtr, subjectPtr *string
		if req.CEFRLevel != "" {
			cefrPtr = &req.CEFRLevel
		}
		if req.Subject != "" {
			subjectPtr = &req.Subject
		}
		gSess := &domain.AIGenerationSession{
			TeacherID:        req.TeacherID,
			SourceDocumentID: &req.DocumentID,
			GroupID:          &req.GroupID,
			GeneratedCount:   req.NumQuestions,
			CEFRLevel:        cefrPtr,
			Subject:          subjectPtr,
			ModelUsed:        resp.Model,
			PromptTokens:     resp.TokensIn,
			CompletionTokens: resp.TokensOut,
			CreatedAt:        time.Now(),
		}
		if err2 := s.genSess.CreateSession(ctx, gSess); err2 == nil {
			sessionID = gSess.ID
		}
	}

	// Parse GPT response.
	var generated struct {
		Title     string `json:"title"`
		Questions []struct {
			Text        string   `json:"text"`
			TopicTags   []string `json:"topic_tags"`
			Difficulty  string   `json:"difficulty"`
			Explanation string   `json:"explanation"`
			Options     []struct {
				Text      string `json:"text"`
				IsCorrect bool   `json:"is_correct"`
			} `json:"options"`
		} `json:"questions"`
	}
	if err = json.Unmarshal([]byte(resp.Content), &generated); err != nil {
		return s.failJob(ctx, jobID, fmt.Errorf("parse quiz json: %w", err))
	}

	title := req.Title
	if title == "" {
		title = generated.Title
	}

	qcReq := QuizCreateRequest{
		GroupID:   req.GroupID,
		TeacherID: req.TeacherID,
		Title:     title,
	}
	for _, q := range generated.Questions {
		gq := GeneratedQuestion{
			Text:        q.Text,
			Explanation: q.Explanation,
			Difficulty:  q.Difficulty,
			TopicTags:   q.TopicTags,
		}
		for _, o := range q.Options {
			gq.Options = append(gq.Options, GeneratedOption{Text: o.Text, IsCorrect: o.IsCorrect})
		}
		qcReq.Questions = append(qcReq.Questions, gq)
	}

	quizID, err := s.quiz.CreateQuizWithQuestions(ctx, qcReq)
	if err != nil {
		return s.failJob(ctx, jobID, fmt.Errorf("persist quiz: %w", err))
	}

	// Link session → quiz so TAR queries can join via quiz_id.
	if s.genSess != nil && sessionID != uuid.Nil {
		_ = s.genSess.LinkQuiz(ctx, sessionID, quizID)
	}

	result, _ := json.Marshal(map[string]string{"quiz_id": quizID.String()})
	return s.jobs.UpdateJob(ctx, jobID, domain.AIJobDone, result, nil)
}

// ─── AI Mentor session ────────────────────────────────────────────────────────

type SendMessageRequest struct {
	SessionID uuid.UUID
	UserID    uuid.UUID
	Content   string
}

// SendMentorMessage adds a user message and returns the AI response.
func (s *Service) SendMentorMessage(ctx context.Context, req SendMessageRequest) (*domain.AIMessage, error) {
	sess, err := s.sessions.GetSession(ctx, req.SessionID)
	if err != nil || sess == nil {
		return nil, domain.ErrNotFound
	}
	if sess.UserID != req.UserID {
		return nil, domain.ErrForbidden
	}

	if !s.ai.IsConfigured() {
		return nil, fmt.Errorf("AI not configured")
	}

	history, err := s.sessions.ListMessages(ctx, req.SessionID)
	if err != nil {
		return nil, fmt.Errorf("ai.SendMentorMessage fetch history: %w", err)
	}

	// Persist user message.
	userMsg := &domain.AIMessage{
		ID:        uuid.New(),
		SessionID: req.SessionID,
		Role:      "user",
		Content:   req.Content,
		CreatedAt: time.Now(),
	}
	if err = s.sessions.AddMessage(ctx, userMsg); err != nil {
		return nil, fmt.Errorf("ai.SendMentorMessage save user msg: %w", err)
	}

	// Build OpenAI messages from history.
	var cefrLevel, subject string
	if sess.CEFRLevel != nil {
		cefrLevel = *sess.CEFRLevel
	}
	if sess.Subject != nil {
		subject = *sess.Subject
	}

	var msgs []oai.ChatMessage
	for _, m := range history {
		if m.Role == "system" {
			continue
		}
		msgs = append(msgs, oai.ChatMessage{Role: m.Role, Content: m.Content})
	}
	msgs = append(msgs, oai.ChatMessage{Role: "user", Content: req.Content})

	resp, err := s.ai.Chat(ctx, oai.ChatRequest{
		SystemPrompt: prompts.MentorSystem(cefrLevel, subject),
		Messages:     msgs,
		MaxTokens:    500,
		Temperature:  0.7,
	})
	if err != nil {
		return nil, fmt.Errorf("ai.SendMentorMessage openai: %w", err)
	}

	_ = s.logTokens(ctx, req.UserID, resp.Model, resp.TokensIn, resp.TokensOut, "mentor")

	assistantMsg := &domain.AIMessage{
		ID:        uuid.New(),
		SessionID: req.SessionID,
		Role:      "assistant",
		Content:   resp.Content,
		TokensIn:  resp.TokensIn,
		TokensOut: resp.TokensOut,
		Model:     resp.Model,
		CreatedAt: time.Now(),
	}
	if err = s.sessions.AddMessage(ctx, assistantMsg); err != nil {
		return nil, fmt.Errorf("ai.SendMentorMessage save assistant msg: %w", err)
	}

	return assistantMsg, nil
}

// ─── Writing check ────────────────────────────────────────────────────────────

type WritingCheckRequest struct {
	UserID uuid.UUID
	Text   string
}

type WritingCheckResult struct {
	CEFREstimate  string         `json:"cefr_estimate"`
	Scores        map[string]int `json:"scores"`
	Errors        []WritingError `json:"errors"`
	Strengths     []string       `json:"strengths"`
	OneImprovement string        `json:"one_improvement"`
}

type WritingError struct {
	Fragment string `json:"fragment"`
	Hint     string `json:"hint"`
}

func (s *Service) CheckWriting(ctx context.Context, req WritingCheckRequest) (*WritingCheckResult, error) {
	if !s.ai.IsConfigured() {
		return nil, fmt.Errorf("AI not configured")
	}

	resp, err := s.ai.Chat(ctx, oai.ChatRequest{
		SystemPrompt: prompts.WritingEvaluatorSystem,
		Messages:     []oai.ChatMessage{{Role: "user", Content: req.Text}},
		MaxTokens:    1000,
		JSONMode:     true,
	})
	if err != nil {
		return nil, fmt.Errorf("ai.CheckWriting: %w", err)
	}

	_ = s.logTokens(ctx, req.UserID, resp.Model, resp.TokensIn, resp.TokensOut, "writing")

	var result WritingCheckResult
	if err = json.Unmarshal([]byte(resp.Content), &result); err != nil {
		return nil, fmt.Errorf("ai.CheckWriting parse: %w", err)
	}
	return &result, nil
}

// ─── Session helpers ──────────────────────────────────────────────────────────

func (s *Service) CreateSession(ctx context.Context, sess *domain.AISession) error {
	sess.ID = uuid.New()
	sess.CreatedAt = time.Now()
	sess.UpdatedAt = time.Now()
	return s.sessions.CreateSession(ctx, sess)
}

func (s *Service) GetSession(ctx context.Context, id, userID uuid.UUID) (*domain.AISession, error) {
	sess, err := s.sessions.GetSession(ctx, id)
	if err != nil || sess == nil {
		return nil, domain.ErrNotFound
	}
	if sess.UserID != userID {
		return nil, domain.ErrForbidden
	}
	return sess, nil
}

func (s *Service) GetSessionMessages(ctx context.Context, sessionID, userID uuid.UUID) ([]*domain.AIMessage, error) {
	sess, err := s.sessions.GetSession(ctx, sessionID)
	if err != nil || sess == nil {
		return nil, domain.ErrNotFound
	}
	if sess.UserID != userID {
		return nil, domain.ErrForbidden
	}
	return s.sessions.ListMessages(ctx, sessionID)
}

// ─── Jobs ─────────────────────────────────────────────────────────────────────

func (s *Service) CreateJob(ctx context.Context, jobType string, userID uuid.UUID, refID *uuid.UUID) (*domain.AIJob, error) {
	job := &domain.AIJob{
		ID:        uuid.New(),
		Type:      jobType,
		UserID:    userID,
		RefID:     refID,
		Status:    domain.AIJobPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.jobs.CreateJob(ctx, job); err != nil {
		return nil, fmt.Errorf("ai.CreateJob: %w", err)
	}
	return job, nil
}

func (s *Service) GetJob(ctx context.Context, id uuid.UUID) (*domain.AIJob, error) {
	return s.jobs.GetJob(ctx, id)
}

// ─── Gamification ─────────────────────────────────────────────────────────────

func (s *Service) GetGamification(ctx context.Context, studentID uuid.UUID) (*domain.StudentGamification, error) {
	return s.gamif.GetOrCreate(ctx, studentID)
}

func (s *Service) AwardXP(ctx context.Context, studentID uuid.UUID, amount int, reason string) error {
	if err := s.gamif.AddXP(ctx, studentID, amount, reason); err != nil {
		return err
	}
	return s.gamif.UpdateStreak(ctx, studentID)
}

// ─── Spaced repetition ────────────────────────────────────────────────────────

func (s *Service) RecordAnswer(ctx context.Context, studentID, topicID uuid.UUID, correct bool, quality int) error {
	m, err := s.mastery.GetMastery(ctx, studentID, topicID)
	if err != nil {
		return err
	}
	isNew := m == nil
	if isNew {
		m = &domain.TopicMastery{
			ID:           uuid.New(),
			StudentID:    studentID,
			TopicID:      topicID,
			IntervalDays: 1,
			EaseFactor:   2.5,
		}
	}
	if correct {
		m.CorrectCount++
	}
	m.TotalCount++
	sm2.Update(m, correct, quality)

	xpAmount := domain.XPQuizCorrect
	acc := float64(m.CorrectCount) / float64(m.TotalCount)
	if acc < 0.6 {
		xpAmount = domain.XPWeakTopicCorrect // bonus for practicing weak topics
	}
	if correct {
		_ = s.AwardXP(ctx, studentID, xpAmount, "quiz_correct")
	}

	return s.mastery.UpsertMastery(ctx, m)
}

func (s *Service) GetReviewQueue(ctx context.Context, studentID uuid.UUID) ([]*domain.TopicMastery, error) {
	return s.mastery.ListDueMastery(ctx, studentID, 20)
}

func (s *Service) GetWeakTopics(ctx context.Context, studentID uuid.UUID) ([]*domain.TopicMastery, error) {
	return s.mastery.ListWeakTopics(ctx, studentID, 10)
}

// ─── Internals ────────────────────────────────────────────────────────────────

// buildContext retrieves the most relevant chunks for quiz generation.
// Falls back to first N chunks when embeddings are not available.
func (s *Service) buildContext(ctx context.Context, docID uuid.UUID, query string, limit int) (string, error) {
	var chunks []*domain.AIDocumentChunk

	if s.ai.IsConfigured() && query != "" {
		queryEmb, err := s.ai.Embed(ctx, query)
		if err == nil {
			chunks, err = s.docs.SearchSimilarChunks(ctx, docID, queryEmb.Vector, limit)
			if err != nil {
				return "", fmt.Errorf("buildContext search: %w", err)
			}
		}
	}

	// Fallback: if embedding search failed or returned nothing, use first N chunks.
	if len(chunks) == 0 {
		// SearchSimilarChunks with nil embedding returns first N by index.
		var err error
		chunks, err = s.docs.SearchSimilarChunks(ctx, docID, nil, limit)
		if err != nil {
			return "", fmt.Errorf("buildContext fallback: %w", err)
		}
	}

	var sb strings.Builder
	for _, c := range chunks {
		sb.WriteString(c.Content)
		sb.WriteString("\n\n")
	}
	return sb.String(), nil
}

func (s *Service) logTokens(ctx context.Context, userID uuid.UUID, model string, in, out int, feature string) error {
	return s.tokens.LogUsage(ctx, &domain.AITokenUsage{
		ID:        uuid.New(),
		UserID:    userID,
		Model:     model,
		TokensIn:  in,
		TokensOut: out,
		Feature:   feature,
		CostUSD:   oai.CalcCostUSD(model, in, out),
		CreatedAt: time.Now(),
	})
}

func (s *Service) failJob(ctx context.Context, jobID uuid.UUID, err error) error {
	msg := err.Error()
	_ = s.jobs.UpdateJob(ctx, jobID, domain.AIJobFailed, nil, &msg)
	return err
}

// ─── Teacher Dashboard ────────────────────────────────────────────────────────

// GetClassInsights fetches raw analytics, runs the Rule Engine, persists recommendations,
// and lazily measures outcomes for recommendations that are 7+ days old.
func (s *Service) GetClassInsights(ctx context.Context, groupID, teacherID uuid.UUID) (*domain.ClassInsights, error) {
	if s.insights == nil {
		return nil, fmt.Errorf("insights not configured")
	}

	// Demo groups return hardcoded rich insights — no DB queries needed.
	if s.insights != nil {
		isDemoGroup, _ := s.insights.IsDemo(ctx, groupID)
		if isDemoGroup {
			return DemoInsights(groupID, teacherID), nil
		}
	}

	ins, err := s.insights.ClassInsights(ctx, groupID)
	if err != nil {
		return nil, err
	}

	// Skip rule engine if the persistence layer isn't wired up.
	if s.recRepo == nil {
		return ins, nil
	}

	// Build TAR context for the rule engine.
	tar, tarTotal := 0.0, 0
	if s.feedback != nil {
		tar, tarTotal, _ = s.feedback.AcceptanceRate(ctx, teacherID)
	}

	rc := RuleContext{
		GroupID:   groupID,
		TeacherID: teacherID,
		Insights:  ins,
		TAR:       tar,
		TARTotal:  tarTotal,
	}
	rawRecs := RunRules(rc)

	// Persist — replaces all pending recs for this group.
	_ = s.recRepo.ReplaceForGroup(ctx, groupID, rawRecs)

	// Convert to the TeacherRecommendation view (with IDs).
	recs := make([]domain.TeacherRecommendation, 0, len(rawRecs))
	for _, r := range rawRecs {
		recs = append(recs, domain.TeacherRecommendation{
			ID:           r.ID,
			Priority:     r.Priority,
			Action:       r.Action,
			Topic:        r.Topic,
			Reason:       r.Reason,
			StudentCount: r.StudentCount,
		})
	}
	ins.Recommendations = recs

	// Lazily measure outcomes for old recommendations (non-blocking best-effort).
	go s.measurePendingOutcomes(context.Background(), groupID)

	return ins, nil
}

// measurePendingOutcomes runs in the background — measures and stores outcome deltas
// for any recommendations that are 7+ days old with no outcome recorded yet.
func (s *Service) measurePendingOutcomes(ctx context.Context, groupID uuid.UUID) {
	recs, err := s.recRepo.PendingForOutcome(ctx, groupID)
	if err != nil || len(recs) == 0 {
		return
	}

	for _, rec := range recs {
		if rec.Topic == "" {
			// Non-topic rules (at_risk, tar_low, celebrate) — store zero-delta placeholder.
			o := &domain.RecommendationOutcome{
				ID:               uuid.New(),
				RecommendationID: rec.ID,
				MeasuredAt:       time.Now(),
			}
			_ = s.recRepo.SaveOutcome(ctx, o)
			continue
		}

		var snapshot struct {
			AvgAccuracy float64 `json:"avg_accuracy"`
		}
		_ = json.Unmarshal(rec.RuleData, &snapshot)

		o := s.recRepo.MeasureOutcomeForTopic(ctx, groupID, rec.Topic, snapshot.AvgAccuracy)
		if o != nil {
			o.RecommendationID = rec.ID
			_ = s.recRepo.SaveOutcome(ctx, o)
		}
	}
}

// RecordRecommendationAction records the teacher's response to a recommendation.
func (s *Service) RecordRecommendationAction(ctx context.Context, recID, teacherID uuid.UUID, status, action string) error {
	if s.recRepo == nil {
		return nil
	}
	rec, err := s.recRepo.GetRecommendation(ctx, recID)
	if err != nil {
		return fmt.Errorf("RecordRecommendationAction: %w", err)
	}
	if rec == nil {
		return domain.ErrNotFound
	}
	if rec.TeacherID != teacherID {
		return domain.ErrForbidden
	}
	return s.recRepo.RecordAction(ctx, recID, status, action)
}

// ExplainRecommendation generates a GPT-written explanation for why a recommendation
// was triggered and what the teacher should do. Falls back to the rule reason if GPT
// is not configured or fails.
func (s *Service) ExplainRecommendation(ctx context.Context, recID, teacherID uuid.UUID) (string, error) {
	if s.recRepo == nil {
		return "", domain.ErrNotFound
	}
	rec, err := s.recRepo.GetRecommendation(ctx, recID)
	if err != nil {
		return "", fmt.Errorf("ExplainRecommendation: %w", err)
	}
	if rec == nil {
		return "", domain.ErrNotFound
	}
	if rec.TeacherID != teacherID {
		return "", domain.ErrForbidden
	}

	if !s.ai.IsConfigured() {
		return rec.Reason, nil
	}

	resp, err := s.ai.Chat(ctx, oai.ChatRequest{
		SystemPrompt: prompts.RecommendationExplainerSystem,
		Messages:     []oai.ChatMessage{{Role: "user", Content: prompts.RecommendationExplainerPrompt(rec)}},
		MaxTokens:    200,
		Temperature:  0.5,
	})
	if err != nil {
		return rec.Reason, nil // graceful degradation
	}
	_ = s.logTokens(ctx, teacherID, resp.Model, resp.TokensIn, resp.TokensOut, "rec_explain")
	return resp.Content, nil
}

func (s *Service) GetStudentProgress(ctx context.Context, groupID, studentID uuid.UUID) (*domain.StudentProgress, error) {
	if s.insights == nil {
		return nil, fmt.Errorf("insights not configured")
	}
	return s.insights.StudentProgress(ctx, groupID, studentID)
}

func (s *Service) SubmitQuestionFeedback(ctx context.Context, questionID, teacherID uuid.UUID, accepted bool) error {
	if s.feedback == nil {
		return nil
	}
	fb := &domain.QuizQuestionFeedback{
		ID:         uuid.New(),
		QuestionID: questionID,
		TeacherID:  teacherID,
		Accepted:   accepted,
		CreatedAt:  time.Now(),
	}
	if err := s.feedback.UpsertFeedback(ctx, fb); err != nil {
		return err
	}
	// Lazily sync session counts so TeacherStats stays accurate.
	if s.genSess != nil {
		_ = s.genSess.SyncCountsByQuestion(ctx, questionID)
	}
	return nil
}

func (s *Service) GetTeacherStats(ctx context.Context, teacherID uuid.UUID) (*domain.TeacherGenerationStats, error) {
	if s.genSess == nil {
		return &domain.TeacherGenerationStats{}, nil
	}
	return s.genSess.TeacherStats(ctx, teacherID)
}

func (s *Service) GetMyAcceptanceRate(ctx context.Context, teacherID uuid.UUID) (float64, int, error) {
	if s.feedback == nil {
		return 0, 0, nil
	}
	return s.feedback.AcceptanceRate(ctx, teacherID)
}

// GenerateQuizFromPayload is the entry point for the Asynq worker.
// It parses string IDs and delegates to GenerateQuiz.
func (s *Service) GenerateQuizFromPayload(ctx context.Context, docID, groupID, teacherID, jobIDStr, title string, numQ int, cefrLevel, subject string) error {
	dID, err := uuid.Parse(docID)
	if err != nil {
		return fmt.Errorf("GenerateQuizFromPayload docID: %w", err)
	}
	gID, err := uuid.Parse(groupID)
	if err != nil {
		return fmt.Errorf("GenerateQuizFromPayload groupID: %w", err)
	}
	tID, err := uuid.Parse(teacherID)
	if err != nil {
		return fmt.Errorf("GenerateQuizFromPayload teacherID: %w", err)
	}
	jID, err := uuid.Parse(jobIDStr)
	if err != nil {
		return fmt.Errorf("GenerateQuizFromPayload jobID: %w", err)
	}
	return s.GenerateQuiz(ctx, GenerateQuizRequest{
		DocumentID:   dID,
		GroupID:      gID,
		TeacherID:    tID,
		Title:        title,
		NumQuestions: numQ,
		CEFRLevel:    cefrLevel,
		Subject:      subject,
	}, jID)
}
