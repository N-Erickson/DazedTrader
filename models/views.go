package models

import (
	"dazedtrader/ui"
	"fmt"
	"strings"
)

// LoginView renders the API key setup form
func (m *AppModel) LoginView() string {
	title := ui.HeaderStyle.Render("🔐 ROBINHOOD CRYPTO API SETUP")

	var content strings.Builder

	if m.Error != "" {
		content.WriteString(ui.NegativeStyle.Render("❌ " + m.Error + "\n\n"))
	}

	if m.Loading {
		content.WriteString("🔄 Verifying API key...\n\n")
	} else {
		content.WriteString(ui.PositiveStyle.Render("Enter your Robinhood API credentials:") + "\n")
		content.WriteString("Format: apikey:privatekey (privatekey in base64)\n")
		content.WriteString("(Get from: https://docs.robinhood.com/crypto/trading/)\n\n")

		apiKeyInput := m.APIKeyForm.APIKey
		if !m.ShowAPIKey && apiKeyInput != "" {
			// Show masked API key
			apiKeyInput = strings.Repeat("*", len(apiKeyInput))
		}

		if apiKeyInput == "" {
			apiKeyInput = "│"
		} else {
			apiKeyInput += "│"
		}

		content.WriteString(ui.InputStyle.Render(apiKeyInput) + "\n\n")
		content.WriteString("Press Ctrl+V to paste, Tab to toggle visibility, Enter to verify, Esc to cancel\n")
		content.WriteString("Ctrl+A to clear all text\n")
	}

	footer := ui.InfoStyle.Render("Tip: Your API key is stored securely locally and used for Robinhood Crypto API access")

	return fmt.Sprintf("%s\n%s\n%s", title, ui.MenuStyle.Render(content.String()), footer)
}

