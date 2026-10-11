package netflix

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type telegramMessage struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

type telegramResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
}

// SendTelegramMessage sends a message to the configured Telegram chat.
func SendTelegramMessage(token, chatID, text string) error {
	token = strings.TrimSpace(token)
	chatID = strings.TrimSpace(chatID)
	if token == "" || chatID == "" {
		return fmt.Errorf("Telegram bot token and chat ID must both be set")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	return sendTelegramMessage(client, "https://api.telegram.org", token, chatID, text)
}

func sendTelegramMessage(client *http.Client, apiURL, token, chatID, text string) error {
	body, err := json.Marshal(telegramMessage{ChatID: chatID, Text: text})
	if err != nil {
		return fmt.Errorf("encode Telegram message: %w", err)
	}

	url := strings.TrimRight(apiURL, "/") + "/bot" + token + "/sendMessage"
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create Telegram request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send Telegram request: %w", err)
	}
	defer resp.Body.Close()

	var result telegramResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&result); err != nil {
		return fmt.Errorf("decode Telegram response (HTTP %s): %w", resp.Status, err)
	}
	if resp.StatusCode != http.StatusOK || !result.OK {
		if result.Description != "" {
			return fmt.Errorf("Telegram API returned HTTP %s: %s", resp.Status, result.Description)
		}
		return fmt.Errorf("Telegram API returned HTTP %s", resp.Status)
	}

	return nil
}
