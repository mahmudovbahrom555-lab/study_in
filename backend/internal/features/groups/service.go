package groups

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

// safeAlphabet is uppercase-only, excluding visually ambiguous chars: 0/O, 1/I.
// Codes stored and compared as uppercase, so users can type in any case.
const safeAlphabet = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"

const inviteCodeLen = 8

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// CreateGroup creates a new group owned by the calling teacher.
func (s *Service) CreateGroup(ctx context.Context, teacherID uuid.UUID, req CreateGroupRequest) (*GroupResponse, error) {
	code, err := generateInviteCode()
	if err != nil {
		return nil, fmt.Errorf("groups.CreateGroup generate invite: %w", err)
	}

	now := time.Now()
	g := &domain.Group{
		ID:          uuid.New(),
		TeacherID:   teacherID,
		Name:        strings.TrimSpace(req.Name),
		Subject:     req.Subject,
		Description: req.Description,
		InviteCode:  code,
		IsArchived:  false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.CreateGroup(ctx, g); err != nil {
		return nil, fmt.Errorf("groups.CreateGroup: %w", err)
	}

	resp := groupToResponse(g)
	return &resp, nil
}

// ListMyGroups returns groups relevant to the caller.
// Teachers see their own groups; students see groups they joined.
func (s *Service) ListMyGroups(ctx context.Context, userID uuid.UUID, role domain.Role) ([]GroupResponse, error) {
	var gs []*domain.Group
	var err error

	switch role {
	case domain.RoleTeacher:
		gs, err = s.repo.ListTeacherGroups(ctx, userID)
	case domain.RoleStudent, domain.RoleParent:
		gs, err = s.repo.ListStudentGroups(ctx, userID)
	default:
		return nil, domain.ErrForbidden
	}
	if err != nil {
		return nil, fmt.Errorf("groups.ListMyGroups: %w", err)
	}

	out := make([]GroupResponse, len(gs))
	for i, g := range gs {
		out[i] = groupToResponse(g)
	}
	return out, nil
}

// GetGroup returns a group by ID, enforcing visibility.
func (s *Service) GetGroup(ctx context.Context, id, callerID uuid.UUID, role domain.Role) (*GroupResponse, error) {
	g, err := s.repo.GetGroupByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("groups.GetGroup: %w", err)
	}
	if g == nil {
		return nil, domain.ErrNotFound
	}

	if err := s.assertVisible(ctx, g, callerID, role); err != nil {
		return nil, err
	}

	resp := groupToResponse(g)
	return &resp, nil
}

// UpdateGroup allows the owning teacher to rename/describe their group.
func (s *Service) UpdateGroup(ctx context.Context, id, teacherID uuid.UUID, req UpdateGroupRequest) (*GroupResponse, error) {
	g, err := s.repo.GetGroupByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("groups.UpdateGroup fetch: %w", err)
	}
	if g == nil {
		return nil, domain.ErrNotFound
	}
	if g.TeacherID != teacherID {
		return nil, domain.ErrForbidden
	}

	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		g.Name = trimmed
	}
	if req.Subject != nil {
		g.Subject = req.Subject
	}
	if req.Description != nil {
		g.Description = req.Description
	}
	g.UpdatedAt = time.Now()

	if err := s.repo.UpdateGroup(ctx, g); err != nil {
		return nil, fmt.Errorf("groups.UpdateGroup save: %w", err)
	}

	resp := groupToResponse(g)
	return &resp, nil
}

// ArchiveGroup toggles the archived flag (teacher only).
func (s *Service) ArchiveGroup(ctx context.Context, id, teacherID uuid.UUID) (*GroupResponse, error) {
	g, err := s.repo.GetGroupByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("groups.ArchiveGroup fetch: %w", err)
	}
	if g == nil {
		return nil, domain.ErrNotFound
	}
	if g.TeacherID != teacherID {
		return nil, domain.ErrForbidden
	}

	g.IsArchived = !g.IsArchived
	g.UpdatedAt = time.Now()

	if err := s.repo.UpdateGroup(ctx, g); err != nil {
		return nil, fmt.Errorf("groups.ArchiveGroup save: %w", err)
	}

	resp := groupToResponse(g)
	return &resp, nil
}

// DeleteGroup soft-deletes a group (teacher only).
func (s *Service) DeleteGroup(ctx context.Context, id, teacherID uuid.UUID) error {
	g, err := s.repo.GetGroupByID(ctx, id)
	if err != nil {
		return fmt.Errorf("groups.DeleteGroup fetch: %w", err)
	}
	if g == nil {
		return domain.ErrNotFound
	}
	if g.TeacherID != teacherID {
		return domain.ErrForbidden
	}

	if err := s.repo.SoftDeleteGroup(ctx, id); err != nil {
		return fmt.Errorf("groups.DeleteGroup: %w", err)
	}
	return nil
}

