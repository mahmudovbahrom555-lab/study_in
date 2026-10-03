package auth

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// Кириллица переводит SMS в кодировку UCS-2: одно сообщение — максимум 70 символов.
// Длиннее — Eskiz режет на сегменты по 67 символов, и каждый оплачивается отдельно.
func TestVerificationMessage_FitsSingleSMS(t *testing.T) {
	msg := verificationMessage("123456")

	if n := utf8.RuneCountInString(msg); n > 70 {
		t.Fatalf("SMS is %d chars, max 70 for one UCS-2 segment: %q", n, msg)
	}
	if !strings.HasPrefix(msg, "123456") {
		t.Errorf("code must come first (iOS/Android autofill): %q", msg)
	}
	lines := strings.Split(msg, "\n")
	if len(lines) != 2 || !strings.Contains(lines[0], "code") || !strings.Contains(lines[1], "Код") {
		t.Errorf("expected English line first, Russian second: %q", msg)
	}
}