// DashboardView renders the portfolio dashboard
func (m *AppModel) DashboardView() string {
	if !m.Authenticated {
		return "Please login first!"
	}

	title := ui.HeaderStyle.Render("📊 PORTFOLIO DASHBOARD")

	var content strings.Builder

	if m.Error != "" {
		content.WriteString(ui.NegativeStyle.Render("❌ " + m.Error + "\n\n"))
	}

	// Show data source indicator
	if m.DataSource != "" {
		var sourceIcon string
		if m.DataSource == "Robinhood API" {
			sourceIcon = "🟢"
		} else {
			sourceIcon = "🟡"
		}
		content.WriteString(ui.InfoStyle.Render(sourceIcon + " Data: " + m.DataSource + "\n\n"))
	}

	if m.Loading {
		content.WriteString(ui.LoadingStyle.Render("🔄 Loading portfolio data...\n\n"))
	} else if m.Portfolio == nil {
		content.WriteString("📊 Loading your portfolio...\n")
		content.WriteString("This may take a moment on first load.\n\n")
	} else {
		// Portfolio Summary
		content.WriteString("📈 PORTFOLIO SUMMARY\n")
		content.WriteString("═════════════════════\n")
		// Calculate total value from holdings
		totalValue := 0.0
		for _, holding := range m.Portfolio.Holdings {
			totalValue += holding.MarketValue
		}
		content.WriteString(fmt.Sprintf("Total Value:     %s\n", ui.FormatMarketValue(totalValue)))
		// Calculate total day change from holdings
		totalDayChange := 0.0
		for _, holding := range m.Portfolio.Holdings {
			totalDayChange += holding.DayChange
		}
		dayChangePct := 0.0
		if totalValue > 0 {
			dayChangePct = (totalDayChange / (totalValue - totalDayChange)) * 100
		}
		content.WriteString(fmt.Sprintf("Day Change:      %s (%s)\n",
			ui.FormatCurrency(totalDayChange),
			ui.FormatPercentage(dayChangePct)))
		content.WriteString(fmt.Sprintf("Buying Power:    %s\n\n", ui.FormatValue(m.Portfolio.BuyingPower)))

		// Holdings
		if len(m.Portfolio.Holdings) > 0 {
			content.WriteString("💰 CRYPTO HOLDINGS\n")
			content.WriteString("══════════════════\n")
			content.WriteString("Asset          Quantity        Price               Market Value        Day Change\n")
			content.WriteString("─────────────────────────────────────────────────────────────────────────────────────\n")

			for _, pos := range m.Portfolio.Holdings {
				// Show all holdings, even with 0 quantity
				priceStr := "Loading..."
				if pos.CurrentPrice > 0 {
					priceStr = ui.FormatPrice(pos.CurrentPrice)
				}

				content.WriteString(fmt.Sprintf("%-10s    %12.4f    %-15s    %-15s    %s\n",
					pos.AssetCode,
					pos.Quantity,
					priceStr,
					ui.FormatMarketValue(pos.MarketValue),
					ui.FormatCurrency(pos.DayChange),
				))
			}
			content.WriteString("\n")
		} else {
			content.WriteString("No crypto holdings found.\n")
			content.WriteString("Your buying power is available for trading.\n\n")
		}

		// Recent Orders
		if len(m.Portfolio.Orders) > 0 {
			content.WriteString("📋 RECENT ORDERS\n")
			content.WriteString("═══════════════\n")
			content.WriteString("Symbol    Side  Quantity     Avg Price    State\n")
			content.WriteString("───────────────────────────────────────────────\n")

			maxOrders := 3
			if len(m.Portfolio.Orders) < maxOrders {
				maxOrders = len(m.Portfolio.Orders)
			}

			for i := 0; i < maxOrders; i++ {
				order := m.Portfolio.Orders[i]
				avgPriceStr := "Market"
				if order.AveragePrice > 0 {
					avgPriceStr = ui.FormatValue(order.AveragePrice)
				}
				content.WriteString(fmt.Sprintf("%-8s  %-4s  %8.4f  %-10s  %s\n",
					order.Symbol,
					strings.ToUpper(order.Side),
					order.FilledQuantity,
					avgPriceStr,
					order.State,
				))
			}
			content.WriteString("\n")
		}

		// Last updated
		if !m.Portfolio.LastUpdated.IsZero() {
			content.WriteString(fmt.Sprintf("Last updated: %s\n",
				m.Portfolio.LastUpdated.Format("3:04 PM")))
		}
	}

	footer := ui.InfoStyle.Render("Press 'R' or 'F5' to refresh • 'Esc' to return to menu • Auto-refresh every 5s")

	return fmt.Sprintf("%s\n%s\n%s", title, ui.MenuStyle.Render(content.String()), footer)
}

// marketDataView renders real-time crypto market data
func (m *AppModel) marketDataView() string {
	title := ui.HeaderStyle.Render("📊 CRYPTO MARKET DATA")

	var content strings.Builder

	if m.Error != "" {
		content.WriteString(ui.NegativeStyle.Render("❌ " + m.Error + "\n\n"))
	}

	if m.Loading {
		content.WriteString(ui.LoadingStyle.Render("🔄 Loading market data...\n\n"))
	} else if m.MarketData == nil {
		content.WriteString("📊 Loading market data...\n")
		content.WriteString("This may take a moment on first load.\n\n")
	} else {
		// Top Gainers
		content.WriteString("🚀 TOP GAINERS (24H)\n")
		content.WriteString("═══════════════════════\n")
		content.WriteString("Symbol    Price        Change      Volume         Market Cap\n")
		content.WriteString("──────────────────────────────────────────────────────────────\n")

		for _, crypto := range m.MarketData.TopGainers {
			changeStr := ui.PositiveStyle.Render(fmt.Sprintf("+%.2f%%", crypto.ChangePercent24h))
			content.WriteString(fmt.Sprintf("%-8s  %-11s  %-10s  %-13s  %s\n",
				crypto.Symbol,
				ui.FormatValue(crypto.Price),
				changeStr,
				ui.FormatCompact(crypto.Volume24h),
				ui.FormatCompact(crypto.MarketCap),
			))
		}

		content.WriteString("\n")

		// Top Losers
		content.WriteString("📉 TOP LOSERS (24H)\n")
		content.WriteString("══════════════════════\n")
		content.WriteString("Symbol    Price        Change      Volume         Market Cap\n")
		content.WriteString("──────────────────────────────────────────────────────────────\n")

		for _, crypto := range m.MarketData.TopLosers {
			changeStr := ui.NegativeStyle.Render(fmt.Sprintf("%.2f%%", crypto.ChangePercent24h))
			content.WriteString(fmt.Sprintf("%-8s  %-11s  %-10s  %-13s  %s\n",
				crypto.Symbol,
				ui.FormatValue(crypto.Price),
				changeStr,
				ui.FormatCompact(crypto.Volume24h),
				ui.FormatCompact(crypto.MarketCap),
			))
		}

		content.WriteString("\n")

		// Last updated
		if !m.MarketData.LastUpdated.IsZero() {
			content.WriteString(fmt.Sprintf("Last updated: %s\n",
				m.MarketData.LastUpdated.Format("3:04 PM")))
		}

		content.WriteString("\n")
		content.WriteString(ui.InfoStyle.Render("💡 All cryptocurrencies shown are available for trading on Robinhood"))
	}

	footer := ui.InfoStyle.Render("Press 'R' or 'F5' to refresh • 'Esc' to return to menu • Auto-refresh every 30s")

	return fmt.Sprintf("%s\n%s\n%s", title, ui.MenuStyle.Render(content.String()), footer)
}

