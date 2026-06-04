package tock

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/theopenconversationkit/tock-mcp-server/config"
)

func TestAsk_HeaderOverrideIsCaseInsensitive(t *testing.T) {
	var gotFilter string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotFilter = r.Header.Get("X-Toki-Filter")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"responses":[{"text":"ok"}]}`))
	}))
	defer ts.Close()

	client := NewClient(config.TockConfig{
		BaseURL:   ts.URL,
		Namespace: "app",
		Bot:       "bot",
		Connector: "web",
		UserID:    "u1",
		ExtraHeaders: map[string]string{
			"x-toki-filter": "",
		},
	})

	_, err := client.Ask(context.Background(), "hello", map[string]string{
		"X-Toki-Filter": "category:faq",
	})
	if err != nil {
		t.Fatalf("Ask() error: %v", err)
	}

	if gotFilter != "category:faq" {
		t.Fatalf("expected X-Toki-Filter override to be applied, got %q", gotFilter)
	}
}
