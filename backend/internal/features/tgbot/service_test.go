package tgbot_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mahmudovbahrom555-lab/study_in/backend/internal/features/tgbot"
)

type sent struct {
	chat int64
	text string
}

type fakeBot struct{ out []sent }

func (b *fakeBot) SendMessage(_ context.Context, chatID int64, text string, _ any) error {
	b.out = append(b.out, sent{chatID, text})
	return nil
}

type fakeRepo struct{ chats map[string]int64 }

func (r *fakeRepo) SaveChat(_ context.Context, phone string, chatID, _ int64) error {
	r.chats[phone] = chatID
	return nil
}

func (r *fakeRepo) ChatByPhone(_ context.Context, phone string) (int64, bool, error) {
	id, ok := r.chats[phone]
	return id, ok, nil
}

type fakeCodes map[string]string

func (c fakeCodes) ActiveCode(_ context.Context, phone string) (string, bool, error) {
	code, ok := c[phone]
	return code, ok, nil
}

func newSvc(codes fakeCodes) (*tgbot.Service, *fakeBot, *fakeRepo) {
	bot, repo := &fakeBot{}, &fakeRepo{chats: map[string]int64{}}
	return tgbot.NewService(repo, bot, codes), bot, repo
}

func contactUpdate(chatID, fromID, ownerID int64, phone string) tgbot.Update {
	m := &tgbot.Message{Contact: &tgbot.Contact{PhoneNumber: phone, UserID: ownerID}}
	m.Chat.ID = chatID
	m.From = &struct {
		ID int64 `json:"id"`
	}{ID: fromID}
	return tgbot.Update{Message: m}
}

func TestOwnContact_SavesPhoneAndSendsPendingCode(t *testing.T) {
	svc, bot, repo := newSvc(fakeCodes{"+998901234567": "123456"})

	// Telegram присылает номер без плюса.
	require.NoError(t, svc.HandleUpdate(context.Background(), contactUpdate(10, 7, 7, "998901234567")))
	assert.Equal(t, int64(10), repo.chats["+998901234567"])
	require.Len(t, bot.out, 2)
	assert.Contains(t, bot.out[1].text, "123456")
}

func TestForeignContact_Rejected(t *testing.T) {
	svc, bot, repo := newSvc(fakeCodes{"+998901234567": "123456"})

	// Переслали чужой контакт: владелец номера — другой пользователь.
	require.NoError(t, svc.HandleUpdate(context.Background(), contactUpdate(10, 7, 99, "+998901234567")))
	assert.Empty(t, repo.chats)
	require.Len(t, bot.out, 1)
	assert.NotContains(t, bot.out[0].text, "123456")
}

func TestSendCode_OnlyToConfirmedPhone(t *testing.T) {
	svc, bot, repo := newSvc(fakeCodes{})

	ok, err := svc.SendCode(context.Background(), "+998901234567", "654321")
	require.NoError(t, err)
	assert.False(t, ok, "номер не подтверждён в боте")
	assert.Empty(t, bot.out)

	repo.chats["+998901234567"] = 10
	ok, err = svc.SendCode(context.Background(), "+998901234567", "654321")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, int64(10), bot.out[0].chat)
	assert.Contains(t, bot.out[0].text, "654321")
}

func TestWebhook_RequiresSecret(t *testing.T) {
	svc, bot, _ := newSvc(fakeCodes{})
	r := chi.NewRouter()
	tgbot.NewHandler(svc, "s3cret").RegisterRoutes(r)
	body := `{"message":{"chat":{"id":1},"from":{"id":1},"text":"/start"}}`

	req := httptest.NewRequest(http.MethodPost, "/telegram/webhook", strings.NewReader(body))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Empty(t, bot.out)

	req = httptest.NewRequest(http.MethodPost, "/telegram/webhook", strings.NewReader(body))
	req.Header.Set("X-Telegram-Bot-Api-Secret-Token", "s3cret")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
	require.Len(t, bot.out, 1, "на /start бот просит номер")
}
