package groups_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/features/groups"
)

// --- in-memory repository ---

type memRepo struct {
	groups   map[uuid.UUID]*domain.Group
	members  map[string]*domain.GroupMember // "groupID:studentID" → member
	bans     map[string]bool
	requests map[uuid.UUID]*domain.GroupJoinRequest
}

func newMemRepo() *memRepo {
	return &memRepo{
		groups:   make(map[uuid.UUID]*domain.Group),
		members:  make(map[string]*domain.GroupMember),
		bans:     make(map[string]bool),
		requests: make(map[uuid.UUID]*domain.GroupJoinRequest),
	}
}

func (r *memRepo) GetGroupPreviewByCode(_ context.Context, code string) (*domain.GroupPreview, error) {
	for _, g := range r.groups {
		if g.InviteCode == code && g.DeletedAt == nil {
			return &domain.GroupPreview{ID: g.ID, Name: g.Name, IsArchived: g.IsArchived, TeacherName: "Teacher"}, nil
		}
	}
	return nil, nil
}

func (r *memRepo) UpsertJoinRequest(_ context.Context, groupID, studentID uuid.UUID) (*domain.GroupJoinRequest, error) {
	for _, jr := range r.requests {
		if jr.GroupID == groupID && jr.StudentID == studentID {
			jr.Status = domain.JoinRequestPending
			cp := *jr
			return &cp, nil
		}
	}
	jr := &domain.GroupJoinRequest{ID: uuid.New(), GroupID: groupID, StudentID: studentID, Status: domain.JoinRequestPending}
	r.requests[jr.ID] = jr
	cp := *jr
	return &cp, nil
}

func (r *memRepo) GetJoinRequest(_ context.Context, id uuid.UUID) (*domain.GroupJoinRequest, error) {
	jr, ok := r.requests[id]
	if !ok {
		return nil, nil
	}
	cp := *jr
	return &cp, nil
}

func (r *memRepo) ListPendingJoinRequests(_ context.Context, groupID uuid.UUID) ([]*domain.JoinRequestDetail, error) {
	var out []*domain.JoinRequestDetail
	for _, jr := range r.requests {
		if jr.GroupID == groupID && jr.Status == domain.JoinRequestPending {
			out = append(out, &domain.JoinRequestDetail{GroupJoinRequest: *jr, StudentName: "Student"})
		}
	}
	return out, nil
}

func (r *memRepo) ListMyPendingJoinRequests(_ context.Context, studentID uuid.UUID) ([]*domain.MyJoinRequest, error) {
	var out []*domain.MyJoinRequest
	for _, jr := range r.requests {
		if jr.StudentID == studentID && jr.Status == domain.JoinRequestPending {
			out = append(out, &domain.MyJoinRequest{GroupJoinRequest: *jr, GroupName: r.groups[jr.GroupID].Name})
		}
	}
	return out, nil
}

func (r *memRepo) AcceptJoinRequest(_ context.Context, requestID uuid.UUID, m *domain.GroupMember) error {
	cp := *m
	r.members[memberKey(m.GroupID, m.StudentID)] = &cp
	r.requests[requestID].Status = domain.JoinRequestAccepted
	return nil
}

func (r *memRepo) DeclineJoinRequest(_ context.Context, requestID uuid.UUID) error {
	r.requests[requestID].Status = domain.JoinRequestDeclined
	return nil
}

func memberKey(gid, sid uuid.UUID) string {
	return gid.String() + ":" + sid.String()
}

func (r *memRepo) CreateGroup(_ context.Context, g *domain.Group) error {
	cp := *g
	r.groups[g.ID] = &cp
	return nil
}

func (r *memRepo) GetGroupByID(_ context.Context, id uuid.UUID) (*domain.Group, error) {
	g, ok := r.groups[id]
	if !ok || g.DeletedAt != nil {
		return nil, nil
	}
	cp := *g
	return &cp, nil
}

