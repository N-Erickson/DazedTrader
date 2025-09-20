package api

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	BaseURL = "https://api.robinhood.com"
)

type Client struct {
	HTTPClient *http.Client
	Token      string
	BaseURL    string
}

type LoginRequest struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	MFACode      string `json:"mfa_code,omitempty"`
	GrantType    string `json:"grant_type"`
	ClientID     string `json:"client_id"`
	Scope        string `json:"scope"`
	DeviceToken  string `json:"device_token"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	Token        string `json:"token"`
	MFARequired  bool   `json:"mfa_required"`
	MFAType      string `json:"mfa_type"`
	Detail       string `json:"detail"`
	Error        string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type User struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type Account struct {
	URL                    string  `json:"url"`
	AccountNumber          string  `json:"account_number"`
	Type                   string  `json:"type"`
	BuyingPower            string  `json:"buying_power"`
	DayTradeBuyingPower   string  `json:"day_trade_buying_power"`
	MaxACHEarlyAccessAmount string `json:"max_ach_early_access_amount"`
	Portfolio              string  `json:"portfolio"`
}

type Portfolio struct {
	URL                      string `json:"url"`
	MarketValue             string `json:"market_value"`
	TotalReturnToday        string `json:"total_return_today"`
	WithdrawableAmount      string `json:"withdrawable_amount"`
	StartDate               string `json:"start_date"`
	LastCoreMarketTrading   string `json:"last_core_market_trading_day"`
}

type Position struct {
	URL        string `json:"url"`
	Instrument string `json:"instrument"`
	Quantity   string `json:"quantity"`
	Account    string `json:"account"`
	SharesHeld string `json:"shares_held_for_sells"`
}

type Instrument struct {
	URL        string `json:"url"`
	Symbol     string `json:"symbol"`
	Name       string `json:"name"`
	SimpleName string `json:"simple_name"`
}

type Quote struct {
	Symbol           string `json:"symbol"`
	LastTradePrice  string `json:"last_trade_price"`
	PreviousClose   string `json:"previous_close"`
	BidPrice        string `json:"bid_price"`
	AskPrice        string `json:"ask_price"`
	BidSize         string `json:"bid_size"`
	AskSize         string `json:"ask_size"`
	UpdatedAt       string `json:"updated_at"`
}

type Order struct {
	ID           string `json:"id"`
	URL          string `json:"url"`
	Account      string `json:"account"`
	Instrument   string `json:"instrument"`
	Side         string `json:"side"`
	Type         string `json:"type"`
	Quantity     string `json:"quantity"`
	Price        string `json:"price"`
	State        string `json:"state"`
	TimeInForce  string `json:"time_in_force"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

func NewClient() *Client {
	return &Client{
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		BaseURL: BaseURL,
	}
}

func (c *Client) SetToken(token string) {
	c.Token = token
}

func (c *Client) makeRequest(method, endpoint string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	var bodyContent []byte
	if body != nil {
		var err error
		bodyContent, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(bodyContent)
	}

	fullURL := c.BaseURL + endpoint
	req, err := http.NewRequest(method, fullURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://robinhood.com")
	req.Header.Set("Referer", "https://robinhood.com/login/")
	req.Header.Set("Sec-Ch-Ua", "\"Google Chrome\";v=\"121\", \"Not A(Brand\";v=\"99\", \"Chromium\";v=\"121\"")
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", "\"macOS\"")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	if c.Token != "" {
		req.Header.Set("Authorization", "Token "+c.Token)
	}

	// Debug: Print full request details
	fmt.Printf("DEBUG: %s %s\n", method, fullURL)
	fmt.Printf("DEBUG: Request headers: %+v\n", req.Header)
	if bodyContent != nil {
		fmt.Printf("DEBUG: Request body: %s\n", string(bodyContent))
	}

	return c.HTTPClient.Do(req)
}

// generateDeviceToken creates a random hex string for device identification
func generateDeviceToken() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func (c *Client) Login(username, password, mfaCode string) (*LoginResponse, error) {
	loginReq := LoginRequest{
		Username:    username,
		Password:    password,
		MFACode:     mfaCode,
		GrantType:   "password",
		ClientID:    "rZDCFrI7Lgu8kgB2vKPu8rRcJYdGEq2lPcHD6zJu",
		Scope:       "internal",
		DeviceToken: generateDeviceToken(),
	}

	// Debug: Print request details
	fmt.Printf("DEBUG: Making request to %s/oauth2/token/\n", c.BaseURL)
	fmt.Printf("DEBUG: Login request for user: %s\n", loginReq.Username)

	resp, err := c.makeRequest("POST", "/oauth2/token/", loginReq)
	if err != nil {
		return nil, fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	// Debug: Print response details
	fmt.Printf("DEBUG: Response status: %d\n", resp.StatusCode)
	fmt.Printf("DEBUG: Response headers: %+v\n", resp.Header)

	// Read the response body for debugging
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	fmt.Printf("DEBUG: Response body: %s\n", string(body))

	// Check HTTP status code
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	// Check if body is empty
	if len(body) == 0 {
		return nil, fmt.Errorf("empty response from server")
	}

	var loginResp LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return nil, fmt.Errorf("failed to decode login response (body: %s): %w", string(body), err)
	}

	// Handle OAuth2 response
	if loginResp.Error != "" {
		return nil, fmt.Errorf("authentication error: %s - %s", loginResp.Error, loginResp.ErrorDescription)
	}

	// Use access_token if available, fallback to token field
	token := loginResp.AccessToken
	if token == "" {
		token = loginResp.Token
	}

	if token != "" {
		c.SetToken(token)
		loginResp.Token = token // Ensure token field is set for compatibility
	}

	return &loginResp, nil
}

