// Package files — единое хранилище загружаемых файлов (G10).
//
// Все разделы (сдачи ДЗ, позже вложения учителя и др.) ссылаются на File.ID.
// Файл загружается отдельно (status=pending) и прикрепляется к сущности
// атомарно вместе с ней; неприкреплённые файлы удаляет CleanupPending.
package files

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

const (
	// MaxFileSize — лимит одного файла (DECISIONS.md, 2026-10-07).
	MaxFileSize = 10 << 20
	// PendingTTL — сколько живёт загруженный, но не прикреплённый файл.
	PendingTTL = 24 * time.Hour

	cleanupBatch   = 100
	maxNameRunes   = 200
)

type Repository interface {
	Create(ctx context.Context, f *domain.File) error
	ListStalePending(ctx context.Context, before time.Time, limit int) ([]*domain.File, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type ObjectStore interface {
	PutObject(ctx context.Context, key string, r io.Reader, size int64, contentType string) error
	RemoveObject(ctx context.Context, key string) error
}

type Signer interface {
	PresignedGetURL(ctx context.Context, objectKey string) (string, error)
}

type Service struct {
	repo   Repository
	store  ObjectStore
	signer Signer
}

func NewService(repo Repository, store ObjectStore, signer Signer) *Service {
	return &Service{repo: repo, store: store, signer: signer}
}

// Upload сохраняет файл в статусе pending. Тип определяется по содержимому.
func (s *Service) Upload(ctx context.Context, ownerID uuid.UUID, purpose domain.FilePurpose, name string, r io.Reader) (*domain.File, error) {
	data, err := io.ReadAll(io.LimitReader(r, MaxFileSize+1))
	if err != nil {
		return nil, fmt.Errorf("files.Upload read: %w", err)
	}
	if len(data) > MaxFileSize {
		return nil, domain.NewError("FILE_TOO_LARGE", "Файл больше 10 МБ", domain.ErrValidation)
	}
	if len(data) == 0 {
		return nil, domain.NewError("FILE_EMPTY", "Пустой файл", domain.ErrValidation)
	}
	d, ok := detect(name, data)
	if !ok {
		return nil, domain.NewError("FILE_TYPE_NOT_ALLOWED", "Разрешены фото JPG и PNG, PDF и Word", domain.ErrValidation)
	}

	sum := sha256.Sum256(data)
	hash := hex.EncodeToString(sum[:])
	id := uuid.New()
	f := &domain.File{
		ID:           id,
		OwnerID:      ownerID,
		Purpose:      purpose,
		ObjectKey:    fmt.Sprintf("files/%s/%s%s", ownerID, id, d.ext),
		OriginalName: cleanName(name),
		MimeType:     d.mime,
		SizeBytes:    int64(len(data)),
		SHA256:       &hash,
		Width:        d.width,
		Height:       d.height,
		Status:       domain.FileStatusPending,
		CreatedAt:    time.Now(),
	}

	if err := s.store.PutObject(ctx, f.ObjectKey, bytes.NewReader(data), f.SizeBytes, f.MimeType); err != nil {
		return nil, fmt.Errorf("files.Upload put: %w", err)
	}
	if err := s.repo.Create(ctx, f); err != nil {
		// Без записи в БД объект никто не найдёт и не удалит — убираем сразу.
		_ = s.store.RemoveObject(ctx, f.ObjectKey)
		return nil, fmt.Errorf("files.Upload save: %w", err)
	}
	return f, nil
}

// SignedURL — временная ссылка на скачивание (бакет приватный).
func (s *Service) SignedURL(ctx context.Context, f *domain.File) (string, error) {
	return s.signer.PresignedGetURL(ctx, f.ObjectKey)
}

// CleanupPending удаляет файлы, загруженные больше PendingTTL назад и так и не
// прикреплённые (прерванная сдача, заменённые при пересдаче). Возвращает число удалённых.
func (s *Service) CleanupPending(ctx context.Context, now time.Time) (int, error) {
	stale, err := s.repo.ListStalePending(ctx, now.Add(-PendingTTL), cleanupBatch)
	if err != nil {
		return 0, fmt.Errorf("files.CleanupPending list: %w", err)
	}
	removed := 0
	for _, f := range stale {
		if err := s.store.RemoveObject(ctx, f.ObjectKey); err != nil {
			continue // попробуем в следующий запуск
		}
		if err := s.repo.Delete(ctx, f.ID); err != nil {
			return removed, fmt.Errorf("files.CleanupPending delete: %w", err)
		}
		removed++
	}
	return removed, nil
}

// cleanName оставляет только имя файла (без пути) разумной длины — для показа людям.
func cleanName(name string) string {
	// Телефоны с Windows-путями присылают «C:\...\photo.jpg»; filepath на Linux
	// обратный слэш разделителем не считает, поэтому режем по обоим.
	n := name
	if i := strings.LastIndexAny(n, `/\`); i >= 0 {
		n = n[i+1:]
	}
	if n == "" || n == "." || n == ".." {
		n = "file"
	}
	if utf8.RuneCountInString(n) > maxNameRunes {
		n = string([]rune(n)[:maxNameRunes])
	}
	return n
}
