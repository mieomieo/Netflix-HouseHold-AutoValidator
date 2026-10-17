package logging

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestShouldNotifyTelegram(t *testing.T) {
	tests := []struct {
		name    string
		level   logrus.Level
		message string
		want    bool
	}{
		{name: "warning", level: logrus.WarnLevel, message: "page race failed", want: true},
		{name: "error", level: logrus.ErrorLevel, message: "IMAP failed", want: true},
		{name: "startup", level: logrus.InfoLevel, message: "Starting Netflix email verification process", want: true},
		{name: "cycle summary", level: logrus.InfoLevel, message: "Processing cycle completed in 1s", want: true},
		{name: "routine info", level: logrus.InfoLevel, message: "Found 2 unseen emails", want: false},
		{name: "debug", level: logrus.DebugLevel, message: "details", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entry := logrus.NewEntry(logrus.New())
			entry.Level = test.level
			entry.Message = test.message
			if got := shouldNotifyTelegram(entry); got != test.want {
				t.Errorf("shouldNotifyTelegram() = %v, want %v", got, test.want)
			}
		})
	}
}
