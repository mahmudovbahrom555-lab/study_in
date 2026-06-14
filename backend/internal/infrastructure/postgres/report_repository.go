package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/features/reports"
)

type ReportRepository struct {
	db *sqlx.DB
}

func NewReportRepository(db *sqlx.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

// ComputeMetrics runs SQL aggregations for a student in a group over [from, to].
// The "previous period" is the same-length window ending just before from.
//
// Fix: mastery CTEs now filter by group via question_topics→questions→quizzes,
// so only topics that appear in this group's quizzes are included.
func (r *ReportRepository) ComputeMetrics(
	ctx context.Context,
	studentID, groupID uuid.UUID,
	from, to time.Time,
) (*reports.RawMetrics, error) {
	prevFrom := from.AddDate(0, -1, 0)
	prevTo := from.AddDate(0, 0, -1)

	const q = `
WITH
-- Topics that belong to this group (via question_topics → questions → quizzes)
group_topics AS (
    SELECT DISTINCT qt.topic_id
    FROM question_topics qt
    JOIN questions qn ON qn.id = qt.question_id
    JOIN quizzes qz    ON qz.id = qn.quiz_id
    WHERE qz.group_id = $2
),
-- Attendance for current period
att_cur AS (
    SELECT
        COUNT(*) FILTER (WHERE status = 'present' OR status = 'late') * 100.0
        / NULLIF(COUNT(*), 0) AS pct
    FROM attendance
    WHERE student_id  = $1
      AND group_id    = $2
      AND lesson_date BETWEEN $3 AND $4
),
-- Quiz scores current period
quiz_cur AS (
    SELECT
        COALESCE(AVG(
            qa.score::float * 100.0 / NULLIF(qa.max_score, 0)
        ), 0) AS avg_score,
        COUNT(qa.id) AS attempts
    FROM quiz_attempts qa
    JOIN quizzes qz ON qz.id = qa.quiz_id
    WHERE qa.student_id   = $1
      AND qz.group_id     = $2
      AND qa.finished_at  BETWEEN $3 AND $4
      AND qa.finished_at  IS NOT NULL
),
-- Quiz scores previous period
quiz_prev AS (
    SELECT
        COALESCE(AVG(
            qa.score::float * 100.0 / NULLIF(qa.max_score, 0)
        ), 0) AS avg_score
    FROM quiz_attempts qa
    JOIN quizzes qz ON qz.id = qa.quiz_id
    WHERE qa.student_id   = $1
      AND qz.group_id     = $2
      AND qa.finished_at  BETWEEN $5 AND $6
      AND qa.finished_at  IS NOT NULL
),
-- Mastery at end of current period — computed from skill_logs for a true point-in-time snapshot.
-- topic_mastery is a mutable aggregate (one row per topic); using it with updated_at <= $4
-- gives 0 for any topic updated after that cutoff, inflating the delta to 100% every month.
mastery_cur AS (
    SELECT COALESCE(AVG(per_topic.ratio), 0) AS avg_mastery
    FROM (
        SELECT
            sl.topic_id,
            COUNT(*) FILTER (WHERE sl.is_correct)::float / NULLIF(COUNT(*), 0) AS ratio
        FROM skill_logs sl
        JOIN group_topics gt ON gt.topic_id = sl.topic_id
        WHERE sl.student_id = $1
          AND sl.answered_at <= $4
        GROUP BY sl.topic_id
    ) per_topic
),
-- Mastery at end of previous period — same logic, earlier cutoff.
mastery_prev AS (
    SELECT COALESCE(AVG(per_topic.ratio), 0) AS avg_mastery
    FROM (
        SELECT
            sl.topic_id,
            COUNT(*) FILTER (WHERE sl.is_correct)::float / NULLIF(COUNT(*), 0) AS ratio
        FROM skill_logs sl
        JOIN group_topics gt ON gt.topic_id = sl.topic_id
        WHERE sl.student_id = $1
          AND sl.answered_at <= $6
        GROUP BY sl.topic_id
    ) per_topic
),
-- Homework completion
hw AS (
    SELECT
        COUNT(s.id) * 100.0 / NULLIF(COUNT(a.id), 0) AS completion_pct
    FROM assignments a
    LEFT JOIN submissions s
        ON s.assignment_id = a.id AND s.student_id = $1
    WHERE a.group_id   = $2
      AND a.due_date   BETWEEN $3 AND $4
      AND a.deleted_at IS NULL
)
SELECT
    COALESCE(att.pct,            0) AS attendance_pct,
    COALESCE(qc.avg_score,       0) AS quiz_score_avg,
    COALESCE(qp.avg_score,       0) AS quiz_score_prev_avg,
    COALESCE(mc.avg_mastery,     0) AS mastery_avg,
    COALESCE(mp.avg_mastery,     0) AS mastery_prev_avg,
    COALESCE(hw.completion_pct,  0) AS homework_completion_pct,
    COALESCE(qc.attempts,        0) AS quiz_attempts_count
FROM att_cur att, quiz_cur qc, quiz_prev qp, mastery_cur mc, mastery_prev mp, hw;
`

	var row struct {
		AttendancePct         float64 `db:"attendance_pct"`
		QuizScoreAvg          float64 `db:"quiz_score_avg"`
		QuizScorePrevAvg      float64 `db:"quiz_score_prev_avg"`
		MasteryAvg            float64 `db:"mastery_avg"`
		MasteryPrevAvg        float64 `db:"mastery_prev_avg"`
		HomeworkCompletionPct float64 `db:"homework_completion_pct"`
		QuizAttemptsCount     int     `db:"quiz_attempts_count"`
	}

	if err := r.db.GetContext(ctx, &row, q,
		studentID, groupID,
		from, to,
		prevFrom, prevTo,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &reports.RawMetrics{}, nil
		}
		return nil, fmt.Errorf("ReportRepository.ComputeMetrics: %w", err)
	}

	return &reports.RawMetrics{
		AttendancePct:         row.AttendancePct,
		QuizScoreAvg:          row.QuizScoreAvg,
		QuizScorePrevAvg:      row.QuizScorePrevAvg,
		MasteryAvg:            row.MasteryAvg,
		MasteryPrevAvg:        row.MasteryPrevAvg,
		HomeworkCompletionPct: row.HomeworkCompletionPct,
		QuizAttemptsCount:     row.QuizAttemptsCount,
	}, nil
}

