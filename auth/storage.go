package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type TokenData struct {
	Token     string `json:"token"`
	Username  string `json:"username"`
	ExpiresAt int64  `json:"expires_at"`
}

func getTokenFile() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "dazedtrader")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return filepath.Join(configDir, "auth_token.json"), nil
}

func SaveToken(token, username string, expiresAt int64) error {
	tokenFile, err := getTokenFile()
	if err != nil {
		return err
	}

	tokenData := TokenData{
		Token:     token,
		Username:  username,
		ExpiresAt: expiresAt,
	}

	data, err := json.Marshal(tokenData)
	if err != nil {
		return fmt.Errorf("failed to marshal token data: %w", err)
	}

	if err := os.WriteFile(tokenFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	return nil
}

func LoadToken() (*TokenData, error) {
	tokenFile, err := getTokenFile()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(tokenFile); os.IsNotExist(err) {
		return nil, nil // No token file exists
	}

	data, err := os.ReadFile(tokenFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read token file: %w", err)
	}

	var tokenData TokenData
	if err := json.Unmarshal(data, &tokenData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token data: %w", err)
	}

	return &tokenData, nil
}

func ClearToken() error {
	tokenFile, err := getTokenFile()
	if err != nil {
		return err
	}

	if _, err := os.Stat(tokenFile); os.IsNotExist(err) {
		return nil // Token file doesn't exist
	}

	if err := os.Remove(tokenFile); err != nil {
		return fmt.Errorf("failed to remove token file: %w", err)
	}

	return nil
}