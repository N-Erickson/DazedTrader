package models

import (
	"dazedtrader/ui"
	"fmt"
	"strconv"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *AppModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		if m.State == StateMenu {
			return m, tea.Quit
		}
		// Only quit from menu, otherwise go back
		m.State = StateMenu
		m.Error = ""
		return m, nil

	case "esc":
		// Always go back or to menu
		// Reset trading form when leaving trading state
		if m.State == StateTrading {
			m.TradingForm = TradingForm{}
			m.TradingStep = 0
		}
		m.State = StateMenu
		m.Error = ""
		return m, nil

	case "f5":
		// Refresh data based on current state
		if (m.State == StateDashboard || m.State == StatePortfolio || m.State == StateOrderHistory) && m.Authenticated && !m.Loading {
			m.Error = ""
			return m, m.loadCryptoPortfolioCmd()
		} else if m.State == StateMarketData && !m.Loading {
			m.Error = ""
			return m, m.loadMarketDataCmd()
		} else if m.State == StateNews && !m.Loading {
			m.Error = ""
			return m, m.loadNewsDataCmd()
		}
		return m, nil

	case "r":
		// Only handle 'r' for refresh if NOT in API key setup
		if m.State != StateLogin && (m.State == StateDashboard || m.State == StatePortfolio || m.State == StateOrderHistory) && m.Authenticated && !m.Loading {
			m.Error = ""
			return m, m.loadCryptoPortfolioCmd()
		} else if m.State == StateMarketData && !m.Loading {
			m.Error = ""
			return m, m.loadMarketDataCmd()
		} else if m.State == StateNews && !m.Loading {
			m.Error = ""
			return m, m.loadNewsDataCmd()
		}
		// If in login state, don't handle it globally - let it fall through to login handler
		if m.State == StateLogin {
			break // Break out of this switch to continue to state-specific handlers
		}
		return m, nil
	}

	// Handle state-specific key presses
	switch m.State {
	case StateMenu:
		return m.handleMenuKeys(msg)
	case StateLogin:
		return m.handleLoginKeys(msg)
	case StateDashboard:
		return m.handleDashboardKeys(msg)
	case StatePortfolio:
		return m.handlePortfolioKeys(msg)
	case StateTrading:
		return m.handleTradingKeys(msg)
	case StateAlgoTrading:
		return m.handleAlgoTradingKeys(msg)
	case StateMarketData:
		return m.handleMarketDataKeys(msg)
	case StateOrderHistory:
		return m.handleOrderHistoryKeys(msg)
	case StateNews:
		return m.handleNewsKeys(msg)
	}

	return m, nil
}

func (m *AppModel) handleMenuKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.Cursor > 0 {
			m.Cursor--
		}
	case "down", "j":
		if m.Cursor < len(m.Choices)-1 {
			m.Cursor++
		}
	case "enter", " ":
		return m.handleMenuSelection()
	case "1":
		if m.Authenticated {
			m.Cursor = 0
			return m.handleMenuSelection()
		}
	case "2":
		if m.Authenticated {
			m.Cursor = 1
			return m.handleMenuSelection()
		}
	case "3":
		m.Cursor = 2
		return m.handleMenuSelection()
	case "4":
		if m.Authenticated {
			m.Cursor = 3
			return m.handleMenuSelection()
		}
	case "5":
		m.Cursor = 4
		return m.handleMenuSelection()
	case "6":
		if !m.Authenticated {
			m.Cursor = 5
			return m.handleMenuSelection()
		}
	}
	return m, nil
}

func (m *AppModel) handleMenuSelection() (tea.Model, tea.Cmd) {
	switch m.Cursor {
	case 0: // Crypto Portfolio
		if m.Authenticated {
			m.State = StateDashboard
			m.Error = ""
			if m.Portfolio == nil {
				return m, m.loadCryptoPortfolioCmd()
			}
		}
	case 1: // Crypto Trading
		if m.Authenticated {
			m.State = StateTrading
		}
	case 2: // Algorithmic Trading
		if m.Authenticated {
			m.State = StateAlgoTrading
		}
	case 3: // Market Data
		m.State = StateMarketData
		if m.MarketData == nil {
			return m, m.loadMarketDataCmd()
		}
	case 4: // Order History
		if m.Authenticated {
			m.State = StateOrderHistory
			if m.Portfolio == nil {
				return m, m.loadCryptoPortfolioCmd()
			}
		}
	case 5: // Crypto News
		m.State = StateNews
		if m.NewsData == nil {
			return m, m.loadNewsDataCmd()
		}
	case 6: // API Key Setup
		if !m.Authenticated {
			m.State = StateLogin
			m.Error = ""
		}
	case 7: // Help
		m.State = StateHelp
	case 8: // Logout
		if m.Authenticated {
			m.HandleLogout()
		}
	case 9: // Exit
		return m, tea.Quit
	}
	return m, nil
}

