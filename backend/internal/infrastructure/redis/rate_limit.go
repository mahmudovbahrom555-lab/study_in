package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/domain"
)

// RateLimiter реализует счётчик запросов через Redis INCR+EXPIRE.
type RateLimiter struct {
	client *redis.Client
}

func NewRateLimiter(client *redis.Client) *RateLimiter {
	return &RateLimiter{client: client}
}

// Allow возвращает nil если действие разрешено, domain.ErrRateLimit — если лимит превышен.
// key — уникальный ключ действия (например "sms:+998901234567").
// limit — максимальное число запросов за window.
func (rl *RateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) error {
	pipe := rl.client.TxPipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("rate limiter exec: %w", err)
	}
	if incr.Val() > int64(limit) {
		return domain.ErrRateLimit
	}
	return nil
}
