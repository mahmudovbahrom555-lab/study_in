package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

// InsightsRepository implements ai.InsightsRepository using raw SQL analytics queries.
type InsightsRepository struct{ db *sqlx.DB }

func NewInsightsRepository(db *sqlx.DB) *InsightsRepository {
	return &InsightsRepository{db: db}
}

// ClassInsights returns the Teacher Dashboard aggregate for a group.
// Queries: group members → topic mastery → gamification → quiz attempts → feedback.
func (r *InsightsRepository) ClassInsights(ctx context.Context, groupID uuid.UUID) (*domain.ClassInsights, error) {
	out := &domain.ClassInsights{
		GroupID: groupID,
		Period:  "last_30_days",
	}

	// 1. Student count
	if err := r.db.GetContext(ctx, &out.StudentCount,
		`SELECT COUNT(*) FROM group_members WHERE group_id=$1`, groupID); err != nil {
		return nil, fmt.Errorf("ClassInsights student count: %w", err)
	}

	// 2. Class-level topic weaknesses (topics where avg accuracy < 70%).
	//    Join group_members → topic_mastery → topic_tags.
	type weakRow struct {
		Topic              string  `db:"topic_name"`
		AvgAccuracy        float64 `db:"avg_accuracy"`
		StudentsStruggling int     `db:"students_struggling"`
	}
	var weakRows []weakRow
	err := r.db.SelectContext(ctx, &weakRows, `
		SELECT
		    tt.name                                    AS topic_name,
		    AVG(CASE WHEN tm.total_count > 0
		             THEN tm.correct_count::float / tm.total_count
		             ELSE 0 END)                       AS avg_accuracy,
		    COUNT(*) FILTER (WHERE tm.total_count > 0 AND
		                    tm.correct_count::float / tm.total_count < 0.6) AS students_struggling
		FROM group_members gm
		JOIN topic_mastery  tm ON tm.student_id = gm.student_id
		JOIN topic_tags     tt ON tt.id          = tm.topic_id
		WHERE gm.group_id = $1
		  AND tm.total_count > 0
		GROUP BY tt.name
		HAVING AVG(CASE WHEN tm.total_count > 0
		                THEN tm.correct_count::float / tm.total_count
		                ELSE 0 END) < 0.70
		ORDER BY avg_accuracy ASC
		LIMIT 10`, groupID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("ClassInsights weak topics: %w", err)
	}
	for _, w := range weakRows {
		out.ClassWeakness = append(out.ClassWeakness, domain.TopicWeakness{
			Topic:              w.Topic,
			AvgAccuracy:        w.AvgAccuracy,
			StudentsStruggling: w.StudentsStruggling,
			TotalStudents:      out.StudentCount,
		})
	}

	// 3. Student summaries: XP, streak, top weakness, last_active, at-risk flag.
	type studentRow struct {
		StudentID  uuid.UUID  `db:"student_id"`
		Name       string     `db:"name"`
		XPTotal    int        `db:"xp_total"`
		StreakDays int        `db:"streak_days"`
		LastActive *time.Time `db:"last_activity_date"`
	}
	var stuRows []studentRow
	err = r.db.SelectContext(ctx, &stuRows, `
		SELECT
		    gm.student_id,
		    u.name,
		    COALESCE(sg.xp_total, 0)        AS xp_total,
		    COALESCE(sg.streak_days, 0)     AS streak_days,
		    sg.last_activity_date
		FROM group_members gm
		JOIN users u ON u.id = gm.student_id
		LEFT JOIN student_gamification sg ON sg.student_id = gm.student_id
		WHERE gm.group_id = $1
		ORDER BY xp_total DESC`, groupID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("ClassInsights students: %w", err)
	}

	// Top weakness per student (lowest accuracy topic).
	topWeakness := r.topWeaknessPerStudent(ctx, groupID)
	threeDaysAgo := time.Now().AddDate(0, 0, -3)

	for _, s := range stuRows {
		var lastActiveStr *string
		isAtRisk := true
		if s.LastActive != nil {
			str := s.LastActive.Format("2006-01-02")
			lastActiveStr = &str
			isAtRisk = s.LastActive.Before(threeDaysAgo)
		}
		out.Students = append(out.Students, domain.StudentSummary{
			StudentID:   s.StudentID,
			Name:        s.Name,
			XPTotal:     s.XPTotal,
			StreakDays:  s.StreakDays,
			TopWeakness: topWeakness[s.StudentID],
			LastActive:  lastActiveStr,
			IsAtRisk:    isAtRisk,
		})
	}

	// 4. Quiz stats for this group.
	type quizStatRow struct {
		Generated     int     `db:"quiz_count"`
		TotalAttempts int     `db:"attempt_count"`
		AvgScore      float64 `db:"avg_score"`
	}
	var qs quizStatRow
	err = r.db.GetContext(ctx, &qs, `
		SELECT
		    COUNT(DISTINCT q.id)              AS quiz_count,
		    COUNT(qa.id)                       AS attempt_count,
		    COALESCE(AVG(
		        CASE WHEN qa.max_score > 0
		             THEN qa.score::float / qa.max_score * 100
		             ELSE 0 END), 0)           AS avg_score
		FROM quizzes q
		LEFT JOIN quiz_attempts qa ON qa.quiz_id = q.id AND qa.finished_at IS NOT NULL
		WHERE q.group_id = $1
		  AND q.deleted_at IS NULL
		  AND q.created_at >= NOW() - interval '30 days'`, groupID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("ClassInsights quiz stats: %w", err)
	}

	// Acceptance rate for this group's teacher (across all their quizzes).
	var accepted, total int
	_ = r.db.GetContext(ctx, &accepted, `
		SELECT COUNT(*) FROM quiz_question_feedback qf
		JOIN questions q ON q.id = qf.question_id
		JOIN quizzes qz ON qz.id = q.quiz_id
		WHERE qz.group_id = $1 AND qf.accepted = TRUE`, groupID)
	_ = r.db.GetContext(ctx, &total, `
		SELECT COUNT(*) FROM quiz_question_feedback qf
		JOIN questions q ON q.id = qf.question_id
		JOIN quizzes qz ON qz.id = q.quiz_id
		WHERE qz.group_id = $1`, groupID)

	rate := 0.0
	if total > 0 {
		rate = float64(accepted) / float64(total)
	}

	out.QuizStats = domain.QuizStats{
		Generated:      qs.Generated,
		AcceptanceRate: rate,
		TotalAttempts:  qs.TotalAttempts,
		AvgScore:       qs.AvgScore,
	}

	// 5. "Next Best Action" recommendations — deterministic heuristics, no GPT.
	out.Recommendations = r.buildRecommendations(out)

	return out, nil
}

