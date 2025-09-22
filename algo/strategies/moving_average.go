package strategies

import (
	"dazedtrader/algo"
	"fmt"
	"time"
)

// MovingAverageStrategy implements a moving average crossover strategy
type MovingAverageStrategy struct {
	name           string
	symbol         string
	shortPeriod    int
	longPeriod     int
	minVolume      float64

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

// NewMovingAverageStrategy creates a new moving average strategy
func NewMovingAverageStrategy(name, symbol string) *MovingAverageStrategy {
	return &MovingAverageStrategy{
		name:        name,
		symbol:      symbol,
		shortPeriod: 5,  // Default 5-period short MA
		longPeriod:  20, // Default 20-period long MA
		minVolume:   1000, // Minimum volume filter
		prices:      make([]float64, 0),
		shortBuffer: make([]float64, 0),
		longBuffer:  make([]float64, 0),
		lastSignal:  algo.SignalHold,
	}
}

// GetName returns the strategy name
func (mas *MovingAverageStrategy) GetName() string {
	return mas.name
}

// GetSymbol returns the symbol this strategy trades
func (mas *MovingAverageStrategy) GetSymbol() string {
	return mas.symbol
}

// Initialize initializes the strategy with configuration
func (mas *MovingAverageStrategy) Initialize(config algo.StrategyConfig) error {
	mas.config = config

	// Parse strategy-specific parameters
	if params, ok := config.Parameters["short_period"].(float64); ok {
		mas.shortPeriod = int(params)
	}
	if params, ok := config.Parameters["long_period"].(float64); ok {
		mas.longPeriod = int(params)
	}
	if params, ok := config.Parameters["min_volume"].(float64); ok {
		mas.minVolume = params
	}

	// Validate parameters
	if mas.shortPeriod >= mas.longPeriod {
		return fmt.Errorf("short period (%d) must be less than long period (%d)", mas.shortPeriod, mas.longPeriod)
	}
	if mas.shortPeriod < 2 || mas.longPeriod < 2 {
		return fmt.Errorf("periods must be at least 2")
	}

	// Initialize buffers
	mas.shortBuffer = make([]float64, mas.shortPeriod)
	mas.longBuffer = make([]float64, mas.longPeriod)

	return nil
}

// ProcessTick processes a new price tick and generates trading signals
func (mas *MovingAverageStrategy) ProcessTick(tick algo.PriceTick) ([]algo.Signal, error) {
	// Add price to history
	mas.prices = append(mas.prices, tick.Price)

	// Keep only what we need for calculations
	maxHistory := mas.longPeriod * 2
	if len(mas.prices) > maxHistory {
		mas.prices = mas.prices[len(mas.prices)-maxHistory:]
	}

	// Update moving averages
	mas.updateMovingAverages()

	// Generate signals
	signals := mas.generateSignals(tick)

	return signals, nil
}

// updateMovingAverages calculates the current moving averages
func (mas *MovingAverageStrategy) updateMovingAverages() {
	if len(mas.prices) < mas.longPeriod {
		return // Not enough data yet
	}

	// Calculate short MA
	shortSum := 0.0
	for i := len(mas.prices) - mas.shortPeriod; i < len(mas.prices); i++ {
		shortSum += mas.prices[i]
	}
	mas.shortMA = shortSum / float64(mas.shortPeriod)

	// Calculate long MA
	longSum := 0.0
	for i := len(mas.prices) - mas.longPeriod; i < len(mas.prices); i++ {
		longSum += mas.prices[i]
	}
	mas.longMA = longSum / float64(mas.longPeriod)
}

// generateSignals generates trading signals based on MA crossover
func (mas *MovingAverageStrategy) generateSignals(tick algo.PriceTick) []algo.Signal {
	signals := make([]algo.Signal, 0)

	// Need both MAs calculated
	if mas.shortMA == 0 || mas.longMA == 0 {
		return signals
	}

	// Calculate previous MAs to detect actual crossovers
	var prevShortMA, prevLongMA float64
	var signal algo.SignalType
	confidence := 0.5 // Base confidence

	if len(mas.prices) > mas.longPeriod {
		// Calculate previous short MA
		shortSum := 0.0
		for i := len(mas.prices) - mas.shortPeriod - 1; i < len(mas.prices) - 1; i++ {
			shortSum += mas.prices[i]
		}
		prevShortMA = shortSum / float64(mas.shortPeriod)

		// Calculate previous long MA
		longSum := 0.0
		for i := len(mas.prices) - mas.longPeriod - 1; i < len(mas.prices) - 1; i++ {
			longSum += mas.prices[i]
		}
		prevLongMA = longSum / float64(mas.longPeriod)
	}

	// Detect actual crossovers
	if prevShortMA != 0 && prevLongMA != 0 {
		// Bullish crossover: short MA crossed above long MA
		if prevShortMA <= prevLongMA && mas.shortMA > mas.longMA {
			signal = algo.SignalBuy
			spread := (mas.shortMA - mas.longMA) / mas.longMA
			confidence = 0.5 + (spread * 10)
			if confidence > 1.0 {
				confidence = 1.0
			}
		} else if prevShortMA >= prevLongMA && mas.shortMA < mas.longMA {
			// Bearish crossover: short MA crossed below long MA
			signal = algo.SignalSell
			spread := (mas.longMA - mas.shortMA) / mas.longMA
			confidence = 0.5 + (spread * 10)
			if confidence > 1.0 {
				confidence = 1.0
			}
		} else {
			signal = algo.SignalHold
		}
	} else {
		signal = algo.SignalHold
	}

	// Only generate signal if it's not hold and enough time has passed
	if signal != algo.SignalHold {
		// Add cooldown period to prevent excessive trading
		if time.Since(mas.lastSignalTime) >= time.Minute*5 {
			// Calculate position size based on confidence and risk parameters
			quantity := mas.calculatePositionSize(tick.Price, confidence)

			if quantity > 0 {
				signals = append(signals, algo.Signal{
					Type:       signal,
					Symbol:     mas.symbol,
					Price:      tick.Price,
					Quantity:   quantity,
					Timestamp:  tick.Timestamp,
					Confidence: confidence,
					Metadata: map[string]interface{}{
						"short_ma":     mas.shortMA,
						"long_ma":      mas.longMA,
						"spread":       mas.shortMA - mas.longMA,
						"spread_pct":   ((mas.shortMA - mas.longMA) / mas.longMA) * 100,
						"strategy":     "moving_average",
					},
				})

				mas.lastSignal = signal
				mas.lastSignalTime = tick.Timestamp
			}
		}
	}

	return signals
}

// calculatePositionSize calculates the position size based on confidence and risk
func (mas *MovingAverageStrategy) calculatePositionSize(price, confidence float64) float64 {
	// Base position size from risk limits
	maxPositionValue := mas.config.RiskLimits.MaxPositionSize

	// Scale by confidence (higher confidence = larger position)
	scaledPositionValue := maxPositionValue * confidence

	// Convert to quantity
	quantity := scaledPositionValue / price

	// Apply minimum position size filter (reduced for ultra conservative mode)
	minPositionValue := 5.0 // Minimum $5 position (was $10)
	if quantity*price < minPositionValue {
		return 0
	}

	return quantity
}

// GetState returns the current strategy state
func (mas *MovingAverageStrategy) GetState() algo.StrategyState {
	state := algo.StrategyState{
		Name:           mas.name,
		Symbol:         mas.symbol,
		IsRunning:      true,
		Position:       mas.position,
		LastUpdateTime: time.Now(),
		Metadata: map[string]interface{}{
			"short_ma":      mas.shortMA,
			"long_ma":       mas.longMA,
			"short_period":  mas.shortPeriod,
			"long_period":   mas.longPeriod,
			"prices_count":  len(mas.prices),
			"last_signal":   string(mas.lastSignal),
		},
	}

	return state
}

// Stop stops the strategy
func (mas *MovingAverageStrategy) Stop() error {
	// Clean up resources if needed
	mas.prices = nil
	mas.shortBuffer = nil
	mas.longBuffer = nil
	return nil
}

// Reset resets the strategy state
func (mas *MovingAverageStrategy) Reset() error {
	mas.prices = make([]float64, 0)
	mas.shortBuffer = make([]float64, mas.shortPeriod)
	mas.longBuffer = make([]float64, mas.longPeriod)
	mas.shortMA = 0
	mas.longMA = 0
	mas.lastSignal = algo.SignalHold
	mas.lastSignalTime = time.Time{}
	return nil
}