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
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/features/assignments"
)

// AssignmentRepository implements assignments.Repository.
type AssignmentRepository struct {
	db *sqlx.DB
}

func NewAssignmentRepository(db *sqlx.DB) assignments.Repository {
	return &AssignmentRepository{db: db}
}

func (r *AssignmentRepository) CreateAssignment(ctx context.Context, a *domain.Assignment) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO assignments (id, group_id, teacher_id, title, description, due_date, created_at, updated_at)
		VALUES (:id, :group_id, :teacher_id, :title, :description, :due_date, :created_at, :updated_at)`, a)
	if err != nil {
		return fmt.Errorf("AssignmentRepository.CreateAssignment: %w", err)
	}
	return nil
}

func (r *AssignmentRepository) GetAssignmentByID(ctx context.Context, id uuid.UUID) (*domain.Assignment, error) {
	var a domain.Assignment
	err := r.db.GetContext(ctx, &a,
		`SELECT * FROM assignments WHERE id=$1 AND deleted_at IS NULL`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("AssignmentRepository.GetAssignmentByID: %w", err)
	}
	return &a, nil
}

func (r *AssignmentRepository) ListGroupAssignments(ctx context.Context, groupID uuid.UUID) ([]*domain.Assignment, error) {
	var as []*domain.Assignment
	err := r.db.SelectContext(ctx, &as,
		`SELECT * FROM assignments WHERE group_id=$1 AND deleted_at IS NULL ORDER BY created_at DESC`, groupID)
	if err != nil {
		return nil, fmt.Errorf("AssignmentRepository.ListGroupAssignments: %w", err)
	}
	return as, nil
}

func (r *AssignmentRepository) UpdateAssignment(ctx context.Context, a *domain.Assignment) error {
	_, err := r.db.NamedExecContext(ctx, `
		UPDATE assignments SET title=:title, description=:description, due_date=:due_date, updated_at=:updated_at
		WHERE id=:id AND deleted_at IS NULL`, a)
	if err != nil {
		return fmt.Errorf("AssignmentRepository.UpdateAssignment: %w", err)
	}
	return nil
}

func (r *AssignmentRepository) SoftDeleteAssignment(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE assignments SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("AssignmentRepository.SoftDeleteAssignment: %w", err)
	}
	return nil
}

func (r *AssignmentRepository) AddAssignmentAttachment(ctx context.Context, a *domain.AssignmentAttachment) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO assignment_attachments (id, assignment_id, object_key, filename, mime_type, size_bytes, created_at)
		VALUES (:id, :assignment_id, :object_key, :filename, :mime_type, :size_bytes, :created_at)`, a)
	if err != nil {
		return fmt.Errorf("AssignmentRepository.AddAssignmentAttachment: %w", err)
	}
	return nil
}

func (r *AssignmentRepository) ListAssignmentAttachments(ctx context.Context, assignmentID uuid.UUID) ([]*domain.AssignmentAttachment, error) {
	var atts []*domain.AssignmentAttachment
	err := r.db.SelectContext(ctx, &atts,
		`SELECT * FROM assignment_attachments WHERE assignment_id=$1 ORDER BY created_at ASC`, assignmentID)
	if err != nil {
		return nil, fmt.Errorf("AssignmentRepository.ListAssignmentAttachments: %w", err)
	}
	return atts, nil
}