// orderHistoryView renders the order history page
func (m *AppModel) orderHistoryView() string {
	title := ui.HeaderStyle.Render("📋 ORDER HISTORY")

	var content strings.Builder

	if !m.Authenticated {
		content.WriteString("Please login first!")
		footer := ui.InfoStyle.Render("Press 'Esc' to return to menu")
		return fmt.Sprintf("%s\n%s\n%s", title, ui.MenuStyle.Render(content.String()), footer)
	}

	if m.Error != "" {
		content.WriteString(ui.NegativeStyle.Render("❌ " + m.Error + "\n\n"))
	}

	if m.Loading {
		content.WriteString(ui.LoadingStyle.Render("🔄 Loading order history...\n\n"))
	} else if m.Portfolio == nil {
		content.WriteString("📊 No portfolio data available.\n")
		content.WriteString("Press 'R' or 'F5' to refresh.\n\n")
	} else if len(m.Portfolio.Orders) == 0 {
		content.WriteString("📊 No orders found.\n")
		content.WriteString("Your order history will appear here once you start trading.\n\n")
	} else {
		// Order History
		content.WriteString("📋 RECENT ORDERS\n")
		content.WriteString("═══════════════════\n")
		content.WriteString("Date/Time         Symbol      Side   Type    Quantity      Avg Price    State\n")
		content.WriteString("─────────────────────────────────────────────────────────────────────────────\n")

		for _, order := range m.Portfolio.Orders {
			// Parse and format the timestamp
			createdTime := order.CreatedAt
			if len(createdTime) > 16 {
				createdTime = createdTime[:16] // Truncate for display
			}

			avgPriceStr := "Market"
			if order.AveragePrice > 0 {
				avgPriceStr = ui.FormatValue(order.AveragePrice)
			}

			// Color code the side
			sideStr := strings.ToUpper(order.Side)
			if order.Side == "buy" {
				sideStr = ui.PositiveStyle.Render(sideStr)
			} else {
				sideStr = ui.NegativeStyle.Render(sideStr)
			}

			// Color code the state
			stateStr := order.State
			switch order.State {
			case "filled", "confirmed":
				stateStr = ui.PositiveStyle.Render(order.State)
			case "cancelled", "rejected":
				stateStr = ui.NegativeStyle.Render(order.State)
			case "pending", "queued":
				stateStr = ui.LoadingStyle.Render(order.State)
			}

			content.WriteString(fmt.Sprintf("%-16s  %-10s  %-5s  %-6s  %8.4f      %-12s %s\n",
				createdTime,
				order.Symbol,
				sideStr,
				strings.ToUpper(order.Type),
				order.FilledQuantity,
				avgPriceStr,
				stateStr,
			))
		}

		content.WriteString("\n")
		content.WriteString(fmt.Sprintf("Showing %d most recent orders\n", len(m.Portfolio.Orders)))

		// Last updated
		if !m.Portfolio.LastUpdated.IsZero() {
			content.WriteString(fmt.Sprintf("Last updated: %s\n",
				m.Portfolio.LastUpdated.Format("3:04 PM")))
		}
	}

	footer := ui.InfoStyle.Render("Press 'R' or 'F5' to refresh • 'Esc' to return to menu")

	return fmt.Sprintf("%s\n%s\n%s", title, ui.MenuStyle.Render(content.String()), footer)
}

