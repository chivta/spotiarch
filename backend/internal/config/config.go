package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	DatabaseURL string      `json:"database_url"`
	Token       TokenConfig `json:"token"`
	Log         LogConfig   `json:"log"`
}

type TokenConfig struct {
	SecretKey     string `json:"secret_key"`
	TokenDuration int    `json:"duration_hours"` // in hours
}

type LogConfig struct {
	Debug       bool   `json:"debug"`
	EnableInfo  bool   `json:"enable_info"`
	ErrorOutput string `json:"error_output"`
	InfoOutput  string `json:"info_output"`
	DebugOutput string `json:"debug_output"`
}

func Load() (*Config, error) {
	data, err := os.ReadFile("./config.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read config.json: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config.json: %w", err)
	}

	return &config, nil
}
