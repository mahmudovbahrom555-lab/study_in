package feed

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

func TestReactionsToResponse_NamesOnlyForTeacher(t *testing.T) {
	me, other := uuid.New(), uuid.New()
	rs := []domain.PostReaction{
		{UserID: me, Kind: domain.ReactionUnderstood, UserName: "Ali"},
		{UserID: other, Kind: domain.ReactionQuestion, UserName: "Vali"},
	}

	student := reactionsToResponse(rs, me, false)
	assert.Equal(t, 1, student.Understood)
	assert.Equal(t, 1, student.Question)
	assert.Equal(t, domain.ReactionUnderstood, student.Mine)
	assert.Empty(t, student.UnderstoodBy)
	assert.Empty(t, student.QuestionBy)

	teacher := reactionsToResponse(rs, uuid.New(), true)
	assert.Equal(t, []string{"Ali"}, teacher.UnderstoodBy)
	assert.Equal(t, []string{"Vali"}, teacher.QuestionBy)
	assert.Empty(t, teacher.Mine)
}
