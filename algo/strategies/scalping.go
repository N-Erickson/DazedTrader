package strategies

import (
	"dazedtrader/algo"
	"fmt"
	"time"
)

// ScalpingStrategy implements a scalping strategy for quick trades
type ScalpingStrategy struct {
	name           string
	symbol         string
	shortPeriod    int
	longPeriod     int
	volatilityThreshold float64

	// State
	prices         []float64
	shortMA        float64
	longMA         float64
	lastSignal     algo.SignalType
	lastSignalTime time.Time
	position       algo.Position
	config         algo.StrategyConfig

	// Buffers for MA calculation
	shortBuffer    []float64
	longBuffer     []float64
}

// NewScalpingStrategy creates a new scalping strategy
func NewScalpingStrategy(name, symbol string) *ScalpingStrategy {
	return &ScalpingStrategy{
		name:        name,
		symbol:      symbol,
		shortPeriod: 3,  // Very short period for scalping
		longPeriod:  10, // Short long period for quick signals
		volatilityThreshold: 0.01, // 1% volatility threshold
		prices:      make([]float64, 0),
		shortBuffer: make([]float64, 0),
		longBuffer:  make([]float64, 0),
		lastSignal:  algo.SignalHold,
	}
}

// GetName returns the strategy name
func (ss *ScalpingStrategy) GetName() string {
	return ss.name
}

// GetSymbol returns the symbol this strategy trades
func (ss *ScalpingStrategy) GetSymbol() string {
	return ss.symbol
}

// Initialize initializes the strategy with configuration
func (ss *ScalpingStrategy) Initialize(config algo.StrategyConfig) error {
	ss.config = config

	// Parse strategy-specific parameters
	if params, ok := config.Parameters["short_period"].(float64); ok {
		ss.shortPeriod = int(params)
	}
	if params, ok := config.Parameters["long_period"].(float64); ok {
		ss.longPeriod = int(params)
	}
	if params, ok := config.Parameters["volatility_threshold"].(float64); ok {
		ss.volatilityThreshold = params
	}

	// Validate parameters
	if ss.shortPeriod >= ss.longPeriod {
		return fmt.Errorf("short period (%d) must be less than long period (%d)", ss.shortPeriod, ss.longPeriod)
	}

	// Initialize buffers
	ss.shortBuffer = make([]float64, ss.shortPeriod)
	ss.longBuffer = make([]float64, ss.longPeriod)

	return nil
}

// ProcessTick processes a new price tick and generates trading signals
func (ss *ScalpingStrategy) ProcessTick(tick algo.PriceTick) ([]algo.Signal, error) {
	// Add price to history
	ss.prices = append(ss.prices, tick.Price)

	// Keep only what we need for calculations
	maxHistory := ss.longPeriod * 2
	if len(ss.prices) > maxHistory {
		ss.prices = ss.prices[len(ss.prices)-maxHistory:]
	}

	// Update moving averages
	ss.updateMovingAverages()

	// Generate signals (more aggressive than regular MA strategy)
	signals := ss.generateScalpingSignals(tick)

	return signals, nil
}

// updateMovingAverages calculates the current moving averages
func (ss *ScalpingStrategy) updateMovingAverages() {
	if len(ss.prices) < ss.longPeriod {
		return // Not enough data yet
	}

	// Calculate short MA
	shortSum := 0.0
	for i := len(ss.prices) - ss.shortPeriod; i < len(ss.prices); i++ {
		shortSum += ss.prices[i]
	}
	ss.shortMA = shortSum / float64(ss.shortPeriod)

	// Calculate long MA
	longSum := 0.0
	for i := len(ss.prices) - ss.longPeriod; i < len(ss.prices); i++ {
		longSum += ss.prices[i]
	}
	ss.longMA = longSum / float64(ss.longPeriod)
}

