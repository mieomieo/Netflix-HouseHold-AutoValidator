package logging

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestTelegramHookFireSendsCompleteStructuredLog(t *testing.T) {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	hook := &telegramHook{messages: make(chan string, 1)}
	message := strings.Repeat("complete-log-", 50)

	endTrace := BeginTelegramTrace("trace-123")
	defer endTrace()

	entry := logrus.NewEntry(logger).WithField("trace_id", "trace-123")
	entry.Level = logrus.DebugLevel
	entry.Message = message
	if err := hook.Fire(entry); err != nil {
		t.Fatalf("Fire() returned error: %v", err)
	}

	select {
	case got := <-hook.messages:
		var fields map[string]interface{}
		if err := json.Unmarshal([]byte(got), &fields); err != nil {
			t.Fatalf("Telegram log is not valid JSON: %v", err)
		}
		if fields["msg"] != message {
			t.Errorf("log message was truncated or changed")
		}
		if fields["level"] != "debug" {
			t.Errorf("log level = %v, want debug", fields["level"])
		}
		if fields["trace_id"] != "trace-123" {
			t.Errorf("trace_id = %v, want trace-123", fields["trace_id"])
		}
	default:
		t.Fatal("Fire() did not queue the log entry")
	}
}

func TestTelegramHookFireIgnoresLogsOutsideNetflixTrace(t *testing.T) {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	hook := &telegramHook{messages: make(chan string, 1)}

	entry := logrus.NewEntry(logger).WithField("trace_id", "unrelated-trace")
	entry.Level = logrus.InfoLevel
	entry.Message = "application log"
	if err := hook.Fire(entry); err != nil {
		t.Fatalf("Fire() returned error: %v", err)
	}

	if len(hook.messages) != 0 {
		t.Fatal("Fire() queued a log outside an active Netflix trace")
	}
}

func TestSplitTelegramMessageRespectsMessageLimit(t *testing.T) {
	message := strings.Repeat("界", maxTelegramTextRunes+20)
	parts := splitTelegramMessage(message)
	if len(parts) < 2 {
		t.Fatalf("splitTelegramMessage() returned %d part(s), want multiple parts", len(parts))
	}

	var reconstructed strings.Builder
	for i, part := range parts {
		if got := len([]rune(part)); got > maxTelegramTextRunes {
			t.Errorf("part %d has %d runes, maximum is %d", i+1, got, maxTelegramTextRunes)
		}
		prefix := fmt.Sprintf("[part %d/%d]\n", i+1, len(parts))
		if !strings.HasPrefix(part, prefix) {
			t.Errorf("part %d is missing prefix %q", i+1, prefix)
			continue
		}
		reconstructed.WriteString(strings.TrimPrefix(part, prefix))
	}

	if reconstructed.String() != message {
		t.Error("split message parts do not reconstruct the original log")
	}
}
