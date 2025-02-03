package main

import (
	"encoding/json"
	"fmt"
)

type Config struct {
	SenderChain              string  `json:"senderChain"`
	ReceiverChain            string  `json:"receiverChain"`
	FetchTime                int     `json:"fetchTime"`
	APIPort                  int     `json:"apiPort"`
	PurgeOldBlocks           bool    `json:"purgeOldBlocks"`
	PurgeOldMessages         bool    `json:"purgeOldMessages"`
	AggregateBlockAmount     uint64  `json:"aggregateBlockAmount"`
	AlertAvgLatencyMin       float64 `json:"alertAvgLatencyMin"`
	AlertMissingRelayMin     uint64  `json:"alertMissingRelayMin"`
	AlertMissingReceptionMin uint64  `json:"alertMissingReceptionMin"`
	TelegramToken            string  `json:"telegramToken"`
	TelegramChatId           string  `json:"telegramChatId"`
	DiscordWebhookURL        string  `json:"discordWebhookURL"`
	CustomWebhookURL         string  `json:"customWebhookURL"`
}

func parseConfig(data []byte) (*Config, error) {
	// Create config with default values
	config := &Config{
		SenderChain:              "",
		ReceiverChain:            "",
		FetchTime:                1,
		APIPort:                  8800,
		PurgeOldBlocks:           false,
		PurgeOldMessages:         true,
		AggregateBlockAmount:     10,
		AlertAvgLatencyMin:       0,
		AlertMissingRelayMin:     0,
		AlertMissingReceptionMin: 0,
		TelegramToken:            "",
		TelegramChatId:           "",
		DiscordWebhookURL:        "",
		CustomWebhookURL:         "",
	}

	if err := json.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Validate required fields
	if config.SenderChain == "" || config.ReceiverChain == "" {
		return nil, fmt.Errorf("senderChain and receiverChain are required")
	}

	if config.TelegramToken != "" && config.TelegramChatId == "" {
		return nil, fmt.Errorf("telegramChatId must be provided for Telegram alerts")
	}

	return config, nil
}
