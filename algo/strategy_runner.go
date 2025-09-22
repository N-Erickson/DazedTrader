package algo

import (
	"context"
	"sync"
	"time"
)

// StrategyRunner manages the execution of a single strategy
type StrategyRunner struct {
	strategy     Strategy
	orderManager *OrderManager
	riskManager  *RiskManager
	config       StrategyConfig

	// State
	state        StrategyState
	position     Position
	trades       []Trade
	pnl          float64

	// Control
	stopChan     chan struct{}
	priceChan    chan PriceTick
	signalChan   chan Signal

	// Synchronization
	mu           sync.RWMutex
	isRunning    bool
}

// Trade represents a completed trade
type Trade struct {
	ID           string    `json:"id"`
	Symbol       string    `json:"symbol"`
	Side         string    `json:"side"`  // "buy" or "sell"
	Quantity     float64   `json:"quantity"`
	Price        float64   `json:"price"`
	Timestamp    time.Time `json:"timestamp"`
	StrategyName string    `json:"strategy_name"`
	PnL          float64   `json:"pnl"`
	Commission   float64   `json:"commission"`
}

// NewStrategyRunner creates a new strategy runner
func NewStrategyRunner(strategy Strategy, orderManager *OrderManager, riskManager *RiskManager, config StrategyConfig) *StrategyRunner {
	return &StrategyRunner{
		strategy:     strategy,
		orderManager: orderManager,
		riskManager:  riskManager,
		config:       config,
		stopChan:     make(chan struct{}),
		priceChan:    make(chan PriceTick, 100),
		signalChan:   make(chan Signal, 50),
		state: StrategyState{
			Name:           config.Name,
			Symbol:         config.Symbol,
			IsRunning:      false,
			Position:       Position{Symbol: config.Symbol},
			StartTime:      time.Now(),
			LastUpdateTime: time.Now(),
			Metadata:       make(map[string]interface{}),
		},
		trades: make([]Trade, 0),
	}
}

// Run starts the strategy runner main loop
func (sr *StrategyRunner) Run(ctx context.Context) {
	sr.mu.Lock()
	sr.isRunning = true
	sr.state.IsRunning = true
	sr.state.StartTime = time.Now()
	sr.mu.Unlock()

	defer func() {
		sr.mu.Lock()
		sr.isRunning = false
		sr.state.IsRunning = false
		sr.mu.Unlock()
	}()

	// Main strategy loop
	for {
		select {
		case <-ctx.Done():
			return
		case <-sr.stopChan:
			return
		case tick := <-sr.priceChan:
			sr.processPriceTick(tick)
		case signal := <-sr.signalChan:
			sr.processSignal(signal)
		}
	}
}

// Stop stops the strategy runner
func (sr *StrategyRunner) Stop() {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	if sr.isRunning {
		close(sr.stopChan)
		sr.strategy.Stop()
	}
}

// ProcessPriceTick feeds a price tick to the strategy
func (sr *StrategyRunner) ProcessPriceTick(tick PriceTick) {
	select {
	case sr.priceChan <- tick:
	default:
		// Channel full, skip this tick
	}
}

// processPriceTick handles a price tick
func (sr *StrategyRunner) processPriceTick(tick PriceTick) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	// Update position market value
	if sr.position.Quantity != 0 {
		sr.position.MarketValue = sr.position.Quantity * tick.Price
		sr.position.UnrealizedPnL = sr.position.MarketValue - (sr.position.Quantity * sr.position.AveragePrice)
	}

	// Let strategy process the tick
	signals, err := sr.strategy.ProcessTick(tick)
	if err != nil {
		// Log error
		return
	}

	// Process any generated signals
	for _, signal := range signals {
		sr.processSignal(signal)
	}

	sr.state.LastUpdateTime = time.Now()
}

// processSignal handles a trading signal
func (sr *StrategyRunner) processSignal(signal Signal) {
	// Apply risk management checks
	if !sr.riskManager.ValidateSignal(signal, sr.config.RiskLimits, sr.state) {
		return
	}

	// Check daily trade limits
	todayTrades := sr.getTodayTradeCount()
	if todayTrades >= sr.config.RiskLimits.MaxDailyTrades {
		return
	}

	// Execute the signal
	sr.executeSignal(signal)
	sr.state.LastSignal = &signal
}

// executeSignal executes a trading signal
func (sr *StrategyRunner) executeSignal(signal Signal) {
	switch signal.Type {
	case SignalBuy:
		sr.executeBuySignal(signal)
	case SignalSell:
		sr.executeSellSignal(signal)
	case SignalHold:
		// Do nothing for hold signals
	}
}

