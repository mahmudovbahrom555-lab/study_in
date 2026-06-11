// Package main — точка входа в API.
//
// Жизненный цикл:
//  1. Загрузка конфигурации
//  2. Инициализация логгера
//  3. Подключение к PostgreSQL и Redis
//  4. Создание HTTP сервера
//  5. Запуск сервера
//  6. Ожидание сигнала остановки
//  7. Graceful shutdown
package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/config"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/infrastructure/postgres"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/infrastructure/redis"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/pkg/logger"
	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/server"
)

func main() {
	if err := run(); err != nil {
		// Логгер ещё может быть не инициализирован — используем stderr.
		slog.Error("application failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log := logger.New(cfg.Log.Level, cfg.Log.Format)
	slog.SetDefault(log)

	log.Info("starting repetapp api",
		slog.String("version", cfg.Server.Version),
		slog.String("env", cfg.Server.Env),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Auto-migrate when AUTO_MIGRATE=true (default in production Docker image)
	if os.Getenv("AUTO_MIGRATE") == "true" {
		if err := autoMigrate(cfg.Database.URL, "file://migrations"); err != nil {
			return err
		}
		log.Info("migrations applied")
	}

	// PostgreSQL
	db, err := postgres.Connect(ctx, cfg.Database.URL, cfg.Database.MaxConns)
	if err != nil {
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Error("close db", slog.String("error", err.Error()))
		}
	}()
	log.Info("postgres connected")

	// Redis
	redisClient, err := redis.Connect(ctx, cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		return err
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Error("close redis", slog.String("error", err.Error()))
		}
	}()
	log.Info("redis connected")

	// HTTP server
	srv := server.New(cfg, log, db, redisClient)

	// Запускаем сервер в горутине, чтобы main мог ждать сигнал остановки.
	errCh := make(chan error, 1)
	go func() {
		if err := srv.Start(); err != nil {
			errCh <- err
		}
	}()

	// Graceful shutdown по SIGINT/SIGTERM.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		log.Info("shutdown signal received", slog.String("signal", sig.String()))
	case err := <-errCh:
		log.Error("server error", slog.String("error", err.Error()))
		return err
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown error", slog.String("error", err.Error()))
		return err
	}

	log.Info("shutdown complete")
	return nil
}

func autoMigrate(dbURL, migrationsPath string) error {
	m, err := migrate.New(migrationsPath, dbURL)
	if err != nil {
		return err
	}
	defer m.Close()
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}