// SaveReport upserts a monthly report and scans the DB-assigned id and created_at back
// into rep. created_at is not reset on conflict so its value reflects first-insert time.
func (r *ReportRepository) SaveReport(ctx context.Context, rep *domain.ParentReport) error {
	const q = `
INSERT INTO parent_reports
    (student_id, group_id, period_start, period_end,
     attendance_pct, quiz_score_avg, quiz_score_prev_avg,
     mastery_avg, mastery_prev_avg,
     homework_completion_pct, quiz_attempts_count, ai_summary)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
ON CONFLICT (student_id, group_id, period_start)
DO UPDATE SET
    attendance_pct          = EXCLUDED.attendance_pct,
    quiz_score_avg          = EXCLUDED.quiz_score_avg,
    quiz_score_prev_avg     = EXCLUDED.quiz_score_prev_avg,
    mastery_avg             = EXCLUDED.mastery_avg,
    mastery_prev_avg        = EXCLUDED.mastery_prev_avg,
    homework_completion_pct = EXCLUDED.homework_completion_pct,
    quiz_attempts_count     = EXCLUDED.quiz_attempts_count,
    ai_summary              = EXCLUDED.ai_summary
RETURNING id, created_at
`
	return r.db.QueryRowContext(ctx, q,
		rep.StudentID, rep.GroupID,
		rep.PeriodStart, rep.PeriodEnd,
		rep.AttendancePct, rep.QuizScoreAvg, rep.QuizScorePrevAvg,
		rep.MasteryAvg, rep.MasteryPrevAvg,
		rep.HomeworkCompletionPct, rep.QuizAttemptsCount, rep.AISummary,
	).Scan(&rep.ID, &rep.CreatedAt)
}

