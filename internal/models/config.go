package models

// Config represents the application configuration
type Config struct {
	Email         EmailConfig `yaml:"email"`
	TargetFrom    string      `yaml:"targetFrom"`
	TargetSubject string      `yaml:"targetSubject"`
	Telegram      TelegramConfig `yaml:"telegram"`
}

// EmailConfig represents IMAP email configuration
type EmailConfig struct {
	Imap     string `yaml:"imap"`
	Login    string `yaml:"login"`
	Password string `yaml:"password"`
	MailBox  string `yaml:"mailbox"`
}

// TelegramConfig represents Telegram notification configuration
type TelegramConfig struct {
	BotToken string `yaml:"botToken"`
	ChatID   string `yaml:"chatID"`
}
