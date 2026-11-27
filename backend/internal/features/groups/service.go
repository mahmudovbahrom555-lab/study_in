package groups

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

// safeAlphabet is uppercase-only, excluding visually ambiguous chars: 0/O, 1/I.
// Codes stored and compared as uppercase, so users can type in any case.
const safeAlphabet = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"

const inviteCodeLen = 8

type Service struct {
	repo     Repository
	unread   UnreadCounter
	avatars  AvatarSigner
	notifier Notifier
}

// Notifier — уведомления в приложении (модуль notifications).
type Notifier interface {
	Send(ctx context.Context, userID uuid.UUID, title, body string, data any) error
}

// WithNotifier включает уведомление родителя о заявке ребёнка.
func (s *Service) WithNotifier(n Notifier) *Service {
	s.notifier = n
	return s
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

var errRequestDecided = domain.NewError("REQUEST_DECIDED", "Заявка уже рассмотрена", domain.ErrConflict)

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
		IsPublic:    req.IsPublic,
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
	if req.IsPublic != nil {
		g.IsPublic = *req.IsPublic
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
	g, err := s.joinableGroup(ctx, req.InviteCode, studentID)
	if err != nil {
		return nil, err
	}

	applicant, err := s.repo.GetApplicant(ctx, studentID)
	if err != nil {
		return nil, fmt.Errorf("groups.JoinGroup applicant: %w", err)
	}
	// Ребёнка младше 13 в группу записывает родитель: заявка сначала ждёт его.
	status := domain.JoinRequestPending
	var parents []domain.ParentRef
	if applicant.IsChild {
		status = domain.JoinRequestParentPending
		if parents, err = s.repo.ListParents(ctx, studentID); err != nil {
			return nil, fmt.Errorf("groups.JoinGroup parents: %w", err)
		}
	}

	jr, err := s.repo.UpsertJoinRequest(ctx, g.ID, studentID, status, nil)
	if err != nil {
		return nil, fmt.Errorf("groups.JoinGroup request: %w", err)
	}
	if len(parents) > 0 {
		s.notifyParents(ctx, parents, applicant.Name, g)
	}
	return &JoinRequestResponse{
		ID: jr.ID, GroupID: g.ID, GroupName: g.Name, Status: jr.Status,
		HasParent: len(parents) > 0,
	}, nil
}

// joinableGroup — группа по коду, если ученик может подать в неё заявку:
// группа есть и не в архиве, ученик в ней не состоит и не исключён.
func (s *Service) joinableGroup(ctx context.Context, code string, studentID uuid.UUID) (*domain.Group, error) {
	g, err := s.repo.GetGroupByInviteCode(ctx, normalizeCode(code))
	if err != nil {
		return nil, fmt.Errorf("groups.joinableGroup lookup: %w", err)
	}
	if g == nil {
		return nil, domain.ErrNotFound
	}
	if g.IsArchived {
		return nil, domain.ErrForbidden
	}

	existing, err := s.repo.GetMember(ctx, g.ID, studentID)
	if err != nil {
		return nil, fmt.Errorf("groups.joinableGroup check member: %w", err)
	}
	if existing != nil {
		return nil, domain.ErrAlreadyMember
	}
	banned, err := s.repo.IsBanned(ctx, g.ID, studentID)
	if err != nil {
		return nil, fmt.Errorf("groups.joinableGroup check ban: %w", err)
	}
	if banned {
		return nil, errBannedFromGroup
	}
	return g, nil
}

// EnrollChild — родитель записывает своего ребёнка по ссылке или коду группы.
// Согласие уже есть (заявку подал родитель), поэтому заявка сразу у репетитора
// с пометкой, кто из родителей её подал.
func (s *Service) EnrollChild(ctx context.Context, parentID uuid.UUID, req EnrollChildRequest) (*JoinRequestResponse, error) {
	parents, err := s.repo.ListParents(ctx, req.StudentID)
	if err != nil {
		return nil, fmt.Errorf("groups.EnrollChild parents: %w", err)
	}
	if !slices.ContainsFunc(parents, func(p domain.ParentRef) bool { return p.ID == parentID }) {
		return nil, domain.ErrNotFound // не свой ребёнок
	}
	g, err := s.joinableGroup(ctx, req.InviteCode, req.StudentID)
	if err != nil {
		return nil, err
	}
	jr, err := s.repo.UpsertJoinRequest(ctx, g.ID, req.StudentID, domain.JoinRequestPending, &parentID)
	if err != nil {
		return nil, fmt.Errorf("groups.EnrollChild request: %w", err)
	}
	return &JoinRequestResponse{ID: jr.ID, GroupID: g.ID, GroupName: g.Name, Status: jr.Status, HasParent: true}, nil
}

// notifyParents: ошибка уведомления не отменяет заявку — родитель увидит её
// и в разделе «Дети».
func (s *Service) notifyParents(ctx context.Context, parents []domain.ParentRef, child string, g *domain.Group) {
	if s.notifier == nil {
		return
	}
	teacher := ""
	if p, err := s.repo.GetGroupPreviewByCode(ctx, g.InviteCode); err == nil && p != nil {
		teacher = p.TeacherName
	}
	for _, p := range parents {
		title, body := consentMessage(p.Language, child, g.Name, teacher)
		_ = s.notifier.Send(ctx, p.ID, title, body, map[string]string{"type": "parent_consent"})
	}
}

// ParentConsents — заявки детей, ждущие согласия этого родителя.
func (s *Service) ParentConsents(ctx context.Context, parentID uuid.UUID) ([]ParentConsentItem, error) {
	list, err := s.repo.ListParentConsents(ctx, parentID)
	if err != nil {
		return nil, fmt.Errorf("groups.ParentConsents: %w", err)
	}
	out := make([]ParentConsentItem, len(list))
	for i, c := range list {
		out[i] = ParentConsentItem{
			ID: c.ID, StudentID: c.StudentID, StudentName: c.StudentName,
			GroupID: c.GroupID, GroupName: c.GroupName, Subject: c.Subject,
			TeacherName: c.TeacherName, CreatedAt: c.CreatedAt,
		}
	}
	return out, nil
}

// DecideParentConsent — родитель разрешает (заявка уходит репетитору) или
// отказывает. Решать может только привязанный к ребёнку родитель.
func (s *Service) DecideParentConsent(ctx context.Context, requestID, parentID uuid.UUID, approve bool) error {
	jr, err := s.repo.GetJoinRequest(ctx, requestID)
	if err != nil {
		return fmt.Errorf("groups.DecideParentConsent: %w", err)
	}
	if jr == nil {
		return domain.ErrNotFound
	}
	parents, err := s.repo.ListParents(ctx, jr.StudentID)
	if err != nil {
		return fmt.Errorf("groups.DecideParentConsent parents: %w", err)
	}
	if !slices.ContainsFunc(parents, func(p domain.ParentRef) bool { return p.ID == parentID }) {
		return domain.ErrNotFound // чужая заявка: не подтверждаем, что она есть
	}
	if jr.Status != domain.JoinRequestParentPending {
		return errRequestDecided
	}
	ok, err := s.repo.DecideParentConsent(ctx, requestID, parentID, approve)
	if err != nil {
		return fmt.Errorf("groups.DecideParentConsent update: %w", err)
	}
	if !ok {
		return errRequestDecided // второй родитель успел ответить раньше
	}
	return nil
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
	resp := s.previewResponse(ctx, p)
	return &resp, nil
}

func (s *Service) previewResponse(ctx context.Context, p *domain.GroupPreview) GroupPreviewResponse {
	return GroupPreviewResponse{
		ID:                  p.ID,
		Name:                p.Name,
		Subject:             p.Subject,
		Description:         p.Description,
		TeacherName:         p.TeacherName,
		TeacherAvatarURL:    s.signAvatar(ctx, p.TeacherAvatarURL),
		TeacherAvatarPreset: p.TeacherAvatarPreset,
		MemberCount:         p.MemberCount,
	}
}

// searchLimit — сколько групп отдаёт поиск: это подбор, а не каталог.
const searchLimit = 20

// SearchGroups ищет открытые группы: "@username" — группы этого репетитора,
// иначе — по названию или предмету (и по username, если запрос похож на него).
// Ищутся только группы и репетиторы, ученики в поиск не попадают.
func (s *Service) SearchGroups(ctx context.Context, query string) ([]GroupSearchItem, error) {
	q := strings.TrimSpace(query)
	username := ""
	pattern := ""
	if strings.HasPrefix(q, "@") {
		username = domain.NormalizeUsername(q)
	} else {
		if utf8.RuneCountInString(q) < 2 {
			return nil, domain.NewError("QUERY_TOO_SHORT", "query must be at least 2 characters", domain.ErrValidation)
		}
		pattern = "%" + likeEscaper.Replace(q) + "%"
		if domain.ValidUsername(strings.ToLower(q)) {
			username = strings.ToLower(q)
		}
	}
	if username == "" && pattern == "" {
		return []GroupSearchItem{}, nil
	}

	found, err := s.repo.SearchPublicGroups(ctx, pattern, username, searchLimit)
	if err != nil {
		return nil, fmt.Errorf("groups.SearchGroups: %w", err)
	}
	out := make([]GroupSearchItem, len(found))
	for i, p := range found {
		out[i] = GroupSearchItem{
			GroupPreviewResponse: s.previewResponse(ctx, p),
			InviteCode:           p.InviteCode,
			TeacherUsername:      p.TeacherUsername,
		}
	}
	return out, nil
}

// likeEscaper экранирует символы шаблона ILIKE в запросе пользователя.
var likeEscaper = strings.NewReplacer(`\`, `\\`, "%", `\%`, "_", `\_`)

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
			ParentName:   r.ApprovedByParentName,
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
		return errRequestDecided
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