func (m *AppModel) handleLoginKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Submit API key
		if m.APIKeyForm.APIKey != "" {
			m.Error = ""
			return m, m.apiKeySetupCmd()
		}
		return m, nil

	case "tab":
		m.ShowAPIKey = !m.ShowAPIKey
		return m, nil

	case "ctrl+v":
		// Paste from clipboard
		clipboardText, err := clipboard.ReadAll()
		if err == nil && clipboardText != "" {
			// Clean the clipboard text (remove newlines and whitespace)
			clipboardText = strings.ReplaceAll(clipboardText, "\n", "")
			clipboardText = strings.ReplaceAll(clipboardText, "\r", "")
			clipboardText = strings.TrimSpace(clipboardText)
			m.APIKeyForm.APIKey = clipboardText
		}
		return m, nil

	case "backspace":
		if len(m.APIKeyForm.APIKey) > 0 {
			m.APIKeyForm.APIKey = m.APIKeyForm.APIKey[:len(m.APIKeyForm.APIKey)-1]
		}
		return m, nil

	case "ctrl+a":
		// Clear all text (select all + delete)
		m.APIKeyForm.APIKey = ""
		return m, nil

	default:
		// Handle text input for API key
		if len(msg.String()) == 1 {
			char := msg.String()
			if char[0] >= 32 && char[0] <= 126 { // Printable ASCII
				m.APIKeyForm.APIKey += char
			}
		}
	}
	return m, nil
}

func (m *AppModel) handleDashboardKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Dashboard-specific shortcuts can go here
	return m, nil
}

func (m *AppModel) handlePortfolioKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Portfolio view-specific shortcuts can go here
	return m, nil
}

func (m *AppModel) handleMarketDataKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Market data view-specific shortcuts can go here
	return m, nil
}

func (m *AppModel) handleOrderHistoryKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Order history view-specific shortcuts can go here
	return m, nil
}

func (m *AppModel) handleAlgoTradingKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "s", "S":
		// Start trading engine
		if m.TradingEngine != nil && !m.TradingEngine.IsRunning() {
			if err := m.TradingEngine.Start(); err != nil {
				m.Error = fmt.Sprintf("Failed to start trading engine: %v", err)
			} else {
				m.updateAlgoTradingState()
			}
		}
	case "t", "T":
		// Stop trading engine
		if m.TradingEngine != nil && m.TradingEngine.IsRunning() {
			if err := m.TradingEngine.Stop(); err != nil {
				m.Error = fmt.Sprintf("Failed to stop trading engine: %v", err)
			} else {
				m.updateAlgoTradingState()
			}
		}
	case "p", "P":
		// Pause/Resume trading engine
		if m.TradingEngine != nil && m.TradingEngine.IsRunning() {
			if m.TradingEngine.IsPaused() {
				m.TradingEngine.Resume()
			} else {
				m.TradingEngine.Pause()
			}
			m.updateAlgoTradingState()
		}
	case "e", "E":
		// Emergency stop
		if m.TradingEngine != nil {
			if err := m.TradingEngine.EmergencyStop(); err != nil {
				m.Error = fmt.Sprintf("Emergency stop failed: %v", err)
			} else {
				m.updateAlgoTradingState()
			}
		}
	case "1", "2", "3", "4", "5":
		// Start/stop individual strategies
		strategyIndex, _ := strconv.Atoi(msg.String())
		if strategyIndex > 0 && strategyIndex <= len(m.StrategyConfigs) {
			config := m.StrategyConfigs[strategyIndex-1]
			if m.TradingEngine != nil {
				if _, exists := m.AlgoState.ActiveStrategies[config.Name]; exists {
					// Stop strategy
					m.TradingEngine.StopStrategy(config.Name)
				} else {
					// Start strategy - toggle enabled state first
					m.StrategyConfigs[strategyIndex-1].Enabled = !m.StrategyConfigs[strategyIndex-1].Enabled
					if m.StrategyConfigs[strategyIndex-1].Enabled {
						m.TradingEngine.StartStrategy(config.Name, m.StrategyConfigs[strategyIndex-1])
					}
				}
				m.updateAlgoTradingState()
			}
		}
	case "r", "R":
		// Refresh algo trading status
		m.updateAlgoTradingState()
	}
	return m, nil
}

