// Package tgbot — бот MaMu Learn в Telegram: человек делится своим номером,
// и бот присылает ему коды входа вместо SMS (бесплатно).
package tgbot

import (
	"context"
	"fmt"
	"strings"
)

// Messenger — отправка сообщений бота (infrastructure/telegram.Bot).
type Messenger interface {
	SendMessage(ctx context.Context, chatID int64, text string, markup any) error
}

// Repository — номера, подтверждённые в боте.
type Repository interface {
	SaveChat(ctx context.Context, phone string, chatID, tgUserID int64) error
	ChatByPhone(ctx context.Context, phone string) (chatID int64, ok bool, err error)
}

// CodeSource — действующий код входа для номера (модуль auth), чтобы
// прислать его сразу, как только человек поделился номером.
type CodeSource interface {
	ActiveCode(ctx context.Context, phone string) (code string, ok bool, err error)
}

type Service struct {
	repo  Repository
	bot   Messenger
	codes CodeSource
}

func NewService(repo Repository, bot Messenger, codes CodeSource) *Service {
	return &Service{repo: repo, bot: bot, codes: codes}
}

// Update — нужная нам часть обновления Telegram.
type Update struct {
	Message *Message `json:"message"`
}

type Message struct {
	Chat struct {
		ID int64 `json:"id"`
	} `json:"chat"`
	From *struct {
		ID int64 `json:"id"`
	} `json:"from"`
	Text    string   `json:"text"`
	Contact *Contact `json:"contact"`
}

type Contact struct {
	PhoneNumber string `json:"phone_number"`
	UserID      int64  `json:"user_id"`
}

// Тексты на двух языках сразу: язык человека бот ещё не знает.
const (
	msgWelcome = "MaMu Learn: kirish kodini olish uchun quyidagi tugma orqali raqamingizni yuboring.\n\n" +
		"MaMu Learn: чтобы получать коды входа, отправьте свой номер кнопкой ниже."
	msgShareButton = "📱 Raqamni yuborish / Отправить номер"
	msgNotOwn      = "Faqat o'z raqamingizni tugma orqali yuboring.\n\n" +
		"Отправьте свой номер — кнопкой ниже, а не чужой контакт."
	msgLinked = "Raqam tasdiqlandi. Ilovaga qayting va kodni kiriting.\n\n" +
		"Номер подтверждён. Вернитесь в приложение и введите код."
)

func shareKeyboard() any {
	return map[string]any{
		"keyboard": [][]map[string]any{{
			{"text": msgShareButton, "request_contact": true},
		}},
		"resize_keyboard":   true,
		"one_time_keyboard": true,
	}
}

func removeKeyboard() any { return map[string]any{"remove_keyboard": true} }

// HandleUpdate: /start и любой текст — просим номер; свой контакт —
// запоминаем номер и сразу шлём ожидающий код, если человек его уже запросил.
func (s *Service) HandleUpdate(ctx context.Context, u Update) error {
	m := u.Message
	if m == nil || m.From == nil {
		return nil
	}
	if m.Contact == nil {
		return s.bot.SendMessage(ctx, m.Chat.ID, msgWelcome, shareKeyboard())
	}
	// Контакт можно переслать чужой — Telegram ставит user_id владельца номера.
	if m.Contact.UserID != m.From.ID {
		return s.bot.SendMessage(ctx, m.Chat.ID, msgNotOwn, shareKeyboard())
	}
	phone := NormalizePhone(m.Contact.PhoneNumber)
	if err := s.repo.SaveChat(ctx, phone, m.Chat.ID, m.From.ID); err != nil {
		return fmt.Errorf("tgbot save chat: %w", err)
	}
	if err := s.bot.SendMessage(ctx, m.Chat.ID, msgLinked, removeKeyboard()); err != nil {
		return err
	}
	code, ok, err := s.codes.ActiveCode(ctx, phone)
	if err != nil || !ok {
		return err
	}
	return s.bot.SendMessage(ctx, m.Chat.ID, codeMessage(code), nil)
}

// SendCode присылает код входа в Telegram. false — номер ещё не подтверждён
// в боте: приложение предложит открыть бота, а код придёт после этого.
func (s *Service) SendCode(ctx context.Context, phone, code string) (bool, error) {
	chatID, ok, err := s.repo.ChatByPhone(ctx, NormalizePhone(phone))
	if err != nil || !ok {
		return false, err
	}
	if err := s.bot.SendMessage(ctx, chatID, codeMessage(code), nil); err != nil {
		return false, err
	}
	return true, nil
}

func codeMessage(code string) string {
	return fmt.Sprintf("MaMu Learn: kirish kodi / код входа — %s\n\n"+
		"Kodni hech kimga bermang. / Никому не сообщайте код.", code)
}

// NormalizePhone приводит номер к виду +998901234567: Telegram присылает его
// то с плюсом, то без, приложение — с плюсом.
func NormalizePhone(p string) string {
	var b strings.Builder
	for _, r := range p {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return "+" + b.String()
}
