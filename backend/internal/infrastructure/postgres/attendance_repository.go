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
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/features/attendance"
)

type AttendanceRepository struct {
	db *sqlx.DB
}

func NewAttendanceRepository(db *sqlx.DB) attendance.Repository {
	return &AttendanceRepository{db: db}
}

func (r *AttendanceRepository) Upsert(ctx context.Context, a *domain.Attendance) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO attendance (id, group_id, student_id, teacher_id, lesson_date, status, note, created_at, updated_at)
		VALUES (:id, :group_id, :student_id, :teacher_id, :lesson_date, :status, :note, :created_at, :updated_at)
		ON CONFLICT (group_id, student_id, lesson_date)
		DO UPDATE SET status=EXCLUDED.status, note=EXCLUDED.note, updated_at=NOW()`, a)
	if err != nil {
		return fmt.Errorf("AttendanceRepository.Upsert: %w", err)
	}
	return nil
}

func (r *AttendanceRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Attendance, error) {
	var a domain.Attendance
	err := r.db.GetContext(ctx, &a, `SELECT * FROM attendance WHERE id=$1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("AttendanceRepository.GetByID: %w", err)
	}
	return &a, nil
}

func (r *AttendanceRepository) ListByGroupDate(ctx context.Context, groupID uuid.UUID, date time.Time) ([]*domain.Attendance, error) {
	var list []*domain.Attendance
	err := r.db.SelectContext(ctx, &list,
		`SELECT * FROM attendance WHERE group_id=$1 AND lesson_date=$2 ORDER BY student_id`,
		groupID, date.Format("2006-01-02"))
	if err != nil {
		return nil, fmt.Errorf("AttendanceRepository.ListByGroupDate: %w", err)
	}
	return list, nil
}

func (r *AttendanceRepository) ListByGroupStudent(ctx context.Context, groupID, studentID uuid.UUID) ([]*domain.Attendance, error) {
	var list []*domain.Attendance
	err := r.db.SelectContext(ctx, &list,
		`SELECT * FROM attendance WHERE group_id=$1 AND student_id=$2 ORDER BY lesson_date DESC`,
		groupID, studentID)
	if err != nil {
		return nil, fmt.Errorf("AttendanceRepository.ListByGroupStudent: %w", err)
	}
	return list, nil
}

func (r *AttendanceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM attendance WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("AttendanceRepository.Delete: %w", err)
	}
	return nil
}
