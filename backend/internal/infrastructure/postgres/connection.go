// Package postgres предоставляет подключение к PostgreSQL через pgx + sqlx.
package postgres

import (
	"context"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

// Connect открывает пул соединений к PostgreSQL и проверяет доступность.
func Connect(ctx context.Context, url string, maxConns int) (*sqlx.DB, error) {
	db, err := sqlx.ConnectContext(ctx, "pgx", url)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}

	db.SetMaxOpenConns(maxConns)
	db.SetMaxIdleConns(maxConns / 2)
	db.SetConnMaxLifetime(30 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return db, nil
}
