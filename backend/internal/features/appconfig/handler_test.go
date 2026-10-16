package appconfig

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestAppConfig_FeaturesAsMap(t *testing.T) {
	r := chi.NewRouter()
	NewHandler(Config{MinVersion: "1.2.0", Features: []string{"quizzes"}}).RegisterRoutes(r)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/app-config", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}

	var body struct {
		Data struct {
			MinVersion string          `json:"min_version"`
			Features   map[string]bool `json:"features"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Data.MinVersion != "1.2.0" || !body.Data.Features["quizzes"] || len(body.Data.Features) != 1 {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestAppConfig_NoFeaturesIsEmptyObject(t *testing.T) {
	r := chi.NewRouter()
	NewHandler(Config{MinVersion: "0.0.0"}).RegisterRoutes(r)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/app-config", nil))

	// Приложение ждёт объект, а не null.
	want := `{"data":{"min_version":"0.0.0","features":{}}}` + "\n"
	if rec.Body.String() != want {
		t.Fatalf("got %q, want %q", rec.Body.String(), want)
	}
}