func (c *Client) GetUser() (*User, error) {
	resp, err := c.makeRequest("GET", "/user/", nil)
	if err != nil {
		return nil, fmt.Errorf("get user request failed: %w", err)
	}
	defer resp.Body.Close()

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode user response: %w", err)
	}

	return &user, nil
}

func (c *Client) GetAccounts() ([]Account, error) {
	resp, err := c.makeRequest("GET", "/accounts/", nil)
	if err != nil {
		return nil, fmt.Errorf("get accounts request failed: %w", err)
	}
	defer resp.Body.Close()

	var accountsResp struct {
		Results []Account `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&accountsResp); err != nil {
		return nil, fmt.Errorf("failed to decode accounts response: %w", err)
	}

	return accountsResp.Results, nil
}

func (c *Client) GetPortfolio(portfolioURL string) (*Portfolio, error) {
	// Extract path from full URL
	endpoint := portfolioURL
	if len(portfolioURL) > len(c.BaseURL) {
		endpoint = portfolioURL[len(c.BaseURL):]
	}

	resp, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("get portfolio request failed: %w", err)
	}
	defer resp.Body.Close()

	var portfolio Portfolio
	if err := json.NewDecoder(resp.Body).Decode(&portfolio); err != nil {
		return nil, fmt.Errorf("failed to decode portfolio response: %w", err)
	}

	return &portfolio, nil
}

func (c *Client) GetPositions() ([]Position, error) {
	resp, err := c.makeRequest("GET", "/positions/", nil)
	if err != nil {
		return nil, fmt.Errorf("get positions request failed: %w", err)
	}
	defer resp.Body.Close()

	var positionsResp struct {
		Results []Position `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&positionsResp); err != nil {
		return nil, fmt.Errorf("failed to decode positions response: %w", err)
	}

	return positionsResp.Results, nil
}

func (c *Client) GetInstrument(instrumentURL string) (*Instrument, error) {
	endpoint := instrumentURL
	if len(instrumentURL) > len(c.BaseURL) {
		endpoint = instrumentURL[len(c.BaseURL):]
	}

	resp, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("get instrument request failed: %w", err)
	}
	defer resp.Body.Close()

	var instrument Instrument
	if err := json.NewDecoder(resp.Body).Decode(&instrument); err != nil {
		return nil, fmt.Errorf("failed to decode instrument response: %w", err)
	}

	return &instrument, nil
}

func (c *Client) GetQuotes(symbols []string) ([]Quote, error) {
	symbolsParam := ""
	for i, symbol := range symbols {
		if i > 0 {
			symbolsParam += ","
		}
		symbolsParam += symbol
	}

	endpoint := fmt.Sprintf("/quotes/?symbols=%s", symbolsParam)
	resp, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("get quotes request failed: %w", err)
	}
	defer resp.Body.Close()

	var quotesResp struct {
		Results []Quote `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&quotesResp); err != nil {
		return nil, fmt.Errorf("failed to decode quotes response: %w", err)
	}

	return quotesResp.Results, nil
}

func (c *Client) GetOrders() ([]Order, error) {
	resp, err := c.makeRequest("GET", "/orders/", nil)
	if err != nil {
		return nil, fmt.Errorf("get orders request failed: %w", err)
	}
	defer resp.Body.Close()

	var ordersResp struct {
		Results []Order `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&ordersResp); err != nil {
		return nil, fmt.Errorf("failed to decode orders response: %w", err)
	}

	return ordersResp.Results, nil
}

func (c *Client) PlaceOrder(order Order) (*Order, error) {
	resp, err := c.makeRequest("POST", "/orders/", order)
	if err != nil {
		return nil, fmt.Errorf("place order request failed: %w", err)
	}
	defer resp.Body.Close()

	var placedOrder Order
	if err := json.NewDecoder(resp.Body).Decode(&placedOrder); err != nil {
		return nil, fmt.Errorf("failed to decode order response: %w", err)
	}

	return &placedOrder, nil
}

func (c *Client) CancelOrder(orderID string) error {
	endpoint := fmt.Sprintf("/orders/%s/cancel/", orderID)
	resp, err := c.makeRequest("POST", endpoint, nil)
	if err != nil {
		return fmt.Errorf("cancel order request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return fmt.Errorf("failed to cancel order: status %d", resp.StatusCode)
	}

	return nil
}