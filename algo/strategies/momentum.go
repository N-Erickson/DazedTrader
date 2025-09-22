package strategies

import (
	"dazedtrader/algo"
	"fmt"
	"math"
	"time"
)

// MomentumStrategy implements a momentum-based trading strategy
type MomentumStrategy struct {
	name           string
	symbol         string
	lookbackPeriod int
	momentumThreshold float64
	volumeMultiplier  float64

	// State
	prices         []float64
	volumes        []float64
	returns        []float64
	avgVolume      float64
	momentum       float64
	volatility     float64
	lastSignal     algo.SignalType
	lastSignalTime time.Time
	position       algo.Position
	config         algo.StrategyConfig
}

// NewMomentumStrategy creates a new momentum strategy
func NewMomentumStrategy(name, symbol string) *MomentumStrategy {
	return &MomentumStrategy{
		name:              name,
		symbol:            symbol,
		lookbackPeriod:    10,   // Default 10-period lookback
		momentumThreshold: 0.02, // 2% momentum threshold
		volumeMultiplier:  1.5,  // Volume must be 1.5x average
		prices:            make([]float64, 0),
		volumes:           make([]float64, 0),
		returns:           make([]float64, 0),
		lastSignal:        algo.SignalHold,
	}
}

// GetName returns the strategy name
func (ms *MomentumStrategy) GetName() string {
	return ms.name
}

// GetSymbol returns the symbol this strategy trades
func (ms *MomentumStrategy) GetSymbol() string {
	return ms.symbol
}

// Initialize initializes the strategy with configuration
func (ms *MomentumStrategy) Initialize(config algo.StrategyConfig) error {
	ms.config = config

	// Parse strategy-specific parameters
	if params, ok := config.Parameters["lookback_period"].(float64); ok {
		ms.lookbackPeriod = int(params)
	}
	if params, ok := config.Parameters["momentum_threshold"].(float64); ok {
		ms.momentumThreshold = params
	}
	if params, ok := config.Parameters["volume_multiplier"].(float64); ok {
		ms.volumeMultiplier = params
	}

	return nil
}

// ProcessTick processes a new price tick and generates trading signals
func (ms *MomentumStrategy) ProcessTick(tick algo.PriceTick) ([]algo.Signal, error) {
	// Add price and volume to history
	ms.prices = append(ms.prices, tick.Price)
	ms.volumes = append(ms.volumes, tick.Volume)

	// Keep only lookback period * 2 for calculations
	maxHistory := ms.lookbackPeriod * 2
	if len(ms.prices) > maxHistory {
		ms.prices = ms.prices[len(ms.prices)-maxHistory:]
		ms.volumes = ms.volumes[len(ms.volumes)-maxHistory:]
	}

	// Calculate returns
	if len(ms.prices) >= 2 {
		lastReturn := (ms.prices[len(ms.prices)-1] - ms.prices[len(ms.prices)-2]) / ms.prices[len(ms.prices)-2]
		ms.returns = append(ms.returns, lastReturn)

		if len(ms.returns) > maxHistory {
			ms.returns = ms.returns[len(ms.returns)-maxHistory:]
		}
	}

	// Update momentum indicators
	ms.updateMomentumIndicators()

	// Generate signals
	signals := ms.generateSignals(tick)

	return signals, nil
}

// updateMomentumIndicators calculates momentum and volatility indicators
func (ms *MomentumStrategy) updateMomentumIndicators() {
	if len(ms.prices) < ms.lookbackPeriod {
		return
	}

	// Calculate momentum (rate of change over lookback period)
	currentPrice := ms.prices[len(ms.prices)-1]
	pastPrice := ms.prices[len(ms.prices)-ms.lookbackPeriod]
	ms.momentum = (currentPrice - pastPrice) / pastPrice

	// Calculate average volume
	if len(ms.volumes) >= ms.lookbackPeriod {
		volumeSum := 0.0
		for i := len(ms.volumes) - ms.lookbackPeriod; i < len(ms.volumes); i++ {
			volumeSum += ms.volumes[i]
		}
		ms.avgVolume = volumeSum / float64(ms.lookbackPeriod)
	}

	// Calculate volatility (standard deviation of returns)
	if len(ms.returns) >= ms.lookbackPeriod {
		ms.volatility = ms.calculateVolatility()
	}
}

// calculateVolatility calculates the volatility of returns
func (ms *MomentumStrategy) calculateVolatility() float64 {
	if len(ms.returns) < ms.lookbackPeriod {
		return 0
	}

	// Calculate mean return
	meanReturn := 0.0
	recentReturns := ms.returns[len(ms.returns)-ms.lookbackPeriod:]
	for _, ret := range recentReturns {
		meanReturn += ret
	}
	meanReturn /= float64(len(recentReturns))

	// Calculate variance
	variance := 0.0
	for _, ret := range recentReturns {
		diff := ret - meanReturn
		variance += diff * diff
	}
	variance /= float64(len(recentReturns) - 1)

	return math.Sqrt(variance)
}

