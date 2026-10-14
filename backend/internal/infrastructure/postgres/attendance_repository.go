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

const attendanceColumns = `id, group_id, student_id, teacher_id, lesson_date, status, note,
	created_at, updated_at`

// При повторной отметке за тот же день строка обновляется, и RETURNING отдаёт
// её настоящий id — а не новый uuid, сгенерированный в сервисе.
const upsertAttendance = `
	INSERT INTO attendance (id, group_id, student_id, teacher_id, lesson_date, status, note, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	ON CONFLICT (group_id, student_id, lesson_date)
	DO UPDATE SET status = EXCLUDED.status, note = EXCLUDED.note,
	              teacher_id = EXCLUDED.teacher_id, updated_at = NOW()
	RETURNING id, created_at, updated_at`

type AttendanceRepository struct {
	db *sqlx.DB
}

func NewAttendanceRepository(db *sqlx.DB) attendance.Repository {
	return &AttendanceRepository{db: db}
}

func (r *AttendanceRepository) Upsert(ctx context.Context, a *domain.Attendance) error {
	if err := upsertOne(ctx, r.db, a); err != nil {
		return fmt.Errorf("AttendanceRepository.Upsert: %w", err)
	}
	return nil
}

func (r *AttendanceRepository) UpsertMany(ctx context.Context, list []*domain.Attendance) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("AttendanceRepository.UpsertMany begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	for _, a := range list {
		if err := upsertOne(ctx, tx, a); err != nil {
			return fmt.Errorf("AttendanceRepository.UpsertMany: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("AttendanceRepository.UpsertMany commit: %w", err)
	}
	return nil
}

func upsertOne(ctx context.Context, q sqlx.QueryerContext, a *domain.Attendance) error {
	return q.QueryRowxContext(ctx, upsertAttendance,
		a.ID, a.GroupID, a.StudentID, a.TeacherID, a.LessonDate.Format("2006-01-02"),
		a.Status, a.Note, a.CreatedAt, a.UpdatedAt,
	).Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt)
}

func (r *AttendanceRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Attendance, error) {
	var a domain.Attendance
	err := r.db.GetContext(ctx, &a, `SELECT `+attendanceColumns+` FROM attendance WHERE id=$1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("AttendanceRepository.GetByID: %w", err)
	}
	return &a, nil
}

// Roster: нынешние ученики группы плюс те, кого уже нет в группе,
// но отметили в этот день, — чтобы отметка не «пропадала» из списка.
func (r *AttendanceRepository) Roster(ctx context.Context, groupID uuid.UUID, date time.Time) ([]*domain.RosterEntry, error) {
	var list []*domain.RosterEntry
	err := r.db.SelectContext(ctx, &list, `
		WITH students AS (
			SELECT student_id FROM group_members WHERE group_id = $1
			UNION
			SELECT student_id FROM attendance WHERE group_id = $1 AND lesson_date = $2
		)
		SELECT s.student_id, u.name AS student_name,
		       a.id AS attendance_id, a.status, a.note
		FROM students s
		JOIN users u ON u.id = s.student_id
		LEFT JOIN attendance a
		       ON a.group_id = $1 AND a.student_id = s.student_id AND a.lesson_date = $2
		ORDER BY lower(u.name), s.student_id`,
		groupID, date.Format("2006-01-02"))
	if err != nil {
		return nil, fmt.Errorf("AttendanceRepository.Roster: %w", err)
	}
	return list, nil
}

func (r *AttendanceRepository) ListByGroupStudent(ctx context.Context, groupID, studentID uuid.UUID) ([]*domain.Attendance, error) {
	var list []*domain.Attendance
	err := r.db.SelectContext(ctx, &list,
		`SELECT `+attendanceColumns+` FROM attendance
		 WHERE group_id=$1 AND student_id=$2 ORDER BY lesson_date DESC`,
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