// generateScalpingSignals generates more aggressive trading signals for scalping
func (ss *ScalpingStrategy) generateScalpingSignals(tick algo.PriceTick) []algo.Signal {
	signals := make([]algo.Signal, 0)

	// Need both MAs calculated
	if ss.shortMA == 0 || ss.longMA == 0 {
		return signals
	}

	// Check for crossover signals with volatility filter
	var signal algo.SignalType
	confidence := 0.6 // Higher base confidence for scalping

	// Calculate price volatility
	if len(ss.prices) >= 5 {
		recentPrices := ss.prices[len(ss.prices)-5:]
		volatility := ss.calculateVolatility(recentPrices)

		// Only trade if volatility is above threshold (ensures movement)
		if volatility < ss.volatilityThreshold {
			return signals
		}
	}

	// Bullish crossover: short MA crosses above long MA
	if ss.shortMA > ss.longMA {
		signal = algo.SignalBuy
		spread := (ss.shortMA - ss.longMA) / ss.longMA
		confidence = 0.6 + (spread * 15) // More aggressive scaling
	} else if ss.shortMA < ss.longMA {
		// Bearish crossover: short MA crosses below long MA
		signal = algo.SignalSell
		spread := (ss.longMA - ss.shortMA) / ss.longMA
		confidence = 0.6 + (spread * 15)
	} else {
		signal = algo.SignalHold
	}

	if confidence > 1.0 {
		confidence = 1.0
	}

	// Very short cooldown for scalping (30 seconds instead of 5 minutes)
	if signal != ss.lastSignal && signal != algo.SignalHold {
		if time.Since(ss.lastSignalTime) >= time.Second*30 {
			quantity := ss.calculatePositionSize(tick.Price, confidence)

			if quantity > 0 {
				signals = append(signals, algo.Signal{
					Type:       signal,
					Symbol:     ss.symbol,
					Price:      tick.Price,
					Quantity:   quantity,
					Timestamp:  tick.Timestamp,
					Confidence: confidence,
					Metadata: map[string]interface{}{
						"short_ma":     ss.shortMA,
						"long_ma":      ss.longMA,
						"spread":       ss.shortMA - ss.longMA,
						"spread_pct":   ((ss.shortMA - ss.longMA) / ss.longMA) * 100,
						"strategy":     "scalping",
					},
				})

				ss.lastSignal = signal
				ss.lastSignalTime = tick.Timestamp
			}
		}
	}

	return signals
}

// calculateVolatility calculates price volatility
func (ss *ScalpingStrategy) calculateVolatility(prices []float64) float64 {
	if len(prices) < 2 {
		return 0
	}

	mean := 0.0
	for _, price := range prices {
		mean += price
	}
	mean /= float64(len(prices))

	variance := 0.0
	for _, price := range prices {
		variance += (price - mean) * (price - mean)
	}
	variance /= float64(len(prices))

	return variance / (mean * mean) // Relative volatility
}

// calculatePositionSize calculates the position size for scalping
func (ss *ScalpingStrategy) calculatePositionSize(price, confidence float64) float64 {
	// Smaller position sizes for scalping (more trades, less risk per trade)
	maxPositionValue := ss.config.RiskLimits.MaxPositionSize * 0.8 // Use 80% of max

	// Scale by confidence
	scaledPositionValue := maxPositionValue * confidence

	// Convert to quantity
	quantity := scaledPositionValue / price

	// Apply minimum position size filter (reduced for ultra conservative mode)
	minPositionValue := 3.0 // Even lower minimum for scalping
	if quantity*price < minPositionValue {
		return 0
	}

	return quantity
}

// GetState returns the current strategy state
func (ss *ScalpingStrategy) GetState() algo.StrategyState {
	state := algo.StrategyState{
		Name:           ss.name,
		Symbol:         ss.symbol,
		IsRunning:      true,
		Position:       ss.position,
		LastUpdateTime: time.Now(),
		Metadata: map[string]interface{}{
			"short_ma":          ss.shortMA,
			"long_ma":           ss.longMA,
			"short_period":      ss.shortPeriod,
			"long_period":       ss.longPeriod,
			"volatility_threshold": ss.volatilityThreshold,
			"prices_count":      len(ss.prices),
			"last_signal":       string(ss.lastSignal),
		},
	}

	return state
}

// Stop stops the strategy
func (ss *ScalpingStrategy) Stop() error {
	ss.prices = nil
	ss.shortBuffer = nil
	ss.longBuffer = nil
	return nil
}

// Reset resets the strategy state
func (ss *ScalpingStrategy) Reset() error {
	ss.prices = make([]float64, 0)
	ss.shortBuffer = make([]float64, ss.shortPeriod)
	ss.longBuffer = make([]float64, ss.longPeriod)
	ss.shortMA = 0
	ss.longMA = 0
	ss.lastSignal = algo.SignalHold
	ss.lastSignalTime = time.Time{}
	return nil
}