func (m *AppModel) handleNewsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "left", "a":
		// Previous page
		if m.NewsPage > 0 {
			m.NewsPage--
		}
	case "right", "d":
		// Next page - calculate dynamic total pages
		if m.NewsData != nil {
			allNews := m.getAllNewsArticles()
			articlesPerPage := 8
			totalPages := (len(allNews) + articlesPerPage - 1) / articlesPerPage
			if totalPages > 0 && m.NewsPage < totalPages-1 {
				m.NewsPage++
			}
		}
	}
	return m, nil
}

func (m *AppModel) handleTradingKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.TradingStep {
	case TradingStepSymbol:
		return m.handleTradingSymbolInput(msg)
	case TradingStepSide:
		return m.handleTradingSideSelection(msg)
	case TradingStepType:
		return m.handleTradingTypeSelection(msg)
	case TradingStepQuantity:
		return m.handleTradingQuantityInput(msg)
	case TradingStepPrice:
		return m.handleTradingPriceInput(msg)
	case TradingStepConfirm:
		return m.handleTradingConfirmation(msg)
	default:
		// Default trading view
		switch msg.String() {
		case "enter":
			m.TradingStep = TradingStepSymbol
			m.TradingForm = TradingForm{} // Reset form
			return m, nil
		case "1":
			m.TradingForm.Symbol = "BTC-USD"
			if price, err := m.GetLivePrice("BTC-USD"); err == nil {
				m.TradingForm.CurrentPrice = price
			}
			m.TradingStep = TradingStepSide
			m.TradingForm.Side = "buy"
			return m, nil
		case "2":
			m.TradingForm.Symbol = "ETH-USD"
			if price, err := m.GetLivePrice("ETH-USD"); err == nil {
				m.TradingForm.CurrentPrice = price
			}
			m.TradingStep = TradingStepSide
			m.TradingForm.Side = "buy"
			return m, nil
		case "3":
			m.TradingForm.Symbol = "DOGE-USD"
			if price, err := m.GetLivePrice("DOGE-USD"); err == nil {
				m.TradingForm.CurrentPrice = price
			}
			m.TradingStep = TradingStepSide
			m.TradingForm.Side = "buy"
			return m, nil
		case "4":
			m.TradingForm.Symbol = "ADA-USD"
			if price, err := m.GetLivePrice("ADA-USD"); err == nil {
				m.TradingForm.CurrentPrice = price
			}
			m.TradingStep = TradingStepSide
			m.TradingForm.Side = "buy"
			return m, nil
		}
	}
	return m, nil
}

func (m *AppModel) handleTradingSymbolInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.TradingForm.Symbol != "" {
			// Fetch live price for the symbol
			if price, err := m.GetLivePrice(m.TradingForm.Symbol); err == nil {
				m.TradingForm.CurrentPrice = price
			}
			m.TradingStep = TradingStepSide
			m.TradingForm.Side = "buy" // Default to buy
		}
		return m, nil
	case "backspace":
		if len(m.TradingForm.Symbol) > 0 {
			m.TradingForm.Symbol = m.TradingForm.Symbol[:len(m.TradingForm.Symbol)-1]
		}
		return m, nil
	default:
		if len(msg.String()) == 1 {
			char := msg.String()
			// Allow letters, numbers, and hyphens for crypto symbols
			if (char[0] >= 'A' && char[0] <= 'Z') || (char[0] >= 'a' && char[0] <= 'z') ||
			   (char[0] >= '0' && char[0] <= '9') || char[0] == '-' {
				m.TradingForm.Symbol += strings.ToUpper(char)
			}
		}
	}
	return m, nil
}

