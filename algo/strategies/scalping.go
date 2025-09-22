package strategies

import (
	"dazedtrader/algo"
	"fmt"
	"log"
	"os"
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

	// Calculate price volatility - more lenient for first position
	if len(ss.prices) >= 5 {
		recentPrices := ss.prices[len(ss.prices)-5:]
		volatility := ss.calculateVolatility(recentPrices)

		// Use lower threshold if we don't have a position (need to enter market)
		hasPosition := false // TODO: get actual position from runner
		threshold := ss.volatilityThreshold
		if !hasPosition {
			threshold = ss.volatilityThreshold * 0.1 // 10x more permissive for first entry
		}

		if volatility < threshold {
			msg := fmt.Sprintf("[%s] Volatility too low: %.8f < %.8f (entry threshold)", ss.name, volatility, threshold)
			fmt.Println(msg)
			ss.logToFile(msg)
			return signals
		}
		msg := fmt.Sprintf("[%s] Volatility check passed: %.8f >= %.8f", ss.name, volatility, threshold)
		fmt.Println(msg)
		ss.logToFile(msg)
	}

	// Calculate previous MAs to detect actual crossovers
	var prevShortMA, prevLongMA float64
	if len(ss.prices) > ss.longPeriod {
		// Calculate previous short MA
		shortSum := 0.0
		for i := len(ss.prices) - ss.shortPeriod - 1; i < len(ss.prices) - 1; i++ {
			shortSum += ss.prices[i]
		}
		prevShortMA = shortSum / float64(ss.shortPeriod)

		// Calculate previous long MA
		longSum := 0.0
		for i := len(ss.prices) - ss.longPeriod - 1; i < len(ss.prices) - 1; i++ {
			longSum += ss.prices[i]
		}
		prevLongMA = longSum / float64(ss.longPeriod)
	}

	// Detect actual crossovers
	if prevShortMA != 0 && prevLongMA != 0 {
		msg := fmt.Sprintf("[%s] CROSSOVER CHECK: Prev(Short=%.6f, Long=%.6f) Current(Short=%.6f, Long=%.6f) Price=%.6f",
			ss.name, prevShortMA, prevLongMA, ss.shortMA, ss.longMA, tick.Price)
		fmt.Println(msg)
		ss.logToFile(msg)

		// TREND FOLLOWING with PROFIT VALIDATION
		if prevShortMA <= prevLongMA && ss.shortMA > ss.longMA {
			// Bullish crossover: Short MA crossed above long MA = uptrend starting
			signal = algo.SignalBuy
			spread := (ss.shortMA - ss.longMA) / ss.longMA
			confidence = 0.6 + (spread * 15)
			msg = fmt.Sprintf("[%s] 📈 BULLISH CROSSOVER -> BUY (riding uptrend)", ss.name)
			fmt.Println(msg)
			ss.logToFile(msg)
		} else if prevShortMA >= prevLongMA && ss.shortMA < ss.longMA {
			// Bearish crossover: Short MA crossed below long MA = downtrend starting
			signal = algo.SignalSell
			spread := (ss.longMA - ss.shortMA) / ss.longMA
			confidence = 0.6 + (spread * 15)
			msg = fmt.Sprintf("[%s] 📉 BEARISH CROSSOVER -> SELL (riding downtrend)", ss.name)
			fmt.Println(msg)
			ss.logToFile(msg)
		} else {
			signal = algo.SignalHold
		}
	} else {
		// If no previous data for crossover detection, use trend strength
		// INVERTED LOGIC: When trending up, prepare to sell. When trending down, prepare to buy.
		if ss.shortMA > ss.longMA {
			spread := (ss.shortMA - ss.longMA) / ss.longMA
			//REMOVED: fmt.Printf("[%s] Bullish trend: Short=%.6f > Long=%.6f, Spread=%.4f%% -> SELL SIGNAL\n", ss.name, ss.shortMA, ss.longMA, spread*100)
			// Only trade if spread indicates profitable opportunity
			minProfitableSpread := 0.008 // 0.8% minimum for profitable trade
			if spread >= minProfitableSpread {
				signal = algo.SignalSell  // CHANGED: was SignalBuy - sell at peak
				confidence = 0.5 + (spread * 10)
				//REMOVED: fmt.Printf("[%s] PROFITABLE SELL: Spread %.4f%% >= %.1f%% (selling at trend peak)\n", ss.name, spread*100, minProfitableSpread*100)
			} else {
				signal = algo.SignalHold
				//REMOVED: fmt.Printf("[%s] Not profitable: Spread %.4f%% < %.1f%% profit threshold\n", ss.name, spread*100, minProfitableSpread*100)
			}
		} else if ss.shortMA < ss.longMA {
			spread := (ss.longMA - ss.shortMA) / ss.longMA
			//REMOVED: fmt.Printf("[%s] Bearish trend: Short=%.6f < Long=%.6f, Spread=%.4f%% -> BUY SIGNAL\n", ss.name, ss.shortMA, ss.longMA, spread*100)
			minProfitableSpread := 0.008 // 0.8% minimum for profitable trade
			if spread >= minProfitableSpread {
				signal = algo.SignalBuy   // CHANGED: was SignalSell - buy at bottom
				confidence = 0.5 + (spread * 10)
				//REMOVED: fmt.Printf("[%s] PROFITABLE BUY: Spread %.4f%% >= %.1f%% (buying at trend bottom)\n", ss.name, spread*100, minProfitableSpread*100)
			} else {
				signal = algo.SignalHold
				//REMOVED: fmt.Printf("[%s] Not profitable: Spread %.4f%% < %.1f%% profit threshold\n", ss.name, spread*100, minProfitableSpread*100)
			}
		} else {
			signal = algo.SignalHold
			//REMOVED: fmt.Printf("[%s] MAs equal: Short=%.6f = Long=%.6f\n", ss.name, ss.shortMA, ss.longMA)
		}
	}

	if confidence > 1.0 {
		confidence = 1.0
	}

	// Very short cooldown for scalping (15 seconds, not 30)
	if signal != algo.SignalHold {
		if time.Since(ss.lastSignalTime) >= time.Second*30 {
			quantity := ss.calculatePositionSize(tick.Price, confidence)

			if quantity > 0 {
				//REMOVED: fmt.Printf("[%s] ✅ SIGNAL CREATED: %s %.8f at $%.2f (confidence: %.2f)\n", ss.name, signal, quantity, tick.Price, confidence)
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
			} else {
				//REMOVED: fmt.Printf("[%s] ❌ SIGNAL BLOCKED: Position size too small (%.8f)\n", ss.name, quantity)
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

// logToFile logs debug messages to a file
func (ss *ScalpingStrategy) logToFile(message string) {
	logFile := fmt.Sprintf("/tmp/dazedtrader_%s.log", ss.name)
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	logger := log.New(f, "", log.LstdFlags)
	logger.Println(message)
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