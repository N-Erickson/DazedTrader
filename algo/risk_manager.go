package algo

import (
	"time"
)

// RiskManager handles risk management for trading strategies
type RiskManager struct {
	config           *EngineConfig
	dailyTradeCount  map[string]int
	dailyLoss        map[string]float64
	lastResetTime    time.Time
}

// NewRiskManager creates a new risk manager
func NewRiskManager(config *EngineConfig) *RiskManager {
	return &RiskManager{
		config:          config,
		dailyTradeCount: make(map[string]int),
		dailyLoss:       make(map[string]float64),
		lastResetTime:   time.Now().Truncate(24 * time.Hour),
	}
}

// ValidateSignal validates a trading signal against risk limits
func (rm *RiskManager) ValidateSignal(signal Signal, limits RiskLimits, state StrategyState) bool {
	// Reset daily counters if needed
	rm.resetDailyCountersIfNeeded()

	strategyKey := state.Name

	// Check daily trade limit
	if rm.dailyTradeCount[strategyKey] >= limits.MaxDailyTrades {
		return false
	}

	// Check daily loss limit
	if rm.dailyLoss[strategyKey] >= limits.MaxDailyLoss {
		return false
	}

	// Check position size limit
	if signal.Type == SignalBuy {
		positionValue := signal.Quantity * signal.Price
		if positionValue > limits.MaxPositionSize {
			return false
		}

		// Check if this would exceed total position limit
		currentPositionValue := state.Position.Quantity * signal.Price
		if currentPositionValue+positionValue > limits.MaxPositionSize {
			return false
		}
	}

	// Check cooldown period
	if state.LastSignal != nil {
		timeSinceLastSignal := time.Since(state.LastSignal.Timestamp)
		if timeSinceLastSignal < limits.CooldownPeriod {
			return false
		}
	}

	// Check confidence threshold (optional)
	if signal.Confidence < 0.5 { // Minimum 50% confidence
		return false
	}

	return true
}

// RecordTrade records a trade for risk tracking
func (rm *RiskManager) RecordTrade(strategyName string, trade Trade) {
	rm.resetDailyCountersIfNeeded()

	// Increment trade count
	rm.dailyTradeCount[strategyName]++

	// Record loss if applicable
	if trade.PnL < 0 {
		rm.dailyLoss[strategyName] += -trade.PnL
	}
}

// ShouldStopStrategy determines if a strategy should be stopped due to risk limits
func (rm *RiskManager) ShouldStopStrategy(strategyName string, limits RiskLimits, currentPnL float64) bool {
	rm.resetDailyCountersIfNeeded()

	// Check daily loss limit
	if rm.dailyLoss[strategyName] >= limits.MaxDailyLoss {
		return true
	}

	// Check stop loss percentage
	if limits.StopLossPercent > 0 && currentPnL < 0 {
		// This would need initial capital tracking per strategy
		// For now, use a simple absolute loss check
		if -currentPnL >= limits.MaxDailyLoss {
			return true
		}
	}

	return false
}

// GetDailyStats returns daily statistics for a strategy
func (rm *RiskManager) GetDailyStats(strategyName string) (int, float64) {
	rm.resetDailyCountersIfNeeded()
	return rm.dailyTradeCount[strategyName], rm.dailyLoss[strategyName]
}

// resetDailyCountersIfNeeded resets daily counters if a new day has started
func (rm *RiskManager) resetDailyCountersIfNeeded() {
	now := time.Now().Truncate(24 * time.Hour)
	if now.After(rm.lastResetTime) {
		rm.dailyTradeCount = make(map[string]int)
		rm.dailyLoss = make(map[string]float64)
		rm.lastResetTime = now
	}
}

// CalculatePositionSize calculates appropriate position size based on risk parameters
func (rm *RiskManager) CalculatePositionSize(price float64, limits RiskLimits, accountBalance float64) float64 {
	// Simple position sizing based on maximum position size
	maxShares := limits.MaxPositionSize / price

	// Kelly Criterion could be implemented here for more sophisticated sizing
	// For now, use simple fixed percentage of max position size
	return maxShares * 0.5 // Use 50% of max allowed position
}

// ValidateOrder performs final validation before order submission
func (rm *RiskManager) ValidateOrder(order OrderRequest, limits RiskLimits) bool {
	// Basic order validation
	if order.Quantity <= 0 {
		return false
	}

	// Check minimum order value (avoid dust orders)
	if order.Price > 0 && order.Quantity*order.Price < 1.0 { // Minimum $1 order
		return false
	}

	return true
}