func (m *AppModel) handleTradingSideSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.TradingStep = TradingStepType
		m.TradingForm.Type = "market" // Default to market
		return m, nil
	case "up", "down":
		if m.TradingForm.Side == "buy" {
			m.TradingForm.Side = "sell"
		} else {
			m.TradingForm.Side = "buy"
		}
		return m, nil
	case "backspace":
		// Go back to symbol selection
		m.TradingStep = TradingStepSymbol
		return m, nil
	}
	return m, nil
}

func (m *AppModel) handleTradingTypeSelection(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.TradingStep = TradingStepQuantity
		// Trigger async price fetch if we don't have a current price
		if m.TradingForm.CurrentPrice == 0 && m.TradingForm.Symbol != "" {
			return m, m.fetchTradingPriceCmd()
		}
		return m, nil
	case "up", "down":
		if m.TradingForm.Type == "market" {
			m.TradingForm.Type = "limit"
		} else {
			m.TradingForm.Type = "market"
		}
		return m, nil
	case "backspace":
		// Go back to side selection
		m.TradingStep = TradingStepSide
		return m, nil
	}
	return m, nil
}

func (m *AppModel) handleTradingQuantityInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.TradingForm.Quantity != "" {
			if m.TradingForm.Type == "limit" {
				m.TradingStep = TradingStepPrice
			} else {
				m.TradingStep = TradingStepConfirm
			}
		}
		return m, nil
	case "backspace":
		if len(m.TradingForm.Quantity) > 0 {
			m.TradingForm.Quantity = m.TradingForm.Quantity[:len(m.TradingForm.Quantity)-1]
			// Update estimated cost when deleting characters
			m.updateEstimatedCost()
		} else {
			// Go back to type selection if quantity is empty
			m.TradingStep = TradingStepType
		}
		return m, nil
	default:
		if len(msg.String()) == 1 {
			char := msg.String()
			// Allow numbers and decimal point
			if (char[0] >= '0' && char[0] <= '9') || char[0] == '.' {
				m.TradingForm.Quantity += char
				// Update estimated cost in real-time
				m.updateEstimatedCost()
			}
		}
	}
	return m, nil
}

func (m *AppModel) handleTradingPriceInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.TradingForm.Price != "" {
			m.TradingStep = TradingStepConfirm
		}
		return m, nil
	case "backspace":
		if len(m.TradingForm.Price) > 0 {
			m.TradingForm.Price = m.TradingForm.Price[:len(m.TradingForm.Price)-1]
		} else {
			// Go back to quantity if price is empty
			m.TradingStep = TradingStepQuantity
		}
		return m, nil
	default:
		if len(msg.String()) == 1 {
			char := msg.String()
			// Allow numbers and decimal point
			if (char[0] >= '0' && char[0] <= '9') || char[0] == '.' {
				m.TradingForm.Price += char
				// Update estimated cost in real-time
				m.updateEstimatedCost()
			}
		}
	}
	return m, nil
}

func (m *AppModel) handleTradingConfirmation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		// Place the order
		return m, m.placeOrderCmd()
	case "backspace":
		// Go back to previous step
		if m.TradingForm.Type == "limit" {
			m.TradingStep = TradingStepPrice
		} else {
			m.TradingStep = TradingStepQuantity
		}
		return m, nil
	case "esc":
		// Cancel and return to menu
		m.State = StateMenu
		m.TradingStep = 0
		m.TradingForm = TradingForm{}
		return m, nil
	}
	return m, nil
}

