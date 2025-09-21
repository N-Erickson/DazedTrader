package algo

import (
	"math"
	"time"
)

// PerformanceMetrics tracks trading performance metrics
type PerformanceMetrics struct {
	// Portfolio metrics
	InitialCapital    float64   `json:"initial_capital"`
	TotalPnL          float64   `json:"total_pnl"`
	DailyPnL          float64   `json:"daily_pnl"`
	UnrealizedPnL     float64   `json:"unrealized_pnl"`
	RealizedPnL       float64   `json:"realized_pnl"`

	// Trading metrics
	TotalTrades       int       `json:"total_trades"`
	WinningTrades     int       `json:"winning_trades"`
	LosingTrades      int       `json:"losing_trades"`
	WinRate           float64   `json:"win_rate"`
	AverageWin        float64   `json:"average_win"`
	AverageLoss       float64   `json:"average_loss"`
	ProfitFactor      float64   `json:"profit_factor"`

	// Risk metrics
	MaxDrawdown       float64   `json:"max_drawdown"`
	MaxDrawdownPercent float64  `json:"max_drawdown_percent"`
	SharpeRatio       float64   `json:"sharpe_ratio"`
	SortinoRatio      float64   `json:"sortino_ratio"`
	VaR95             float64   `json:"var_95"` // Value at Risk 95%

	// Time tracking
	StartTime         time.Time `json:"start_time"`
	LastUpdate        time.Time `json:"last_update"`
	DayStartTime      time.Time `json:"day_start_time"`
	DayStartPnL       float64   `json:"day_start_pnl"`

	// Historical data for calculations
	dailyReturns      []float64
	peakValue         float64
	currentDrawdown   float64
}

// NewPerformanceMetrics creates a new performance metrics tracker
func NewPerformanceMetrics() *PerformanceMetrics {
	return &PerformanceMetrics{
		InitialCapital:  10000.0, // Default starting capital
		StartTime:       time.Now(),
		DayStartTime:    time.Now().Truncate(24 * time.Hour),
		dailyReturns:    make([]float64, 0),
		peakValue:       10000.0,
	}
}

// UpdateFromTrade updates metrics based on a completed trade
func (pm *PerformanceMetrics) UpdateFromTrade(trade Trade) {
	pm.TotalTrades++
	pm.RealizedPnL += trade.PnL

	// Update win/loss statistics
	if trade.PnL > 0 {
		pm.WinningTrades++
		pm.AverageWin = ((pm.AverageWin * float64(pm.WinningTrades-1)) + trade.PnL) / float64(pm.WinningTrades)
	} else if trade.PnL < 0 {
		pm.LosingTrades++
		pm.AverageLoss = ((pm.AverageLoss * float64(pm.LosingTrades-1)) + math.Abs(trade.PnL)) / float64(pm.LosingTrades)
	}

	// Update win rate
	if pm.TotalTrades > 0 {
		pm.WinRate = float64(pm.WinningTrades) / float64(pm.TotalTrades)
	}

	// Update profit factor
	if pm.AverageLoss > 0 {
		pm.ProfitFactor = (pm.AverageWin * float64(pm.WinningTrades)) / (pm.AverageLoss * float64(pm.LosingTrades))
	}

	pm.LastUpdate = time.Now()
}

// UpdatePortfolioValue updates the current portfolio value and calculates metrics
func (pm *PerformanceMetrics) UpdatePortfolioValue(currentValue float64) {
	pm.TotalPnL = currentValue - pm.InitialCapital

	// Update drawdown
	if currentValue > pm.peakValue {
		pm.peakValue = currentValue
	}

	pm.currentDrawdown = pm.peakValue - currentValue
	if pm.currentDrawdown > pm.MaxDrawdown {
		pm.MaxDrawdown = pm.currentDrawdown
		pm.MaxDrawdownPercent = (pm.MaxDrawdown / pm.peakValue) * 100
	}

	// Calculate daily return
	if time.Since(pm.DayStartTime) >= 24*time.Hour {
		dailyReturn := pm.TotalPnL - pm.DayStartPnL
		pm.dailyReturns = append(pm.dailyReturns, dailyReturn)

		// Keep only last 252 days (trading year)
		if len(pm.dailyReturns) > 252 {
			pm.dailyReturns = pm.dailyReturns[1:]
		}

		// Reset daily tracking
		pm.DayStartTime = time.Now().Truncate(24 * time.Hour)
		pm.DayStartPnL = pm.TotalPnL
	}

	pm.DailyPnL = pm.TotalPnL - pm.DayStartPnL

	// Update risk metrics
	pm.calculateRiskMetrics()

	pm.LastUpdate = time.Now()
}

