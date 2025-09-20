package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type APIKeyData struct {
	APIKey    string `json:"api_key"`
	Username  string `json:"username"`
	ExpiresAt int64  `json:"expires_at"`
}

func getAPIKeyFile() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "dazedtrader")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return filepath.Join(configDir, "api_key.json"), nil
}

func SaveAPIKey(apiKey, username string, expiresAt int64) error {
	apiKeyFile, err := getAPIKeyFile()
	if err != nil {
		return err
	}

	apiKeyData := APIKeyData{
		APIKey:    apiKey,
		Username:  username,
		ExpiresAt: expiresAt,
	}

	data, err := json.Marshal(apiKeyData)
	if err != nil {
		return fmt.Errorf("failed to marshal API key data: %w", err)
	}

	if err := os.WriteFile(apiKeyFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write API key file: %w", err)
	}

	return nil
}

func LoadAPIKey() (*APIKeyData, error) {
	apiKeyFile, err := getAPIKeyFile()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(apiKeyFile); os.IsNotExist(err) {
		return nil, nil // No API key file exists
	}

	data, err := os.ReadFile(apiKeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read API key file: %w", err)
	}

	var apiKeyData APIKeyData
	if err := json.Unmarshal(data, &apiKeyData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal API key data: %w", err)
	}

	return &apiKeyData, nil
}

func ClearAPIKey() error {
	apiKeyFile, err := getAPIKeyFile()
	if err != nil {
		return err
	}

	if _, err := os.Stat(apiKeyFile); os.IsNotExist(err) {
		return nil // API key file doesn't exist
	}

	if err := os.Remove(apiKeyFile); err != nil {
		return fmt.Errorf("failed to remove API key file: %w", err)
	}

	return nil
}