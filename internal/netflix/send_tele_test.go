package netflix

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSendTelegramMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("request method = %q, want %q", r.Method, http.MethodPost)
		}
		if r.URL.Path != "/bottest-token/sendMessage" {
			t.Errorf("request path = %q, want %q", r.URL.Path, "/bottest-token/sendMessage")
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", got)
		}

		var message telegramMessage
		if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
			t.Errorf("decode request body: %v", err)
			return
		}
		if message.ChatID != "12345" || message.Text != "confirmed" {
			t.Errorf("message = %+v, want chat_id 12345 and text confirmed", message)
		}
		_, _ = w.Write([]byte(`{"ok":true,"result":{}}`))
	}))
	defer server.Close()

	if err := sendTelegramMessage(server.Client(), server.URL, "test-token", "12345", "confirmed"); err != nil {
		t.Fatalf("sendTelegramMessage returned error: %v", err)
	}
}

func TestSendTelegramMessageReturnsTelegramAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":false,"description":"chat not found"}`))
	}))
	defer server.Close()

	err := sendTelegramMessage(server.Client(), server.URL, "test-token", "12345", "confirmed")
	if err == nil || !strings.Contains(err.Error(), "chat not found") {
		t.Fatalf("error = %v, want Telegram API error containing %q", err, "chat not found")
	}
}

func TestSendTelegramMessageRequiresConfiguration(t *testing.T) {
	err := SendTelegramMessage("", "", "confirmed")
	if err == nil || !strings.Contains(err.Error(), "must both be set") {
		t.Fatalf("error = %v, want missing Telegram configuration error", err)
	}
}
