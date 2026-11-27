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
		INSERT INTO groups (id, teacher_id, name, subject, description, invite_code, is_archived, is_public, created_at, updated_at)
		VALUES (:id, :teacher_id, :name, :subject, :description, :invite_code, :is_archived, :is_public, :created_at, :updated_at)`
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
		    is_archived=:is_archived, is_public=:is_public, updated_at=:updated_at
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

func (r *GroupRepository) IsBanned(ctx context.Context, groupID, userID uuid.UUID) (bool, error) {
	var banned bool
	err := r.db.GetContext(ctx, &banned,
		`SELECT EXISTS (SELECT 1 FROM group_bans WHERE group_id = $1 AND user_id = $2)`, groupID, userID)
	if err != nil {
		return false, fmt.Errorf("GroupRepository.IsBanned: %w", err)
	}
	return banned, nil
}

func (r *GroupRepository) ListMembers(ctx context.Context, groupID uuid.UUID) ([]*domain.GroupMemberDetail, error) {
	var ms []*domain.GroupMemberDetail
	err := r.db.SelectContext(ctx, &ms, `
		SELECT
			gm.group_id,
			gm.student_id,
			gm.joined_at,
			gm.payment_status,
			gm.chat_muted,
			u.name,
			u.phone,
			u.avatar_url,
			u.avatar_preset
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

// --- Заявки на вступление ---

func (r *GroupRepository) GetGroupPreviewByCode(ctx context.Context, code string) (*domain.GroupPreview, error) {
	var p domain.GroupPreview
	err := r.db.GetContext(ctx, &p, `
		SELECT g.id, g.name, g.subject, g.description, g.is_archived,
		       u.name AS teacher_name,
		       u.avatar_url AS teacher_avatar_url,
		       u.avatar_preset AS teacher_avatar_preset,
		       (SELECT COUNT(*) FROM group_members m WHERE m.group_id = g.id) AS member_count
		FROM groups g
		JOIN users u ON u.id = g.teacher_id
		WHERE g.invite_code = $1 AND g.deleted_at IS NULL`, code)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("GroupRepository.GetGroupPreviewByCode: %w", err)
	}
	return &p, nil
}

// SearchPublicGroups ищет открытые группы по названию или предмету (pattern —
// готовый шаблон ILIKE, "" — не искать) или по @username репетитора.
func (r *GroupRepository) SearchPublicGroups(ctx context.Context, pattern, username string, limit int) ([]*domain.GroupPreview, error) {
	var out []*domain.GroupPreview
	err := r.db.SelectContext(ctx, &out, `
		SELECT g.id, g.name, g.subject, g.description, g.is_archived, g.invite_code,
		       u.name AS teacher_name,
		       u.username AS teacher_username,
		       u.avatar_url AS teacher_avatar_url,
		       u.avatar_preset AS teacher_avatar_preset,
		       (SELECT COUNT(*) FROM group_members m WHERE m.group_id = g.id) AS member_count
		FROM groups g
		JOIN users u ON u.id = g.teacher_id
		WHERE g.is_public AND NOT g.is_archived AND g.deleted_at IS NULL
		  AND u.deleted_at IS NULL
		  AND (($1 <> '' AND (g.name ILIKE $1 ESCAPE '\' OR g.subject ILIKE $1 ESCAPE '\'))
		       OR ($2 <> '' AND u.username = $2))
		ORDER BY g.created_at DESC
		LIMIT $3`, pattern, username, limit)
	if err != nil {
		return nil, fmt.Errorf("GroupRepository.SearchPublicGroups: %w", err)
	}
	return out, nil
}

func (r *GroupRepository) UpsertJoinRequest(ctx context.Context, groupID, studentID uuid.UUID, status domain.JoinRequestStatus, approvedBy *uuid.UUID) (*domain.GroupJoinRequest, error) {
	var jr domain.GroupJoinRequest
	err := r.db.GetContext(ctx, &jr, `
		INSERT INTO group_join_requests (group_id, student_id, status, approved_by_parent) VALUES ($1, $2, $3, $4)
		ON CONFLICT (group_id, student_id) DO UPDATE
		    SET status = $3, created_at = NOW(), decided_at = NULL, approved_by_parent = $4
		RETURNING *`, groupID, studentID, status, approvedBy)
	if err != nil {
		return nil, fmt.Errorf("GroupRepository.UpsertJoinRequest: %w", err)
	}
	return &jr, nil
}