func (m *AppModel) menuView() string {
	title := ui.TitleStyle.Render("🚀 DAZED TRADER 🚀\nRobinhood Terminal Interface")

	var menu string
	menu += "Choose an option:\n\n"

	for i, choice := range m.Choices {
		cursor := " "
		if m.Cursor == i {
			cursor = ">"
			choice = ui.SelectedStyle.Render(choice)
		} else {
			choice = ui.UnselectedStyle.Render(choice)
		}

		// Disable options if not authenticated
		needsAuth := (i >= 1 && i <= 4) // Dashboard, Trading, Positions, Orders
		logoutOption := i == 6

		if needsAuth && !m.Authenticated {
			choice = ui.DisabledStyle.Render(choice + " (Login Required)")
		} else if logoutOption && !m.Authenticated {
			choice = ui.DisabledStyle.Render(choice + " (Not Logged In)")
		}

		menu += fmt.Sprintf("%s %s\n", cursor, choice)
	}

	authStatus := "🔴 Not Authenticated"
	if m.Authenticated {
		authStatus = fmt.Sprintf("🟢 Authenticated as %s", m.Username)
	}

	footer := ui.InfoStyle.Render(fmt.Sprintf("\nStatus: %s\nPress 'q' to quit • Use ↑↓ to navigate • Enter to select\nShortcuts: 1-Login 2-Dashboard 3-Trading 4-Portfolio", authStatus))

	return fmt.Sprintf("%s\n\n%s\n%s", title, ui.MenuStyle.Render(menu), footer)
}

func (m *AppModel) portfolioView() string {
	if !m.Authenticated {
		return "Please login first!"
	}

	title := ui.HeaderStyle.Render("📈 DETAILED POSITIONS")

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
		content.WriteString(ui.LoadingStyle.Render("🔄 Loading positions...\n\n"))
	} else if m.Portfolio == nil || len(m.Portfolio.Holdings) == 0 {
		content.WriteString("No positions found.\n")
		content.WriteString("Start trading to see your holdings here!\n\n")
	} else {
		content.WriteString("Symbol         Shares              Price         Market Value      Day Change      % Change\n")
		content.WriteString("────────────────────────────────────────────────────────────────────────────────────────\n")

		for _, pos := range m.Portfolio.Holdings {
			changePercent := 0.0
			if pos.CurrentPrice > 0 && pos.Quantity > 0 {
				previousPrice := pos.CurrentPrice - (pos.DayChange / pos.Quantity)
				if previousPrice > 0 {
					changePercent = ((pos.CurrentPrice - previousPrice) / previousPrice) * 100
				}
			}

			// Format with proper padding and color coding
			content.WriteString(fmt.Sprintf("%-8s %15.4f %18s %18s %15s %12s\n",
				pos.AssetCode,
				pos.Quantity,
				ui.FormatPrice(pos.CurrentPrice),
				ui.FormatMarketValue(pos.MarketValue),
				ui.FormatCurrency(pos.DayChange),
				ui.FormatPercentage(changePercent),
			))
		}

		// Calculate total value from holdings
		totalValue := 0.0
		for _, holding := range m.Portfolio.Holdings {
			totalValue += holding.MarketValue
		}
		total := fmt.Sprintf("\nTOTAL PORTFOLIO VALUE: %s", ui.FormatValue(totalValue))
		content.WriteString(total)
	}

	footer := ui.InfoStyle.Render("Press 'Esc' to return to menu • 'R' or 'F5' to refresh")

	return fmt.Sprintf("%s\n%s\n%s", title, ui.MenuStyle.Render(content.String()), footer)
}