// generateSignals generates trading signals based on momentum
func (ms *MomentumStrategy) generateSignals(tick algo.PriceTick) []algo.Signal {
	signals := make([]algo.Signal, 0)

	// Need enough data
	if len(ms.prices) < ms.lookbackPeriod {
		return signals
	}

	var signal algo.SignalType
	confidence := 0.5

	// MOMENTUM TREND FOLLOWING with profit protection
	if ms.momentum > ms.momentumThreshold {
		// Strong positive momentum = uptrend = BUY
		if tick.Volume > ms.avgVolume*ms.volumeMultiplier {
			signal = algo.SignalBuy
			confidence = 0.5 + (ms.momentum * 5)
			if confidence > 1.0 {
				confidence = 1.0
			}
			fmt.Printf("[%s] 📈 STRONG POSITIVE MOMENTUM -> BUY (riding momentum up)\n", ms.name)
		}
	} else if ms.momentum < -ms.momentumThreshold {
		// Strong negative momentum = downtrend = SELL (with profit protection)
		if tick.Volume > ms.avgVolume*ms.volumeMultiplier {
			signal = algo.SignalSell
			confidence = 0.5 + (math.Abs(ms.momentum) * 5)
			if confidence > 1.0 {
				confidence = 1.0
			}
			fmt.Printf("[%s] 📉 STRONG NEGATIVE MOMENTUM -> SELL (with profit protection)\n", ms.name)
		}
	} else {
		signal = algo.SignalHold
	}

	// Adjust confidence based on volatility (lower volatility = higher confidence)
	if ms.volatility > 0 {
		volatilityAdjustment := 1.0 / (1.0 + ms.volatility*10)
		confidence *= volatilityAdjustment
	}

	// Only generate signal if different from last and enough time passed
	if signal != ms.lastSignal && signal != algo.SignalHold {
		if time.Since(ms.lastSignalTime) >= time.Minute*2 { // Shorter cooldown for momentum
			quantity := ms.calculatePositionSize(tick.Price, confidence)

			if quantity > 0 {
				signals = append(signals, algo.Signal{
					Type:       signal,
					Symbol:     ms.symbol,
					Price:      tick.Price,
					Quantity:   quantity,
					Timestamp:  tick.Timestamp,
					Confidence: confidence,
					Metadata: map[string]interface{}{
						"momentum":       ms.momentum,
						"momentum_pct":   ms.momentum * 100,
						"volatility":     ms.volatility,
						"avg_volume":     ms.avgVolume,
						"current_volume": tick.Volume,
						"volume_ratio":   tick.Volume / ms.avgVolume,
						"strategy":       "momentum",
					},
				})

				ms.lastSignal = signal
				ms.lastSignalTime = tick.Timestamp
			}
		}
	}

	return signals
}

// calculatePositionSize calculates position size based on momentum strength and volatility
func (ms *MomentumStrategy) calculatePositionSize(price, confidence float64) float64 {
	maxPositionValue := ms.config.RiskLimits.MaxPositionSize

	// Scale by confidence and inverse volatility
	volatilityAdjustment := 1.0
	if ms.volatility > 0 {
		volatilityAdjustment = 1.0 / (1.0 + ms.volatility*5) // Reduce size for high volatility
	}

	scaledPositionValue := maxPositionValue * confidence * volatilityAdjustment

	// Scale by momentum strength
	momentumMultiplier := 1.0 + math.Abs(ms.momentum)*2
	if momentumMultiplier > 2.0 {
		momentumMultiplier = 2.0 // Cap at 2x
	}

	scaledPositionValue *= momentumMultiplier

	quantity := scaledPositionValue / price

	// Minimum position filter
	if quantity*price < 10.0 {
		return 0
	}

	return quantity
}

// GetState returns the current strategy state
func (ms *MomentumStrategy) GetState() algo.StrategyState {
	state := algo.StrategyState{
		Name:           ms.name,
		Symbol:         ms.symbol,
		IsRunning:      true,
		Position:       ms.position,
		LastUpdateTime: time.Now(),
		Metadata: map[string]interface{}{
			"momentum":          ms.momentum,
			"momentum_pct":      ms.momentum * 100,
			"volatility":        ms.volatility,
			"avg_volume":        ms.avgVolume,
			"lookback_period":   ms.lookbackPeriod,
			"momentum_threshold": ms.momentumThreshold,
			"prices_count":      len(ms.prices),
			"last_signal":       string(ms.lastSignal),
		},
	}

	return state
}

// Stop stops the strategy
func (ms *MomentumStrategy) Stop() error {
	ms.prices = nil
	ms.volumes = nil
	ms.returns = nil
	return nil
}

// Reset resets the strategy state
func (ms *MomentumStrategy) Reset() error {
	ms.prices = make([]float64, 0)
	ms.volumes = make([]float64, 0)
	ms.returns = make([]float64, 0)
	ms.momentum = 0
	ms.volatility = 0
	ms.avgVolume = 0
	ms.lastSignal = algo.SignalHold
	ms.lastSignalTime = time.Time{}
	return nil
}