func (r *GroupRepository) GetJoinRequest(ctx context.Context, id uuid.UUID) (*domain.GroupJoinRequest, error) {
	var jr domain.GroupJoinRequest
	err := r.db.GetContext(ctx, &jr, `SELECT * FROM group_join_requests WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("GroupRepository.GetJoinRequest: %w", err)
	}
	return &jr, nil
}

func (r *GroupRepository) ListPendingJoinRequests(ctx context.Context, groupID uuid.UUID) ([]*domain.JoinRequestDetail, error) {
	var out []*domain.JoinRequestDetail
	err := r.db.SelectContext(ctx, &out, `
		SELECT jr.*, u.name AS student_name, u.avatar_url, u.avatar_preset,
		       p.name AS approved_by_parent_name
		FROM group_join_requests jr
		JOIN users u ON u.id = jr.student_id
		LEFT JOIN users p ON p.id = jr.approved_by_parent
		WHERE jr.group_id = $1 AND jr.status = 'pending'
		ORDER BY jr.created_at`, groupID)
	if err != nil {
		return nil, fmt.Errorf("GroupRepository.ListPendingJoinRequests: %w", err)
	}
	return out, nil
}

func (r *GroupRepository) ListMyPendingJoinRequests(ctx context.Context, studentID uuid.UUID) ([]*domain.MyJoinRequest, error) {
	var out []*domain.MyJoinRequest
	err := r.db.SelectContext(ctx, &out, `
		SELECT jr.*, g.name AS group_name
		FROM group_join_requests jr
		JOIN groups g ON g.id = jr.group_id AND g.deleted_at IS NULL
		WHERE jr.student_id = $1 AND jr.status IN ('pending', 'parent_pending')
		ORDER BY jr.created_at DESC`, studentID)
	if err != nil {
		return nil, fmt.Errorf("GroupRepository.ListMyPendingJoinRequests: %w", err)
	}
	return out, nil
}

func (r *GroupRepository) GetApplicant(ctx context.Context, studentID uuid.UUID) (*domain.Applicant, error) {
	var a domain.Applicant
	err := r.db.GetContext(ctx, &a, `SELECT name, is_child FROM users WHERE id = $1`, studentID)
	if err != nil {
		return nil, fmt.Errorf("GroupRepository.GetApplicant: %w", err)
	}
	return &a, nil
}

func (r *GroupRepository) ListParents(ctx context.Context, studentID uuid.UUID) ([]domain.ParentRef, error) {
	var out []domain.ParentRef
	err := r.db.SelectContext(ctx, &out, `
		SELECT u.id, u.language
		FROM parent_links pl
		JOIN users u ON u.id = pl.parent_id
		WHERE pl.student_id = $1`, studentID)
	if err != nil {
		return nil, fmt.Errorf("GroupRepository.ListParents: %w", err)
	}
	return out, nil
}

func (r *GroupRepository) ListParentConsents(ctx context.Context, parentID uuid.UUID) ([]*domain.ParentConsent, error) {
	var out []*domain.ParentConsent
	err := r.db.SelectContext(ctx, &out, `
		SELECT jr.id, jr.student_id, s.name AS student_name, jr.group_id,
		       g.name AS group_name, g.subject, t.name AS teacher_name, jr.created_at
		FROM group_join_requests jr
		JOIN parent_links pl ON pl.student_id = jr.student_id AND pl.parent_id = $1
		JOIN users s ON s.id = jr.student_id
		JOIN groups g ON g.id = jr.group_id AND g.deleted_at IS NULL
		JOIN users t ON t.id = g.teacher_id
		WHERE jr.status = 'parent_pending'
		ORDER BY jr.created_at DESC`, parentID)
	if err != nil {
		return nil, fmt.Errorf("GroupRepository.ListParentConsents: %w", err)
	}
	return out, nil
}

// DecideParentConsent переводит заявку из parent_pending: разрешил — к
// репетитору (pending), отказал — declined. false — заявка уже не ждала родителя.
func (r *GroupRepository) DecideParentConsent(ctx context.Context, requestID, parentID uuid.UUID, approve bool) (bool, error) {
	res, err := r.db.ExecContext(ctx, `
		UPDATE group_join_requests
		SET status = CASE WHEN $3 THEN 'pending' ELSE 'declined' END,
		    approved_by_parent = CASE WHEN $3 THEN $2::uuid END,
		    decided_at = CASE WHEN $3 THEN NULL ELSE NOW() END
		WHERE id = $1 AND status = 'parent_pending'
		  AND EXISTS (SELECT 1 FROM parent_links pl
		              WHERE pl.parent_id = $2 AND pl.student_id = group_join_requests.student_id)`,
		requestID, parentID, approve)
	if err != nil {
		return false, fmt.Errorf("GroupRepository.DecideParentConsent: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("GroupRepository.DecideParentConsent rows: %w", err)
	}
	return n == 1, nil
}

func (r *GroupRepository) AcceptJoinRequest(ctx context.Context, requestID uuid.UUID, m *domain.GroupMember) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("GroupRepository.AcceptJoinRequest begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }() // после Commit — no-op

	if _, err := tx.NamedExecContext(ctx, `
		INSERT INTO group_members (group_id, student_id, joined_at, payment_status)
		VALUES (:group_id, :student_id, :joined_at, :payment_status)
		ON CONFLICT DO NOTHING`, m); err != nil {
		return fmt.Errorf("GroupRepository.AcceptJoinRequest member: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE group_join_requests SET status = 'accepted', decided_at = NOW()
		WHERE id = $1`, requestID); err != nil {
		return fmt.Errorf("GroupRepository.AcceptJoinRequest status: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("GroupRepository.AcceptJoinRequest commit: %w", err)
	}
	return nil
}

func (r *GroupRepository) DeclineJoinRequest(ctx context.Context, requestID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE group_join_requests SET status = 'declined', decided_at = NOW()
		WHERE id = $1`, requestID)
	if err != nil {
		return fmt.Errorf("GroupRepository.DeclineJoinRequest: %w", err)
	}
	return nil
}
