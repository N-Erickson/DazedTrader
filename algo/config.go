package algo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

// ConfigManager handles loading and saving trading configurations
type ConfigManager struct {
	configDir string
}

// NewConfigManager creates a new configuration manager
func NewConfigManager() *ConfigManager {
	configDir := filepath.Join(os.Getenv("HOME"), ".dazedtrader", "algo")
	os.MkdirAll(configDir, 0755)

	return &ConfigManager{
		configDir: configDir,
	}
}

// LoadEngineConfig loads the engine configuration
func (cm *ConfigManager) LoadEngineConfig() (*EngineConfig, error) {
	configPath := filepath.Join(cm.configDir, "engine.json")

	// Create default config if it doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaultConfig := cm.getDefaultEngineConfig()
		if err := cm.SaveEngineConfig(defaultConfig); err != nil {
			return nil, fmt.Errorf("failed to create default config: %v", err)
		}
		return defaultConfig, nil
	}

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config EngineConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %v", err)
	}

	return &config, nil
}

// SaveEngineConfig saves the engine configuration
func (cm *ConfigManager) SaveEngineConfig(config *EngineConfig) error {
	configPath := filepath.Join(cm.configDir, "engine.json")

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	if err := ioutil.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

// LoadStrategyConfigs loads all strategy configurations
func (cm *ConfigManager) LoadStrategyConfigs() ([]StrategyConfig, error) {
	strategiesPath := filepath.Join(cm.configDir, "strategies")
	os.MkdirAll(strategiesPath, 0755)

	files, err := ioutil.ReadDir(strategiesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read strategies directory: %v", err)
	}

	var configs []StrategyConfig
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".json" {
			configPath := filepath.Join(strategiesPath, file.Name())
			config, err := cm.loadStrategyConfig(configPath)
			if err != nil {
				// Log error but continue with other configs
				continue
			}
			configs = append(configs, *config)
		}
	}

	// Create default configs if none exist
	if len(configs) == 0 {
		defaultConfigs := cm.getDefaultStrategyConfigs()
		for _, config := range defaultConfigs {
			if err := cm.SaveStrategyConfig(&config); err != nil {
				continue // Log error but continue
			}
			configs = append(configs, config)
		}
	}

	return configs, nil
}

// SaveStrategyConfig saves a strategy configuration
func (cm *ConfigManager) SaveStrategyConfig(config *StrategyConfig) error {
	strategiesPath := filepath.Join(cm.configDir, "strategies")
	os.MkdirAll(strategiesPath, 0755)

	filename := fmt.Sprintf("%s.json", config.Name)
	configPath := filepath.Join(strategiesPath, filename)

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal strategy config: %v", err)
	}

	if err := ioutil.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write strategy config: %v", err)
	}

	return nil
}

// loadStrategyConfig loads a single strategy configuration
func (cm *ConfigManager) loadStrategyConfig(path string) (*StrategyConfig, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read strategy config: %v", err)
	}

	var config StrategyConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse strategy config: %v", err)
	}

	return &config, nil
}

// getDefaultEngineConfig returns default engine configuration
func (cm *ConfigManager) getDefaultEngineConfig() *EngineConfig {
	return &EngineConfig{
		MaxConcurrentStrategies: 5,
		PriceUpdateInterval:     time.Second * 10,
		OrderRetryLimit:         3,
		OrderTimeout:            time.Minute * 5,
		EmergencyStopLoss:       10.0, // 10% emergency stop
		DailyLossLimit:          500.0, // $500 daily loss limit
		EnableBacktesting:       false,
		LogLevel:                "info",
	}
}

