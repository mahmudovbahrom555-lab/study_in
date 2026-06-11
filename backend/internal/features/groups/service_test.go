package groups_test

import (
	"context"
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
	groups  map[uuid.UUID]*domain.Group
	members map[string]*domain.GroupMember // "groupID:studentID" → member
}

func newMemRepo() *memRepo {
	return &memRepo{
		groups:  make(map[uuid.UUID]*domain.Group),
		members: make(map[string]*domain.GroupMember),
	}
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
			out = append(out, &domain.GroupMemberDetail{
				GroupMember: *m,
				Name:        "Test Student",
				Phone:       "+998901234567",
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

// --- helpers ---

func buildSvc() *groups.Service {
	return groups.NewService(newMemRepo())
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

func TestJoinGroup_Success(t *testing.T) {
	svc := buildSvc()
	g, _ := svc.CreateGroup(context.Background(), teacherID, groups.CreateGroupRequest{Name: "Physics"})

	resp, err := svc.JoinGroup(context.Background(), studentID, groups.JoinGroupRequest{InviteCode: g.InviteCode})
	require.NoError(t, err)
	assert.Equal(t, g.ID, resp.ID)
}

func TestJoinGroup_AlreadyMember(t *testing.T) {
	svc := buildSvc()
	g, _ := svc.CreateGroup(context.Background(), teacherID, groups.CreateGroupRequest{Name: "Bio"})

	_, _ = svc.JoinGroup(context.Background(), studentID, groups.JoinGroupRequest{InviteCode: g.InviteCode})
	_, err := svc.JoinGroup(context.Background(), studentID, groups.JoinGroupRequest{InviteCode: g.InviteCode})
	assert.ErrorIs(t, err, domain.ErrAlreadyMember)
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

func TestUpdatePayment_Success(t *testing.T) {
	svc := buildSvc()
	g, _ := svc.CreateGroup(context.Background(), teacherID, groups.CreateGroupRequest{Name: "Math"})
	_, _ = svc.JoinGroup(context.Background(), studentID, groups.JoinGroupRequest{InviteCode: g.InviteCode})

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
	_, _ = svc.JoinGroup(context.Background(), studentID, groups.JoinGroupRequest{InviteCode: g.InviteCode})

	err := svc.LeaveGroup(context.Background(), g.ID, studentID)
	require.NoError(t, err)

	// Joining again should succeed after leaving
	_, err = svc.JoinGroup(context.Background(), studentID, groups.JoinGroupRequest{InviteCode: g.InviteCode})
	require.NoError(t, err)
}
