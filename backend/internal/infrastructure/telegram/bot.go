// Package telegram — минимальный клиент Telegram Bot API: отправка сообщений
// и регистрация вебхука. Сообщения бота бесплатны.
package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Bot struct {
	token string
	http  *http.Client
	base  string
}

func NewBot(token string) *Bot {
	return &Bot{token: token, http: &http.Client{Timeout: 10 * time.Second}, base: "https://api.telegram.org"}
}

// SendMessage отправляет текст в чат; markup — клавиатура (nil — без неё).
func (b *Bot) SendMessage(ctx context.Context, chatID int64, text string, markup any) error {
	body := map[string]any{"chat_id": chatID, "text": text}
	if markup != nil {
		body["reply_markup"] = markup
	}
	return b.call(ctx, "sendMessage", body)
}

// SetWebhook направляет обновления бота на url; secret Telegram присылает в
// заголовке X-Telegram-Bot-Api-Secret-Token — по нему вебхук отличает Telegram.
func (b *Bot) SetWebhook(ctx context.Context, url, secret string) error {
	return b.call(ctx, "setWebhook", map[string]any{
		"url":             url,
		"secret_token":    secret,
		"allowed_updates": []string{"message"},
	})
}

func (b *Bot) call(ctx context.Context, method string, body any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("telegram %s marshal: %w", method, err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/bot%s/%s", b.base, b.token, method), bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("telegram %s request: %w", method, err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := b.http.Do(req)
	if err != nil {
		// Без текста ошибки: в нём URL с токеном бота.
		return fmt.Errorf("telegram %s: request failed", method)
	}
	defer resp.Body.Close()
	var out struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return fmt.Errorf("telegram %s decode: %w", method, err)
	}
	if !out.OK {
		return fmt.Errorf("telegram %s: %s", method, out.Description)
	}
	return nil
}
