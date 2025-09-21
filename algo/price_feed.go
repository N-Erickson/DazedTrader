package algo

import (
	"context"
	"dazedtrader/api"
	"sync"
	"time"
)

// PriceFeed provides real-time price updates for a symbol
type PriceFeed struct {
	symbol        string
	client        *api.CryptoClient
	updateInterval time.Duration
	subscribers   []chan<- PriceTick
	lastTick      PriceTick
	isRunning     bool
	stopChan      chan struct{}
	mu            sync.RWMutex
}

// NewPriceFeed creates a new price feed for a symbol
func NewPriceFeed(symbol string, client *api.CryptoClient, updateInterval time.Duration) *PriceFeed {
	return &PriceFeed{
		symbol:         symbol,
		client:         client,
		updateInterval: updateInterval,
		subscribers:    make([]chan<- PriceTick, 0),
		stopChan:       make(chan struct{}),
	}
}

// Start starts the price feed
func (pf *PriceFeed) Start(ctx context.Context) {
	pf.mu.Lock()
	if pf.isRunning {
		pf.mu.Unlock()
		return
	}
	pf.isRunning = true
	pf.mu.Unlock()

	ticker := time.NewTicker(pf.updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-pf.stopChan:
			return
		case <-ticker.C:
			pf.fetchAndBroadcastPrice()
		}
	}
}

// Stop stops the price feed
func (pf *PriceFeed) Stop() {
	pf.mu.Lock()
	defer pf.mu.Unlock()

	if pf.isRunning {
		close(pf.stopChan)
		pf.isRunning = false
	}
}

// Subscribe adds a subscriber to receive price updates
func (pf *PriceFeed) Subscribe(ch chan<- PriceTick) {
	pf.mu.Lock()
	defer pf.mu.Unlock()

	pf.subscribers = append(pf.subscribers, ch)
}

// Unsubscribe removes a subscriber
func (pf *PriceFeed) Unsubscribe(ch chan<- PriceTick) {
	pf.mu.Lock()
	defer pf.mu.Unlock()

	for i, sub := range pf.subscribers {
		if sub == ch {
			// Remove subscriber
			pf.subscribers = append(pf.subscribers[:i], pf.subscribers[i+1:]...)
			break
		}
	}
}

// GetLastTick returns the last price tick
func (pf *PriceFeed) GetLastTick() PriceTick {
	pf.mu.RLock()
	defer pf.mu.RUnlock()

	return pf.lastTick
}

// fetchAndBroadcastPrice fetches the latest price and broadcasts to subscribers
func (pf *PriceFeed) fetchAndBroadcastPrice() {
	// Fetch current price from Robinhood API
	quotes, err := pf.client.GetBestBidAsk([]string{pf.symbol})
	if err != nil || len(quotes) == 0 {
		return
	}

	quote := quotes[0]

	// Create price tick
	tick := PriceTick{
		Symbol:    pf.symbol,
		Price:     quote.Price,
		BidPrice:  quote.BidPrice,
		AskPrice:  quote.AskPrice,
		Volume:    0, // Not available in current API
		Timestamp: time.Now(),
	}

	pf.mu.Lock()
	pf.lastTick = tick
	subscribers := make([]chan<- PriceTick, len(pf.subscribers))
	copy(subscribers, pf.subscribers)
	pf.mu.Unlock()

	// Broadcast to all subscribers
	for _, sub := range subscribers {
		select {
		case sub <- tick:
		default:
			// Subscriber channel is full, skip
		}
	}
}

// IsRunning returns whether the price feed is running
func (pf *PriceFeed) IsRunning() bool {
	pf.mu.RLock()
	defer pf.mu.RUnlock()

	return pf.isRunning
}