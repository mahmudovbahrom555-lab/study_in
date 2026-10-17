// Package account — удаление учётной записи со всеми данными (требование
// Google Play): из приложения (DELETE /me) и с веб-страницы без установки
// приложения (/account/delete).
package account

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
)

// Repository удаляет пользователя и всё, что к нему относится.
type Repository interface {
	// DeleteUser удаляет пользователя (каскады БД: его группы, если он
	// репетитор, сдачи, оценки, посещаемость, связи с родителями…) и
	// возвращает ключи объектов хранилища, которые после этого надо стереть.
	// domain.ErrNotFound — пользователя нет.
	DeleteUser(ctx context.Context, userID uuid.UUID) (objectKeys []string, err error)
}

// ObjectRemover стирает файлы из хранилища (MinIO / S3).
type ObjectRemover interface {
	RemoveObject(ctx context.Context, key string) error
}

type Service struct {
	repo  Repository
	store ObjectRemover
}

func NewService(repo Repository, store ObjectRemover) *Service {
	return &Service{repo: repo, store: store}
}

// DeleteAccount безвозвратно удаляет аккаунт. Данные в БД удаляются одной
// транзакцией; файлы — после неё. Сбой удаления файла не отменяет удаление
// аккаунта: файл уже ни к чему не привязан, ошибку пишем в лог.
func (s *Service) DeleteAccount(ctx context.Context, userID uuid.UUID) error {
	keys, err := s.repo.DeleteUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("account.DeleteAccount: %w", err)
	}
	failed := 0
	for _, k := range keys {
		if err := s.store.RemoveObject(ctx, k); err != nil {
			failed++
			slog.Warn("account delete: object not removed", slog.String("key", k), slog.String("err", err.Error()))
		}
	}
	slog.Info("account deleted", slog.Int("files", len(keys)), slog.Int("files_failed", failed))
	return nil
}
