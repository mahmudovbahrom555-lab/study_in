package sms

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const eskizBaseURL = "https://notify.eskiz.uz/api"

// EskizSender отправляет SMS через Eskiz.uz API.
type EskizSender struct {
	email    string
	password string
	from     string
	client   *http.Client

	mu        sync.Mutex
	token     string
	tokenExp  time.Time
}

func NewEskizSender(email, password, from string) *EskizSender {
	return &EskizSender{
		email:    email,
		password: password,
		from:     from,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

func (e *EskizSender) Send(ctx context.Context, phone, message string) error {
	token, err := e.getToken(ctx)
	if err != nil {
		return fmt.Errorf("eskiz auth: %w", err)
	}

	payload := map[string]string{
		"mobile_phone": phone,
		"message":      message,
		"from":         e.from,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, eskizBaseURL+"/message/sms/send", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("eskiz build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := e.client.Do(req)
	if err != nil {
		return fmt.Errorf("eskiz send: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("eskiz send: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func (e *EskizSender) getToken(ctx context.Context) (string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.token != "" && time.Now().Before(e.tokenExp) {
		return e.token, nil
	}

	payload := map[string]string{"email": e.email, "password": e.password}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, eskizBaseURL+"/auth/login", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("login failed: status %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode login response: %w", err)
	}

	e.token = result.Data.Token
	// Eskiz токены живут ~29 дней, обновляем каждые 25 дней
	e.tokenExp = time.Now().Add(25 * 24 * time.Hour)
	return e.token, nil
}
