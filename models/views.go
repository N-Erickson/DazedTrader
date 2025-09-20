package models

import (
	"dazedtrader/ui"
	"fmt"
	"strings"
)

// LoginView renders the login form
func (m *AppModel) LoginView() string {
	title := ui.HeaderStyle.Render("🔐 ROBINHOOD LOGIN")

	var content strings.Builder

	if m.Error != "" {
		content.WriteString(ui.NegativeStyle.Render("❌ " + m.Error + "\n\n"))
	}

	if m.Loading {
		content.WriteString("🔄 Authenticating...\n\n")
	} else {
		switch m.LoginStep {
		case 0: // Username
			content.WriteString(ui.PositiveStyle.Render("Enter your Robinhood username/email:") + "\n")
			userInput := m.LoginForm.Username
			if userInput == "" {
				userInput = "│"
			} else {
				userInput += "│"
			}
			content.WriteString(ui.InputStyle.Render(userInput) + "\n\n")
			content.WriteString("Press Enter to continue, Esc to cancel\n")

		case 1: // Password
			content.WriteString(ui.PositiveStyle.Render("Username: ") + m.LoginForm.Username + "\n")
			content.WriteString(ui.PositiveStyle.Render("Enter your password:") + "\n")

			var passDisplay string
			if m.ShowPassword {
				passDisplay = m.LoginForm.Password + "│"
			} else {
				passDisplay = strings.Repeat("*", len(m.LoginForm.Password)) + "│"
			}
			if passDisplay == "│" {
				passDisplay = "│"
			}

			content.WriteString(ui.InputStyle.Render(passDisplay) + "\n\n")
			content.WriteString("Press Tab to toggle visibility, Enter to continue, Esc to go back\n")

		case 2: // MFA
			content.WriteString(ui.PositiveStyle.Render("Username: ") + m.LoginForm.Username + "\n")
			content.WriteString(ui.PositiveStyle.Render("Two-Factor Authentication Required") + "\n\n")
			content.WriteString("Enter your 6-digit MFA code:\n")
			mfaInput := m.LoginForm.MFACode
			if mfaInput == "" {
				mfaInput = "│"
			} else {
				mfaInput += "│"
			}
			content.WriteString(ui.InputStyle.Render(mfaInput) + "\n\n")
			content.WriteString("Press Enter to login, Esc to go back\n")
		}
	}

	footer := ui.InfoStyle.Render("Tip: Your credentials are sent directly to Robinhood and stored securely locally")

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

	if m.Loading {
		content.WriteString(ui.LoadingStyle.Render("🔄 Loading portfolio data...\n\n"))
	} else if m.Portfolio == nil {
		content.WriteString("📊 Loading your portfolio...\n")
		content.WriteString("This may take a moment on first load.\n\n")
	} else {
		// Portfolio Summary
		content.WriteString("📈 PORTFOLIO SUMMARY\n")
		content.WriteString("═════════════════════\n")
		content.WriteString(fmt.Sprintf("Total Value:     %s\n", ui.FormatValue(m.Portfolio.TotalValue)))
		content.WriteString(fmt.Sprintf("Day Change:      %s (%s)\n",
			ui.FormatCurrency(m.Portfolio.DayChange),
			ui.FormatPercentage(m.Portfolio.DayChangePct)))
		content.WriteString(fmt.Sprintf("Buying Power:    %s\n\n", ui.FormatValue(m.Portfolio.BuyingPower)))

		// Top Positions
		if len(m.Portfolio.Positions) > 0 {
			content.WriteString("🏆 TOP POSITIONS\n")
			content.WriteString("═══════════════\n")
			content.WriteString("Symbol    Shares      Price    Market Value   Day Change\n")
			content.WriteString("────────────────────────────────────────────────────────\n")

			maxPositions := 5
			if len(m.Portfolio.Positions) < maxPositions {
				maxPositions = len(m.Portfolio.Positions)
			}

			for i := 0; i < maxPositions; i++ {
				pos := m.Portfolio.Positions[i]
				content.WriteString(fmt.Sprintf("%-8s %8.4f  %s   %s   %s\n",
					pos.Symbol,
					pos.Shares,
					ui.FormatValue(pos.Price),
					ui.FormatValue(pos.MarketValue),
					ui.FormatCurrency(pos.DayChange),
				))
			}
			content.WriteString("\n")
		}

		// Recent Orders
		if len(m.Portfolio.Orders) > 0 {
			content.WriteString("📋 RECENT ORDERS\n")
			content.WriteString("═══════════════\n")
			content.WriteString("Symbol  Side  Quantity    Price      State\n")
			content.WriteString("─────────────────────────────────────────\n")

			maxOrders := 3
			if len(m.Portfolio.Orders) < maxOrders {
				maxOrders = len(m.Portfolio.Orders)
			}

			for i := 0; i < maxOrders; i++ {
				order := m.Portfolio.Orders[i]
				content.WriteString(fmt.Sprintf("%-6s  %-4s  %8.4f  %-9s  %s\n",
					order.Symbol,
					strings.ToUpper(order.Side),
					order.Quantity,
					order.Price,
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

	footer := ui.InfoStyle.Render("Press 'R' or 'F5' to refresh • 'Esc' to return to menu • Auto-refresh every 30s")

	return fmt.Sprintf("%s\n%s\n%s", title, ui.MenuStyle.Render(content.String()), footer)
}