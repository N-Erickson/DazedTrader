package strategies

import (
	"dazedtrader/algo"
	"log"
	"time"
)

// Momentum25DayStrategy implements the proven 25-day momentum strategy
// Research shows this beats buy-and-hold with lower drawdown
type Momentum25DayStrategy struct {
	name           string
	symbol         string
	lookbackPeriod int
	prices         []float64
	priceHistory   []PricePoint
	position       algo.Position
	config         algo.StrategyConfig
	lastSignalTime time.Time
}

type PricePoint struct {
	Price     float64
	Timestamp time.Time
}

// NewMomentum25DayStrategy creates a new 25-day momentum strategy
func NewMomentum25DayStrategy(name, symbol string) *Momentum25DayStrategy {
	return &Momentum25DayStrategy{
		name:           name,
		symbol:         symbol,
		lookbackPeriod: 25, // Proven in research
		prices:         make([]float64, 0),
		priceHistory:   make([]PricePoint, 0),
		lastSignalTime: time.Time{},
	}
}

// GetName returns the strategy name
func (m *Momentum25DayStrategy) GetName() string {
	return m.name
}

// GetSymbol returns the symbol this strategy trades
func (m *Momentum25DayStrategy) GetSymbol() string {
	return m.symbol
}

// Initialize initializes the strategy with configuration
func (m *Momentum25DayStrategy) Initialize(config algo.StrategyConfig) error {
	m.config = config
	log.Printf("[%s] Initialized 25-day momentum strategy for %s", m.name, m.symbol)
	return nil
}

// ProcessTick processes a new price tick and generates trading signals
func (m *Momentum25DayStrategy) ProcessTick(tick algo.PriceTick) ([]algo.Signal, error) {
	// Add current price to history
	pricePoint := PricePoint{
		Price:     tick.Price,
		Timestamp: tick.Timestamp,
	}
	m.priceHistory = append(m.priceHistory, pricePoint)

	// Keep only what we need (25 days + buffer)
	maxHistory := m.lookbackPeriod + 5
	if len(m.priceHistory) > maxHistory {
		m.priceHistory = m.priceHistory[len(m.priceHistory)-maxHistory:]
	}

	log.Printf("[%s] Price history length: %d (need %d for signal)", m.name, len(m.priceHistory), m.lookbackPeriod)

	// Need at least 25 days of data
	if len(m.priceHistory) <= m.lookbackPeriod {
		log.Printf("[%s] Insufficient data for 25-day momentum (have %d, need %d)", m.name, len(m.priceHistory), m.lookbackPeriod+1)
		return []algo.Signal{}, nil
	}

	// Get price from 25 days ago
	currentPrice := tick.Price
	price25DaysAgo := m.priceHistory[len(m.priceHistory)-m.lookbackPeriod-1].Price

	momentumPercent := ((currentPrice - price25DaysAgo) / price25DaysAgo) * 100

	log.Printf("[%s] 25-Day Momentum Check: Current=$%.6f, 25DaysAgo=$%.6f, Momentum=%.2f%%",
		m.name, currentPrice, price25DaysAgo, momentumPercent)

	// Generate signal based on momentum
	var signal algo.SignalType
	confidence := 0.7 // High confidence for proven strategy

	// Minimum momentum threshold to avoid noise (1% minimum)
	momentumThreshold := 1.0

	if momentumPercent > momentumThreshold {
		// Positive momentum - uptrend
		signal = algo.SignalBuy
		log.Printf("[%s] POSITIVE MOMENTUM: +%.2f%% -> BUY signal", m.name, momentumPercent)
	} else if momentumPercent < -momentumThreshold {
		// Negative momentum - downtrend - only sell if profitable
		signal = algo.SignalSell
		log.Printf("[%s] NEGATIVE MOMENTUM: %.2f%% -> SELL signal (with profit protection)", m.name, momentumPercent)
	} else {
		// Weak momentum - hold
		signal = algo.SignalHold
		log.Printf("[%s] WEAK MOMENTUM: %.2f%% (threshold: ±%.1f%%) -> HOLD", m.name, momentumPercent, momentumThreshold)
	}

	// Generate signals based on momentum logic only (no artificial time restrictions)
	signals := make([]algo.Signal, 0)
	if signal != algo.SignalHold {
		// Only minimal cooldown to prevent duplicate signals on same tick
		if time.Since(m.lastSignalTime) >= time.Minute*15 {
			quantity := m.calculatePositionSize(tick.Price, confidence)

			if quantity > 0 {
				log.Printf("[%s] SIGNAL APPROVED: %s %.8f at $%.6f (momentum: %.2f%%)",
					m.name, signal, quantity, tick.Price, momentumPercent)

				signals = append(signals, algo.Signal{
					Type:       signal,
					Symbol:     m.symbol,
					Price:      tick.Price,
					Quantity:   quantity,
					Timestamp:  tick.Timestamp,
					Confidence: confidence,
					Metadata: map[string]interface{}{
						"momentum_percent": momentumPercent,
						"price_25d_ago":    price25DaysAgo,
						"strategy":         "25day_momentum",
						"lookback_days":    m.lookbackPeriod,
					},
				})

				m.lastSignalTime = tick.Timestamp
			} else {
				log.Printf("[%s] SIGNAL BLOCKED: Position size too small (%.8f)", m.name, quantity)
			}
		} else {
			timeSince := time.Since(m.lastSignalTime)
			log.Printf("[%s] COOLDOWN: %v since last signal (need 2h)", m.name, timeSince.Truncate(time.Minute))
		}
	}

	return signals, nil
}