// SaveSubmission — одна транзакция: сдача и её файлы сохраняются вместе или не
// сохраняются вовсе, учитель не увидит работу «наполовину».
func (r *AssignmentRepository) SaveSubmission(ctx context.Context, s *domain.Submission, fileIDs []uuid.UUID) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("AssignmentRepository.SaveSubmission begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }() // после Commit — no-op

	if _, err := tx.NamedExecContext(ctx, `
		INSERT INTO submissions (id, assignment_id, student_id, comment, submitted_at)
		VALUES (:id, :assignment_id, :student_id, :comment, :submitted_at)
		ON CONFLICT (assignment_id, student_id) DO UPDATE
		SET comment = EXCLUDED.comment, submitted_at = EXCLUDED.submitted_at`, s); err != nil {
		return fmt.Errorf("AssignmentRepository.SaveSubmission upsert: %w", err)
	}

	// Прежние файлы отвязываем и возвращаем в pending: если ученик их убрал при
	// пересдаче, их удалит очистка; если оставил — ниже снова станут attached.
	if _, err := tx.ExecContext(ctx, `
		UPDATE files SET status = 'pending'
		WHERE id IN (SELECT file_id FROM submission_files WHERE submission_id = $1)`, s.ID); err != nil {
		return fmt.Errorf("AssignmentRepository.SaveSubmission release: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM submission_files WHERE submission_id = $1`, s.ID); err != nil {
		return fmt.Errorf("AssignmentRepository.SaveSubmission unlink: %w", err)
	}

	for i, fid := range fileIDs {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO submission_files (submission_id, file_id, position) VALUES ($1, $2, $3)`,
			s.ID, fid, i+1); err != nil {
			return fmt.Errorf("AssignmentRepository.SaveSubmission link: %w", err)
		}
	}
	if len(fileIDs) > 0 {
		q, args, err := sqlx.In(`UPDATE files SET status = 'attached' WHERE id IN (?)`, fileIDs)
		if err != nil {
			return fmt.Errorf("AssignmentRepository.SaveSubmission attach build: %w", err)
		}
		if _, err := tx.ExecContext(ctx, tx.Rebind(q), args...); err != nil {
			return fmt.Errorf("AssignmentRepository.SaveSubmission attach: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("AssignmentRepository.SaveSubmission commit: %w", err)
	}
	return nil
}

func (r *AssignmentRepository) GetSubmission(ctx context.Context, assignmentID, studentID uuid.UUID) (*domain.Submission, error) {
	var s domain.Submission
	err := r.db.GetContext(ctx, &s,
		`SELECT * FROM submissions WHERE assignment_id=$1 AND student_id=$2`, assignmentID, studentID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("AssignmentRepository.GetSubmission: %w", err)
	}
	return &s, nil
}

func (r *AssignmentRepository) GetSubmissionByID(ctx context.Context, id uuid.UUID) (*domain.Submission, error) {
	var s domain.Submission
	err := r.db.GetContext(ctx, &s, `SELECT * FROM submissions WHERE id=$1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("AssignmentRepository.GetSubmissionByID: %w", err)
	}
	return &s, nil
}

const submissionColumns = `s.id, s.assignment_id, s.student_id, s.comment, s.grade,
	s.teacher_note, s.submitted_at, s.graded_at`

// ListSubmissions — сначала непроверенные, внутри — по времени сдачи.
func (r *AssignmentRepository) ListSubmissions(ctx context.Context, assignmentID uuid.UUID) ([]*domain.SubmissionWithStudent, error) {
	var subs []*domain.SubmissionWithStudent
	err := r.db.SelectContext(ctx, &subs, `
		SELECT `+submissionColumns+`, u.name AS student_name
		FROM submissions s
		JOIN users u ON u.id = s.student_id
		WHERE s.assignment_id = $1
		ORDER BY (s.grade IS NOT NULL), s.submitted_at`, assignmentID)
	if err != nil {
		return nil, fmt.Errorf("AssignmentRepository.ListSubmissions: %w", err)
	}
	return subs, nil
}

func (r *AssignmentRepository) ListStudentSubmissions(ctx context.Context, studentID uuid.UUID, assignmentIDs []uuid.UUID) (map[uuid.UUID]*domain.Submission, error) {
	out := make(map[uuid.UUID]*domain.Submission, len(assignmentIDs))
	if len(assignmentIDs) == 0 {
		return out, nil
	}
	q, args, err := sqlx.In(`
		SELECT `+submissionColumns+`
		FROM submissions s
		WHERE s.student_id = ? AND s.assignment_id IN (?)`, studentID, assignmentIDs)
	if err != nil {
		return nil, fmt.Errorf("AssignmentRepository.ListStudentSubmissions build: %w", err)
	}
	var subs []*domain.Submission
	if err := r.db.SelectContext(ctx, &subs, r.db.Rebind(q), args...); err != nil {
		return nil, fmt.Errorf("AssignmentRepository.ListStudentSubmissions: %w", err)
	}
	for _, s := range subs {
		out[s.AssignmentID] = s
	}
	return out, nil
}

func (r *AssignmentRepository) SubmissionStats(ctx context.Context, assignmentIDs []uuid.UUID) (map[uuid.UUID]domain.SubmissionStats, error) {
	out := make(map[uuid.UUID]domain.SubmissionStats, len(assignmentIDs))
	if len(assignmentIDs) == 0 {
		return out, nil
	}
	q, args, err := sqlx.In(`
		SELECT assignment_id,
		       COUNT(*)                                AS submitted,
		       COUNT(*) FILTER (WHERE grade IS NULL)   AS ungraded
		FROM submissions
		WHERE assignment_id IN (?)
		GROUP BY assignment_id`, assignmentIDs)
	if err != nil {
		return nil, fmt.Errorf("AssignmentRepository.SubmissionStats build: %w", err)
	}
	var rows []struct {
		AssignmentID uuid.UUID `db:"assignment_id"`
		domain.SubmissionStats
	}
	if err := r.db.SelectContext(ctx, &rows, r.db.Rebind(q), args...); err != nil {
		return nil, fmt.Errorf("AssignmentRepository.SubmissionStats: %w", err)
	}
	for _, row := range rows {
		out[row.AssignmentID] = row.SubmissionStats
	}
	return out, nil
}

func (r *AssignmentRepository) GradeSubmission(ctx context.Context, id uuid.UUID, grade int16, note string) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx,
		`UPDATE submissions SET grade=$1, teacher_note=$2, graded_at=$3 WHERE id=$4`,
		grade, note, now, id)
	if err != nil {
		return fmt.Errorf("AssignmentRepository.GradeSubmission: %w", err)
	}
	return nil
}

// ListSubmissionFiles — файлы нескольких сдач одним запросом, по порядку прикрепления.
func (r *AssignmentRepository) ListSubmissionFiles(ctx context.Context, submissionIDs []uuid.UUID) (map[uuid.UUID][]*domain.File, error) {
	out := make(map[uuid.UUID][]*domain.File, len(submissionIDs))
	if len(submissionIDs) == 0 {
		return out, nil
	}
	q, args, err := sqlx.In(`
		SELECT sf.submission_id, f.id, f.owner_id, f.purpose, f.object_key, f.original_name,
		       f.mime_type, f.size_bytes, f.sha256, f.width, f.height, f.status, f.created_at
		FROM submission_files sf
		JOIN files f ON f.id = sf.file_id
		WHERE sf.submission_id IN (?)
		ORDER BY sf.submission_id, sf.position`, submissionIDs)
	if err != nil {
		return nil, fmt.Errorf("AssignmentRepository.ListSubmissionFiles build: %w", err)
	}
	var rows []struct {
		SubmissionID uuid.UUID `db:"submission_id"`
		domain.File
	}
	if err := r.db.SelectContext(ctx, &rows, r.db.Rebind(q), args...); err != nil {
		return nil, fmt.Errorf("AssignmentRepository.ListSubmissionFiles: %w", err)
	}
	for i := range rows {
		f := rows[i].File
		out[rows[i].SubmissionID] = append(out[rows[i].SubmissionID], &f)
	}
	return out, nil
}
