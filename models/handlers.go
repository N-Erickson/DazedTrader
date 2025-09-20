package models

import (
	"dazedtrader/ui"
	"fmt"
	"strings"

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
		switch m.State {
		case StateLogin:
			if m.LoginStep > 0 {
				m.LoginStep--
				m.Error = ""
				return m, nil
			}
			fallthrough
		default:
			m.State = StateMenu
			m.Error = ""
		}
		return m, nil

	case "f5":
		// Refresh data if on dashboard
		if (m.State == StateDashboard || m.State == StatePortfolio) && m.Authenticated && !m.Loading {
			m.Error = ""
			return m, m.loadPortfolioCmd()
		}
		return m, nil

	case "r":
		// Only handle 'r' for refresh if NOT in login form
		if m.State != StateLogin && (m.State == StateDashboard || m.State == StatePortfolio) && m.Authenticated && !m.Loading {
			m.Error = ""
			return m, m.loadPortfolioCmd()
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
		m.Cursor = 0
		return m.handleMenuSelection()
	case "2":
		if m.Authenticated {
			m.Cursor = 1
			return m.handleMenuSelection()
		}
	case "3":
		if m.Authenticated {
			m.Cursor = 2
			return m.handleMenuSelection()
		}
	case "4":
		if m.Authenticated {
			m.Cursor = 3
			return m.handleMenuSelection()
		}
	}
	return m, nil
}

func (m *AppModel) handleMenuSelection() (tea.Model, tea.Cmd) {
	switch m.Cursor {
	case 0: // Login
		if !m.Authenticated {
			m.State = StateLogin
			m.LoginStep = 0
			m.Error = ""
		}
	case 1: // Dashboard
		if m.Authenticated {
			m.State = StateDashboard
			m.Error = ""
			if m.Portfolio == nil {
				return m, m.loadPortfolioCmd()
			}
		}
	case 2: // Trading
		if m.Authenticated {
			m.State = StateTrading
		}
	case 3: // Positions
		if m.Authenticated {
			m.State = StatePortfolio
			if m.Portfolio == nil {
				return m, m.loadPortfolioCmd()
			}
		}
	case 4: // Orders
		// TODO: Implement orders view
	case 5: // Help
		m.State = StateHelp
	case 6: // Logout
		if m.Authenticated {
			m.HandleLogout()
		}
	case 7: // Exit
		return m, tea.Quit
	}
	return m, nil
}

func (m *AppModel) handleLoginKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.LoginStep == 0 && m.LoginForm.Username != "" {
			m.LoginStep = 1
			return m, nil
		}
		if m.LoginStep == 1 && m.LoginForm.Password != "" {
			// Try login
			m.Error = ""
			return m, m.loginCmd()
		}
		if m.LoginStep == 2 && len(m.LoginForm.MFACode) == 6 {
			// Try login with MFA
			m.Error = ""
			return m, m.loginCmd()
		}
		return m, nil

	case "tab":
		if m.LoginStep == 1 {
			m.ShowPassword = !m.ShowPassword
		}
		return m, nil

	case "backspace":
		switch m.LoginStep {
		case 0:
			if len(m.LoginForm.Username) > 0 {
				m.LoginForm.Username = m.LoginForm.Username[:len(m.LoginForm.Username)-1]
			}
		case 1:
			if len(m.LoginForm.Password) > 0 {
				m.LoginForm.Password = m.LoginForm.Password[:len(m.LoginForm.Password)-1]
			}
		case 2:
			if len(m.LoginForm.MFACode) > 0 {
				m.LoginForm.MFACode = m.LoginForm.MFACode[:len(m.LoginForm.MFACode)-1]
			}
		}
		return m, nil

	default:
		// Handle text input
		if len(msg.String()) == 1 {
			char := msg.String()
			switch m.LoginStep {
			case 0:
				m.LoginForm.Username += char
			case 1:
				m.LoginForm.Password += char
			case 2:
				if len(m.LoginForm.MFACode) < 6 && char >= "0" && char <= "9" {
					m.LoginForm.MFACode += char
				}
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

	if m.Loading {
		content.WriteString(ui.LoadingStyle.Render("🔄 Loading positions...\n\n"))
	} else if m.Portfolio == nil || len(m.Portfolio.Positions) == 0 {
		content.WriteString("No positions found.\n")
		content.WriteString("Start trading to see your holdings here!\n\n")
	} else {
		content.WriteString("Symbol      Shares        Price     Market Value    Day Change     % Change\n")
		content.WriteString("─────────────────────────────────────────────────────────────────────────\n")

		for _, pos := range m.Portfolio.Positions {
			changePercent := 0.0
			if pos.Price > 0 && pos.Shares > 0 {
				previousPrice := pos.Price - (pos.DayChange / pos.Shares)
				if previousPrice > 0 {
					changePercent = ((pos.Price - previousPrice) / previousPrice) * 100
				}
			}

			content.WriteString(fmt.Sprintf("%-10s %10.4f   %s   %s   %s    %s\n",
				pos.Symbol,
				pos.Shares,
				ui.FormatValue(pos.Price),
				ui.FormatValue(pos.MarketValue),
				ui.FormatCurrency(pos.DayChange),
				ui.FormatPercentage(changePercent),
			))
		}

		total := fmt.Sprintf("\nTOTAL PORTFOLIO VALUE: %s", ui.FormatValue(m.Portfolio.TotalValue))
		content.WriteString(total)
	}

	footer := ui.InfoStyle.Render("Press 'Esc' to return to menu • 'R' or 'F5' to refresh")

	return fmt.Sprintf("%s\n%s\n%s", title, ui.MenuStyle.Render(content.String()), footer)
}

func (m *AppModel) tradingView() string {
	title := ui.HeaderStyle.Render("💹 TRADING INTERFACE")

	content := `
🎯 TRADING FEATURES
══════════════════

This section will include:
• Stock quote lookup
• Buy/sell order entry
• Order management
• Real-time market data

⚠️  IMPLEMENTATION IN PROGRESS ⚠️

For now, use the official Robinhood app for trading.
Portfolio viewing and monitoring are fully functional.

Current Status:
✅ Portfolio tracking
✅ Position monitoring
✅ Order history
🔄 Trading interface (coming soon)
`

	if m.Authenticated && m.Portfolio != nil {
		content += fmt.Sprintf("\nYour buying power: %s", ui.FormatValue(m.Portfolio.BuyingPower))
	}

	footer := ui.InfoStyle.Render("Press 'Esc' to return to menu")

	return fmt.Sprintf("%s\n%s\n%s", title, ui.MenuStyle.Render(content), footer)
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