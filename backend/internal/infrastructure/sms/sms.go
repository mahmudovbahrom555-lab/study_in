// Package sms предоставляет абстракцию для отправки SMS.
package sms

import "context"

// Sender — интерфейс отправки SMS.
// Реализации: MockSender (dev), EskizSender (prod).
type Sender interface {
	Send(ctx context.Context, phone, message string) error
}
