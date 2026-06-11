package assignments

import (
	"context"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

const maxFileSize = 50 << 20 // 50 MB

type Service struct {
	repo   Repository
	groups GroupChecker
	signer Signer
	store  ObjectStore
}

// ObjectStore is a write-side interface for MinIO uploads.
type ObjectStore interface {
	PutObject(ctx context.Context, key string, r io.Reader, size int64, contentType string) error
}

func NewService(repo Repository, groups GroupChecker, signer Signer, store ObjectStore) *Service {
	return &Service{repo: repo, groups: groups, signer: signer, store: store}
}

// CreateAssignment publishes a homework assignment to a group (teacher only).
func (s *Service) CreateAssignment(ctx context.Context, teacherID, groupID uuid.UUID, req CreateAssignmentRequest) (*domain.Assignment, error) {
	g, err := s.groups.GetGroupByID(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("assignments.CreateAssignment fetch group: %w", err)
	}
	if g == nil {
		return nil, domain.ErrNotFound
	}
	if g.TeacherID != teacherID {
		return nil, domain.ErrForbidden
	}

	now := time.Now()
	a := &domain.Assignment{
		ID:          uuid.New(),
		GroupID:     groupID,
		TeacherID:   teacherID,
		Title:       strings.TrimSpace(req.Title),
		Description: req.Description,
		DueDate:     req.DueDate,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.CreateAssignment(ctx, a); err != nil {
		return nil, fmt.Errorf("assignments.CreateAssignment: %w", err)
	}
	return a, nil
}

// UploadAssignmentFile uploads a file attachment for an assignment (teacher only).
func (s *Service) UploadAssignmentFile(ctx context.Context, assignmentID, teacherID uuid.UUID, filename string, r io.Reader, size int64) (*domain.AssignmentAttachment, error) {
	a, err := s.repo.GetAssignmentByID(ctx, assignmentID)
	if err != nil {
		return nil, fmt.Errorf("assignments.UploadAssignmentFile fetch: %w", err)
	}
	if a == nil {
		return nil, domain.ErrNotFound
	}
	if a.TeacherID != teacherID {
		return nil, domain.ErrForbidden
	}
	if size > maxFileSize {
		return nil, domain.ErrValidation
	}

	objectKey := fmt.Sprintf("assignments/%s/%s_%s", assignmentID, uuid.New().String(), sanitizeFilename(filename))
	contentType := mime.TypeByExtension(filepath.Ext(filename))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	if err := s.store.PutObject(ctx, objectKey, r, size, contentType); err != nil {
		return nil, fmt.Errorf("assignments.UploadAssignmentFile put: %w", err)
	}

	att := &domain.AssignmentAttachment{
		ID:           uuid.New(),
		AssignmentID: assignmentID,
		ObjectKey:    objectKey,
		Filename:     filename,
		MimeType:     contentType,
		SizeBytes:    size,
		CreatedAt:    time.Now(),
	}
	if err := s.repo.AddAssignmentAttachment(ctx, att); err != nil {
		return nil, fmt.Errorf("assignments.UploadAssignmentFile save: %w", err)
	}
	return att, nil
}

// GetAssignment returns an assignment if the caller can see it.
func (s *Service) GetAssignment(ctx context.Context, id, callerID uuid.UUID, role domain.Role) (*domain.Assignment, []*domain.AssignmentAttachment, error) {
	a, err := s.repo.GetAssignmentByID(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("assignments.GetAssignment: %w", err)
	}
	if a == nil {
		return nil, nil, domain.ErrNotFound
	}
	if err := s.assertVisible(ctx, a.GroupID, callerID, role); err != nil {
		return nil, nil, err
	}

	atts, err := s.repo.ListAssignmentAttachments(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("assignments.GetAssignment atts: %w", err)
	}
	return a, atts, nil
}

// ListAssignments returns all assignments for a group.
func (s *Service) ListAssignments(ctx context.Context, groupID, callerID uuid.UUID, role domain.Role) ([]*domain.Assignment, error) {
	if err := s.assertVisible(ctx, groupID, callerID, role); err != nil {
		return nil, err
	}
	as, err := s.repo.ListGroupAssignments(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("assignments.ListAssignments: %w", err)
	}
	return as, nil
}

// UpdateAssignment edits title/description/due_date (teacher only).
func (s *Service) UpdateAssignment(ctx context.Context, id, teacherID uuid.UUID, req UpdateAssignmentRequest) (*domain.Assignment, error) {
	a, err := s.repo.GetAssignmentByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("assignments.UpdateAssignment fetch: %w", err)
	}
	if a == nil {
		return nil, domain.ErrNotFound
	}
	if a.TeacherID != teacherID {
		return nil, domain.ErrForbidden
	}

	if req.Title != nil {
		a.Title = strings.TrimSpace(*req.Title)
	}
	if req.Description != nil {
		a.Description = req.Description
	}
	if req.DueDate != nil {
		a.DueDate = req.DueDate
	}
	a.UpdatedAt = time.Now()

	if err := s.repo.UpdateAssignment(ctx, a); err != nil {
		return nil, fmt.Errorf("assignments.UpdateAssignment save: %w", err)
	}
	return a, nil
}

// DeleteAssignment soft-deletes an assignment (teacher only).
func (s *Service) DeleteAssignment(ctx context.Context, id, teacherID uuid.UUID) error {
	a, err := s.repo.GetAssignmentByID(ctx, id)
	if err != nil {
		return fmt.Errorf("assignments.DeleteAssignment fetch: %w", err)
	}
	if a == nil {
		return domain.ErrNotFound
	}
	if a.TeacherID != teacherID {
		return domain.ErrForbidden
	}
	if err := s.repo.SoftDeleteAssignment(ctx, id); err != nil {
		return fmt.Errorf("assignments.DeleteAssignment: %w", err)
	}
	return nil
}

// Submit creates or updates a student's submission.
func (s *Service) Submit(ctx context.Context, assignmentID, studentID uuid.UUID, req SubmitRequest) (*domain.Submission, error) {
	a, err := s.repo.GetAssignmentByID(ctx, assignmentID)
	if err != nil {
		return nil, fmt.Errorf("assignments.Submit fetch assignment: %w", err)
	}
	if a == nil {
		return nil, domain.ErrNotFound
	}

	m, err := s.groups.GetMember(ctx, a.GroupID, studentID)
	if err != nil {
		return nil, fmt.Errorf("assignments.Submit check member: %w", err)
	}
	if m == nil {
		return nil, domain.ErrForbidden
	}

	existing, err := s.repo.GetSubmission(ctx, assignmentID, studentID)
	if err != nil {
		return nil, fmt.Errorf("assignments.Submit check existing: %w", err)
	}
	if existing != nil {
		// Re-submission: update comment, keep grade.
		existing.Comment = req.Comment
		if err := s.repo.CreateSubmission(ctx, existing); err != nil {
			// Upsert via ON CONFLICT in repo.
		}
		return existing, nil
	}

	sub := &domain.Submission{
		ID:           uuid.New(),
		AssignmentID: assignmentID,
		StudentID:    studentID,
		Comment:      req.Comment,
		SubmittedAt:  time.Now(),
	}
	if err := s.repo.CreateSubmission(ctx, sub); err != nil {
		return nil, fmt.Errorf("assignments.Submit: %w", err)
	}
	return sub, nil
}

// UploadSubmissionFile uploads a file for a student submission.
func (s *Service) UploadSubmissionFile(ctx context.Context, submissionID, studentID uuid.UUID, filename string, r io.Reader, size int64) (*domain.SubmissionAttachment, error) {
	sub, err := s.repo.GetSubmissionByID(ctx, submissionID)
	if err != nil {
		return nil, fmt.Errorf("assignments.UploadSubmissionFile fetch: %w", err)
	}
	if sub == nil {
		return nil, domain.ErrNotFound
	}
	if sub.StudentID != studentID {
		return nil, domain.ErrForbidden
	}
	if size > maxFileSize {
		return nil, domain.ErrValidation
	}

	objectKey := fmt.Sprintf("submissions/%s/%s_%s", submissionID, uuid.New().String(), sanitizeFilename(filename))
	contentType := mime.TypeByExtension(filepath.Ext(filename))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	if err := s.store.PutObject(ctx, objectKey, r, size, contentType); err != nil {
		return nil, fmt.Errorf("assignments.UploadSubmissionFile put: %w", err)
	}

	att := &domain.SubmissionAttachment{
		ID:           uuid.New(),
		SubmissionID: submissionID,
		ObjectKey:    objectKey,
		Filename:     filename,
		MimeType:     contentType,
		SizeBytes:    size,
		CreatedAt:    time.Now(),
	}
	if err := s.repo.AddSubmissionAttachment(ctx, att); err != nil {
		return nil, fmt.Errorf("assignments.UploadSubmissionFile save: %w", err)
	}
	return att, nil
}

// Grade sets a grade on a submission (teacher only).
func (s *Service) Grade(ctx context.Context, submissionID, teacherID uuid.UUID, req GradeRequest) (*domain.Submission, error) {
	sub, err := s.repo.GetSubmissionByID(ctx, submissionID)
	if err != nil {
		return nil, fmt.Errorf("assignments.Grade fetch: %w", err)
	}
	if sub == nil {
		return nil, domain.ErrNotFound
	}

	a, err := s.repo.GetAssignmentByID(ctx, sub.AssignmentID)
	if err != nil {
		return nil, fmt.Errorf("assignments.Grade fetch assignment: %w", err)
	}
	if a == nil || a.TeacherID != teacherID {
		return nil, domain.ErrForbidden
	}

	note := ""
	if req.Note != nil {
		note = *req.Note
	}
	if err := s.repo.GradeSubmission(ctx, submissionID, req.Grade, note); err != nil {
		return nil, fmt.Errorf("assignments.Grade save: %w", err)
	}

	now := time.Now()
	sub.Grade = &req.Grade
	sub.TeacherNote = req.Note
	sub.GradedAt = &now
	return sub, nil
}

// ListSubmissions returns all submissions for an assignment (teacher only).
func (s *Service) ListSubmissions(ctx context.Context, assignmentID, teacherID uuid.UUID) ([]*domain.Submission, error) {
	a, err := s.repo.GetAssignmentByID(ctx, assignmentID)
	if err != nil {
		return nil, fmt.Errorf("assignments.ListSubmissions fetch: %w", err)
	}
	if a == nil {
		return nil, domain.ErrNotFound
	}
	if a.TeacherID != teacherID {
		return nil, domain.ErrForbidden
	}

	subs, err := s.repo.ListSubmissions(ctx, assignmentID)
	if err != nil {
		return nil, fmt.Errorf("assignments.ListSubmissions: %w", err)
	}
	return subs, nil
}

// SignedURL returns a signed URL for an assignment attachment.
func (s *Service) SignedURL(ctx context.Context, objectKey string) (string, error) {
	return s.signer.PresignedGetURL(ctx, objectKey)
}

// AssignmentAttachments returns attachments for an assignment.
func (s *Service) AssignmentAttachments(ctx context.Context, assignmentID uuid.UUID) ([]*domain.AssignmentAttachment, error) {
	return s.repo.ListAssignmentAttachments(ctx, assignmentID)
}

// SubmissionAttachments returns attachments for a submission.
func (s *Service) SubmissionAttachments(ctx context.Context, submissionID uuid.UUID) ([]*domain.SubmissionAttachment, error) {
	return s.repo.ListSubmissionAttachments(ctx, submissionID)
}

// GetStudentSubmission returns a student's existing submission for an assignment.
func (s *Service) GetStudentSubmission(ctx context.Context, assignmentID, studentID uuid.UUID) (*domain.Submission, error) {
	return s.repo.GetSubmission(ctx, assignmentID, studentID)
}

func (s *Service) assertVisible(ctx context.Context, groupID, callerID uuid.UUID, role domain.Role) error {
	g, err := s.groups.GetGroupByID(ctx, groupID)
	if err != nil {
		return fmt.Errorf("assertVisible: %w", err)
	}
	if g == nil {
		return domain.ErrNotFound
	}
	if role == domain.RoleTeacher {
		if g.TeacherID != callerID {
			return domain.ErrForbidden
		}
		return nil
	}
	m, err := s.groups.GetMember(ctx, groupID, callerID)
	if err != nil {
		return fmt.Errorf("assertVisible: %w", err)
	}
	if m == nil {
		return domain.ErrForbidden
	}
	return nil
}

// sanitizeFilename strips path separators from user-supplied filenames.
func sanitizeFilename(name string) string {
	name = filepath.Base(name)
	name = strings.ReplaceAll(name, "..", "")
	return name
}