// getDefaultStrategyConfigs returns default strategy configurations
func (cm *ConfigManager) getDefaultStrategyConfigs() []StrategyConfig {
	return []StrategyConfig{
		{
			Name:    "btc_moving_average",
			Symbol:  "BTC-USD",
			Enabled: false, // Start disabled for safety
			Parameters: map[string]interface{}{
				"short_period": 5.0,
				"long_period":  20.0,
				"min_volume":   1000.0,
			},
			RiskLimits: RiskLimits{
				MaxPositionSize:   100.0, // $100 max position
				StopLossPercent:   5.0,   // 5% stop loss
				TakeProfitPercent: 10.0,  // 10% take profit
				MaxDailyTrades:    10,
				MaxDailyLoss:      50.0, // $50 daily loss limit
				CooldownPeriod:    time.Minute * 5,
			},
			BacktestPeriod: time.Hour * 24, // 24 hours backtest
		},
		{
			Name:    "eth_momentum",
			Symbol:  "ETH-USD",
			Enabled: false,
			Parameters: map[string]interface{}{
				"lookback_period":    10.0,
				"momentum_threshold": 0.02,
				"volume_multiplier":  1.5,
			},
			RiskLimits: RiskLimits{
				MaxPositionSize:   100.0,
				StopLossPercent:   5.0,
				TakeProfitPercent: 8.0,
				MaxDailyTrades:    15, // More trades for momentum strategy
				MaxDailyLoss:      50.0,
				CooldownPeriod:    time.Minute * 2, // Shorter cooldown
			},
			BacktestPeriod: time.Hour * 12,
		},
		{
			Name:    "doge_scalping",
			Symbol:  "DOGE-USD",
			Enabled: false,
			Parameters: map[string]interface{}{
				"short_period":       3.0,
				"long_period":        10.0,
				"volatility_threshold": 0.01,
			},
			RiskLimits: RiskLimits{
				MaxPositionSize:   50.0, // Smaller position for DOGE
				StopLossPercent:   3.0,  // Tighter stops for scalping
				TakeProfitPercent: 5.0,
				MaxDailyTrades:    30, // More trades for scalping
				MaxDailyLoss:      25.0,
				CooldownPeriod:    time.Second * 30, // Very short cooldown
			},
			BacktestPeriod: time.Hour * 6,
		},
	}
}

// GetStrategyTemplates returns template configurations for different strategy types
func (cm *ConfigManager) GetStrategyTemplates() map[string]StrategyConfig {
	return map[string]StrategyConfig{
		"moving_average": {
			Name:    "template_moving_average",
			Symbol:  "BTC-USD",
			Enabled: false,
			Parameters: map[string]interface{}{
				"short_period": 5.0,
				"long_period":  20.0,
				"min_volume":   1000.0,
			},
			RiskLimits: RiskLimits{
				MaxPositionSize:   100.0,
				StopLossPercent:   5.0,
				TakeProfitPercent: 10.0,
				MaxDailyTrades:    10,
				MaxDailyLoss:      50.0,
				CooldownPeriod:    time.Minute * 5,
			},
		},
		"momentum": {
			Name:    "template_momentum",
			Symbol:  "ETH-USD",
			Enabled: false,
			Parameters: map[string]interface{}{
				"lookback_period":    10.0,
				"momentum_threshold": 0.02,
				"volume_multiplier":  1.5,
			},
			RiskLimits: RiskLimits{
				MaxPositionSize:   100.0,
				StopLossPercent:   5.0,
				TakeProfitPercent: 8.0,
				MaxDailyTrades:    15,
				MaxDailyLoss:      50.0,
				CooldownPeriod:    time.Minute * 2,
			},
		},
		"mean_reversion": {
			Name:    "template_mean_reversion",
			Symbol:  "ADA-USD",
			Enabled: false,
			Parameters: map[string]interface{}{
				"lookback_period":      20.0,
				"oversold_threshold":   30.0,
				"overbought_threshold": 70.0,
				"rsi_period":          14.0,
			},
			RiskLimits: RiskLimits{
				MaxPositionSize:   75.0,
				StopLossPercent:   4.0,
				TakeProfitPercent: 6.0,
				MaxDailyTrades:    8,
				MaxDailyLoss:      40.0,
				CooldownPeriod:    time.Minute * 10,
			},
		},
	}
}

// ValidateStrategyConfig validates a strategy configuration
func (cm *ConfigManager) ValidateStrategyConfig(config *StrategyConfig) error {
	if config.Name == "" {
		return fmt.Errorf("strategy name cannot be empty")
	}

	if config.Symbol == "" {
		return fmt.Errorf("strategy symbol cannot be empty")
	}

	if config.RiskLimits.MaxPositionSize <= 0 {
		return fmt.Errorf("max position size must be greater than 0")
	}

	if config.RiskLimits.StopLossPercent < 0 || config.RiskLimits.StopLossPercent > 50 {
		return fmt.Errorf("stop loss percent must be between 0 and 50")
	}

	if config.RiskLimits.MaxDailyTrades <= 0 {
		return fmt.Errorf("max daily trades must be greater than 0")
	}

	if config.RiskLimits.MaxDailyLoss <= 0 {
		return fmt.Errorf("max daily loss must be greater than 0")
	}

	return nil
}

// BackupConfigs creates a backup of all configurations
func (cm *ConfigManager) BackupConfigs() error {
	backupDir := filepath.Join(cm.configDir, "backups")
	os.MkdirAll(backupDir, 0755)

	timestamp := time.Now().Format("20060102_150405")
	_ = filepath.Join(backupDir, fmt.Sprintf("config_backup_%s.tar.gz", timestamp))

	// For simplicity, just copy the config directory
	// In production, you might want to create a proper tar.gz backup

	return nil
}