// buildRecommendations derives teacher actions from the already-computed insights.
// Rules are ordered by estimated impact. All logic is pure Go — zero GPT calls.
func (r *InsightsRepository) buildRecommendations(ins *domain.ClassInsights) []domain.TeacherRecommendation {
	var recs []domain.TeacherRecommendation
	half := ins.StudentCount / 2
	if half < 1 {
		half = 1
	}

	// P1: create a quiz for every topic where ≥50% of students struggle.
	for _, w := range ins.ClassWeakness {
		if w.StudentsStruggling >= half {
			recs = append(recs, domain.TeacherRecommendation{
				Priority:     1,
				Action:       "create_quiz",
				Topic:        w.Topic,
				Reason:       fmt.Sprintf("%.0f%% средняя точность — %d/%d учеников испытывают трудности", w.AvgAccuracy*100, w.StudentsStruggling, w.TotalStudents),
				StudentCount: w.StudentsStruggling,
			})
		}
	}

	// P1: if TAR < 70% the model is underperforming — remind teacher to rate questions.
	if ins.QuizStats.Generated > 0 && ins.QuizStats.AcceptanceRate < 0.70 {
		recs = append(recs, domain.TeacherRecommendation{
			Priority: 1,
			Action:   "rate_questions",
			Reason:   fmt.Sprintf("TAR %.0f%% — оцените сгенерированные вопросы, чтобы улучшить качество AI", ins.QuizStats.AcceptanceRate*100),
		})
	}

	// P2: check in with at-risk students (3+ days inactive).
	atRisk := 0
	for _, s := range ins.Students {
		if s.IsAtRisk {
			atRisk++
		}
	}
	if atRisk > 0 {
		recs = append(recs, domain.TeacherRecommendation{
			Priority:     2,
			Action:       "check_students",
			Reason:       fmt.Sprintf("%d учеников не заходили 3+ дня — стоит написать им", atRisk),
			StudentCount: atRisk,
		})
	}

	// P2: topics where 25-49% struggle — schedule a review (lighter than a new quiz).
	for _, w := range ins.ClassWeakness {
		if w.StudentsStruggling < half && w.StudentsStruggling > 0 {
			recs = append(recs, domain.TeacherRecommendation{
				Priority:     2,
				Action:       "schedule_review",
				Topic:        w.Topic,
				Reason:       fmt.Sprintf("%d ученик(ов) с трудом справляется — назначьте повторение через 3 дня", w.StudentsStruggling),
				StudentCount: w.StudentsStruggling,
			})
		}
	}

	// P3: celebrate if avg score ≥ 80% and no at-risk students.
	if ins.QuizStats.AvgScore >= 80 && atRisk == 0 && ins.StudentCount > 0 {
		recs = append(recs, domain.TeacherRecommendation{
			Priority: 3,
			Action:   "celebrate",
			Reason:   fmt.Sprintf("Средний балл %.0f%% — группа молодцы! Можно усложнить материал", ins.QuizStats.AvgScore),
		})
	}

	return recs
}