// JoinGroup lets a student join via invite code.
func (s *Service) JoinGroup(ctx context.Context, studentID uuid.UUID, req JoinGroupRequest) (*GroupResponse, error) {
	g, err := s.repo.GetGroupByInviteCode(ctx, strings.ToUpper(strings.TrimSpace(req.InviteCode)))
	if err != nil {
		return nil, fmt.Errorf("groups.JoinGroup lookup: %w", err)
	}
	if g == nil {
		return nil, domain.ErrNotFound
	}
	if g.IsArchived {
		return nil, domain.ErrForbidden
	}

	existing, err := s.repo.GetMember(ctx, g.ID, studentID)
	if err != nil {
		return nil, fmt.Errorf("groups.JoinGroup check member: %w", err)
	}
	if existing != nil {
		return nil, domain.ErrAlreadyMember
	}

	m := &domain.GroupMember{
		GroupID:       g.ID,
		StudentID:     studentID,
		JoinedAt:      time.Now(),
		PaymentStatus: domain.PaymentStatusTrial,
	}
	if err := s.repo.AddMember(ctx, m); err != nil {
		return nil, fmt.Errorf("groups.JoinGroup add member: %w", err)
	}

	resp := groupToResponse(g)
	return &resp, nil
}

// ListMembers returns members of a group (teacher only, or members themselves).
func (s *Service) ListMembers(ctx context.Context, groupID, callerID uuid.UUID, role domain.Role) ([]MemberResponse, error) {
	g, err := s.repo.GetGroupByID(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("groups.ListMembers fetch group: %w", err)
	}
	if g == nil {
		return nil, domain.ErrNotFound
	}

	if err := s.assertVisible(ctx, g, callerID, role); err != nil {
		return nil, err
	}

	members, err := s.repo.ListMembers(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("groups.ListMembers: %w", err)
	}

	out := make([]MemberResponse, len(members))
	for i, m := range members {
		out[i] = memberToResponse(m)
	}
	return out, nil
}

// RemoveMember kicks a student (teacher only).
func (s *Service) RemoveMember(ctx context.Context, groupID, studentID, teacherID uuid.UUID) error {
	g, err := s.repo.GetGroupByID(ctx, groupID)
	if err != nil {
		return fmt.Errorf("groups.RemoveMember fetch group: %w", err)
	}
	if g == nil {
		return domain.ErrNotFound
	}
	if g.TeacherID != teacherID {
		return domain.ErrForbidden
	}

	existing, err := s.repo.GetMember(ctx, groupID, studentID)
	if err != nil {
		return fmt.Errorf("groups.RemoveMember check: %w", err)
	}
	if existing == nil {
		return domain.ErrNotMember
	}

	if err := s.repo.RemoveMember(ctx, groupID, studentID); err != nil {
		return fmt.Errorf("groups.RemoveMember: %w", err)
	}
	return nil
}

// LeaveGroup lets a student leave voluntarily.
func (s *Service) LeaveGroup(ctx context.Context, groupID, studentID uuid.UUID) error {
	existing, err := s.repo.GetMember(ctx, groupID, studentID)
	if err != nil {
		return fmt.Errorf("groups.LeaveGroup check: %w", err)
	}
	if existing == nil {
		return domain.ErrNotMember
	}

	if err := s.repo.RemoveMember(ctx, groupID, studentID); err != nil {
		return fmt.Errorf("groups.LeaveGroup: %w", err)
	}
	return nil
}

// UpdatePayment updates a student's payment status (teacher only).
func (s *Service) UpdatePayment(ctx context.Context, groupID, studentID, teacherID uuid.UUID, req UpdatePaymentRequest) error {
	g, err := s.repo.GetGroupByID(ctx, groupID)
	if err != nil {
		return fmt.Errorf("groups.UpdatePayment fetch group: %w", err)
	}
	if g == nil {
		return domain.ErrNotFound
	}
	if g.TeacherID != teacherID {
		return domain.ErrForbidden
	}

	existing, err := s.repo.GetMember(ctx, groupID, studentID)
	if err != nil {
		return fmt.Errorf("groups.UpdatePayment check member: %w", err)
	}
	if existing == nil {
		return domain.ErrNotMember
	}

	if err := s.repo.UpdateMemberPayment(ctx, groupID, studentID, req.Status); err != nil {
		return fmt.Errorf("groups.UpdatePayment save: %w", err)
	}
	return nil
}

// assertVisible checks whether the caller may see the group.
// Teacher sees their own groups; student/parent must be a member.
func (s *Service) assertVisible(ctx context.Context, g *domain.Group, callerID uuid.UUID, role domain.Role) error {
	if role == domain.RoleTeacher {
		if g.TeacherID != callerID {
			return domain.ErrForbidden
		}
		return nil
	}
	m, err := s.repo.GetMember(ctx, g.ID, callerID)
	if err != nil {
		return fmt.Errorf("assertVisible: %w", err)
	}
	if m == nil {
		return domain.ErrForbidden
	}
	return nil
}

// generateInviteCode returns a cryptographically random invite code from safeAlphabet.
func generateInviteCode() (string, error) {
	b := make([]byte, inviteCodeLen)
	alphabetLen := big.NewInt(int64(len(safeAlphabet)))
	for i := range b {
		n, err := rand.Int(rand.Reader, alphabetLen)
		if err != nil {
			return "", err
		}
		b[i] = safeAlphabet[n.Int64()]
	}
	return string(b), nil
}
