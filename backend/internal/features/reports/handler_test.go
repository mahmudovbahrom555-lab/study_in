package reports

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/middleware"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/pkg/jwt"
)

type reportTestRepo struct {
	Repository
	teacherID, groupID, studentID uuid.UUID
	member                        bool
	reportReads                   int
}

func (r *reportTestRepo) IsGroupTeacher(_ context.Context, groupID, teacherID uuid.UUID) (bool, error) {
	return groupID == r.groupID && teacherID == r.teacherID, nil
}

func (r *reportTestRepo) IsGroupMember(_ context.Context, groupID, studentID uuid.UUID) (bool, error) {
	return r.member && groupID == r.groupID && studentID == r.studentID, nil
}

func (r *reportTestRepo) GetLatestReport(context.Context, uuid.UUID, uuid.UUID) (*domain.ParentReport, error) {
	r.reportReads++
	now := time.Now().UTC()
	return &domain.ParentReport{
		PeriodStart: time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC),
		CreatedAt:   now,
	}, nil
}

type reportTestParents struct{ parentID, studentID uuid.UUID }

func (p reportTestParents) GetLink(_ context.Context, parentID, studentID uuid.UUID) (*domain.ParentLink, error) {
	if parentID == p.parentID && studentID == p.studentID {
		return &domain.ParentLink{}, nil
	}
	return nil, nil
}

func TestParentReportAccess(t *testing.T) {
	parentID, teacherID, studentID, groupID, stranger := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	tokens := jwt.NewManager("access-test-secret", "refresh-test-secret", time.Hour, time.Hour)
	for _, tc := range []struct {
		name   string
		role   domain.Role
		caller uuid.UUID
		member bool
		status int
	}{
		{"linked parent", domain.RoleParent, parentID, true, http.StatusOK},
		{"unlinked parent", domain.RoleParent, stranger, true, http.StatusForbidden},
		{"class owner", domain.RoleTeacher, teacherID, true, http.StatusOK},
		{"other teacher", domain.RoleTeacher, stranger, true, http.StatusForbidden},
		{"student", domain.RoleStudent, studentID, true, http.StatusForbidden},
		{"linked child outside class", domain.RoleParent, parentID, false, http.StatusNotFound},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &reportTestRepo{teacherID: teacherID, groupID: groupID, studentID: studentID, member: tc.member}
			router := chi.NewRouter()
			router.Use(middleware.Auth(tokens))
			NewHandler(NewService(repo), repo, reportTestParents{parentID, studentID}).RegisterRoutes(router)

			token, err := tokens.IssueAccess(tc.caller, tc.role)
			require.NoError(t, err)
			req := httptest.NewRequest(http.MethodGet, "/parent/children/"+studentID.String()+"/groups/"+groupID.String()+"/roi", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			res := httptest.NewRecorder()
			router.ServeHTTP(res, req)
			require.Equal(t, tc.status, res.Code, res.Body.String())
			if tc.status == http.StatusOK {
				require.Equal(t, 1, repo.reportReads)
			} else {
				require.Zero(t, repo.reportReads)
			}
		})
	}
}