// calculateRiskMetrics calculates advanced risk metrics
func (pm *PerformanceMetrics) calculateRiskMetrics() {
	if len(pm.dailyReturns) < 30 {
		return // Need at least 30 days of data
	}

	// Calculate Sharpe Ratio
	meanReturn := pm.calculateMean(pm.dailyReturns)
	stdDev := pm.calculateStdDev(pm.dailyReturns, meanReturn)

	if stdDev > 0 {
		// Annualized Sharpe ratio (assuming risk-free rate of 0)
		pm.SharpeRatio = (meanReturn * 252) / (stdDev * math.Sqrt(252))
	}

	// Calculate Sortino Ratio (only downside deviation)
	downturns := make([]float64, 0)
	for _, ret := range pm.dailyReturns {
		if ret < 0 {
			downturns = append(downturns, ret)
		}
	}

	if len(downturns) > 0 {
		downsideStdDev := pm.calculateStdDev(downturns, 0)
		if downsideStdDev > 0 {
			pm.SortinoRatio = (meanReturn * 252) / (downsideStdDev * math.Sqrt(252))
		}
	}

	// Calculate VaR (95% confidence)
	pm.VaR95 = pm.calculateVaR(pm.dailyReturns, 0.05)
}

// calculateMean calculates the mean of a slice of float64
func (pm *PerformanceMetrics) calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// calculateStdDev calculates the standard deviation
func (pm *PerformanceMetrics) calculateStdDev(values []float64, mean float64) float64 {
	if len(values) <= 1 {
		return 0
	}

	sumSquares := 0.0
	for _, v := range values {
		diff := v - mean
		sumSquares += diff * diff
	}

	return math.Sqrt(sumSquares / float64(len(values)-1))
}

// calculateVaR calculates Value at Risk at given confidence level
func (pm *PerformanceMetrics) calculateVaR(returns []float64, alpha float64) float64 {
	if len(returns) == 0 {
		return 0
	}

	// Sort returns in ascending order
	sorted := make([]float64, len(returns))
	copy(sorted, returns)

	// Simple insertion sort for small arrays
	for i := 1; i < len(sorted); i++ {
		key := sorted[i]
		j := i - 1
		for j >= 0 && sorted[j] > key {
			sorted[j+1] = sorted[j]
			j--
		}
		sorted[j+1] = key
	}

	// Get percentile
	index := int(alpha * float64(len(sorted)))
	if index >= len(sorted) {
		index = len(sorted) - 1
	}

	return math.Abs(sorted[index])
}

// GetSummary returns a summary of key performance metrics
func (pm *PerformanceMetrics) GetSummary() PerformanceSummary {
	return PerformanceSummary{
		TotalReturn:        pm.TotalPnL,
		TotalReturnPercent: (pm.TotalPnL / pm.InitialCapital) * 100,
		DailyPnL:           pm.DailyPnL,
		TotalTrades:        pm.TotalTrades,
		WinRate:            pm.WinRate * 100,
		MaxDrawdown:        pm.MaxDrawdownPercent,
		SharpeRatio:        pm.SharpeRatio,
		ProfitFactor:       pm.ProfitFactor,
		IsProfit:           pm.TotalPnL > 0,
		DaysActive:         int(time.Since(pm.StartTime).Hours() / 24),
	}
}

// PerformanceSummary provides a simplified view of performance metrics
type PerformanceSummary struct {
	TotalReturn        float64 `json:"total_return"`
	TotalReturnPercent float64 `json:"total_return_percent"`
	DailyPnL           float64 `json:"daily_pnl"`
	TotalTrades        int     `json:"total_trades"`
	WinRate            float64 `json:"win_rate"`
	MaxDrawdown        float64 `json:"max_drawdown"`
	SharpeRatio        float64 `json:"sharpe_ratio"`
	ProfitFactor       float64 `json:"profit_factor"`
	IsProfit           bool    `json:"is_profit"`
	DaysActive         int     `json:"days_active"`
}