// topWeaknessPerStudent returns the lowest-accuracy topic name per student in the group.
func (r *InsightsRepository) topWeaknessPerStudent(ctx context.Context, groupID uuid.UUID) map[uuid.UUID]string {
	type row struct {
		StudentID uuid.UUID `db:"student_id"`
		TopicName string    `db:"topic_name"`
	}
	var rows []row
	_ = r.db.SelectContext(ctx, &rows, `
		SELECT DISTINCT ON (tm.student_id)
		    tm.student_id,
		    tt.name AS topic_name
		FROM group_members gm
		JOIN topic_mastery tm ON tm.student_id = gm.student_id
		JOIN topic_tags    tt ON tt.id          = tm.topic_id
		WHERE gm.group_id = $1
		  AND tm.total_count > 0
		ORDER BY tm.student_id,
		         (tm.correct_count::float / tm.total_count) ASC`, groupID)
	result := make(map[uuid.UUID]string, len(rows))
	for _, r := range rows {
		result[r.StudentID] = r.TopicName
	}
	return result
}

// StudentProgress returns detailed analytics for one student inside a group.
func (r *InsightsRepository) StudentProgress(ctx context.Context, groupID, studentID uuid.UUID) (*domain.StudentProgress, error) {
	out := &domain.StudentProgress{StudentID: studentID}

	// Student name.
	if err := r.db.GetContext(ctx, &out.Name,
		`SELECT name FROM users WHERE id=$1`, studentID); err != nil {
		return nil, fmt.Errorf("StudentProgress name: %w", err)
	}

	// Gamification.
	var g domain.StudentGamification
	if err := r.db.GetContext(ctx, &g,
		`SELECT * FROM student_gamification WHERE student_id=$1`, studentID); err == nil {
		out.Gamification = &g
	}

	// Weak topics (bottom 10 by accuracy).
	type topicRow struct {
		TopicID          uuid.UUID `db:"topic_id"`
		TopicName        string    `db:"topic_name"`
		CorrectCount     int       `db:"correct_count"`
		TotalCount       int       `db:"total_count"`
		NextReview       time.Time `db:"next_review"`
		IntervalDays     int       `db:"interval_days"`
		ConfidenceScore  float64   `db:"confidence_score"`
		ConsistencyScore float64   `db:"consistency_score"`
	}
	var topicRows []topicRow
	_ = r.db.SelectContext(ctx, &topicRows, `
		SELECT
		    tm.topic_id,
		    tt.name             AS topic_name,
		    tm.correct_count,
		    tm.total_count,
		    tm.next_review,
		    tm.interval_days,
		    tm.confidence_score,
		    tm.consistency_score
		FROM topic_mastery tm
		JOIN topic_tags tt ON tt.id = tm.topic_id
		WHERE tm.student_id = $1 AND tm.total_count > 0
		ORDER BY (tm.correct_count::float / tm.total_count) ASC
		LIMIT 10`, studentID)

	for _, tr := range topicRows {
		accuracy := 0.0
		if tr.TotalCount > 0 {
			accuracy = float64(tr.CorrectCount) / float64(tr.TotalCount)
		}
		out.WeakTopics = append(out.WeakTopics, domain.StudentTopicDetail{
			TopicID:          tr.TopicID,
			TopicName:        tr.TopicName,
			Accuracy:         accuracy,
			TotalAnswers:     tr.TotalCount,
			NextReview:       tr.NextReview,
			IntervalDays:     tr.IntervalDays,
			ConfidenceScore:  tr.ConfidenceScore,
			ConsistencyScore: tr.ConsistencyScore,
		})
	}

	// Quiz attempt history (last 20, only from this group).
	type attemptRow struct {
		QuizTitle  string     `db:"title"`
		Score      *int16     `db:"score"`
		MaxScore   int16      `db:"max_score"`
		FinishedAt *time.Time `db:"finished_at"`
	}
	var attempts []attemptRow
	_ = r.db.SelectContext(ctx, &attempts, `
		SELECT q.title, qa.score, qa.max_score, qa.finished_at
		FROM quiz_attempts qa
		JOIN quizzes q ON q.id = qa.quiz_id
		WHERE qa.student_id = $1
		  AND q.group_id    = $2
		  AND qa.finished_at IS NOT NULL
		ORDER BY qa.finished_at DESC
		LIMIT 20`, studentID, groupID)

	for _, a := range attempts {
		if a.Score == nil || a.FinishedAt == nil {
			continue
		}
		pct := 0.0
		if a.MaxScore > 0 {
			pct = float64(*a.Score) / float64(a.MaxScore) * 100
		}
		out.QuizHistory = append(out.QuizHistory, domain.QuizAttemptSummary{
			QuizTitle:  a.QuizTitle,
			Score:      *a.Score,
			MaxScore:   a.MaxScore,
			Percentage: pct,
			FinishedAt: *a.FinishedAt,
		})
	}

	// Skill assessments — latest per skill.
	type skillRow struct {
		Skill     string    `db:"skill"`
		CEFRLevel *string   `db:"cefr_level"`
		Score     *float64  `db:"score"`
		CreatedAt time.Time `db:"created_at"`
	}
	var skills []skillRow
	_ = r.db.SelectContext(ctx, &skills, `
		SELECT DISTINCT ON (skill)
		    skill, cefr_level, score, created_at
		FROM skill_assessments
		WHERE student_id = $1
		ORDER BY skill, created_at DESC`, studentID)

	if len(skills) > 0 {
		out.SkillLevels = make(map[string]domain.SkillLevel, len(skills))
		for _, sk := range skills {
			sl := domain.SkillLevel{LastAssessed: sk.CreatedAt}
			if sk.CEFRLevel != nil {
				sl.CEFR = *sk.CEFRLevel
			}
			if sk.Score != nil {
				sl.Score = *sk.Score
			}
			out.SkillLevels[sk.Skill] = sl
		}
	}

	return out, nil
}