func (r *memRepo) GetGroupByInviteCode(_ context.Context, code string) (*domain.Group, error) {
	for _, g := range r.groups {
		if g.InviteCode == code && g.DeletedAt == nil {
			cp := *g
			return &cp, nil
		}
	}
	return nil, nil
}

func (r *memRepo) ListTeacherGroups(_ context.Context, teacherID uuid.UUID) ([]*domain.Group, error) {
	var out []*domain.Group
	for _, g := range r.groups {
		if g.TeacherID == teacherID && g.DeletedAt == nil {
			cp := *g
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (r *memRepo) ListStudentGroups(_ context.Context, studentID uuid.UUID) ([]*domain.Group, error) {
	var out []*domain.Group
	for _, m := range r.members {
		if m.StudentID == studentID {
			if g := r.groups[m.GroupID]; g != nil && g.DeletedAt == nil {
				cp := *g
				out = append(out, &cp)
			}
		}
	}
	return out, nil
}

func (r *memRepo) UpdateGroup(_ context.Context, g *domain.Group) error {
	if _, ok := r.groups[g.ID]; !ok {
		return domain.ErrNotFound
	}
	cp := *g
	r.groups[g.ID] = &cp
	return nil
}

func (r *memRepo) SoftDeleteGroup(_ context.Context, id uuid.UUID) error {
	if g, ok := r.groups[id]; ok {
		now := time.Now()
		g.DeletedAt = &now
	}
	return nil
}

func (r *memRepo) AddMember(_ context.Context, m *domain.GroupMember) error {
	cp := *m
	r.members[memberKey(m.GroupID, m.StudentID)] = &cp
	return nil
}

func (r *memRepo) RemoveMember(_ context.Context, groupID, studentID uuid.UUID) error {
	delete(r.members, memberKey(groupID, studentID))
	return nil
}

func (r *memRepo) GetMember(_ context.Context, groupID, studentID uuid.UUID) (*domain.GroupMember, error) {
	m, ok := r.members[memberKey(groupID, studentID)]
	if !ok {
		return nil, nil
	}
	cp := *m
	return &cp, nil
}

func (r *memRepo) ListMembers(_ context.Context, groupID uuid.UUID) ([]*domain.GroupMemberDetail, error) {
	var out []*domain.GroupMemberDetail
	for _, m := range r.members {
		if m.GroupID == groupID {
			phone := "+998901234567"
			out = append(out, &domain.GroupMemberDetail{
				GroupMember: *m,
				Name:        "Test Student",
				Phone:       &phone,
			})
		}
	}
	return out, nil
}

func (r *memRepo) UpdateMemberPayment(_ context.Context, groupID, studentID uuid.UUID, status domain.PaymentStatus) error {
	key := memberKey(groupID, studentID)
	if m, ok := r.members[key]; ok {
		m.PaymentStatus = status
	}
	return nil
}

func (r *memRepo) IsBanned(_ context.Context, groupID, userID uuid.UUID) (bool, error) {
	return r.bans[memberKey(groupID, userID)], nil
}

// --- helpers ---

func buildSvc() *groups.Service {
	return groups.NewService(newMemRepo())
}

// join — заявка ученика и её одобрение репетитором (так ученик попадает в группу).
func join(t *testing.T, svc *groups.Service, g *groups.GroupResponse, sid uuid.UUID) {
	t.Helper()
	jr, err := svc.JoinGroup(context.Background(), sid, groups.JoinGroupRequest{InviteCode: g.InviteCode})
	require.NoError(t, err)
	require.NoError(t, svc.DecideJoinRequest(context.Background(), g.ID, jr.ID, teacherID, true))
}

var (
	teacherID = uuid.New()
	studentID = uuid.New()
)

// --- tests ---

func TestCreateGroup_Success(t *testing.T) {
	svc := buildSvc()
	req := groups.CreateGroupRequest{Name: "Математика 7А"}
	resp, err := svc.CreateGroup(context.Background(), teacherID, req)
	require.NoError(t, err)
	assert.Equal(t, "Математика 7А", resp.Name)
	assert.Equal(t, teacherID, resp.TeacherID)
	assert.Len(t, resp.InviteCode, 8)
}

func TestCreateGroup_InviteCodeUnique(t *testing.T) {
	svc := buildSvc()
	codes := map[string]bool{}
	for i := 0; i < 20; i++ {
		r, err := svc.CreateGroup(context.Background(), teacherID, groups.CreateGroupRequest{Name: "G"})
		require.NoError(t, err)
		codes[r.InviteCode] = true
	}
	// All 20 must be unique (collision probability negligible with 8-char safe alphabet)
	assert.Len(t, codes, 20)
}

func TestListMyGroups_Teacher(t *testing.T) {
	svc := buildSvc()
	_, _ = svc.CreateGroup(context.Background(), teacherID, groups.CreateGroupRequest{Name: "G1"})
	_, _ = svc.CreateGroup(context.Background(), teacherID, groups.CreateGroupRequest{Name: "G2"})

	list, err := svc.ListMyGroups(context.Background(), teacherID, domain.RoleTeacher)
	require.NoError(t, err)
	assert.Len(t, list, 2)
}

// Телефоны и оплату учеников (детей) видит только репетитор, не одноклассники.
func TestListMembers_PrivateFieldsOnlyForOwner(t *testing.T) {
	svc := buildSvc()
	g, _ := svc.CreateGroup(context.Background(), teacherID, groups.CreateGroupRequest{Name: "Physics"})
	join(t, svc, g, studentID)

	asTeacher, err := svc.ListMembers(context.Background(), g.ID, teacherID, domain.RoleTeacher)
	require.NoError(t, err)
	require.Len(t, asTeacher, 1)
	assert.NotEmpty(t, asTeacher[0].Phone)
	assert.NotEmpty(t, asTeacher[0].PaymentStatus)

	asStudent, err := svc.ListMembers(context.Background(), g.ID, studentID, domain.RoleStudent)
	require.NoError(t, err)
	require.Len(t, asStudent, 1)
	assert.NotEmpty(t, asStudent[0].Name)
	assert.Empty(t, asStudent[0].Phone)
	assert.Empty(t, asStudent[0].PaymentStatus)
}

func TestJoinGroup_Success(t *testing.T) {
	svc := buildSvc()
	g, _ := svc.CreateGroup(context.Background(), teacherID, groups.CreateGroupRequest{Name: "Physics"})

	resp, err := svc.JoinGroup(context.Background(), studentID, groups.JoinGroupRequest{InviteCode: g.InviteCode})
	require.NoError(t, err)
	assert.Equal(t, g.ID, resp.GroupID)
	assert.Equal(t, domain.JoinRequestPending, resp.Status)

	// До одобрения ученик не в группе; заявка видна репетитору и ученику.
	members, err := svc.ListMembers(context.Background(), g.ID, teacherID, domain.RoleTeacher)
	require.NoError(t, err)
	assert.Empty(t, members)
	reqs, err := svc.ListJoinRequests(context.Background(), g.ID, teacherID)
	require.NoError(t, err)
	require.Len(t, reqs, 1)
	mine, err := svc.MyJoinRequests(context.Background(), studentID)
	require.NoError(t, err)
	require.Len(t, mine, 1)
	assert.Equal(t, "Physics", mine[0].GroupName)

	require.NoError(t, svc.DecideJoinRequest(context.Background(), g.ID, resp.ID, teacherID, true))
	members, _ = svc.ListMembers(context.Background(), g.ID, teacherID, domain.RoleTeacher)
	assert.Len(t, members, 1)
	reqs, _ = svc.ListJoinRequests(context.Background(), g.ID, teacherID)
	assert.Empty(t, reqs)
}

func TestJoinRequest_DeclineThenRequestAgain(t *testing.T) {
	svc := buildSvc()
	ctx := context.Background()
	g, _ := svc.CreateGroup(ctx, teacherID, groups.CreateGroupRequest{Name: "Chem"})
	jr, _ := svc.JoinGroup(ctx, studentID, groups.JoinGroupRequest{InviteCode: g.InviteCode})

	require.NoError(t, svc.DecideJoinRequest(ctx, g.ID, jr.ID, teacherID, false))
	members, _ := svc.ListMembers(ctx, g.ID, teacherID, domain.RoleTeacher)
	assert.Empty(t, members)

	again, err := svc.JoinGroup(ctx, studentID, groups.JoinGroupRequest{InviteCode: g.InviteCode})
	require.NoError(t, err)
	assert.Equal(t, domain.JoinRequestPending, again.Status)
}

func TestJoinRequest_OnlyOwnerDecides(t *testing.T) {
	svc := buildSvc()
	ctx := context.Background()
	g, _ := svc.CreateGroup(ctx, teacherID, groups.CreateGroupRequest{Name: "Geo"})
	jr, _ := svc.JoinGroup(ctx, studentID, groups.JoinGroupRequest{InviteCode: g.InviteCode})

	err := svc.DecideJoinRequest(ctx, g.ID, jr.ID, uuid.New(), true)
	assert.ErrorIs(t, err, domain.ErrForbidden)
	_, err = svc.ListJoinRequests(ctx, g.ID, studentID)
	assert.ErrorIs(t, err, domain.ErrForbidden)

	require.NoError(t, svc.DecideJoinRequest(ctx, g.ID, jr.ID, teacherID, true))
	err = svc.DecideJoinRequest(ctx, g.ID, jr.ID, teacherID, false)
	assert.ErrorIs(t, err, domain.ErrConflict, "already decided")
}

func TestPreviewGroup_ByCodeCaseInsensitive(t *testing.T) {
	svc := buildSvc()
	g, _ := svc.CreateGroup(context.Background(), teacherID, groups.CreateGroupRequest{Name: "History"})

	p, err := svc.PreviewGroup(context.Background(), " "+strings.ToLower(g.InviteCode)+" ")
	require.NoError(t, err)
	assert.Equal(t, "History", p.Name)

	_, err = svc.PreviewGroup(context.Background(), "NOPE1234")
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestJoinGroup_AlreadyMember(t *testing.T) {
	svc := buildSvc()
	g, _ := svc.CreateGroup(context.Background(), teacherID, groups.CreateGroupRequest{Name: "Bio"})

	join(t, svc, g, studentID)
	_, err := svc.JoinGroup(context.Background(), studentID, groups.JoinGroupRequest{InviteCode: g.InviteCode})
	assert.ErrorIs(t, err, domain.ErrAlreadyMember)
}

func TestJoinGroup_BannedCannotRejoin(t *testing.T) {
	repo := newMemRepo()
	svc := groups.NewService(repo)
	g, _ := svc.CreateGroup(context.Background(), teacherID, groups.CreateGroupRequest{Name: "Bio"})
	repo.bans[memberKey(g.ID, studentID)] = true

	_, err := svc.JoinGroup(context.Background(), studentID, groups.JoinGroupRequest{InviteCode: g.InviteCode})
	var de *domain.Error
	require.ErrorAs(t, err, &de)
	assert.Equal(t, "BANNED_FROM_GROUP", de.Code)
}

type fixedUnread map[uuid.UUID]int

func (f fixedUnread) UnreadCounts(_ context.Context, _ uuid.UUID, _ []uuid.UUID) (map[uuid.UUID]int, error) {
	return f, nil
}

func TestListMyGroups_UnreadCount(t *testing.T) {
	repo := newMemRepo()
	g, _ := groups.NewService(repo).CreateGroup(context.Background(), teacherID, groups.CreateGroupRequest{Name: "Bio"})
	svc := groups.NewService(repo).WithUnread(fixedUnread{g.ID: 3})

	list, err := svc.ListMyGroups(context.Background(), teacherID, domain.RoleTeacher)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, 3, list[0].UnreadCount)
}

func TestJoinGroup_InvalidCode(t *testing.T) {
	svc := buildSvc()
	_, err := svc.JoinGroup(context.Background(), studentID, groups.JoinGroupRequest{InviteCode: "NOTEXIST"})
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestDeleteGroup_ForbiddenForNonOwner(t *testing.T) {
	svc := buildSvc()
	g, _ := svc.CreateGroup(context.Background(), teacherID, groups.CreateGroupRequest{Name: "Chem"})

	otherTeacher := uuid.New()
	err := svc.DeleteGroup(context.Background(), g.ID, otherTeacher)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestArchiveGroup_Toggle(t *testing.T) {
	svc := buildSvc()
	g, _ := svc.CreateGroup(context.Background(), teacherID, groups.CreateGroupRequest{Name: "History"})

	resp, err := svc.ArchiveGroup(context.Background(), g.ID, teacherID)
	require.NoError(t, err)
	assert.True(t, resp.IsArchived)

	resp, err = svc.ArchiveGroup(context.Background(), g.ID, teacherID)
	require.NoError(t, err)
	assert.False(t, resp.IsArchived)
}

func TestUpdateGroup_EditsAndClearsFields(t *testing.T) {
	svc := buildSvc()
	subject, desc := "Математика", "7 класс"
	g, _ := svc.CreateGroup(context.Background(), teacherID,
		groups.CreateGroupRequest{Name: "Old", Subject: &subject, Description: &desc})

	name, newSubject, empty := "  Алгебра  ", "Алгебра", "  "
	resp, err := svc.UpdateGroup(context.Background(), g.ID, teacherID,
		groups.UpdateGroupRequest{Name: &name, Subject: &newSubject, Description: &empty})
	require.NoError(t, err)
	assert.Equal(t, "Алгебра", resp.Name)
	require.NotNil(t, resp.Subject)
	assert.Equal(t, "Алгебра", *resp.Subject)
	assert.Nil(t, resp.Description, "очищенное поле хранится как NULL, а не пустая строка")
}

func TestUpdateGroup_ForbiddenForNonOwner(t *testing.T) {
	svc := buildSvc()
	g, _ := svc.CreateGroup(context.Background(), teacherID, groups.CreateGroupRequest{Name: "Math"})
	name := "Hacked"
	_, err := svc.UpdateGroup(context.Background(), g.ID, uuid.New(), groups.UpdateGroupRequest{Name: &name})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestUpdatePayment_Success(t *testing.T) {
	svc := buildSvc()
	g, _ := svc.CreateGroup(context.Background(), teacherID, groups.CreateGroupRequest{Name: "Math"})
	join(t, svc, g, studentID)

	err := svc.UpdatePayment(context.Background(), g.ID, studentID, teacherID,
		groups.UpdatePaymentRequest{Status: domain.PaymentStatusPaid})
	require.NoError(t, err)
}

func TestUpdatePayment_NotMember(t *testing.T) {
	svc := buildSvc()
	g, _ := svc.CreateGroup(context.Background(), teacherID, groups.CreateGroupRequest{Name: "Art"})
	stranger := uuid.New()

	err := svc.UpdatePayment(context.Background(), g.ID, stranger, teacherID,
		groups.UpdatePaymentRequest{Status: domain.PaymentStatusPaid})
	assert.ErrorIs(t, err, domain.ErrNotMember)
}

func TestLeaveGroup_Success(t *testing.T) {
	svc := buildSvc()
	g, _ := svc.CreateGroup(context.Background(), teacherID, groups.CreateGroupRequest{Name: "PE"})
	join(t, svc, g, studentID)

	err := svc.LeaveGroup(context.Background(), g.ID, studentID)
	require.NoError(t, err)

	// После выхода можно снова подать заявку
	_, err = svc.JoinGroup(context.Background(), studentID, groups.JoinGroupRequest{InviteCode: g.InviteCode})
	require.NoError(t, err)
}
