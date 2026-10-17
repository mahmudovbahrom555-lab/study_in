package parents

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

// Раньше родителю уходили доменные структуры как есть: ключи «ID», «Name»
// (приложение не могло их прочитать) и код приглашения в группу.
func keys(t *testing.T, v any) map[string]any {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	var m map[string]any
	require.NoError(t, json.Unmarshal(b, &m))
	return m
}

func TestChildGroup_SnakeCaseWithoutInviteCode(t *testing.T) {
	subject := "English"
	m := keys(t, toChildGroup(&domain.Group{
		ID: uuid.New(), TeacherID: uuid.New(), Name: "B1", Subject: &subject,
		InviteCode: "SECRET12", CreatedAt: time.Now(),
	}))
	for _, k := range []string{"id", "teacher_id", "name", "subject", "is_archived", "created_at"} {
		assert.Contains(t, m, k)
	}
	assert.NotContains(t, m, "invite_code")
	assert.NotContains(t, m, "ID")
}

func TestChildGrade_SnakeCase(t *testing.T) {
	m := keys(t, toChildGrade(&domain.Grade{ID: uuid.New(), Subject: "Test", Value: 85, MaxValue: 100}))
	for _, k := range []string{"id", "group_id", "student_id", "teacher_id", "subject", "value", "max_value", "graded_at", "created_at"} {
		assert.Contains(t, m, k)
	}
}

func TestChildAttendance_DateAsDay(t *testing.T) {
	m := keys(t, toChildAttendance(&domain.Attendance{
		ID: uuid.New(), Status: domain.AttendanceAbsent,
		LessonDate: time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC),
	}))
	assert.Equal(t, "2026-10-14", m["lesson_date"])
	assert.Equal(t, "absent", m["status"])
}
