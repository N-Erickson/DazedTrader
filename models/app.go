package models

import (
	"dazedtrader/api"
	"dazedtrader/auth"
	"fmt"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type AppModel struct {
	State         int
	Choices       []string
	Cursor        int
	Width         int
	Height        int
	Authenticated bool
	Username      string
	Client        *api.Client
	Portfolio     *Portfolio
	Error         string
	Loading       bool

	// Login form state
	LoginStep    int // 0=username, 1=password, 2=mfa
	LoginForm    LoginForm
	ShowPassword bool
}

type LoginForm struct {
	Username string
	Password string
	MFACode  string
}

type Portfolio struct {
	TotalValue   float64
	DayChange    float64
	DayChangePct float64
	BuyingPower  float64
	Positions    []Position
	Orders       []Order
	LastUpdated  time.Time
}

type Position struct {
	Symbol      string
	Shares      float64
	Price       float64
	MarketValue float64
	DayChange   float64
}

type Order struct {
	ID       string
	Symbol   string
	Side     string
	Quantity float64
	Price    string
	State    string
	Created  time.Time
}

func NewAppModel() *AppModel {
	client := api.NewClient()

	// Try to load existing token
	tokenData, err := auth.LoadToken()
	authenticated := false
	username := ""

	if err == nil && tokenData != nil {
		// Check if token is still valid (assume 24 hour expiry)
		if time.Now().Unix() < tokenData.ExpiresAt {
			client.SetToken(tokenData.Token)
			authenticated = true
			username = tokenData.Username
		} else {
			// Token expired, clear it
			auth.ClearToken()
		}
	}

	return &AppModel{
		State: StateMenu,
		Choices: []string{
			"🔐 Login to Robinhood",
			"📊 Portfolio Dashboard",
			"💹 Trading Interface",
			"📈 View Positions",
			"📋 Recent Orders",
			"❓ Help",
			"🔓 Logout",
			"🚪 Exit",
		},
		Cursor:        0,
		Authenticated: authenticated,
		Username:      username,
		Client:        client,
		LoginStep:     0,
	}
}

// App states
const (
	StateMenu = iota
	StateLogin
	StateDashboard
	StatePortfolio
	StateTrading
	StateHelp
)

// LoadPortfolioData loads real portfolio data from Robinhood API
func (m *AppModel) LoadPortfolioData() error {
	if !m.Authenticated || m.Loading {
		return nil
	}

	m.Loading = true
	m.Error = ""

	defer func() {
		m.Loading = false
	}()

	// Get accounts
	accounts, err := m.Client.GetAccounts()
	if err != nil {
		m.Error = fmt.Sprintf("Failed to get accounts: %v", err)
		return err
	}

	if len(accounts) == 0 {
		m.Error = "No accounts found"
		return fmt.Errorf("no accounts found")
	}

	account := accounts[0]

	// Get portfolio
	portfolio, err := m.Client.GetPortfolio(account.Portfolio)
	if err != nil {
		m.Error = fmt.Sprintf("Failed to get portfolio: %v", err)
		return err
	}

	// Parse portfolio values
	totalValue, _ := strconv.ParseFloat(portfolio.MarketValue, 64)
	dayChange, _ := strconv.ParseFloat(portfolio.TotalReturnToday, 64)
	buyingPower, _ := strconv.ParseFloat(account.BuyingPower, 64)

	var dayChangePct float64
	if totalValue > 0 {
		dayChangePct = (dayChange / (totalValue - dayChange)) * 100
	}

	// Get positions
	positions, err := m.Client.GetPositions()
	if err != nil {
		m.Error = fmt.Sprintf("Failed to get positions: %v", err)
		return err
	}

	// Process positions
	var portfolioPositions []Position
	var symbols []string

	for _, pos := range positions {
		quantity, _ := strconv.ParseFloat(pos.Quantity, 64)
		if quantity <= 0 {
			continue
		}

		// Get instrument to get symbol
		instrument, err := m.Client.GetInstrument(pos.Instrument)
		if err != nil {
			continue
		}

		symbols = append(symbols, instrument.Symbol)
		portfolioPositions = append(portfolioPositions, Position{
			Symbol: instrument.Symbol,
			Shares: quantity,
		})
	}

	// Get quotes for symbols
	if len(symbols) > 0 {
		quotes, err := m.Client.GetQuotes(symbols)
		if err == nil {
			// Update positions with current prices
			quoteMap := make(map[string]float64)
			previousCloseMap := make(map[string]float64)

			for _, quote := range quotes {
				price, _ := strconv.ParseFloat(quote.LastTradePrice, 64)
				prevClose, _ := strconv.ParseFloat(quote.PreviousClose, 64)
				quoteMap[quote.Symbol] = price
				previousCloseMap[quote.Symbol] = prevClose
			}

			for i := range portfolioPositions {
				pos := &portfolioPositions[i]
				if price, exists := quoteMap[pos.Symbol]; exists {
					pos.Price = price
					pos.MarketValue = pos.Shares * price

					if prevClose, exists := previousCloseMap[pos.Symbol]; exists {
						pos.DayChange = (price - prevClose) * pos.Shares
					}
				}
			}
		}
	}

	// Get recent orders
	orders, err := m.Client.GetOrders()
	var portfolioOrders []Order
	if err == nil {
		maxOrders := 10
		if len(orders) < maxOrders {
			maxOrders = len(orders)
		}

		for i := 0; i < maxOrders; i++ {
			order := orders[i]
			quantity, _ := strconv.ParseFloat(order.Quantity, 64)

			// Get symbol for order
			instrument, err := m.Client.GetInstrument(order.Instrument)
			symbol := "Unknown"
			if err == nil {
				symbol = instrument.Symbol
			}

			price := "Market"
			if order.Price != "" {
				if priceVal, err := strconv.ParseFloat(order.Price, 64); err == nil {
					price = fmt.Sprintf("$%.2f", priceVal)
				}
			}

			createdAt, _ := time.Parse(time.RFC3339, order.CreatedAt)

			portfolioOrders = append(portfolioOrders, Order{
				ID:       order.ID,
				Symbol:   symbol,
				Side:     order.Side,
				Quantity: quantity,
				Price:    price,
				State:    order.State,
				Created:  createdAt,
			})
		}
	}

	// Update portfolio
	m.Portfolio = &Portfolio{
		TotalValue:   totalValue,
		DayChange:    dayChange,
		DayChangePct: dayChangePct,
		BuyingPower:  buyingPower,
		Positions:    portfolioPositions,
		Orders:       portfolioOrders,
		LastUpdated:  time.Now(),
	}

	return nil
}

// HandleLogin processes login with Robinhood API
func (m *AppModel) HandleLogin() error {
	if m.Loading {
		return nil // Already processing
	}

	m.Loading = true
	m.Error = ""

	// Perform login
	loginResp, err := m.Client.Login(
		m.LoginForm.Username,
		m.LoginForm.Password,
		m.LoginForm.MFACode,
	)

	m.Loading = false

	if err != nil {
		m.Error = fmt.Sprintf("Login failed: %v", err)
		return nil
	}

	if loginResp.Token != "" {
		// Login successful
		m.Authenticated = true
		m.Username = m.LoginForm.Username

		// Save token (expires in 24 hours)
		expiresAt := time.Now().Add(24 * time.Hour).Unix()
		if err := auth.SaveToken(loginResp.Token, m.Username, expiresAt); err != nil {
			m.Error = fmt.Sprintf("Failed to save token: %v", err)
			return nil
		}

		// Clear login form
		m.LoginForm = LoginForm{}
		m.LoginStep = 0
		m.State = StateDashboard

		return nil
	}

	if loginResp.MFARequired {
		// MFA required
		if m.LoginStep < 2 {
			m.LoginStep = 2
			return nil
		} else {
			m.Error = "Invalid MFA code. Please try again."
			m.LoginForm.MFACode = ""
			return nil
		}
	}

	m.Error = loginResp.Detail
	if m.Error == "" {
		m.Error = "Login failed for unknown reason"
	}

	return nil
}

// HandleLogout logs out and clears stored token
func (m *AppModel) HandleLogout() {
	m.Authenticated = false
	m.Username = ""
	m.Client.SetToken("")
	m.Portfolio = nil
	m.Error = ""

	// Clear stored token
	auth.ClearToken()

	// Reset login form
	m.LoginForm = LoginForm{}
	m.LoginStep = 0

	m.State = StateMenu
}

// Bubble Tea interface methods
func (m *AppModel) Init() tea.Cmd {
	// If already authenticated, start loading portfolio data
	if m.Authenticated {
		return tea.Batch(
			m.loadPortfolioCmd(),
			tickEvery(30*time.Second),
		)
	}
	return nil
}

func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case tickMsg:
		// Auto-refresh portfolio data if on dashboard and authenticated
		if m.State == StateDashboard && m.Authenticated && !m.Loading {
			return m, tea.Batch(
				m.loadPortfolioCmd(),
				tickEvery(30*time.Second),
			)
		}
		return m, tickEvery(30*time.Second)

	case portfolioLoadedMsg:
		// Portfolio data loaded, clear any loading state
		if msg.err != nil && m.Error == "" {
			m.Error = fmt.Sprintf("Failed to load portfolio: %v", msg.err)
		}
		return m, nil

	case loginCompletedMsg:
		// Login completed, start loading portfolio if successful
		if msg.err == nil && m.Authenticated {
			return m, tea.Batch(
				m.loadPortfolioCmd(),
				tickEvery(30*time.Second),
			)
		}
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}

	return m, nil
}

func (m *AppModel) View() string {
	switch m.State {
	case StateMenu:
		return m.menuView()
	case StateLogin:
		return m.LoginView()
	case StateDashboard:
		return m.DashboardView()
	case StatePortfolio:
		return m.portfolioView()
	case StateTrading:
		return m.tradingView()
	case StateHelp:
		return m.helpView()
	default:
		return m.menuView()
	}
}

// Message types for Bubble Tea
type tickMsg time.Time
type portfolioLoadedMsg struct{ err error }
type loginCompletedMsg struct{ err error }

func tickEvery(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m *AppModel) loadPortfolioCmd() tea.Cmd {
	return func() tea.Msg {
		err := m.LoadPortfolioData()
		return portfolioLoadedMsg{err: err}
	}
}

func (m *AppModel) loginCmd() tea.Cmd {
	return func() tea.Msg {
		err := m.HandleLogin()
		return loginCompletedMsg{err: err}
	}
}