// GetLatestReport returns the most recent report for student+group.
func (r *ReportRepository) GetLatestReport(
	ctx context.Context,
	studentID, groupID uuid.UUID,
) (*domain.ParentReport, error) {
	const q = `
SELECT id, student_id, group_id, period_start, period_end,
       attendance_pct, quiz_score_avg, quiz_score_prev_avg,
       mastery_avg, mastery_prev_avg,
       homework_completion_pct, quiz_attempts_count, ai_summary, created_at
FROM parent_reports
WHERE student_id = $1 AND group_id = $2
ORDER BY period_start DESC
LIMIT 1
`
	var rep domain.ParentReport
	if err := r.db.GetContext(ctx, &rep, q, studentID, groupID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("ReportRepository.GetLatestReport: %w", err)
	}
	return &rep, nil
}

// ComputeRiskAlerts finds at-risk students across all active groups owned by teacherID.
func (r *ReportRepository) ComputeRiskAlerts(
	ctx context.Context,
	teacherID uuid.UUID,
) ([]*domain.RiskAlert, error) {
	const q = `
WITH teacher_groups AS (
    SELECT id, name FROM groups
    WHERE teacher_id = $1 AND deleted_at IS NULL AND is_archived = FALSE
),
group_students AS (
    SELECT gm.student_id, g.id AS group_id, g.name AS group_name,
           u.name AS student_name
    FROM group_members gm
    JOIN teacher_groups g ON g.id = gm.group_id
    JOIN users u ON u.id = gm.student_id
),
absences AS (
    SELECT student_id, group_id,
           COUNT(*) FILTER (WHERE status = 'absent') AS absent_count
    FROM attendance
    WHERE group_id IN (SELECT id FROM teacher_groups)
      AND lesson_date >= NOW() - INTERVAL '14 days'
    GROUP BY student_id, group_id
),
missing_hw AS (
    SELECT gm.student_id, a.group_id,
           COUNT(a.id) - COUNT(s.id) AS missing_count
    FROM assignments a
    JOIN group_members gm ON gm.group_id = a.group_id
    LEFT JOIN submissions s
        ON s.assignment_id = a.id AND s.student_id = gm.student_id
    WHERE a.group_id IN (SELECT id FROM teacher_groups)
      AND a.due_date >= NOW() - INTERVAL '14 days'
      AND a.due_date  < NOW()
      AND a.deleted_at IS NULL
    GROUP BY gm.student_id, a.group_id
)
SELECT
    gs.student_id,
    gs.student_name,
    gs.group_id,
    gs.group_name,
    COALESCE(ab.absent_count,  0) AS absent_count,
    COALESCE(mh.missing_count, 0) AS missing_count,
    CASE
        WHEN COALESCE(ab.absent_count, 0) >= 3 THEN 'high'
        WHEN COALESCE(ab.absent_count, 0) >= 2
          OR COALESCE(mh.missing_count, 0) >= 2 THEN 'medium'
        ELSE 'low'
    END AS risk_level
FROM group_students gs
LEFT JOIN absences ab    ON ab.student_id = gs.student_id AND ab.group_id = gs.group_id
LEFT JOIN missing_hw mh  ON mh.student_id = gs.student_id AND mh.group_id = gs.group_id
WHERE COALESCE(ab.absent_count, 0) >= 2
   OR COALESCE(mh.missing_count, 0) >= 2
ORDER BY CASE risk_level WHEN 'high' THEN 1 WHEN 'medium' THEN 2 ELSE 3 END, gs.student_name
`

	type row struct {
		StudentID    uuid.UUID `db:"student_id"`
		StudentName  string    `db:"student_name"`
		GroupID      uuid.UUID `db:"group_id"`
		GroupName    string    `db:"group_name"`
		AbsentCount  int       `db:"absent_count"`
		MissingCount int       `db:"missing_count"`
		RiskLevel    string    `db:"risk_level"`
	}

	var rows []row
	if err := r.db.SelectContext(ctx, &rows, q, teacherID); err != nil {
		return nil, fmt.Errorf("ReportRepository.ComputeRiskAlerts: %w", err)
	}

	alerts := make([]*domain.RiskAlert, 0, len(rows))
	for _, rw := range rows {
		reason, rec := buildRiskText(rw.AbsentCount, rw.MissingCount)
		alerts = append(alerts, &domain.RiskAlert{
			TeacherID:      teacherID,
			StudentID:      rw.StudentID,
			GroupID:        rw.GroupID,
			StudentName:    rw.StudentName,
			GroupName:      rw.GroupName,
			RiskLevel:      rw.RiskLevel,
			TriggerReason:  reason,
			Recommendation: rec,
			Status:         "new",
		})
	}
	return alerts, nil
}

