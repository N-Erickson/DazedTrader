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
	} else if m.Portfolio == nil || len(m.Portfolio.Orders) == 0 {
		content.WriteString("📊 Loading your order history...\n")
		content.WriteString("This may take a moment on first load.\n\n")
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

// getTokenPriceIndicator returns a formatted token with price change indicator
func (m *AppModel) getTokenPriceIndicator(symbol string) string {
	// Try to find price data from portfolio holdings
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

	// Return symbol without indicator if no price data found
	return ui.NeutralStyle.Render(symbol)
}