func (m *AppModel) tradingView() string {
	title := ui.HeaderStyle.Render("💹 CRYPTO TRADING")

	var content strings.Builder

	if !m.Authenticated {
		content.WriteString("Please login first!")
		footer := ui.InfoStyle.Render("Press 'Esc' to return to menu")
		return fmt.Sprintf("%s\n%s\n%s", title, ui.MenuStyle.Render(content.String()), footer)
	}

	if m.Error != "" {
		content.WriteString(ui.NegativeStyle.Render("❌ " + m.Error + "\n\n"))
	}

	if m.TradingForm.Submitting {
		content.WriteString(ui.LoadingStyle.Render("🔄 Placing order...\n\n"))
		footer := ui.InfoStyle.Render("Please wait...")
		return fmt.Sprintf("%s\n%s\n%s", title, ui.MenuStyle.Render(content.String()), footer)
	}

	// Show current step
	switch m.TradingStep {
	case TradingStepSymbol:
		content.WriteString("📝 **STEP 1: SELECT SYMBOL**\n\n")
		content.WriteString("Enter crypto symbol (e.g., BTC-USD, ETH-USD):\n")
		content.WriteString(ui.InputStyle.Render(m.TradingForm.Symbol + "│") + "\n\n")
		content.WriteString("Popular symbols: BTC-USD, ETH-USD, DOGE-USD, ADA-USD\n")
		if m.TradingForm.CurrentPrice > 0 {
			content.WriteString(fmt.Sprintf("\n💰 Current Price: %s", ui.FormatValue(m.TradingForm.CurrentPrice)))
		}

	case TradingStepSide:
		content.WriteString("📈 **STEP 2: BUY OR SELL**\n\n")
		content.WriteString(fmt.Sprintf("Symbol: %s", m.TradingForm.Symbol))
		if m.TradingForm.CurrentPrice > 0 {
			content.WriteString(fmt.Sprintf(" | Price: %s", ui.FormatValue(m.TradingForm.CurrentPrice)))
		}
		content.WriteString("\n\n")
		content.WriteString("Choose order side:\n")
		if m.TradingForm.Side == "buy" {
			content.WriteString(ui.SelectedStyle.Render("► BUY") + "\n")
			content.WriteString(ui.UnselectedStyle.Render("  SELL") + "\n")
		} else {
			content.WriteString(ui.UnselectedStyle.Render("  BUY") + "\n")
			content.WriteString(ui.SelectedStyle.Render("► SELL") + "\n")
		}

	case TradingStepType:
		content.WriteString("⚙️ **STEP 3: ORDER TYPE**\n\n")
		content.WriteString(fmt.Sprintf("Symbol: %s | Side: %s", m.TradingForm.Symbol, strings.ToUpper(m.TradingForm.Side)))
		if m.TradingForm.CurrentPrice > 0 {
			content.WriteString(fmt.Sprintf(" | Price: %s", ui.FormatValue(m.TradingForm.CurrentPrice)))
		}
		content.WriteString("\n\n")
		content.WriteString("Choose order type:\n")
		if m.TradingForm.Type == "market" {
			content.WriteString(ui.SelectedStyle.Render("► MARKET (Execute immediately at current price)") + "\n")
			content.WriteString(ui.UnselectedStyle.Render("  LIMIT (Set your own price)") + "\n")
		} else {
			content.WriteString(ui.UnselectedStyle.Render("  MARKET (Execute immediately at current price)") + "\n")
			content.WriteString(ui.SelectedStyle.Render("► LIMIT (Set your own price)") + "\n")
		}

	case TradingStepQuantity:
		content.WriteString("🔢 **STEP 4: QUANTITY**\n\n")
		content.WriteString(fmt.Sprintf("Symbol: %s | Side: %s | Type: %s",
			m.TradingForm.Symbol, strings.ToUpper(m.TradingForm.Side), strings.ToUpper(m.TradingForm.Type)))
		if m.TradingForm.CurrentPrice > 0 {
			content.WriteString(fmt.Sprintf(" | Price: %s", ui.FormatValue(m.TradingForm.CurrentPrice)))
		}
		content.WriteString("\n\n")
		content.WriteString("Enter quantity:\n")
		content.WriteString(ui.InputStyle.Render(m.TradingForm.Quantity + "│") + "\n\n")
		if m.TradingForm.EstimatedCost > 0 {
			content.WriteString(fmt.Sprintf("💰 Estimated Cost: %s\n", ui.FormatValue(m.TradingForm.EstimatedCost)))
		}
		if m.Portfolio != nil {
			content.WriteString(fmt.Sprintf("Available buying power: %s\n", ui.FormatValue(m.Portfolio.BuyingPower)))
		}

	case TradingStepPrice:
		content.WriteString("💰 **STEP 5: LIMIT PRICE**\n\n")
		content.WriteString(fmt.Sprintf("Symbol: %s | Side: %s | Quantity: %s",
			m.TradingForm.Symbol, strings.ToUpper(m.TradingForm.Side), m.TradingForm.Quantity))
		if m.TradingForm.CurrentPrice > 0 {
			content.WriteString(fmt.Sprintf(" | Market Price: %s", ui.FormatValue(m.TradingForm.CurrentPrice)))
		}
		content.WriteString("\n\n")
		content.WriteString("Enter limit price (USD):\n")
		content.WriteString(ui.InputStyle.Render(m.TradingForm.Price + "│") + "\n\n")
		if m.TradingForm.EstimatedCost > 0 {
			content.WriteString(fmt.Sprintf("💰 Estimated Cost: %s\n", ui.FormatValue(m.TradingForm.EstimatedCost)))
		}

	case TradingStepConfirm:
		content.WriteString("✅ **STEP 6: CONFIRM ORDER**\n\n")
		content.WriteString("Please review your order:\n\n")
		content.WriteString(fmt.Sprintf("Symbol:      %s\n", m.TradingForm.Symbol))
		content.WriteString(fmt.Sprintf("Side:        %s\n", strings.ToUpper(m.TradingForm.Side)))
		content.WriteString(fmt.Sprintf("Type:        %s\n", strings.ToUpper(m.TradingForm.Type)))
		content.WriteString(fmt.Sprintf("Quantity:    %s\n", m.TradingForm.Quantity))
		if m.TradingForm.CurrentPrice > 0 {
			content.WriteString(fmt.Sprintf("Market Price: %s\n", ui.FormatValue(m.TradingForm.CurrentPrice)))
		}
		if m.TradingForm.Type == "limit" {
			if price, err := strconv.ParseFloat(m.TradingForm.Price, 64); err == nil {
				content.WriteString(fmt.Sprintf("Limit Price:  %s\n", ui.FormatValue(price)))
			} else {
				content.WriteString(fmt.Sprintf("Limit Price:  $%s\n", m.TradingForm.Price))
			}
		}
		if m.TradingForm.EstimatedCost > 0 {
			content.WriteString(fmt.Sprintf("💰 Est. Cost:  %s\n", ui.FormatValue(m.TradingForm.EstimatedCost)))
		}
		content.WriteString("\n")
		content.WriteString(ui.PositiveStyle.Render("Press ENTER to place order") + "\n")
		content.WriteString(ui.NegativeStyle.Render("Press ESC to cancel") + "\n")

	default:
		// Default view - show quick trading options
		content.WriteString("🚀 **QUICK TRADE**\n\n")
		content.WriteString("Popular crypto pairs:\n")
		content.WriteString("1. BTC-USD - Bitcoin\n")
		content.WriteString("2. ETH-USD - Ethereum\n")
		content.WriteString("3. DOGE-USD - Dogecoin\n")
		content.WriteString("4. ADA-USD - Cardano\n\n")
		if m.Portfolio != nil {
			content.WriteString(fmt.Sprintf("💰 Available buying power: %s\n\n", ui.FormatValue(m.Portfolio.BuyingPower)))
		}
		content.WriteString("Press ENTER to start trading")
	}

	footer := ui.InfoStyle.Render("Use ↑↓ to navigate • Enter to confirm • Esc to go back • Backspace to edit")
	return fmt.Sprintf("%s\n%s\n%s", title, ui.MenuStyle.Render(content.String()), footer)
}

