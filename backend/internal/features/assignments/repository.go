package assignments

import (
	"context"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

// Repository — хранилище заданий и сдач.
type Repository interface {
	// Assignments
	CreateAssignment(ctx context.Context, a *domain.Assignment) error
	GetAssignmentByID(ctx context.Context, id uuid.UUID) (*domain.Assignment, error)
	ListGroupAssignments(ctx context.Context, groupID uuid.UUID) ([]*domain.Assignment, error)
	UpdateAssignment(ctx context.Context, a *domain.Assignment) error
	SoftDeleteAssignment(ctx context.Context, id uuid.UUID) error

	// Assignment attachments
	AddAssignmentAttachment(ctx context.Context, a *domain.AssignmentAttachment) error
	ListAssignmentAttachments(ctx context.Context, assignmentID uuid.UUID) ([]*domain.AssignmentAttachment, error)

	// Submissions
	CreateSubmission(ctx context.Context, s *domain.Submission) error
	GetSubmission(ctx context.Context, assignmentID, studentID uuid.UUID) (*domain.Submission, error)
	GetSubmissionByID(ctx context.Context, id uuid.UUID) (*domain.Submission, error)
	ListSubmissions(ctx context.Context, assignmentID uuid.UUID) ([]*domain.Submission, error)
	GradeSubmission(ctx context.Context, id uuid.UUID, grade int16, note string) error

	// Submission attachments
	AddSubmissionAttachment(ctx context.Context, a *domain.SubmissionAttachment) error
	ListSubmissionAttachments(ctx context.Context, submissionID uuid.UUID) ([]*domain.SubmissionAttachment, error)
}

// GroupChecker — минимальный интерфейс для проверки владения группой.
type GroupChecker interface {
	GetGroupByID(ctx context.Context, id uuid.UUID) (*domain.Group, error)
	GetMember(ctx context.Context, groupID, userID uuid.UUID) (*domain.GroupMember, error)
}

// Signer генерирует подписанные URL для MinIO.
type Signer interface {
	PresignedGetURL(ctx context.Context, objectKey string) (string, error)
}
