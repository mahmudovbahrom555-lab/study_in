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
	repo    Repository
	unread  UnreadCounter
	avatars AvatarSigner
}

// UnreadCounter — непрочитанные сообщения чата по классам (модуль chat).
type UnreadCounter interface {
	UnreadCounts(ctx context.Context, userID uuid.UUID, groupIDs []uuid.UUID) (map[uuid.UUID]int, error)
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// WithUnread добавляет число непрочитанных сообщений в список классов.
// AvatarSigner подписывает ссылки на фото профиля (ключ наружу не отдаём).
type AvatarSigner interface {
	PresignedGetURL(ctx context.Context, objectKey, filename string) (string, error)
}

// WithAvatarSigner включает фото участников в списке.
func (s *Service) WithAvatarSigner(a AvatarSigner) *Service {
	s.avatars = a
	return s
}

func (s *Service) WithUnread(u UnreadCounter) *Service {
	s.unread = u
	return s
}

var errBannedFromGroup = domain.NewError("BANNED_FROM_GROUP", "you were removed from this class", domain.ErrForbidden)

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
		// Те же значения, что DEFAULT в БД, — чтобы ответ совпадал с сохранённым.
		ChatTTLSeconds: domain.DefaultChatTTL,
		ChatMode:       domain.ChatModeOpen,
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
	ids := make([]uuid.UUID, len(gs))
	for i, g := range gs {
		out[i] = groupToResponse(g)
		ids[i] = g.ID
	}
	if s.unread != nil && len(ids) > 0 {
		// Счётчик — подсказка: если он не посчитался, список всё равно нужен.
		if counts, err := s.unread.UnreadCounts(ctx, userID, ids); err == nil {
			for i := range out {
				out[i].UnreadCount = counts[out[i].ID]
			}
		}
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
	// Пустая строка — поле очищено в форме.
	if req.Subject != nil {
		g.Subject = nilIfBlank(*req.Subject)
	}
	if req.Description != nil {
		g.Description = nilIfBlank(*req.Description)
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

// JoinGroup — заявка ученика в группу по коду или ссылке. В группу ученик
// попадает, только когда репетитор примет заявку (решение владельцев,
// 2026-11-26). Повторная заявка возвращает ту же ожидающую.
func (s *Service) JoinGroup(ctx context.Context, studentID uuid.UUID, req JoinGroupRequest) (*JoinRequestResponse, error) {
	g, err := s.repo.GetGroupByInviteCode(ctx, normalizeCode(req.InviteCode))
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
	banned, err := s.repo.IsBanned(ctx, g.ID, studentID)
	if err != nil {
		return nil, fmt.Errorf("groups.JoinGroup check ban: %w", err)
	}
	if banned {
		return nil, errBannedFromGroup
	}

	jr, err := s.repo.UpsertJoinRequest(ctx, g.ID, studentID)
	if err != nil {
		return nil, fmt.Errorf("groups.JoinGroup request: %w", err)
	}
	return &JoinRequestResponse{ID: jr.ID, GroupID: g.ID, GroupName: g.Name, Status: jr.Status}, nil
}

// PreviewGroup — карточка группы по коду: показывается ученику до заявки.
func (s *Service) PreviewGroup(ctx context.Context, code string) (*GroupPreviewResponse, error) {
	p, err := s.repo.GetGroupPreviewByCode(ctx, normalizeCode(code))
	if err != nil {
		return nil, fmt.Errorf("groups.PreviewGroup: %w", err)
	}
	if p == nil || p.IsArchived {
		return nil, domain.ErrNotFound
	}
	return &GroupPreviewResponse{
		ID:                  p.ID,
		Name:                p.Name,
		Subject:             p.Subject,
		Description:         p.Description,
		TeacherName:         p.TeacherName,
		TeacherAvatarURL:    s.signAvatar(ctx, p.TeacherAvatarURL),
		TeacherAvatarPreset: p.TeacherAvatarPreset,
		MemberCount:         p.MemberCount,
	}, nil
}

// ListJoinRequests — ожидающие заявки (только репетитор группы).
func (s *Service) ListJoinRequests(ctx context.Context, groupID, teacherID uuid.UUID) ([]JoinRequestItem, error) {
	if err := s.assertOwner(ctx, groupID, teacherID); err != nil {
		return nil, err
	}
	reqs, err := s.repo.ListPendingJoinRequests(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("groups.ListJoinRequests: %w", err)
	}
	out := make([]JoinRequestItem, len(reqs))
	for i, r := range reqs {
		out[i] = JoinRequestItem{
			ID:           r.ID,
			StudentID:    r.StudentID,
			StudentName:  r.StudentName,
			AvatarURL:    s.signAvatar(ctx, r.AvatarURL),
			AvatarPreset: r.AvatarPreset,
			CreatedAt:    r.CreatedAt,
		}
	}
	return out, nil
}

// MyJoinRequests — свои ожидающие заявки ученика.
func (s *Service) MyJoinRequests(ctx context.Context, studentID uuid.UUID) ([]JoinRequestResponse, error) {
	reqs, err := s.repo.ListMyPendingJoinRequests(ctx, studentID)
	if err != nil {
		return nil, fmt.Errorf("groups.MyJoinRequests: %w", err)
	}
	out := make([]JoinRequestResponse, len(reqs))
	for i, r := range reqs {
		out[i] = JoinRequestResponse{ID: r.ID, GroupID: r.GroupID, GroupName: r.GroupName, Status: r.Status}
	}
	return out, nil
}

// DecideJoinRequest — репетитор принимает или отклоняет заявку.
func (s *Service) DecideJoinRequest(ctx context.Context, groupID, requestID, teacherID uuid.UUID, accept bool) error {
	if err := s.assertOwner(ctx, groupID, teacherID); err != nil {
		return err
	}
	jr, err := s.repo.GetJoinRequest(ctx, requestID)
	if err != nil {
		return fmt.Errorf("groups.DecideJoinRequest: %w", err)
	}
	if jr == nil || jr.GroupID != groupID {
		return domain.ErrNotFound
	}
	if jr.Status != domain.JoinRequestPending {
		return domain.NewError("REQUEST_DECIDED", "Заявка уже рассмотрена", domain.ErrConflict)
	}
	if !accept {
		if err := s.repo.DeclineJoinRequest(ctx, requestID); err != nil {
			return fmt.Errorf("groups.DecideJoinRequest decline: %w", err)
		}
		return nil
	}
	m := &domain.GroupMember{
		GroupID:       groupID,
		StudentID:     jr.StudentID,
		JoinedAt:      time.Now(),
		PaymentStatus: domain.PaymentStatusTrial,
	}
	if err := s.repo.AcceptJoinRequest(ctx, requestID, m); err != nil {
		return fmt.Errorf("groups.DecideJoinRequest accept: %w", err)
	}
	return nil
}

func (s *Service) assertOwner(ctx context.Context, groupID, teacherID uuid.UUID) error {
	g, err := s.repo.GetGroupByID(ctx, groupID)
	if err != nil {
		return fmt.Errorf("groups.assertOwner: %w", err)
	}
	if g == nil {
		return domain.ErrNotFound
	}
	if g.TeacherID != teacherID {
		return domain.ErrForbidden
	}
	return nil
}

func normalizeCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
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

	// Телефоны и статус оплаты учеников (это дети) видит только репетитор группы;
	// одноклассникам — только имена.
	isOwner := g.TeacherID == callerID
	out := make([]MemberResponse, len(members))
	for i, m := range members {
		out[i] = memberToResponse(m)
		out[i].AvatarURL = s.signAvatar(ctx, m.AvatarURL)
		if !isOwner {
			out[i].Phone = ""
			out[i].PaymentStatus = ""
			out[i].ChatMuted = false
		}
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

// nilIfBlank обрезает пробелы; пустое значение хранится как NULL.
func nilIfBlank(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}

func (s *Service) signAvatar(ctx context.Context, key *string) *string {
	if key == nil || s.avatars == nil {
		return nil
	}
	url, err := s.avatars.PresignedGetURL(ctx, *key, "")
	if err != nil {
		return nil
	}
	return &url
}
