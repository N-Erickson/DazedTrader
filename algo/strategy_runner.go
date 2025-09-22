package algo

import (
	"context"
	"fmt"
	"log"
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
	msg := fmt.Sprintf("[%s] Processing %s signal: %.8f %s at $%.6f", sr.config.Name, signal.Type, signal.Quantity, signal.Symbol, signal.Price)
	fmt.Println(msg)
	log.Println(msg)

	// Apply risk management checks
	if !sr.riskManager.ValidateSignal(signal, sr.config.RiskLimits, sr.state) {
		msg := fmt.Sprintf("[%s] BLOCKED: Risk management rejected signal", sr.config.Name)
		fmt.Println(msg)
		log.Println(msg)
		return
	}
	msg = fmt.Sprintf("[%s] Risk management check passed", sr.config.Name)
	fmt.Println(msg)
	log.Println(msg)

	// Check daily trade limits
	todayTrades := sr.getTodayTradeCount()
	if todayTrades >= sr.config.RiskLimits.MaxDailyTrades {
		msg := fmt.Sprintf("[%s] BLOCKED: Daily trade limit reached (%d/%d)", sr.config.Name, todayTrades, sr.config.RiskLimits.MaxDailyTrades)
		fmt.Println(msg)
		log.Println(msg)
		return
	}
	msg = fmt.Sprintf("[%s] Daily trade limit check passed (%d/%d)", sr.config.Name, todayTrades, sr.config.RiskLimits.MaxDailyTrades)
	fmt.Println(msg)
	log.Println(msg)

	// Execute the signal
	msg = fmt.Sprintf("[%s] Executing signal...", sr.config.Name)
	fmt.Println(msg)
	log.Println(msg)
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
	msg := fmt.Sprintf("[%s] Executing BUY signal", sr.config.Name)
	fmt.Println(msg)
	log.Println(msg)

	// POSITION CHECK: Only buy if we don't already have a position
	if sr.position.Quantity > 0 {
		msg := fmt.Sprintf("[%s] BLOCKED: Already have position %.8f DOGE - cannot buy more", sr.config.Name, sr.position.Quantity)
		fmt.Println(msg)
		log.Println(msg)
		return
	}

	msg = fmt.Sprintf("[%s] POSITION CHECK PASSED: No existing position, safe to buy", sr.config.Name)
	fmt.Println(msg)
	log.Println(msg)

	// Calculate position size based on risk limits
	maxPositionSize := sr.config.RiskLimits.MaxPositionSize
	quantity := signal.Quantity
	msg = fmt.Sprintf("[%s] Original quantity: %.8f, Max position: $%.2f", sr.config.Name, quantity, maxPositionSize)
	fmt.Println(msg)
	log.Println(msg)

	// Adjust quantity if it would exceed max position size
	if quantity*signal.Price > maxPositionSize {
		quantity = maxPositionSize / signal.Price
		msg = fmt.Sprintf("[%s] Adjusted quantity to: %.8f (to fit max position size)", sr.config.Name, quantity)
		fmt.Println(msg)
		log.Println(msg)
	}

	// Log adjusted quantity
	msg = fmt.Sprintf("[%s] Final quantity after adjustments: %.8f", sr.config.Name, quantity)
	fmt.Println(msg)
	log.Println(msg)

	orderValue := quantity * signal.Price
	msg = fmt.Sprintf("[%s] Order value: %.8f * $%.6f = $%.2f", sr.config.Name, quantity, signal.Price, orderValue)
	fmt.Println(msg)
	log.Println(msg)

	// Submit buy order
	order := OrderRequest{
		Symbol:   signal.Symbol,
		Side:     "buy",
		Type:     "market", // Using market orders for simplicity
		Quantity: quantity,
	}

	msg = fmt.Sprintf("[%s] Submitting order to API: %s %.8f %s", sr.config.Name, order.Side, order.Quantity, order.Symbol)
	fmt.Println(msg)
	log.Println(msg)

	trade, err := sr.orderManager.SubmitOrder(order)
	if err != nil {
		msg = fmt.Sprintf("[%s] ORDER FAILED: %v", sr.config.Name, err)
		fmt.Println(msg)
		log.Println(msg)
		return
	}
	msg = fmt.Sprintf("[%s] ORDER SUCCESS: Trade ID %s", sr.config.Name, trade.ID)
	fmt.Println(msg)
	log.Println(msg)

	// Update position
	sr.updatePositionFromTrade(trade)
	sr.trades = append(sr.trades, trade)
	sr.state.TradesCount++
}

// executeSellSignal executes a sell signal
func (sr *StrategyRunner) executeSellSignal(signal Signal) {
	msg := fmt.Sprintf("[%s] Executing SELL signal - Current position: %.8f at cost $%.6f", sr.config.Name, sr.position.Quantity, sr.position.AveragePrice)
	fmt.Println(msg)
	log.Println(msg)

	// Only sell if we have a position
	if sr.position.Quantity <= 0 {
		msg := fmt.Sprintf("[%s] BLOCKED: No position to sell (%.8f)", sr.config.Name, sr.position.Quantity)
		fmt.Println(msg)
		log.Println(msg)
		return
	}

	// PROFIT PROTECTION: Only sell if we can make a profit
	currentValue := sr.position.Quantity * signal.Price
	costBasis := sr.position.Quantity * sr.position.AveragePrice
	potentialProfit := currentValue - costBasis
	profitPercent := (potentialProfit / costBasis) * 100

	msg = fmt.Sprintf("[%s] PROFIT CHECK: Cost=$%.2f Current=$%.2f Profit=$%.2f (%.2f%%)",
		sr.config.Name, costBasis, currentValue, potentialProfit, profitPercent)
	fmt.Println(msg)
	log.Println(msg)

	// STOP LOSS: Force sell if loss exceeds -2%
	stopLossPercent := -2.0
	if profitPercent <= stopLossPercent {
		msg := fmt.Sprintf("[%s] STOP LOSS TRIGGERED: %.2f%% loss >= %.1f%% stop loss - FORCE SELLING",
			sr.config.Name, profitPercent, stopLossPercent)
		fmt.Println(msg)
		log.Println(msg)
		// Continue to execute sell - don't return
	} else {
		// Require minimum 0.5% profit to sell (unless stop loss)
		minProfitPercent := 0.5
		if profitPercent < minProfitPercent {
			msg := fmt.Sprintf("[%s] BLOCKED: Insufficient profit %.2f%% < %.1f%% minimum",
				sr.config.Name, profitPercent, minProfitPercent)
			fmt.Println(msg)
			log.Println(msg)
			return
		}
	}

	msg = fmt.Sprintf("[%s] PROFIT APPROVED: %.2f%% profit >= 0.5%% minimum",
		sr.config.Name, profitPercent)
	fmt.Println(msg)
	log.Println(msg)

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