// executeBuySignal executes a buy signal
func (sr *StrategyRunner) executeBuySignal(signal Signal) {
	// Calculate position size based on risk limits
	maxPositionSize := sr.config.RiskLimits.MaxPositionSize
	quantity := signal.Quantity

	// Adjust quantity if it would exceed max position size
	if quantity*signal.Price > maxPositionSize {
		quantity = maxPositionSize / signal.Price
	}

	// Submit buy order
	order := OrderRequest{
		Symbol:   signal.Symbol,
		Side:     "buy",
		Type:     "market", // Using market orders for simplicity
		Quantity: quantity,
	}

	trade, err := sr.orderManager.SubmitOrder(order)
	if err != nil {
		// Log error
		return
	}

	// Update position
	sr.updatePositionFromTrade(trade)
	sr.trades = append(sr.trades, trade)
	sr.state.TradesCount++
}

// executeSellSignal executes a sell signal
func (sr *StrategyRunner) executeSellSignal(signal Signal) {
	// Only sell if we have a position
	if sr.position.Quantity <= 0 {
		return
	}

	// Determine quantity to sell
	quantity := signal.Quantity
	if quantity > sr.position.Quantity {
		quantity = sr.position.Quantity // Sell entire position
	}

	// Submit sell order
	order := OrderRequest{
		Symbol:   signal.Symbol,
		Side:     "sell",
		Type:     "market",
		Quantity: quantity,
	}

	trade, err := sr.orderManager.SubmitOrder(order)
	if err != nil {
		// Log error
		return
	}

	// Update position
	sr.updatePositionFromTrade(trade)
	sr.trades = append(sr.trades, trade)
	sr.state.TradesCount++
}

// updatePositionFromTrade updates the position based on a completed trade
func (sr *StrategyRunner) updatePositionFromTrade(trade Trade) {
	if trade.Side == "buy" {
		// Add to position
		totalCost := sr.position.Quantity*sr.position.AveragePrice + trade.Quantity*trade.Price
		sr.position.Quantity += trade.Quantity
		sr.position.AveragePrice = totalCost / sr.position.Quantity

		if sr.position.OpenTime.IsZero() {
			sr.position.OpenTime = trade.Timestamp
		}
	} else if trade.Side == "sell" {
		// Remove from position
		soldValue := trade.Quantity * trade.Price
		costBasis := trade.Quantity * sr.position.AveragePrice
		realizedPnL := soldValue - costBasis - trade.Commission

		sr.pnl += realizedPnL
		sr.state.PnL = sr.pnl

		sr.position.Quantity -= trade.Quantity

		// If position is closed
		if sr.position.Quantity <= 0 {
			sr.position = Position{Symbol: sr.config.Symbol}
		}
	}

	sr.position.MarketValue = sr.position.Quantity * trade.Price
	if sr.position.Quantity > 0 {
		sr.position.UnrealizedPnL = sr.position.MarketValue - (sr.position.Quantity * sr.position.AveragePrice)
	}

	sr.state.Position = sr.position
}

// GetState returns the current state of the strategy
func (sr *StrategyRunner) GetState() StrategyState {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	// Get current strategy state (includes metadata)
	strategyState := sr.strategy.GetState()

	// Update our state with strategy metadata
	state := sr.state
	state.Position = sr.position
	state.Metadata = strategyState.Metadata
	state.LastUpdateTime = time.Now()

	return state
}

// getTodayTradeCount returns the number of trades executed today
func (sr *StrategyRunner) getTodayTradeCount() int {
	today := time.Now().Truncate(24 * time.Hour)
	count := 0

	for _, trade := range sr.trades {
		if trade.Timestamp.After(today) {
			count++
		}
	}

	return count
}

// GetTrades returns all trades executed by this strategy
func (sr *StrategyRunner) GetTrades() []Trade {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	// Return a copy to avoid race conditions
	trades := make([]Trade, len(sr.trades))
	copy(trades, sr.trades)
	return trades
}

// GetDailyPnL returns the PnL for today
func (sr *StrategyRunner) GetDailyPnL() float64 {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	today := time.Now().Truncate(24 * time.Hour)
	dailyPnL := 0.0

	for _, trade := range sr.trades {
		if trade.Timestamp.After(today) {
			dailyPnL += trade.PnL
		}
	}

	return dailyPnL
}