// ─── QuizFeedbackRepository ───────────────────────────────────────────────────

type QuizFeedbackRepository struct{ db *sqlx.DB }

func NewQuizFeedbackRepository(db *sqlx.DB) *QuizFeedbackRepository {
	return &QuizFeedbackRepository{db: db}
}

func (r *QuizFeedbackRepository) UpsertFeedback(ctx context.Context, fb *domain.QuizQuestionFeedback) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO quiz_question_feedback (id, question_id, teacher_id, accepted, created_at)
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (question_id, teacher_id) DO UPDATE SET accepted=EXCLUDED.accepted`,
		fb.ID, fb.QuestionID, fb.TeacherID, fb.Accepted, fb.CreatedAt)
	return err
}

func (r *QuizFeedbackRepository) AcceptanceRate(ctx context.Context, teacherID uuid.UUID) (float64, int, error) {
	type result struct {
		Total    int `db:"total"`
		Accepted int `db:"accepted"`
	}
	var res result
	err := r.db.GetContext(ctx, &res, `
		SELECT
		    COUNT(*)                             AS total,
		    COUNT(*) FILTER (WHERE accepted)     AS accepted
		FROM quiz_question_feedback
		WHERE teacher_id = $1`, teacherID)
	if err != nil {
		return 0, 0, err
	}
	if res.Total == 0 {
		return 0, 0, nil
	}
	return float64(res.Accepted) / float64(res.Total), res.Total, nil
}

// ─── AIGenerationSessionRepository ───────────────────────────────────────────

type AIGenerationSessionRepository struct{ db *sqlx.DB }

func NewAIGenerationSessionRepository(db *sqlx.DB) *AIGenerationSessionRepository {
	return &AIGenerationSessionRepository{db: db}
}

func (r *AIGenerationSessionRepository) CreateSession(ctx context.Context, s *domain.AIGenerationSession) error {
	s.ID = uuid.New()
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO ai_generation_sessions
		    (id, teacher_id, source_document_id, group_id, generated_count,
		     cefr_level, subject, model_used, prompt_tokens, completion_tokens, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		s.ID, s.TeacherID, s.SourceDocumentID, s.GroupID, s.GeneratedCount,
		s.CEFRLevel, s.Subject, s.ModelUsed, s.PromptTokens, s.CompletionTokens, s.CreatedAt)
	return err
}

func (r *AIGenerationSessionRepository) SyncCountsByQuestion(ctx context.Context, questionID uuid.UUID) error {
	// Resolve question → quiz → session, then re-derive counts.
	var sessionID uuid.UUID
	err := r.db.GetContext(ctx, &sessionID, `
		SELECT gs.id
		FROM ai_generation_sessions gs
		JOIN questions q ON q.quiz_id = gs.quiz_id
		WHERE q.id = $1
		LIMIT 1`, questionID)
	if err != nil {
		return nil // no session linked yet — not an error
	}
	return r.SyncCounts(ctx, sessionID)
}

func (r *AIGenerationSessionRepository) LinkQuiz(ctx context.Context, sessionID, quizID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE ai_generation_sessions SET quiz_id=$1, completed_at=NOW() WHERE id=$2`,
		quizID, sessionID)
	return err
}