func (m *AppModel) helpView() string {
	title := ui.HeaderStyle.Render("❓ HELP & INFORMATION")

	content := `
🚀 DAZED TRADER HELP
═══════════════════

KEYBOARD SHORTCUTS:
  ↑↓ or jk    - Navigate menu options
  Enter/Space - Select option
  Esc         - Go back / Return to main menu
  Q           - Quit application (from main menu)
  R/F5        - Refresh data (on dashboard)
  Tab         - Toggle password visibility (login)

NAVIGATION:
  1 - Quick login
  2 - Quick dashboard (if authenticated)
  3 - Quick trading (if authenticated)
  4 - Quick positions (if authenticated)

FEATURES:
  🔐 Secure Authentication - 2FA/MFA supported
  📊 Real-time Portfolio - Live data from Robinhood
  📈 Position Tracking - Detailed holdings view
  📋 Order History - Recent trading activity
  🔄 Auto-refresh - Updates every 30 seconds

SECURITY NOTES:
  • Uses official Robinhood API endpoints
  • Credentials sent directly to Robinhood (not stored)
  • Authentication tokens stored securely locally
  • All communication over HTTPS

DISCLAIMER:
  This app uses unofficial Robinhood APIs and is provided
  as-is. Always verify trades in the official Robinhood app.
  Trading involves risk - trade responsibly!
`

	footer := ui.InfoStyle.Render("Press 'Esc' to return to menu")

	return fmt.Sprintf("%s\n%s\n%s", title, ui.MenuStyle.Render(content), footer)
}