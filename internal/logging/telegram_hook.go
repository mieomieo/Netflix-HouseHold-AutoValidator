package logging

import (
	"fmt"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/sirupsen/logrus"
)

const (
	telegramBatchWindow  = 2 * time.Second
	telegramBatchSize    = 8
	telegramQueueSize    = 256
	maxTelegramLineRunes = 400
)

type telegramHook struct {
	send     func(string) error
	messages chan string
}

func AddTelegramHook(send func(string) error) {
	hook := &telegramHook{
		send:     send,
		messages: make(chan string, telegramQueueSize),
	}
	Log.AddHook(hook)
	go hook.run()
}

func (h *telegramHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *telegramHook) Fire(entry *logrus.Entry) error {
	if !shouldNotifyTelegram(entry) {
		return nil
	}

	message := fmt.Sprintf("[%s] %s", strings.ToUpper(entry.Level.String()), truncateRunes(entry.Message, maxTelegramLineRunes))
	if traceID, ok := entry.Data["trace_id"]; ok {
		message += " (trace_id: " + fmt.Sprint(traceID) + ")"
	}

	select {
	case h.messages <- message:
	default:
		_, _ = fmt.Fprintln(os.Stderr, "Telegram notification queue full; notification was not queued")
	}
	return nil
}

func shouldNotifyTelegram(entry *logrus.Entry) bool {
	if entry.Level <= logrus.WarnLevel {
		return true
	}
	if entry.Level != logrus.InfoLevel {
		return false
	}

	for _, prefix := range []string{
		"Starting Netflix email verification process",
		"IMAP connection established successfully",
		"Processing cycle completed",
		"Shutting down gracefully",
	} {
		if strings.HasPrefix(entry.Message, prefix) {
			return true
		}
	}
	return false
}

func (h *telegramHook) run() {
	var lastSent time.Time
	for first := range h.messages {
		batch := []string{first}
		timer := time.NewTimer(telegramBatchWindow)
	collect:
		for len(batch) < telegramBatchSize {
			select {
			case message := <-h.messages:
				batch = append(batch, message)
			case <-timer.C:
				break collect
			}
		}
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}

		if wait := time.Until(lastSent.Add(time.Second)); !lastSent.IsZero() && wait > 0 {
			time.Sleep(wait)
		}
		if err := h.send(strings.Join(batch, "\n")); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Telegram notification send failed: %v\n", err)
		}
		lastSent = time.Now()
	}
}

func truncateRunes(value string, limit int) string {
	if utf8.RuneCountInString(value) <= limit {
		return value
	}
	runes := []rune(value)
	return string(runes[:limit]) + "..."
}