// newsView renders the crypto news feed
func (m *AppModel) newsView() string {
	title := ui.HeaderStyle.Render("📰 CRYPTO NEWS FEED")

	var content strings.Builder

	if m.Error != "" {
		content.WriteString(ui.NegativeStyle.Render("❌ " + m.Error + "\n\n"))
	}

	if m.Loading {
		content.WriteString(ui.LoadingStyle.Render("🔄 Loading crypto news...\n\n"))
	} else if m.NewsData == nil {
		content.WriteString("📰 Loading latest crypto news...\n")
		content.WriteString("This may take a moment on first load.\n\n")
	} else {
		// Consolidate all news into single feed
		allNews := m.getAllNewsArticles()

		if len(allNews) == 0 {
			content.WriteString("No news articles available.\n\n")
		} else {
			// Calculate pagination
			articlesPerPage := 8
			totalPages := (len(allNews) + articlesPerPage - 1) / articlesPerPage

			// Ensure NewsPage is within bounds
			if m.NewsPage >= totalPages {
				m.NewsPage = totalPages - 1
			}
			if m.NewsPage < 0 {
				m.NewsPage = 0
			}

			startIdx := m.NewsPage * articlesPerPage
			endIdx := min(startIdx + articlesPerPage, len(allNews))

			// Navigation info
			content.WriteString(fmt.Sprintf("📰 Page %d of %d • Articles %d-%d of %d total\n",
				m.NewsPage+1, totalPages, startIdx+1, endIdx, len(allNews)))
			content.WriteString("═══════════════════════════════════\n\n")

			// Display articles for current page
			for i := startIdx; i < endIdx; i++ {
				content.WriteString(m.formatNewsArticleWithTokens(allNews[i], i-startIdx+1))
			}
		}
	}

	footer := ui.InfoStyle.Render("Press '←/→' or 'A/D' for pages • 'Esc' to return • 'F5' to refresh")
	return fmt.Sprintf("%s\n%s\n%s", title, ui.MenuStyle.Render(content.String()), footer)
}

