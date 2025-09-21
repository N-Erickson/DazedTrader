package algo

import (
	"dazedtrader/api"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// OrderManager handles order execution and tracking
type OrderManager struct {
	client        *api.CryptoClient
	config        *EngineConfig
	pendingOrders map[string]*OrderRecord
	completedOrders []OrderRecord
	mu            sync.RWMutex
}

// OrderRequest represents a trading order request
type OrderRequest struct {
	Symbol   string  `json:"symbol"`
	Side     string  `json:"side"`     // "buy" or "sell"
	Type     string  `json:"type"`     // "market" or "limit"
	Quantity float64 `json:"quantity"`
	Price    float64 `json:"price,omitempty"` // For limit orders
}

// OrderRecord tracks an order through its lifecycle
type OrderRecord struct {
	ID           string    `json:"id"`
	ClientID     string    `json:"client_id"`
	Symbol       string    `json:"symbol"`
	Side         string    `json:"side"`
	Type         string    `json:"type"`
	Quantity     float64   `json:"quantity"`
	Price        float64   `json:"price"`
	Status       string    `json:"status"`
	FilledQty    float64   `json:"filled_qty"`
	AvgPrice     float64   `json:"avg_price"`
	SubmitTime   time.Time `json:"submit_time"`
	CompleteTime time.Time `json:"complete_time"`
	StrategyName string    `json:"strategy_name"`
	RetryCount   int       `json:"retry_count"`
	ErrorMsg     string    `json:"error_msg,omitempty"`
}

// NewOrderManager creates a new order manager
func NewOrderManager(client *api.CryptoClient, config *EngineConfig) *OrderManager {
	return &OrderManager{
		client:          client,
		config:          config,
		pendingOrders:   make(map[string]*OrderRecord),
		completedOrders: make([]OrderRecord, 0),
	}
}

// SubmitOrder submits a trading order and returns the resulting trade
func (om *OrderManager) SubmitOrder(request OrderRequest) (Trade, error) {
	om.mu.Lock()
	defer om.mu.Unlock()

	// Generate client order ID
	clientID := uuid.New().String()

	// Create order record
	record := &OrderRecord{
		ClientID:   clientID,
		Symbol:     request.Symbol,
		Side:       request.Side,
		Type:       request.Type,
		Quantity:   request.Quantity,
		Price:      request.Price,
		Status:     "pending",
		SubmitTime: time.Now(),
	}

	// Submit order to Robinhood API
	var apiOrder *api.CryptoOrder
	var err error

	if request.Type == "market" {
		apiOrder, err = om.client.PlaceCryptoOrderNew(
			clientID,
			request.Side,
			"market",
			request.Symbol,
			fmt.Sprintf("%.8f", request.Quantity),
			"",
		)
	} else {
		apiOrder, err = om.client.PlaceCryptoOrderNew(
			clientID,
			request.Side,
			"limit",
			request.Symbol,
			fmt.Sprintf("%.8f", request.Quantity),
			fmt.Sprintf("%.2f", request.Price),
		)
	}

	if err != nil {
		record.Status = "failed"
		record.ErrorMsg = err.Error()
		record.CompleteTime = time.Now()
		om.completedOrders = append(om.completedOrders, *record)
		return Trade{}, fmt.Errorf("failed to submit order: %v", err)
	}

	// Update record with API response
	record.ID = apiOrder.ID
	record.Status = apiOrder.State
	record.FilledQty = apiOrder.FilledAssetQuantity
	record.AvgPrice = apiOrder.AveragePrice

	// If order is immediately filled (market orders often are)
	if apiOrder.State == "filled" {
		record.CompleteTime = time.Now()
		om.completedOrders = append(om.completedOrders, *record)

		// Create trade record
		trade := Trade{
			ID:        apiOrder.ID,
			Symbol:    request.Symbol,
			Side:      request.Side,
			Quantity:  apiOrder.FilledAssetQuantity,
			Price:     apiOrder.AveragePrice,
			Timestamp: time.Now(),
			PnL:       0, // Will be calculated by strategy runner
			Commission: 0, // Robinhood doesn't charge commission
		}

		return trade, nil
	}

	// Add to pending orders for monitoring
	om.pendingOrders[record.ID] = record

	// Start monitoring the order
	go om.monitorOrder(record.ID)

	// For limit orders, we need to wait for completion
	// For now, return immediately - in practice you might want to wait or use callbacks
	trade := Trade{
		ID:        apiOrder.ID,
		Symbol:    request.Symbol,
		Side:      request.Side,
		Quantity:  request.Quantity,
		Price:     request.Price,
		Timestamp: time.Now(),
		PnL:       0,
		Commission: 0,
	}

	return trade, nil
}

// monitorOrder monitors a pending order until completion
func (om *OrderManager) monitorOrder(orderID string) {
	ticker := time.NewTicker(time.Second * 5) // Check every 5 seconds
	defer ticker.Stop()

	timeout := time.After(om.config.OrderTimeout)

	for {
		select {
		case <-timeout:
			// Order timed out, cancel it
			om.cancelOrder(orderID, "timeout")
			return

		case <-ticker.C:
			// Check order status
			if om.checkAndUpdateOrder(orderID) {
				// Order completed
				return
			}
		}
	}
}

// checkAndUpdateOrder checks and updates the status of a pending order
func (om *OrderManager) checkAndUpdateOrder(orderID string) bool {
	om.mu.Lock()
	defer om.mu.Unlock()

	record, exists := om.pendingOrders[orderID]
	if !exists {
		return true // Order no longer pending
	}

	// Get order status from API (this would need a new API method)
	// For now, simulate completion after some time
	if time.Since(record.SubmitTime) > time.Minute {
		record.Status = "filled"
		record.FilledQty = record.Quantity
		record.AvgPrice = record.Price
		record.CompleteTime = time.Now()

		// Move to completed orders
		om.completedOrders = append(om.completedOrders, *record)
		delete(om.pendingOrders, orderID)
		return true
	}

	return false
}

// cancelOrder cancels a pending order
func (om *OrderManager) cancelOrder(orderID string, reason string) {
	om.mu.Lock()
	defer om.mu.Unlock()

	record, exists := om.pendingOrders[orderID]
	if !exists {
		return
	}

	// Cancel via API
	err := om.client.CancelCryptoOrder(orderID)
	if err != nil {
		record.ErrorMsg = fmt.Sprintf("cancel failed: %v", err)
	}

	record.Status = "cancelled"
	record.CompleteTime = time.Now()
	record.ErrorMsg = reason

	// Move to completed orders
	om.completedOrders = append(om.completedOrders, *record)
	delete(om.pendingOrders, orderID)
}

// CancelAllOrders cancels all pending orders
func (om *OrderManager) CancelAllOrders() {
	om.mu.Lock()
	defer om.mu.Unlock()

	for orderID := range om.pendingOrders {
		go om.cancelOrder(orderID, "emergency_stop")
	}
}

// GetPendingOrders returns all pending orders
func (om *OrderManager) GetPendingOrders() []OrderRecord {
	om.mu.RLock()
	defer om.mu.RUnlock()

	orders := make([]OrderRecord, 0, len(om.pendingOrders))
	for _, record := range om.pendingOrders {
		orders = append(orders, *record)
	}

	return orders
}

// GetCompletedOrders returns all completed orders
func (om *OrderManager) GetCompletedOrders() []OrderRecord {
	om.mu.RLock()
	defer om.mu.RUnlock()

	// Return a copy to avoid race conditions
	orders := make([]OrderRecord, len(om.completedOrders))
	copy(orders, om.completedOrders)
	return orders
}

// GetOrderStats returns order execution statistics
func (om *OrderManager) GetOrderStats() OrderStats {
	om.mu.RLock()
	defer om.mu.RUnlock()

	stats := OrderStats{
		TotalOrders:    len(om.completedOrders),
		PendingOrders:  len(om.pendingOrders),
		FilledOrders:   0,
		CancelledOrders: 0,
		FailedOrders:   0,
	}

	for _, order := range om.completedOrders {
		switch order.Status {
		case "filled":
			stats.FilledOrders++
		case "cancelled":
			stats.CancelledOrders++
		case "failed":
			stats.FailedOrders++
		}
	}

	if stats.TotalOrders > 0 {
		stats.FillRate = float64(stats.FilledOrders) / float64(stats.TotalOrders)
	}

	return stats
}

// OrderStats contains order execution statistics
type OrderStats struct {
	TotalOrders     int     `json:"total_orders"`
	PendingOrders   int     `json:"pending_orders"`
	FilledOrders    int     `json:"filled_orders"`
	CancelledOrders int     `json:"cancelled_orders"`
	FailedOrders    int     `json:"failed_orders"`
	FillRate        float64 `json:"fill_rate"`
}