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

const maxFileSize = 50 << 20 // 50 MB — вложения учителя к заданию

// MaxSubmissionFiles — сколько файлов можно приложить к сдаче (DECISIONS.md, 2026-10-07).
const MaxSubmissionFiles = 5

type Service struct {
	repo   Repository
	groups GroupChecker
	signer Signer
	store  ObjectStore
	files  FileGetter
}

// ObjectStore is a write-side interface for MinIO uploads.
type ObjectStore interface {
	PutObject(ctx context.Context, key string, r io.Reader, size int64, contentType string) error
}

func NewService(repo Repository, groups GroupChecker, signer Signer, store ObjectStore, files FileGetter) *Service {
	return &Service{repo: repo, groups: groups, signer: signer, store: store, files: files}
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

// AssignmentListItem — задание и то, что о нём важно вызывающему:
// ученику — своя сдача, репетитору — сколько сдано и ждёт проверки.
type AssignmentListItem struct {
	Assignment   *domain.Assignment
	MySubmission *domain.Submission      // ученик; nil — ещё не сдавал
	Stats        *domain.SubmissionStats // репетитор
}

// ListAssignments returns all assignments for a group.
func (s *Service) ListAssignments(ctx context.Context, groupID, callerID uuid.UUID, role domain.Role) ([]AssignmentListItem, error) {
	if err := s.assertVisible(ctx, groupID, callerID, role); err != nil {
		return nil, err
	}
	as, err := s.repo.ListGroupAssignments(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("assignments.ListAssignments: %w", err)
	}
	ids := make([]uuid.UUID, len(as))
	items := make([]AssignmentListItem, len(as))
	for i, a := range as {
		ids[i] = a.ID
		items[i].Assignment = a
	}

	switch role {
	case domain.RoleStudent:
		mine, err := s.repo.ListStudentSubmissions(ctx, callerID, ids)
		if err != nil {
			return nil, fmt.Errorf("assignments.ListAssignments my submissions: %w", err)
		}
		for i := range items {
			items[i].MySubmission = mine[items[i].Assignment.ID]
		}
	case domain.RoleTeacher:
		stats, err := s.repo.SubmissionStats(ctx, ids)
		if err != nil {
			return nil, fmt.Errorf("assignments.ListAssignments stats: %w", err)
		}
		for i := range items {
			st := stats[items[i].Assignment.ID]
			items[i].Stats = &st
		}
	}
	return items, nil
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

	hasComment := req.Comment != nil && strings.TrimSpace(*req.Comment) != ""
	if !hasComment && len(req.FileIDs) == 0 {
		return nil, domain.NewError("SUBMISSION_EMPTY", "Добавьте комментарий или файл", domain.ErrValidation)
	}
	if len(req.FileIDs) > MaxSubmissionFiles {
		return nil, domain.NewError("TOO_MANY_FILES", "Не больше 5 файлов", domain.ErrValidation)
	}

	existing, err := s.repo.GetSubmission(ctx, assignmentID, studentID)
	if err != nil {
		return nil, fmt.Errorf("assignments.Submit check existing: %w", err)
	}
	if existing != nil && existing.Grade != nil {
		return nil, domain.NewError("ALREADY_GRADED", "Работа уже оценена — пересдать нельзя", domain.ErrConflict)
	}
	if err := s.checkSubmissionFiles(ctx, existing, studentID, req.FileIDs); err != nil {
		return nil, err
	}

	sub := &domain.Submission{
		ID:           uuid.New(),
		AssignmentID: assignmentID,
		StudentID:    studentID,
		Comment:      req.Comment,
		SubmittedAt:  time.Now(),
	}
	if existing != nil {
		sub.ID = existing.ID // пересдача: та же запись, новые комментарий, файлы и время
	}
	if err := s.repo.SaveSubmission(ctx, sub, req.FileIDs); err != nil {
		return nil, fmt.Errorf("assignments.Submit: %w", err)
	}
	return sub, nil
}

// checkSubmissionFiles: прикрепить можно только свои файлы для сдачи ДЗ, которые
// ещё не прикреплены — либо уже прикреплены к этой же сдаче (пересдача без замены).
func (s *Service) checkSubmissionFiles(ctx context.Context, existing *domain.Submission, studentID uuid.UUID, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	found, err := s.files.GetByIDs(ctx, ids)
	if err != nil {
		return fmt.Errorf("assignments.Submit files: %w", err)
	}
	if len(found) != len(ids) {
		return domain.NewError("FILE_NOT_FOUND", "Файл не найден — загрузите его заново", domain.ErrValidation)
	}

	current := map[uuid.UUID]bool{}
	if existing != nil {
		byID, err := s.repo.ListSubmissionFiles(ctx, []uuid.UUID{existing.ID})
		if err != nil {
			return fmt.Errorf("assignments.Submit current files: %w", err)
		}
		for _, f := range byID[existing.ID] {
			current[f.ID] = true
		}
	}

	for _, f := range found {
		if f.OwnerID != studentID || f.Purpose != domain.FilePurposeSubmission {
			return domain.ErrForbidden
		}
		if f.Status == domain.FileStatusAttached && !current[f.ID] {
			return domain.NewError("FILE_IN_USE", "Файл уже приложен к другой работе", domain.ErrConflict)
		}
	}
	return nil
}

// MySubmission — сдача ученика по заданию вместе с файлами (nil, если ещё не сдавал).
func (s *Service) MySubmission(ctx context.Context, assignmentID, studentID uuid.UUID) (*domain.Submission, []*domain.File, error) {
	sub, err := s.repo.GetSubmission(ctx, assignmentID, studentID)
	if err != nil {
		return nil, nil, fmt.Errorf("assignments.MySubmission: %w", err)
	}
	if sub == nil {
		return nil, nil, nil
	}
	byID, err := s.repo.ListSubmissionFiles(ctx, []uuid.UUID{sub.ID})
	if err != nil {
		return nil, nil, fmt.Errorf("assignments.MySubmission files: %w", err)
	}
	return sub, byID[sub.ID], nil
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
func (s *Service) ListSubmissions(ctx context.Context, assignmentID, teacherID uuid.UUID) ([]*domain.SubmissionWithStudent, error) {
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
func (s *Service) SignedURL(ctx context.Context, objectKey, filename string) (string, error) {
	return s.signer.PresignedGetURL(ctx, objectKey, filename)
}

// AssignmentAttachments returns attachments for an assignment.
func (s *Service) AssignmentAttachments(ctx context.Context, assignmentID uuid.UUID) ([]*domain.AssignmentAttachment, error) {
	return s.repo.ListAssignmentAttachments(ctx, assignmentID)
}

// SubmissionFiles returns files of the given submissions, grouped by submission ID.
func (s *Service) SubmissionFiles(ctx context.Context, submissionIDs []uuid.UUID) (map[uuid.UUID][]*domain.File, error) {
	return s.repo.ListSubmissionFiles(ctx, submissionIDs)
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
