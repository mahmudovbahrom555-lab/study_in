package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/features/feed"
)

// FeedRepository implements feed.Repository.
type FeedRepository struct {
	db *sqlx.DB
}

func NewFeedRepository(db *sqlx.DB) feed.Repository {
	return &FeedRepository{db: db}
}

func (r *FeedRepository) CreatePost(ctx context.Context, p *domain.Post) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO posts (id, group_id, author_id, body, pinned, created_at, updated_at)
		VALUES (:id, :group_id, :author_id, :body, :pinned, :created_at, :updated_at)`, p)
	if err != nil {
		return fmt.Errorf("FeedRepository.CreatePost: %w", err)
	}
	return nil
}

func (r *FeedRepository) GetPostByID(ctx context.Context, id uuid.UUID) (*domain.Post, error) {
	var p domain.Post
	err := r.db.GetContext(ctx, &p,
		`SELECT * FROM posts WHERE id=$1 AND deleted_at IS NULL`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("FeedRepository.GetPostByID: %w", err)
	}
	return &p, nil
}

func (r *FeedRepository) ListGroupPosts(ctx context.Context, groupID uuid.UUID, limit, offset int) ([]*domain.PostWithMeta, error) {
	var posts []*domain.PostWithMeta
	err := r.db.SelectContext(ctx, &posts, `
		SELECT
			p.*,
			u.name AS author_name
		FROM posts p
		JOIN users u ON u.id = p.author_id
		WHERE p.group_id=$1 AND p.deleted_at IS NULL
		ORDER BY p.pinned DESC, p.created_at DESC
		LIMIT $2 OFFSET $3`,
		groupID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("FeedRepository.ListGroupPosts: %w", err)
	}
	return posts, nil
}

func (r *FeedRepository) UpdatePost(ctx context.Context, p *domain.Post) error {
	_, err := r.db.NamedExecContext(ctx, `
		UPDATE posts SET body=:body, pinned=:pinned, updated_at=:updated_at
		WHERE id=:id AND deleted_at IS NULL`, p)
	if err != nil {
		return fmt.Errorf("FeedRepository.UpdatePost: %w", err)
	}
	return nil
}

func (r *FeedRepository) SoftDeletePost(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE posts SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	if err != nil {
		return fmt.Errorf("FeedRepository.SoftDeletePost: %w", err)
	}
	return nil
}

func (r *FeedRepository) PinPost(ctx context.Context, id uuid.UUID, pinned bool) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE posts SET pinned=$1, updated_at=NOW() WHERE id=$2 AND deleted_at IS NULL`,
		pinned, id)
	if err != nil {
		return fmt.Errorf("FeedRepository.PinPost: %w", err)
	}
	return nil
}

func (r *FeedRepository) AddAttachment(ctx context.Context, a *domain.PostAttachment) error {
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO post_attachments (id, post_id, object_key, filename, mime_type, size_bytes, created_at)
		VALUES (:id, :post_id, :object_key, :filename, :mime_type, :size_bytes, :created_at)`, a)
	if err != nil {
		return fmt.Errorf("FeedRepository.AddAttachment: %w", err)
	}
	return nil
}

func (r *FeedRepository) ListAttachments(ctx context.Context, postID uuid.UUID) ([]*domain.PostAttachment, error) {
	var atts []*domain.PostAttachment
	err := r.db.SelectContext(ctx, &atts,
		`SELECT * FROM post_attachments WHERE post_id=$1 ORDER BY created_at ASC`, postID)
	if err != nil {
		return nil, fmt.Errorf("FeedRepository.ListAttachments: %w", err)
	}
	return atts, nil
}
