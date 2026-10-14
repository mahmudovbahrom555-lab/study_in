package attendance

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

type Repository interface {
	// Upsert создаёт отметку или меняет существующую за тот же день;
	// a.ID и a.CreatedAt заполняются значениями из базы.
	Upsert(ctx context.Context, a *domain.Attendance) error
	// UpsertMany — то же для нескольких отметок в одной транзакции.
	UpsertMany(ctx context.Context, list []*domain.Attendance) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Attendance, error)
	// Roster — ученики группы на дату с их отметками; ученики, которых уже нет
	// в группе, но отмеченные в этот день, тоже попадают в список.
	Roster(ctx context.Context, groupID uuid.UUID, date time.Time) ([]*domain.RosterEntry, error)
	ListByGroupStudent(ctx context.Context, groupID, studentID uuid.UUID) ([]*domain.Attendance, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// GroupChecker avoids a circular import on the groups package.
type GroupChecker interface {
	GetGroupByID(ctx context.Context, id uuid.UUID) (*domain.Group, error)
	GetMember(ctx context.Context, groupID, userID uuid.UUID) (*domain.GroupMember, error)
	ListMembers(ctx context.Context, groupID uuid.UUID) ([]*domain.GroupMemberDetail, error)
}
