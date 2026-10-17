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
	maxTelegramTextRunes = 4096
	telegramChunkRunes   = 3900
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
	formatted, err := entry.Logger.Formatter.Format(entry)
	if err != nil {
		return fmt.Errorf("format Telegram log entry: %w", err)
	}

	for _, message := range splitTelegramMessage(strings.TrimRight(string(formatted), "\r\n")) {
		select {
		case h.messages <- message:
		default:
			_, _ = fmt.Fprintln(os.Stderr, "Telegram notification queue full; notification was not queued")
			return nil
		}
	}
	return nil
}

func splitTelegramMessage(message string) []string {
	runes := []rune(message)
	if len(runes) <= maxTelegramTextRunes {
		return []string{message}
	}

	partCount := (len(runes) + telegramChunkRunes - 1) / telegramChunkRunes
	parts := make([]string, 0, partCount)
	for start, partNumber := 0, 1; start < len(runes); start, partNumber = start+telegramChunkRunes, partNumber+1 {
		end := start + telegramChunkRunes
		if end > len(runes) {
			end = len(runes)
		}
		parts = append(parts, fmt.Sprintf("[part %d/%d]\n%s", partNumber, partCount, string(runes[start:end])))
	}
	return parts
}

func (h *telegramHook) run() {
	var lastSent time.Time
	var pending string
	for {
		first := pending
		if first == "" {
			first = <-h.messages
		}
		pending = ""
		batch := []string{first}
		batchRunes := utf8.RuneCountInString(first)
		timer := time.NewTimer(telegramBatchWindow)
	collect:
		for len(batch) < telegramBatchSize {
			select {
			case message := <-h.messages:
				messageRunes := utf8.RuneCountInString(message)
				if batchRunes+1+messageRunes > maxTelegramTextRunes {
					pending = message
					break collect
				}
				batch = append(batch, message)
				batchRunes += 1 + messageRunes
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
