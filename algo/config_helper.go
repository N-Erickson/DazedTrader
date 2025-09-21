package algo

import (
	"fmt"
	"time"
)

// ConservativeSettings provides pre-configured conservative risk settings
type ConservativeSettings struct {
	TotalBudget     float64 // Total amount willing to risk
	DailyBudget     float64 // Daily amount willing to risk
	MaxPositionPct  float64 // Max % of total budget per position (e.g., 0.1 = 10%)
	StopLossPct     float64 // Stop loss percentage (e.g., 0.03 = 3%)
	DailyLossLimit  float64 // Max daily loss in dollars
}

// GetConservativeEngineConfig returns a conservative engine configuration
func GetConservativeEngineConfig(settings ConservativeSettings) *EngineConfig {
	return &EngineConfig{
		MaxConcurrentStrategies: 2,                              // Limit concurrent strategies
		PriceUpdateInterval:     time.Second * 15,               // 15-second updates (less aggressive)
		OrderRetryLimit:         2,                              // Fewer retries
		OrderTimeout:            time.Minute * 3,                // Shorter timeout
		EmergencyStopLoss:       settings.StopLossPct * 100,     // Convert to percentage
		DailyLossLimit:          settings.DailyLossLimit,        // Daily loss limit
		EnableBacktesting:       true,                           // Enable backtesting
		LogLevel:                "info",
	}
}

// GetConservativeStrategyConfig returns a conservative strategy configuration
func GetConservativeStrategyConfig(name, symbol string, settings ConservativeSettings) StrategyConfig {
	maxPositionSize := settings.TotalBudget * settings.MaxPositionPct
	dailyLossPerStrategy := settings.DailyLossLimit / 3 // Divide among 3 potential strategies

	return StrategyConfig{
		Name:    name,
		Symbol:  symbol,
		Enabled: false, // Always start disabled
		RiskLimits: RiskLimits{
			MaxPositionSize:   maxPositionSize,
			StopLossPercent:   settings.StopLossPct * 100,        // Convert to percentage
			TakeProfitPercent: settings.StopLossPct * 200,        // 2x stop loss for take profit
			MaxDailyTrades:    3,                                 // Very conservative trade limit
			MaxDailyLoss:      dailyLossPerStrategy,
			CooldownPeriod:    time.Minute * 15,                  // Long cooldown
		},
		BacktestPeriod: time.Hour * 48, // 48-hour backtest
	}
}

// PresetConservativeSettings provides preset conservative configurations
func PresetConservativeSettings() map[string]ConservativeSettings {
	return map[string]ConservativeSettings{
		"ultra_conservative": {
			TotalBudget:    100.0,  // $100 total budget
			DailyBudget:    10.0,   // $10 daily budget
			MaxPositionPct: 0.05,   // 5% per position = $5 max
			StopLossPct:    0.02,   // 2% stop loss
			DailyLossLimit: 10.0,   // $10 daily loss limit
		},
		"conservative": {
			TotalBudget:    500.0,  // $500 total budget
			DailyBudget:    50.0,   // $50 daily budget
			MaxPositionPct: 0.1,    // 10% per position = $50 max
			StopLossPct:    0.03,   // 3% stop loss
			DailyLossLimit: 25.0,   // $25 daily loss limit
		},
		"moderate": {
			TotalBudget:    1000.0, // $1000 total budget
			DailyBudget:    100.0,  // $100 daily budget
			MaxPositionPct: 0.15,   // 15% per position = $150 max
			StopLossPct:    0.05,   // 5% stop loss
			DailyLossLimit: 50.0,   // $50 daily loss limit
		},
	}
}

// GenerateConservativeConfigs generates configuration files with conservative settings
func (cm *ConfigManager) GenerateConservativeConfigs(preset string) error {
	presets := PresetConservativeSettings()
	settings, exists := presets[preset]
	if !exists {
		return fmt.Errorf("preset '%s' not found. Available: ultra_conservative, conservative, moderate", preset)
	}

	// Generate engine config
	engineConfig := GetConservativeEngineConfig(settings)
	if err := cm.SaveEngineConfig(engineConfig); err != nil {
		return fmt.Errorf("failed to save engine config: %v", err)
	}

	// Generate strategy configs
	strategies := []struct {
		name   string
		symbol string
		params map[string]interface{}
	}{
		{
			name:   "btc_conservative_ma",
			symbol: "BTC-USD",
			params: map[string]interface{}{
				"short_period": 10.0, // Longer periods for less noise
				"long_period":  30.0,
				"min_volume":   2000.0,
			},
		},
		{
			name:   "eth_conservative_momentum",
			symbol: "ETH-USD",
			params: map[string]interface{}{
				"lookback_period":    15.0, // Longer lookback
				"momentum_threshold": 0.03, // Higher threshold
				"volume_multiplier":  2.0,  // Require 2x volume
			},
		},
	}

	for _, strategy := range strategies {
		config := GetConservativeStrategyConfig(strategy.name, strategy.symbol, settings)
		config.Parameters = strategy.params

		if err := cm.SaveStrategyConfig(&config); err != nil {
			return fmt.Errorf("failed to save strategy config %s: %v", strategy.name, err)
		}
	}

	return nil
}

// CalculateTotalExposure calculates the maximum possible exposure across all strategies
func CalculateTotalExposure(configs []StrategyConfig) float64 {
	totalExposure := 0.0
	for _, config := range configs {
		if config.Enabled {
			totalExposure += config.RiskLimits.MaxPositionSize
		}
	}
	return totalExposure
}

// ValidateRiskLimits validates that risk limits are reasonable
func ValidateRiskLimits(configs []StrategyConfig, maxTotalExposure float64) error {
	totalExposure := CalculateTotalExposure(configs)

	if totalExposure > maxTotalExposure {
		return fmt.Errorf("total exposure $%.2f exceeds maximum allowed $%.2f",
			totalExposure, maxTotalExposure)
	}

	for _, config := range configs {
		if config.RiskLimits.MaxPositionSize < 10.0 {
			return fmt.Errorf("strategy %s has position size too small ($%.2f < $10)",
				config.Name, config.RiskLimits.MaxPositionSize)
		}

		if config.RiskLimits.StopLossPercent > 20.0 {
			return fmt.Errorf("strategy %s has stop loss too large (%.1f%% > 20%%)",
				config.Name, config.RiskLimits.StopLossPercent)
		}
	}

	return nil
}