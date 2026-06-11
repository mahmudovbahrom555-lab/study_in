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

func (r *AssignmentRepository) CreateSubmission(ctx context.Context, s *domain.Submission) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO submissions (id, assignment_id, student_id, comment, submitted_at)
		VALUES (:id, :assignment_id, :student_id, :comment, :submitted_at)
		ON CONFLICT (assignment_id, student_id) DO UPDATE
		SET comment=EXCLUDED.comment, submitted_at=EXCLUDED.submitted_at`, s)
	if err != nil {
		return fmt.Errorf("AssignmentRepository.CreateSubmission: %w", err)
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

func (r *AssignmentRepository) ListSubmissions(ctx context.Context, assignmentID uuid.UUID) ([]*domain.Submission, error) {
	var subs []*domain.Submission
	err := r.db.SelectContext(ctx, &subs,
		`SELECT * FROM submissions WHERE assignment_id=$1 ORDER BY submitted_at ASC`, assignmentID)
	if err != nil {
		return nil, fmt.Errorf("AssignmentRepository.ListSubmissions: %w", err)
	}
	return subs, nil
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

func (r *AssignmentRepository) AddSubmissionAttachment(ctx context.Context, a *domain.SubmissionAttachment) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO submission_attachments (id, submission_id, object_key, filename, mime_type, size_bytes, created_at)
		VALUES (:id, :submission_id, :object_key, :filename, :mime_type, :size_bytes, :created_at)`, a)
	if err != nil {
		return fmt.Errorf("AssignmentRepository.AddSubmissionAttachment: %w", err)
	}
	return nil
}

func (r *AssignmentRepository) ListSubmissionAttachments(ctx context.Context, submissionID uuid.UUID) ([]*domain.SubmissionAttachment, error) {
	var atts []*domain.SubmissionAttachment
	err := r.db.SelectContext(ctx, &atts,
		`SELECT * FROM submission_attachments WHERE submission_id=$1 ORDER BY created_at ASC`, submissionID)
	if err != nil {
		return nil, fmt.Errorf("AssignmentRepository.ListSubmissionAttachments: %w", err)
	}
	return atts, nil
}
