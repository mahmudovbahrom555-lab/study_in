// Package monitoring initialises error tracking and observability integrations.
package monitoring

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
)

// InitSentry configures Sentry. No-op when DSN is empty.
func InitSentry(dsn, env, version string) error {
	if dsn == "" {
		slog.Info("sentry disabled (DSN not set)")
		return nil
	}
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		Environment:      env,
		Release:          "repetapp@" + version,
		TracesSampleRate: 0.2,
	}); err != nil {
		return fmt.Errorf("sentry init: %w", err)
	}
	slog.Info("sentry enabled", slog.String("env", env))
	return nil
}

// FlushSentry flushes buffered events before process exit.
func FlushSentry() {
	sentry.Flush(2 * time.Second)
}

// SentryRecovery is an HTTP middleware that captures panics to Sentry.
func SentryRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				hub := sentry.CurrentHub().Clone()
				hub.Scope().SetRequest(r)
				hub.RecoverWithContext(r.Context(), err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