// SyncCounts re-derives accepted/rejected from quiz_question_feedback for the quiz
// linked to this session. Safe to call multiple times (idempotent).
func (r *AIGenerationSessionRepository) SyncCounts(ctx context.Context, sessionID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE ai_generation_sessions gs
		SET
		    accepted_count = COALESCE((
		        SELECT COUNT(*) FROM quiz_question_feedback qf
		        JOIN questions q ON q.id = qf.question_id
		        WHERE q.quiz_id = gs.quiz_id AND qf.accepted
		    ), 0),
		    rejected_count = COALESCE((
		        SELECT COUNT(*) FROM quiz_question_feedback qf
		        JOIN questions q ON q.id = qf.question_id
		        WHERE q.quiz_id = gs.quiz_id AND NOT qf.accepted
		    ), 0)
		WHERE gs.id = $1`, sessionID)
	return err
}

func (r *AIGenerationSessionRepository) TeacherStats(ctx context.Context, teacherID uuid.UUID) (*domain.TeacherGenerationStats, error) {
	type row struct {
		Sessions   int `db:"sessions"`
		Generated  int `db:"generated"`
		Accepted   int `db:"accepted"`
		Edited     int `db:"edited"`
		Rejected   int `db:"rejected"`
	}
	var res row
	err := r.db.GetContext(ctx, &res, `
		SELECT
		    COUNT(*)                  AS sessions,
		    COALESCE(SUM(generated_count), 0) AS generated,
		    COALESCE(SUM(accepted_count), 0)  AS accepted,
		    COALESCE(SUM(edited_count), 0)    AS edited,
		    COALESCE(SUM(rejected_count), 0)  AS rejected
		FROM ai_generation_sessions
		WHERE teacher_id = $1`, teacherID)
	if err != nil {
		return nil, err
	}
	rate := 0.0
	if res.Generated > 0 {
		rate = float64(res.Accepted) / float64(res.Generated)
	}
	return &domain.TeacherGenerationStats{
		TotalSessions:  res.Sessions,
		TotalGenerated: res.Generated,
		TotalAccepted:  res.Accepted,
		TotalEdited:    res.Edited,
		TotalRejected:  res.Rejected,
		AcceptanceRate: rate,
	}, nil
}