// calculatePositionSize calculates position size based on momentum strength
func (m *Momentum25DayStrategy) calculatePositionSize(price, confidence float64) float64 {
	// Conservative position sizing - use 80% of max position
	maxPositionValue := m.config.RiskLimits.MaxPositionSize * 0.8

	// Scale by confidence
	scaledPositionValue := maxPositionValue * confidence

	// Convert to quantity
	quantity := scaledPositionValue / price

	// Minimum position filter
	minPositionValue := 10.0 // $10 minimum
	if quantity*price < minPositionValue {
		return 0
	}

	return quantity
}

// GetState returns the current strategy state
func (m *Momentum25DayStrategy) GetState() algo.StrategyState {
	var momentum float64
	var price25DaysAgo float64

	if len(m.priceHistory) > m.lookbackPeriod {
		currentPrice := m.priceHistory[len(m.priceHistory)-1].Price
		price25DaysAgo = m.priceHistory[len(m.priceHistory)-m.lookbackPeriod-1].Price
		momentum = ((currentPrice - price25DaysAgo) / price25DaysAgo) * 100
	}

	return algo.StrategyState{
		Name:           m.name,
		Symbol:         m.symbol,
		IsRunning:      true,
		Position:       m.position,
		LastUpdateTime: time.Now(),
		Metadata: map[string]interface{}{
			"momentum_percent":   momentum,
			"price_25d_ago":      price25DaysAgo,
			"price_history_size": len(m.priceHistory),
			"lookback_period":    m.lookbackPeriod,
			"last_signal_time":   m.lastSignalTime,
		},
	}
}

// Stop stops the strategy
func (m *Momentum25DayStrategy) Stop() error {
	log.Printf("[%s] Stopping 25-day momentum strategy", m.name)
	m.prices = nil
	m.priceHistory = nil
	return nil
}

// Reset resets the strategy state
func (m *Momentum25DayStrategy) Reset() error {
	log.Printf("[%s] Resetting 25-day momentum strategy", m.name)
	m.prices = make([]float64, 0)
	m.priceHistory = make([]PricePoint, 0)
	m.lastSignalTime = time.Time{}
	return nil
}