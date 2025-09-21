package main

import (
	"dazedtrader/algo"
	"flag"
	"fmt"
	"log"
)

func main() {
	var (
		preset = flag.String("preset", "conservative", "Risk preset: ultra_conservative, conservative, moderate")
		budget = flag.Float64("budget", 0, "Total budget in USD (overrides preset)")
		daily  = flag.Float64("daily", 0, "Daily loss limit in USD (overrides preset)")
		show   = flag.Bool("show", false, "Show current configuration")
		help   = flag.Bool("help", false, "Show help")
	)
	flag.Parse()

	if *help {
		fmt.Println("DazedTrader Algorithmic Trading Configuration Tool")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  ./configure_algo -preset=conservative")
		fmt.Println("  ./configure_algo -budget=500 -daily=25")
		fmt.Println("  ./configure_algo -show")
		fmt.Println()
		fmt.Println("Presets:")
		fmt.Println("  ultra_conservative: $100 budget, $5 max position, 2% stop loss")
		fmt.Println("  conservative:       $500 budget, $50 max position, 3% stop loss")
		fmt.Println("  moderate:          $1000 budget, $150 max position, 5% stop loss")
		fmt.Println()
		fmt.Println("Options:")
		flag.PrintDefaults()
		return
	}

	configManager := algo.NewConfigManager()

	if *show {
		showCurrentConfig(configManager)
		return
	}

	// Generate configuration based on preset or custom values
	if *budget > 0 || *daily > 0 {
		generateCustomConfig(configManager, *budget, *daily)
	} else {
		generatePresetConfig(configManager, *preset)
	}
}

func showCurrentConfig(cm *algo.ConfigManager) {
	fmt.Println("🤖 Current Algorithmic Trading Configuration")
	fmt.Println("═══════════════════════════════════════════")

	// Load engine config
	engineConfig, err := cm.LoadEngineConfig()
	if err != nil {
		fmt.Printf("❌ Error loading engine config: %v\n", err)
		return
	}

	fmt.Printf("💰 Engine-Wide Limits:\n")
	fmt.Printf("   Emergency Stop Loss: %.1f%%\n", engineConfig.EmergencyStopLoss)
	fmt.Printf("   Daily Loss Limit:    $%.2f\n", engineConfig.DailyLossLimit)
	fmt.Printf("   Max Strategies:      %d\n", engineConfig.MaxConcurrentStrategies)
	fmt.Printf("   Price Update:        %v\n", engineConfig.PriceUpdateInterval)

	// Load strategy configs
	strategyConfigs, err := cm.LoadStrategyConfigs()
	if err != nil {
		fmt.Printf("❌ Error loading strategy configs: %v\n", err)
		return
	}

	fmt.Printf("\n📊 Strategy Limits:\n")
	totalExposure := 0.0
	for i, config := range strategyConfigs {
		status := "⚪ Disabled"
		if config.Enabled {
			status = "🟢 Enabled"
			totalExposure += config.RiskLimits.MaxPositionSize
		}

		fmt.Printf("%d. %s (%s) %s\n", i+1, config.Name, config.Symbol, status)
		fmt.Printf("   Max Position:     $%.2f\n", config.RiskLimits.MaxPositionSize)
		fmt.Printf("   Stop Loss:        %.1f%%\n", config.RiskLimits.StopLossPercent)
		fmt.Printf("   Daily Trades:     %d\n", config.RiskLimits.MaxDailyTrades)
		fmt.Printf("   Daily Loss Limit: $%.2f\n", config.RiskLimits.MaxDailyLoss)
		fmt.Printf("   Cooldown:         %v\n", config.RiskLimits.CooldownPeriod)
		fmt.Println()
	}

	fmt.Printf("💡 Total Potential Exposure: $%.2f\n", totalExposure)
	fmt.Printf("📁 Config Location: ~/.dazedtrader/algo/\n")
}

func generatePresetConfig(cm *algo.ConfigManager, preset string) {
	fmt.Printf("🛠️  Generating '%s' configuration...\n", preset)

	err := cm.GenerateConservativeConfigs(preset)
	if err != nil {
		log.Fatalf("❌ Failed to generate config: %v", err)
	}

	fmt.Println("✅ Configuration generated successfully!")
	fmt.Println()

	// Show what was created
	presets := algo.PresetConservativeSettings()
	if settings, exists := presets[preset]; exists {
		fmt.Printf("📋 Generated Configuration:\n")
		fmt.Printf("   Total Budget:      $%.2f\n", settings.TotalBudget)
		fmt.Printf("   Daily Loss Limit:  $%.2f\n", settings.DailyLossLimit)
		fmt.Printf("   Max Position Size: $%.2f (%.0f%% of budget)\n",
			settings.TotalBudget*settings.MaxPositionPct,
			settings.MaxPositionPct*100)
		fmt.Printf("   Stop Loss:         %.1f%%\n", settings.StopLossPct*100)
		fmt.Printf("   Strategies:        2 (disabled by default)\n")
		fmt.Println()
		fmt.Println("💡 Enable strategies in the Algorithmic Trading interface")
		fmt.Println("⚠️  Always test with small amounts first!")
	}
}

func generateCustomConfig(cm *algo.ConfigManager, budget, dailyLimit float64) {
	if budget <= 0 {
		budget = 500.0 // Default budget
	}
	if dailyLimit <= 0 {
		dailyLimit = budget * 0.05 // Default to 5% of budget
	}

	settings := algo.ConservativeSettings{
		TotalBudget:    budget,
		DailyBudget:    dailyLimit * 2, // 2x daily limit for daily budget
		MaxPositionPct: 0.1,            // 10% per position
		StopLossPct:    0.03,           // 3% stop loss
		DailyLossLimit: dailyLimit,
	}

	fmt.Printf("🛠️  Generating custom configuration...\n")
	fmt.Printf("   Budget: $%.2f, Daily Limit: $%.2f\n", budget, dailyLimit)

	// Generate engine config
	engineConfig := algo.GetConservativeEngineConfig(settings)
	if err := cm.SaveEngineConfig(engineConfig); err != nil {
		log.Fatalf("❌ Failed to save engine config: %v", err)
	}

	// Generate strategy configs
	btcConfig := algo.GetConservativeStrategyConfig("btc_custom", "BTC-USD", settings)
	btcConfig.Parameters = map[string]interface{}{
		"short_period": 10.0,
		"long_period":  30.0,
		"min_volume":   2000.0,
	}

	ethConfig := algo.GetConservativeStrategyConfig("eth_custom", "ETH-USD", settings)
	ethConfig.Parameters = map[string]interface{}{
		"lookback_period":    15.0,
		"momentum_threshold": 0.03,
		"volume_multiplier":  2.0,
	}

	if err := cm.SaveStrategyConfig(&btcConfig); err != nil {
		log.Fatalf("❌ Failed to save BTC config: %v", err)
	}

	if err := cm.SaveStrategyConfig(&ethConfig); err != nil {
		log.Fatalf("❌ Failed to save ETH config: %v", err)
	}

	fmt.Println("✅ Custom configuration generated successfully!")
	fmt.Printf("   Max Position Size: $%.2f per strategy\n", settings.TotalBudget*settings.MaxPositionPct)
	fmt.Printf("   Total Exposure:    $%.2f (if both strategies enabled)\n", settings.TotalBudget*settings.MaxPositionPct*2)
	fmt.Println("⚠️  Always test with small amounts first!")
}