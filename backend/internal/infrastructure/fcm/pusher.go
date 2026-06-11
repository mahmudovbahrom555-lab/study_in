// Package fcm provides Firebase Cloud Messaging push notifications.
// When FCM credentials are not configured a no-op pusher is used instead.
package fcm

import (
	"context"
	"fmt"
	"log/slog"
)

// Pusher sends push notifications via FCM.
// The real implementation requires firebase-admin-go; here we stub it
// so the binary compiles without that dependency. To enable real FCM:
//   1. go get firebase.google.com/go/v4
//   2. Replace the stub body below with the real messaging.Client call.
type Pusher struct {
	log *slog.Logger
}

func New(credentialsPath, projectID string, log *slog.Logger) (*Pusher, error) {
	if credentialsPath == "" || projectID == "" {
		return nil, fmt.Errorf("FCM not configured")
	}
	// TODO: initialise firebase app with credentials file
	// app, err := firebase.NewApp(ctx, &firebase.Options{ProjectID: projectID},
	//     option.WithCredentialsFile(credentialsPath))
	log.Info("fcm pusher initialised (stub)", slog.String("project", projectID))
	return &Pusher{log: log}, nil
}

func (p *Pusher) Send(ctx context.Context, token, title, body string, data any) error {
	// TODO: replace with real FCM send
	p.log.Info("fcm push [STUB]",
		slog.String("token_prefix", token[:min(8, len(token))]),
		slog.String("title", title),
	)
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
