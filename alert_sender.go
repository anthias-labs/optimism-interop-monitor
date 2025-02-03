package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func SendAlert(alertType, alertValue string, stats DetailedIntervalStat, config *Config) error {
	statsString, err := json.Marshal(stats)

	if err != nil {
		return err
	}

	message := fmt.Sprintf("Alert: %s at %s\n\n%s", alertType, alertValue, statsString)
	if config.TelegramToken != "" {
		err = telegramMessage(message, config)

		if err != nil {
			return err
		}
	}

	if config.DiscordWebhookURL != "" {
		err = discordMessage(message, config)

		if err != nil {
			return err
		}
	}

	if config.CustomWebhookURL != "" {
		err = customWebhookMessage(message, config)

		if err != nil {
			return err
		}
	}

	return nil
}

func telegramMessage(message string, config *Config) error {
	baseURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", config.TelegramToken)
	params := url.Values{}
	params.Add("text", message)
	params.Add("chat_id", config.TelegramChatId)

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	resp, err := http.Get(fullURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send message, status code: %d", resp.StatusCode)
	}

	return nil
}

func discordMessage(message string, config *Config) error {
	baseURL := config.DiscordWebhookURL

	var body struct {
		Content string `json:"content"`
	}

	body.Content = message

	bodyJson, _ := json.Marshal(body)

	resp, err := http.Post(baseURL, "application/json", bytes.NewBuffer(bodyJson))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send message, status code: %d", resp.StatusCode)
	}

	return nil

}

func customWebhookMessage(message string, config *Config) error {
	baseURL := config.CustomWebhookURL

	var body struct {
		Content string `json:"content"`
	}

	body.Content = message

	bodyJson, _ := json.Marshal(body)

	resp, err := http.Post(baseURL, "application/json", bytes.NewBuffer(bodyJson))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send message, status code: %d", resp.StatusCode)
	}

	return nil

}
