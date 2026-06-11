package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/features/groups"
)

// GroupRepository implements groups.Repository using PostgreSQL.
type GroupRepository struct {
	db *sqlx.DB
}

func NewGroupRepository(db *sqlx.DB) groups.Repository {
	return &GroupRepository{db: db}
}

func (r *GroupRepository) CreateGroup(ctx context.Context, g *domain.Group) error {
	query := `
		INSERT INTO groups (id, teacher_id, name, subject, description, invite_code, is_archived, created_at, updated_at)
		VALUES (:id, :teacher_id, :name, :subject, :description, :invite_code, :is_archived, :created_at, :updated_at)`
	_, err := r.db.NamedExecContext(ctx, query, g)
	if err != nil {
		return fmt.Errorf("GroupRepository.CreateGroup: %w", err)
	}
	return nil
}

func (r *GroupRepository) GetGroupByID(ctx context.Context, id uuid.UUID) (*domain.Group, error) {
	var g domain.Group
	err := r.db.GetContext(ctx, &g, `SELECT * FROM groups WHERE id=$1 AND deleted_at IS NULL`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("GroupRepository.GetGroupByID: %w", err)
	}
	return &g, nil
}

func (r *GroupRepository) GetGroupByInviteCode(ctx context.Context, code string) (*domain.Group, error) {
	var g domain.Group
	err := r.db.GetContext(ctx, &g, `SELECT * FROM groups WHERE invite_code=$1 AND deleted_at IS NULL`, code)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("GroupRepository.GetGroupByInviteCode: %w", err)
	}
	return &g, nil
}

func (r *GroupRepository) ListTeacherGroups(ctx context.Context, teacherID uuid.UUID) ([]*domain.Group, error) {
	var gs []*domain.Group
	err := r.db.SelectContext(ctx, &gs,
		`SELECT * FROM groups WHERE teacher_id=$1 AND deleted_at IS NULL ORDER BY created_at DESC`,
		teacherID)
	if err != nil {
		return nil, fmt.Errorf("GroupRepository.ListTeacherGroups: %w", err)
	}
	return gs, nil
}

func (r *GroupRepository) ListStudentGroups(ctx context.Context, studentID uuid.UUID) ([]*domain.Group, error) {
	var gs []*domain.Group
	err := r.db.SelectContext(ctx, &gs, `
		SELECT g.*
		FROM groups g
		JOIN group_members gm ON gm.group_id = g.id
		WHERE gm.student_id=$1 AND g.deleted_at IS NULL
		ORDER BY gm.joined_at DESC`,
		studentID)
	if err != nil {
		return nil, fmt.Errorf("GroupRepository.ListStudentGroups: %w", err)
	}
	return gs, nil
}

func (r *GroupRepository) UpdateGroup(ctx context.Context, g *domain.Group) error {
	query := `
		UPDATE groups
		SET name=:name, subject=:subject, description=:description,
		    is_archived=:is_archived, updated_at=:updated_at
		WHERE id=:id AND deleted_at IS NULL`
	_, err := r.db.NamedExecContext(ctx, query, g)
	if err != nil {
		return fmt.Errorf("GroupRepository.UpdateGroup: %w", err)
	}
	return nil
}

func (r *GroupRepository) SoftDeleteGroup(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE groups SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("GroupRepository.SoftDeleteGroup: %w", err)
	}
	return nil
}

func (r *GroupRepository) AddMember(ctx context.Context, m *domain.GroupMember) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO group_members (group_id, student_id, joined_at, payment_status)
		VALUES (:group_id, :student_id, :joined_at, :payment_status)`,
		m)
	if err != nil {
		return fmt.Errorf("GroupRepository.AddMember: %w", err)
	}
	return nil
}

func (r *GroupRepository) RemoveMember(ctx context.Context, groupID, studentID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM group_members WHERE group_id=$1 AND student_id=$2`,
		groupID, studentID)
	if err != nil {
		return fmt.Errorf("GroupRepository.RemoveMember: %w", err)
	}
	return nil
}

func (r *GroupRepository) GetMember(ctx context.Context, groupID, studentID uuid.UUID) (*domain.GroupMember, error) {
	var m domain.GroupMember
	err := r.db.GetContext(ctx, &m,
		`SELECT * FROM group_members WHERE group_id=$1 AND student_id=$2`,
		groupID, studentID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("GroupRepository.GetMember: %w", err)
	}
	return &m, nil
}

func (r *GroupRepository) ListMembers(ctx context.Context, groupID uuid.UUID) ([]*domain.GroupMemberDetail, error) {
	var ms []*domain.GroupMemberDetail
	err := r.db.SelectContext(ctx, &ms, `
		SELECT
			gm.group_id,
			gm.student_id,
			gm.joined_at,
			gm.payment_status,
			u.name,
			u.phone,
			u.avatar_url
		FROM group_members gm
		JOIN users u ON u.id = gm.student_id
		WHERE gm.group_id=$1 AND u.deleted_at IS NULL
		ORDER BY gm.joined_at ASC`,
		groupID)
	if err != nil {
		return nil, fmt.Errorf("GroupRepository.ListMembers: %w", err)
	}
	return ms, nil
}

func (r *GroupRepository) UpdateMemberPayment(ctx context.Context, groupID, studentID uuid.UUID, status domain.PaymentStatus) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE group_members SET payment_status=$1 WHERE group_id=$2 AND student_id=$3`,
		status, groupID, studentID)
	if err != nil {
		return fmt.Errorf("GroupRepository.UpdateMemberPayment: %w", err)
	}
	return nil
}
