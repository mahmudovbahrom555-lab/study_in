package sms

import (
	"context"
	"log/slog"
)

// MockSender печатает SMS в лог. Используется в development.
type MockSender struct{}

func NewMockSender() *MockSender { return &MockSender{} }

func (m *MockSender) Send(_ context.Context, phone, message string) error {
	slog.Info("SMS [MOCK]",
		slog.String("phone", phone),
		slog.String("message", message),
	)
	return nil
}