// formatNewsArticle formats a single news article for display
func (m *AppModel) formatNewsArticle(article NewsArticle, index int) string {
	var result strings.Builder

	// Impact indicator
	var impactIcon string
	switch article.Impact {
	case "bullish":
		impactIcon = ui.PositiveStyle.Render("📈")
	case "bearish":
		impactIcon = ui.NegativeStyle.Render("📉")
	default:
		impactIcon = ui.NeutralStyle.Render("📊")
	}

	// Article header with impact
	result.WriteString(fmt.Sprintf("%s %s\n", impactIcon, article.Title))
	result.WriteString(fmt.Sprintf("   %s\n", article.Summary))

	// Source and time info
	sourceInfo := fmt.Sprintf("   %s • %s",
		ui.DisabledStyle.Render(article.Source),
		ui.DisabledStyle.Render(article.PublishedAt))

	// Add symbols if present
	if len(article.Symbols) > 0 {
		symbols := strings.Join(article.Symbols, ", ")
		sourceInfo += ui.DisabledStyle.Render(" • Affects: " + symbols)
	}

	result.WriteString(sourceInfo + "\n\n")
	return result.String()
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// getAllNewsArticles returns all news articles from unified feed
func (m *AppModel) getAllNewsArticles() []NewsArticle {
	// All articles are now stored in BreakingNews - unified feed
	return m.NewsData.BreakingNews
}

// formatNewsArticleWithTokens formats a news article with crypto price indicators
func (m *AppModel) formatNewsArticleWithTokens(article NewsArticle, index int) string {
	var result strings.Builder

	// Impact indicator
	var impactIcon string
	switch article.Impact {
	case "bullish":
		impactIcon = ui.PositiveStyle.Render("📈")
	case "bearish":
		impactIcon = ui.NegativeStyle.Render("📉")
	default:
		impactIcon = ui.NeutralStyle.Render("📊")
	}

	// Article header with number and impact
	result.WriteString(fmt.Sprintf("%d. %s %s\n", index, impactIcon, article.Title))
	result.WriteString(fmt.Sprintf("   %s\n", article.Summary))

	// Source and time info
	sourceInfo := fmt.Sprintf("   %s • %s",
		ui.DisabledStyle.Render(article.Source),
		ui.DisabledStyle.Render(article.PublishedAt))

	// Add crypto tokens with price indicators if present
	if len(article.Symbols) > 0 {
		tokenInfo := m.formatTokenIndicators(article.Symbols)
		if tokenInfo != "" {
			sourceInfo += " • " + tokenInfo
		}
	}

	result.WriteString(sourceInfo + "\n\n")
	return result.String()
}

// formatTokenIndicators creates price indicators for mentioned crypto tokens
func (m *AppModel) formatTokenIndicators(symbols []string) string {
	if len(symbols) == 0 {
		return ""
	}

	var tokenParts []string

	for _, symbol := range symbols {
		// Get current price and change for the token
		tokenDisplay := m.getTokenPriceIndicator(symbol)
		if tokenDisplay != "" {
			tokenParts = append(tokenParts, tokenDisplay)
		} else {
			// Fallback if no price data
			tokenParts = append(tokenParts, ui.NeutralStyle.Render(symbol))
		}
	}

	if len(tokenParts) > 0 {
		return "Tokens: " + strings.Join(tokenParts, " ")
	}
	return ""
}

// algoTradingView renders the algorithmic trading interface
func (m *AppModel) algoTradingView() string {
	if !m.Authenticated {
		title := ui.HeaderStyle.Render("🤖 ALGORITHMIC TRADING")
		content := "Please login first!"
		footer := ui.InfoStyle.Render("Press 'Esc' to return to menu")
		return fmt.Sprintf("%s\n%s\n%s", title, ui.MenuStyle.Render(content), footer)
	}

	title := ui.HeaderStyle.Render("🤖 ALGORITHMIC TRADING")
	var content strings.Builder

	if m.Error != "" {
		content.WriteString(ui.NegativeStyle.Render("❌ " + m.Error + "\n\n"))
	}

	// Update state before rendering
	m.updateAlgoTradingState()

	// Engine Status
	content.WriteString("🔧 **TRADING ENGINE STATUS**\n")
	content.WriteString("══════════════════════════\n")

	engineStatus := "🔴 Stopped"
	if m.AlgoState.IsEngineRunning {
		if m.AlgoState.IsEnginePaused {
			engineStatus = "⏸️ Paused"
		} else {
			engineStatus = "🟢 Running"
		}
	}
	content.WriteString(fmt.Sprintf("Status:          %s\n", engineStatus))
	content.WriteString(fmt.Sprintf("Active Strategies: %d\n", len(m.AlgoState.ActiveStrategies)))
	content.WriteString(fmt.Sprintf("Active Positions:  %d\n", m.AlgoState.ActiveTrades))
	content.WriteString(fmt.Sprintf("Total Trades:      %d\n", m.AlgoState.TotalTrades))

	// Performance Summary
	content.WriteString("\n💰 **PERFORMANCE SUMMARY**\n")
	content.WriteString("═════════════════════════\n")
	content.WriteString(fmt.Sprintf("Total P&L:       %s\n", ui.FormatCurrency(m.AlgoState.TotalPnL)))
	content.WriteString(fmt.Sprintf("Daily P&L:       %s\n", ui.FormatCurrency(m.AlgoState.DailyPnL)))

	if m.AlgoState.PerformanceMetrics != nil {
		metrics := m.AlgoState.PerformanceMetrics
		content.WriteString(fmt.Sprintf("Win Rate:        %.1f%%\n", metrics.WinRate*100))
		content.WriteString(fmt.Sprintf("Profit Factor:   %.2f\n", metrics.ProfitFactor))
		content.WriteString(fmt.Sprintf("Max Drawdown:    %.1f%%\n", metrics.MaxDrawdownPercent))
	}

	// Strategy Status
	content.WriteString("\n📊 **STRATEGY STATUS**\n")
	content.WriteString("═══════════════════════\n")
	content.WriteString("Strategy         Symbol    Status      Position    P&L         Trades\n")
	content.WriteString("────────────────────────────────────────────────────────────────────\n")

	if len(m.StrategyConfigs) == 0 {
		content.WriteString("No strategies configured.\n")
	} else {
		for i, config := range m.StrategyConfigs {
			state, isActive := m.AlgoState.ActiveStrategies[config.Name]

			statusStr := "⚪ Disabled"
			if config.Enabled {
				if isActive {
					statusStr = "🟢 Active"
				} else {
					statusStr = "🔴 Stopped"
				}
			}

			positionStr := "None"
			pnlStr := "$0.00"
			tradesStr := "0"

			if isActive {
				if state.Position.Quantity > 0 {
					positionStr = fmt.Sprintf("%.4f", state.Position.Quantity)
				}
				pnlStr = ui.FormatCurrency(state.PnL)
				tradesStr = fmt.Sprintf("%d", state.TradesCount)
			}

			content.WriteString(fmt.Sprintf("%d. %-12s %-8s %-10s %-10s %-10s %s\n",
				i+1,
				config.Name[:min(12, len(config.Name))],
				config.Symbol,
				statusStr,
				positionStr,
				pnlStr,
				tradesStr,
			))
		}
	}

	// Strategy Details (if any active)
	if len(m.AlgoState.ActiveStrategies) > 0 {
		content.WriteString("\n📈 **ACTIVE STRATEGY DETAILS**\n")
		content.WriteString("════════════════════════════\n")

		for name, state := range m.AlgoState.ActiveStrategies {
			content.WriteString(fmt.Sprintf("\n**%s (%s)**\n", name, state.Symbol))

			// Debug: Show basic state info
			content.WriteString(fmt.Sprintf("  Running:       %v\n", state.IsRunning))
			content.WriteString(fmt.Sprintf("  Last Update:   %s\n", state.LastUpdateTime.Format("15:04:05")))
			content.WriteString(fmt.Sprintf("  Metadata Keys: %d\n", len(state.Metadata)))
			if state.LastSignal != nil {
				content.WriteString(fmt.Sprintf("  Last Signal:   %s at $%.2f (%.1f%% confidence)\n",
					strings.ToUpper(string(state.LastSignal.Type)),
					state.LastSignal.Price,
					state.LastSignal.Confidence*100))
				content.WriteString(fmt.Sprintf("  Signal Time:   %s\n",
					state.LastSignal.Timestamp.Format("15:04:05")))
			}

			if state.Position.Quantity > 0 {
				content.WriteString(fmt.Sprintf("  Position:      %.4f @ $%.2f\n",
					state.Position.Quantity,
					state.Position.AveragePrice))
				content.WriteString(fmt.Sprintf("  Market Value:  %s\n",
					ui.FormatCurrency(state.Position.MarketValue)))
				content.WriteString(fmt.Sprintf("  Unrealized:    %s\n",
					ui.FormatCurrency(state.Position.UnrealizedPnL)))
			}

			// Strategy-specific metadata
			if metadata := state.Metadata; len(metadata) > 0 {
				content.WriteString("  Indicators:    ")

				// Debug: Show all metadata keys
				content.WriteString("[")
				keys := make([]string, 0, len(metadata))
				for k := range metadata {
					keys = append(keys, k)
				}
				for i, k := range keys {
					if i > 0 {
						content.WriteString(", ")
					}
					content.WriteString(k)
				}
				content.WriteString("] ")

				// Moving Average strategy
				if shortMA, ok := metadata["short_ma"].(float64); ok {
					shortPeriod := 10 // default
					if sp, exists := metadata["short_period"].(int); exists {
						shortPeriod = sp
					}
					content.WriteString(fmt.Sprintf("MA(%d): $%.2f ", shortPeriod, shortMA))
				}
				if longMA, ok := metadata["long_ma"].(float64); ok {
					longPeriod := 30 // default
					if lp, exists := metadata["long_period"].(int); exists {
						longPeriod = lp
					}
					content.WriteString(fmt.Sprintf("MA(%d): $%.2f ", longPeriod, longMA))
				}

				// Show crossover status
				if shortMA, shortOK := metadata["short_ma"].(float64); shortOK {
					if longMA, longOK := metadata["long_ma"].(float64); longOK && shortMA > 0 && longMA > 0 {
						if shortMA > longMA {
							content.WriteString("📈BULLISH ")
						} else {
							content.WriteString("📉BEARISH ")
						}
						spread := ((shortMA - longMA) / longMA) * 100
						content.WriteString(fmt.Sprintf("(%.3f%%) ", spread))
					}
				}

				// Momentum strategy
				if momentum, ok := metadata["momentum_pct"].(float64); ok {
					content.WriteString(fmt.Sprintf("Mom: %.2f%% ", momentum))
				}

				// Show price count for debugging
				if pricesCount, ok := metadata["prices_count"].(int); ok {
					content.WriteString(fmt.Sprintf("[%d prices] ", pricesCount))
				}
				content.WriteString("\n")
			}
		}
	}

	// Controls
	content.WriteString("\n🎮 **CONTROLS**\n")
	content.WriteString("═════════════\n")
	content.WriteString("S - Start Engine    T - Stop Engine     P - Pause/Resume\n")
	content.WriteString("E - Emergency Stop   1-5 - Toggle Strategy\n")
	content.WriteString("R - Refresh          Esc - Back to Menu\n")

	// Last update time
	if !m.AlgoState.LastUpdate.IsZero() {
		content.WriteString(fmt.Sprintf("\nLast updated: %s\n",
			m.AlgoState.LastUpdate.Format("15:04:05")))
	}

	footer := ui.InfoStyle.Render("⚠️  CAUTION: Algorithmic trading involves significant risk. Monitor closely!")

	return fmt.Sprintf("%s\n%s\n%s", title, ui.MenuStyle.Render(content.String()), footer)
}

// getTokenPriceIndicator returns a formatted token with price change indicator
func (m *AppModel) getTokenPriceIndicator(symbol string) string {
	// Try to find price data from portfolio holdings first
	if m.Portfolio != nil {
		for _, pos := range m.Portfolio.Holdings {
			if pos.AssetCode == symbol || pos.AssetCode+"-USD" == symbol {
				if pos.DayChange > 0 {
					return ui.PositiveStyle.Render(fmt.Sprintf("%s ↗", symbol))
				} else if pos.DayChange < 0 {
					return ui.NegativeStyle.Render(fmt.Sprintf("%s ↘", symbol))
				} else {
					return ui.NeutralStyle.Render(fmt.Sprintf("%s →", symbol))
				}
			}
		}
	}

	// Check market data if available (both gainers and losers)
	if m.MarketData != nil {
		// Check top gainers
		for _, crypto := range m.MarketData.TopGainers {
			cryptoSymbol := strings.Replace(crypto.Symbol, "-USD", "", 1)
			if cryptoSymbol == symbol || crypto.Symbol == symbol {
				return ui.PositiveStyle.Render(fmt.Sprintf("%s ↗", symbol))
			}
		}
		// Check top losers
		for _, crypto := range m.MarketData.TopLosers {
			cryptoSymbol := strings.Replace(crypto.Symbol, "-USD", "", 1)
			if cryptoSymbol == symbol || crypto.Symbol == symbol {
				return ui.NegativeStyle.Render(fmt.Sprintf("%s ↘", symbol))
			}
		}
	}

	// For tokens not in portfolio or market data, fetch live price change from API
	priceChange := m.getLiveTokenPriceChange(symbol)
	if priceChange > 0 {
		return ui.PositiveStyle.Render(fmt.Sprintf("%s ↗", symbol))
	} else if priceChange < 0 {
		return ui.NegativeStyle.Render(fmt.Sprintf("%s ↘", symbol))
	}

	// Return symbol with neutral indicator if price data unavailable
	return ui.NeutralStyle.Render(fmt.Sprintf("%s →", symbol))
}