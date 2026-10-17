package account

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

type fakeRepo struct {
	keys    []string
	err     error
	deleted []uuid.UUID
}

func (r *fakeRepo) DeleteUser(_ context.Context, id uuid.UUID) ([]string, error) {
	if r.err != nil {
		return nil, r.err
	}
	r.deleted = append(r.deleted, id)
	return r.keys, nil
}

type fakeStore struct {
	removed []string
	fail    map[string]bool
}

func (s *fakeStore) RemoveObject(_ context.Context, key string) error {
	if s.fail[key] {
		return errors.New("minio down")
	}
	s.removed = append(s.removed, key)
	return nil
}

func TestDeleteAccount_RemovesFilesAfterDB(t *testing.T) {
	repo := &fakeRepo{keys: []string{"a.jpg", "b.pdf"}}
	store := &fakeStore{}
	id := uuid.New()

	if err := NewService(repo, store).DeleteAccount(context.Background(), id); err != nil {
		t.Fatal(err)
	}
	if len(repo.deleted) != 1 || repo.deleted[0] != id {
		t.Fatalf("user not deleted: %v", repo.deleted)
	}
	if len(store.removed) != 2 {
		t.Fatalf("files not removed: %v", store.removed)
	}
}

func TestDeleteAccount_FileErrorDoesNotFail(t *testing.T) {
	repo := &fakeRepo{keys: []string{"a.jpg", "b.pdf"}}
	store := &fakeStore{fail: map[string]bool{"a.jpg": true}}

	if err := NewService(repo, store).DeleteAccount(context.Background(), uuid.New()); err != nil {
		t.Fatalf("file error must not fail deletion: %v", err)
	}
	if len(store.removed) != 1 || store.removed[0] != "b.pdf" {
		t.Fatalf("other files must still be removed: %v", store.removed)
	}
}

func TestDeleteAccount_NotFound(t *testing.T) {
	store := &fakeStore{}
	err := NewService(&fakeRepo{err: domain.ErrNotFound}, store).DeleteAccount(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
	if len(store.removed) != 0 {
		t.Fatal("nothing must be removed when the user is missing")
	}
}