func buildRiskText(absences, missing int) (reason, recommendation string) {
	switch {
	case absences >= 3:
		reason = fmt.Sprintf("Пропустил %d занятия за последние 14 дней", absences)
		recommendation = "Свяжитесь с родителями, предложите индивидуальную консультацию для выяснения причины пропусков."
	case absences >= 2 && missing >= 2:
		reason = fmt.Sprintf("Пропустил %d занятия и не сдал %d домашних задания", absences, missing)
		recommendation = "Напомните ученику о домашних заданиях. При необходимости уведомите родителей."
	case absences >= 2:
		reason = fmt.Sprintf("Пропустил %d занятия за последние 14 дней", absences)
		recommendation = "Проверьте, нет ли проблем с мотивацией или расписанием. Свяжитесь с родителями."
	default:
		reason = fmt.Sprintf("Не сдал %d домашних задания за последние 14 дней", missing)
		recommendation = "Напомните ученику о незакрытых заданиях через приложение или напрямую."
	}
	return
}

// SaveRiskAlerts atomically replaces unresolved alerts for teacherID with the fresh set.
// Resolved alerts are preserved. DELETE and INSERT run inside a single transaction so a
// failed INSERT cannot leave the teacher with zero alerts.
func (r *ReportRepository) SaveRiskAlerts(
	ctx context.Context,
	teacherID uuid.UUID,
	alerts []*domain.RiskAlert,
) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("ReportRepository.SaveRiskAlerts begin: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	// Clear new + in_progress alerts. Resolved history is kept intentionally.
	// in_progress is also cleared: the student may no longer be at risk this cycle.
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM owner_risk_alerts WHERE teacher_id = $1 AND status IN ('new', 'in_progress')`,
		teacherID,
	); err != nil {
		return fmt.Errorf("ReportRepository.SaveRiskAlerts delete: %w", err)
	}

	if len(alerts) == 0 {
		return tx.Commit()
	}

	// Single multi-row INSERT with RETURNING id so callers get real DB-assigned UUIDs.
	const paramsPerRow = 6
	placeholders := make([]string, len(alerts))
	args := make([]any, 0, len(alerts)*paramsPerRow)
	for i, a := range alerts {
		base := i * paramsPerRow
		placeholders[i] = fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d,'new')",
			base+1, base+2, base+3, base+4, base+5, base+6)
		args = append(args,
			a.TeacherID, a.StudentID, a.GroupID,
			a.RiskLevel, a.TriggerReason, a.Recommendation,
		)
	}
	q := `INSERT INTO owner_risk_alerts
	    (teacher_id, student_id, group_id, risk_level, trigger_reason, recommendation, status)
	    VALUES ` + strings.Join(placeholders, ",") + ` RETURNING id`

	dbRows, err := tx.QueryContext(ctx, q, args...)
	if err != nil {
		return fmt.Errorf("ReportRepository.SaveRiskAlerts insert: %w", err)
	}
	defer dbRows.Close()

	i := 0
	for dbRows.Next() {
		if err := dbRows.Scan(&alerts[i].ID); err != nil {
			return fmt.Errorf("ReportRepository.SaveRiskAlerts scan: %w", err)
		}
		i++
	}
	if err := dbRows.Err(); err != nil {
		return fmt.Errorf("ReportRepository.SaveRiskAlerts rows: %w", err)
	}

	return tx.Commit()
}

// GetRiskAlerts returns alerts for a teacher, optionally filtered by status.
func (r *ReportRepository) GetRiskAlerts(
	ctx context.Context,
	teacherID uuid.UUID,
	status string,
) ([]*domain.RiskAlert, error) {
	q := `
SELECT ora.id, ora.teacher_id, ora.student_id, ora.group_id,
       u.name  AS student_name,
       g.name  AS group_name,
       ora.risk_level, ora.trigger_reason, ora.recommendation,
       ora.status, ora.created_at
FROM owner_risk_alerts ora
JOIN users  u ON u.id = ora.student_id
JOIN groups g ON g.id = ora.group_id
WHERE ora.teacher_id = $1
`
	args := []any{teacherID}
	if status != "" {
		q += " AND ora.status = $2"
		args = append(args, status)
	}
	q += " ORDER BY CASE ora.risk_level WHEN 'high' THEN 1 WHEN 'medium' THEN 2 ELSE 3 END, ora.created_at DESC LIMIT 100"

	var rows []*domain.RiskAlert
	if err := r.db.SelectContext(ctx, &rows, q, args...); err != nil {
		return nil, fmt.Errorf("ReportRepository.GetRiskAlerts: %w", err)
	}
	return rows, nil
}

// ResolveRiskAlert marks a specific alert as resolved.
func (r *ReportRepository) ResolveRiskAlert(
	ctx context.Context,
	alertID, teacherID uuid.UUID,
) error {
	const q = `
UPDATE owner_risk_alerts
SET status = 'resolved', updated_at = NOW()
WHERE id = $1 AND teacher_id = $2
`
	res, err := r.db.ExecContext(ctx, q, alertID, teacherID)
	if err != nil {
		return fmt.Errorf("ReportRepository.ResolveRiskAlert: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// IsGroupMember verifies the student belongs to the group.
func (r *ReportRepository) IsGroupMember(ctx context.Context, groupID, studentID uuid.UUID) (bool, error) {
	const q = `SELECT EXISTS(SELECT 1 FROM group_members WHERE group_id=$1 AND student_id=$2)`
	var ok bool
	if err := r.db.GetContext(ctx, &ok, q, groupID, studentID); err != nil {
		return false, fmt.Errorf("ReportRepository.IsGroupMember: %w", err)
	}
	return ok, nil
}

// IsGroupTeacher verifies the caller owns the group.
func (r *ReportRepository) IsGroupTeacher(ctx context.Context, groupID, teacherID uuid.UUID) (bool, error) {
	const q = `SELECT EXISTS(SELECT 1 FROM groups WHERE id=$1 AND teacher_id=$2 AND deleted_at IS NULL)`
	var ok bool
	if err := r.db.GetContext(ctx, &ok, q, groupID, teacherID); err != nil {
		return false, fmt.Errorf("ReportRepository.IsGroupTeacher: %w", err)
	}
	return ok, nil
}

// IsParentOfStudent verifies that a parent_links row exists for (parentID, studentID).
func (r *ReportRepository) IsParentOfStudent(ctx context.Context, parentID, studentID uuid.UUID) (bool, error) {
	const q = `SELECT EXISTS(SELECT 1 FROM parent_links WHERE parent_id=$1 AND student_id=$2)`
	var ok bool
	if err := r.db.GetContext(ctx, &ok, q, parentID, studentID); err != nil {
		return false, fmt.Errorf("ReportRepository.IsParentOfStudent: %w", err)
	}
	return ok, nil
}
