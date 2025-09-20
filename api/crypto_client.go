package api

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

const (
	CryptoBaseURL = "https://trading.robinhood.com"
	TradingURL    = "https://trading.robinhood.com/api/v1/crypto/trading"
	MarketDataURL = "https://trading.robinhood.com/api/v1/crypto/marketdata"
)

type CryptoClient struct {
	HTTPClient *http.Client
	APIKey     string
	PrivateKey ed25519.PrivateKey
}

// NewCryptoClient creates a new Robinhood crypto API client
// Input format: "apikey:privatekey" where privatekey is base64-encoded
func NewCryptoClient(credentials string) *CryptoClient {
	parts := strings.Split(credentials, ":")
	if len(parts) != 2 {
		return nil
	}

	apiKey := parts[0]
	privateKeyB64 := parts[1]

	// Decode the private key from base64
	privateKeyBytes, err := base64.StdEncoding.DecodeString(privateKeyB64)
	if err != nil {
		return nil
	}

	// Private key should be 64 bytes for Ed25519
	if len(privateKeyBytes) != ed25519.PrivateKeySize {
		return nil
	}

	privateKey := ed25519.PrivateKey(privateKeyBytes)

	return &CryptoClient{
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		APIKey:     apiKey,
		PrivateKey: privateKey,
	}
}

// Crypto data structures based on actual API responses
type CryptoAccount struct {
	AccountNumber      string `json:"account_number"`
	Status            string `json:"status"`
	BuyingPower       string `json:"buying_power"`
	BuyingPowerCurrency string `json:"buying_power_currency"`
}

type CryptoHolding struct {
	AccountNumber                string  `json:"account_number"`
	AssetCode                    string  `json:"asset_code"`
	TotalQuantity                float64 `json:"total_quantity"`
	QuantityAvailableForTrading  float64 `json:"quantity_available_for_trading"`
}

// Pagination wrapper for API responses
type PaginatedResponse[T any] struct {
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
	Results  []T     `json:"results"`
}

type BestBidAsk struct {
	Symbol                    string  `json:"symbol"`
	Price                     float64 `json:"price"`
	BidPrice                  float64 `json:"bid_inclusive_of_sell_spread"`
	SellSpread               float64 `json:"sell_spread"`
	AskPrice                  float64 `json:"ask_inclusive_of_buy_spread"`
	BuySpread                float64 `json:"buy_spread"`
	Timestamp                string  `json:"timestamp"`
}

type CryptoOrder struct {
	ID                string  `json:"id"`
	AccountNumber     string  `json:"account_number"`
	Symbol            string  `json:"symbol"`
	ClientOrderID     string  `json:"client_order_id"`
	Side              string  `json:"side"`
	Type              string  `json:"type"`
	State             string  `json:"state"`
	AveragePrice      float64 `json:"average_price"`
	FilledAssetQuantity float64 `json:"filled_asset_quantity"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
}

type OrderRequest struct {
	Side        string `json:"side"`
	Type        string `json:"type"`
	TimeInForce string `json:"time_in_force"`
	Quantity    string `json:"quantity"`
	Price       string `json:"price,omitempty"`
	CurrencyID  string `json:"currency_id"`
}

// makeRequest makes HTTP requests to Robinhood crypto API
func (c *CryptoClient) makeRequest(method, endpoint string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	var bodyString string

	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(bodyBytes)
		bodyString = string(bodyBytes)
	}

	req, err := http.NewRequest(method, endpoint, reqBody)
	if err != nil {
		return nil, err
	}

	// Generate timestamp (Unix timestamp in seconds)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	// Extract path from full URL
	path := strings.TrimPrefix(endpoint, CryptoBaseURL)

	// Create message to sign: api_key + timestamp + path + method + body
	message := c.APIKey + timestamp + path + method + bodyString

	// Sign the message with Ed25519
	signature := ed25519.Sign(c.PrivateKey, []byte(message))
	signatureB64 := base64.StdEncoding.EncodeToString(signature)

	// Set the required headers for Robinhood crypto API authentication
	req.Header.Set("x-api-key", c.APIKey)
	req.Header.Set("x-signature", signatureB64)
	req.Header.Set("x-timestamp", timestamp)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "DazedTrader/1.0")


	return c.HTTPClient.Do(req)
}

// GetCryptoAccount retrieves crypto account information
func (c *CryptoClient) GetCryptoAccount() (*CryptoAccount, error) {
	resp, err := c.makeRequest("GET", TradingURL+"/accounts/", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var account CryptoAccount
	if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
		return nil, err
	}

	return &account, nil
}

// GetCryptoHoldings retrieves all crypto holdings
func (c *CryptoClient) GetCryptoHoldings() ([]CryptoHolding, error) {
	resp, err := c.makeRequest("GET", TradingURL+"/holdings/", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var mapResponse struct {
		Results []map[string]interface{} `json:"results"`
	}

	if err := json.Unmarshal(bodyBytes, &mapResponse); err == nil {
		// Convert map to holdings using the actual field names
		var holdings []CryptoHolding
		for _, result := range mapResponse.Results {
			holding := CryptoHolding{}

			// Parse the exact fields from Robinhood API documentation
			if val, ok := result["account_number"].(string); ok {
				holding.AccountNumber = val
			}
			if val, ok := result["asset_code"].(string); ok {
				holding.AssetCode = val
			}

			// Parse total_quantity (API returns as string)
			if val, ok := result["total_quantity"].(string); ok {
				if parsed, err := strconv.ParseFloat(val, 64); err == nil {
					holding.TotalQuantity = parsed
				}
			} else if val, ok := result["total_quantity"].(float64); ok {
				holding.TotalQuantity = val
			}

			// Parse quantity_available_for_trading (API returns as string)
			if val, ok := result["quantity_available_for_trading"].(string); ok {
				if parsed, err := strconv.ParseFloat(val, 64); err == nil {
					holding.QuantityAvailableForTrading = parsed
				}
			} else if val, ok := result["quantity_available_for_trading"].(float64); ok {
				holding.QuantityAvailableForTrading = val
			}

			holdings = append(holdings, holding)
		}
		return holdings, nil
	}

	// Fallback to structured decode
	var response PaginatedResponse[CryptoHolding]
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return nil, err
	}

	return response.Results, nil
}

// GetBestBidAsk retrieves best bid/ask prices for cryptocurrencies
func (c *CryptoClient) GetBestBidAsk(symbols []string) ([]BestBidAsk, error) {
	queryParams := ""
	for i, symbol := range symbols {
		if i == 0 {
			queryParams = "?symbol=" + symbol
		} else {
			queryParams += "&symbol=" + symbol
		}
	}

	endpoint := MarketDataURL + "/best_bid_ask/" + queryParams

	resp, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	// Read the response body first for better error handling
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var response PaginatedResponse[BestBidAsk]
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %v. Response body: %s", err, string(body))
	}

	return response.Results, nil
}

// GetCryptoOrders retrieves crypto order history
func (c *CryptoClient) GetCryptoOrders() ([]CryptoOrder, error) {
	resp, err := c.makeRequest("GET", TradingURL+"/orders/", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var response PaginatedResponse[CryptoOrder]
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Results, nil
}

// GetCryptoOrdersWithParams retrieves crypto order history with query parameters
func (c *CryptoClient) GetCryptoOrdersWithParams(limit int) ([]CryptoOrder, error) {
	// Build query with parameters
	endpoint := TradingURL + "/orders/"
	if limit > 0 {
		endpoint += fmt.Sprintf("?limit=%d", limit)
	}

	resp, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	// Read response body for better error handling
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Try to parse the response structure first
	var rawResponse map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &rawResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	// Extract orders from results field
	var orders []CryptoOrder
	if results, ok := rawResponse["results"].([]interface{}); ok {
		for _, result := range results {
			if orderMap, ok := result.(map[string]interface{}); ok {
				order := CryptoOrder{}

				// Parse all fields according to Robinhood API documentation
				if val, ok := orderMap["id"].(string); ok {
					order.ID = val
				}
				if val, ok := orderMap["account_number"].(string); ok {
					order.AccountNumber = val
				}
				if val, ok := orderMap["symbol"].(string); ok {
					order.Symbol = val
				}
				if val, ok := orderMap["client_order_id"].(string); ok {
					order.ClientOrderID = val
				}
				if val, ok := orderMap["side"].(string); ok {
					order.Side = val
				}
				if val, ok := orderMap["type"].(string); ok {
					order.Type = val
				}
				if val, ok := orderMap["state"].(string); ok {
					order.State = val
				}
				if val, ok := orderMap["created_at"].(string); ok {
					order.CreatedAt = val
				}
				if val, ok := orderMap["updated_at"].(string); ok {
					order.UpdatedAt = val
				}

				// Handle average_price (can be string or float)
				if val, ok := orderMap["average_price"].(string); ok {
					if parsed, err := strconv.ParseFloat(val, 64); err == nil {
						order.AveragePrice = parsed
					}
				} else if val, ok := orderMap["average_price"].(float64); ok {
					order.AveragePrice = val
				}

				// Handle filled_asset_quantity (can be string or float)
				if val, ok := orderMap["filled_asset_quantity"].(string); ok {
					if parsed, err := strconv.ParseFloat(val, 64); err == nil {
						order.FilledAssetQuantity = parsed
					}
				} else if val, ok := orderMap["filled_asset_quantity"].(float64); ok {
					order.FilledAssetQuantity = val
				}

				// Also check for executions array if filled_asset_quantity is not present
				if order.FilledAssetQuantity == 0 {
					if executions, ok := orderMap["executions"].([]interface{}); ok && len(executions) > 0 {
						// Sum up executed quantities
						for _, exec := range executions {
							if execMap, ok := exec.(map[string]interface{}); ok {
								if qty, ok := execMap["quantity"].(string); ok {
									if parsed, err := strconv.ParseFloat(qty, 64); err == nil {
										order.FilledAssetQuantity += parsed
									}
								} else if qty, ok := execMap["quantity"].(float64); ok {
									order.FilledAssetQuantity += qty
								}
							}
						}
					}
				}

				// Check for order config fields to extract quantity if filled_asset_quantity is 0
				if order.FilledAssetQuantity == 0 {
					// Check market_order_config
					if marketConfig, ok := orderMap["market_order_config"].(map[string]interface{}); ok {
						if qty, ok := marketConfig["asset_quantity"].(string); ok {
							if parsed, err := strconv.ParseFloat(qty, 64); err == nil {
								order.FilledAssetQuantity = parsed
							}
						} else if qty, ok := marketConfig["asset_quantity"].(float64); ok {
							order.FilledAssetQuantity = qty
						}
					}
					// Check limit_order_config
					if limitConfig, ok := orderMap["limit_order_config"].(map[string]interface{}); ok {
						if qty, ok := limitConfig["asset_quantity"].(string); ok {
							if parsed, err := strconv.ParseFloat(qty, 64); err == nil {
								order.FilledAssetQuantity = parsed
							}
						} else if qty, ok := limitConfig["asset_quantity"].(float64); ok {
							order.FilledAssetQuantity = qty
						}
						// Also get limit price for average price if not set
						if order.AveragePrice == 0 {
							if price, ok := limitConfig["limit_price"].(string); ok {
								if parsed, err := strconv.ParseFloat(price, 64); err == nil {
									order.AveragePrice = parsed
								}
							} else if price, ok := limitConfig["limit_price"].(float64); ok {
								order.AveragePrice = price
							}
						}
					}
				}

				orders = append(orders, order)
			}
		}
	}

	return orders, nil
}

// PlaceCryptoOrder places a new crypto order (legacy)
func (c *CryptoClient) PlaceCryptoOrder(order OrderRequest) (*CryptoOrder, error) {
	resp, err := c.makeRequest("POST", TradingURL+"/orders/", order)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var cryptoOrder CryptoOrder
	if err := json.NewDecoder(resp.Body).Decode(&cryptoOrder); err != nil {
		return nil, err
	}

	return &cryptoOrder, nil
}

// PlaceCryptoOrderNew places a new crypto order using the correct API format
func (c *CryptoClient) PlaceCryptoOrderNew(clientOrderID, side, orderType, symbol, quantity, price string) (*CryptoOrder, error) {
	// Build order request according to Robinhood API docs
	orderRequest := map[string]interface{}{
		"client_order_id": clientOrderID,
		"side":           side,
		"type":           orderType,
		"symbol":         symbol,
	}

	// Add order type specific configuration
	if orderType == "market" {
		orderRequest["market_order_config"] = map[string]interface{}{
			"asset_quantity": quantity,
		}
	} else if orderType == "limit" {
		orderRequest["limit_order_config"] = map[string]interface{}{
			"asset_quantity": quantity,
			"limit_price":    price,
			"time_in_force":  "gtc",
		}
	}

	resp, err := c.makeRequest("POST", TradingURL+"/orders/", orderRequest)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	// Read response body for manual parsing
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse as map first to handle string/float conversion
	var orderMap map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &orderMap); err != nil {
		return nil, err
	}

	// Convert map to CryptoOrder with proper type handling
	cryptoOrder := CryptoOrder{}

	if val, ok := orderMap["id"].(string); ok {
		cryptoOrder.ID = val
	}
	if val, ok := orderMap["account_number"].(string); ok {
		cryptoOrder.AccountNumber = val
	}
	if val, ok := orderMap["symbol"].(string); ok {
		cryptoOrder.Symbol = val
	}
	if val, ok := orderMap["client_order_id"].(string); ok {
		cryptoOrder.ClientOrderID = val
	}
	if val, ok := orderMap["side"].(string); ok {
		cryptoOrder.Side = val
	}
	if val, ok := orderMap["type"].(string); ok {
		cryptoOrder.Type = val
	}
	if val, ok := orderMap["state"].(string); ok {
		cryptoOrder.State = val
	}
	if val, ok := orderMap["created_at"].(string); ok {
		cryptoOrder.CreatedAt = val
	}
	if val, ok := orderMap["updated_at"].(string); ok {
		cryptoOrder.UpdatedAt = val
	}

	// Handle average_price (can be string or float)
	if val, ok := orderMap["average_price"].(string); ok {
		if parsed, err := strconv.ParseFloat(val, 64); err == nil {
			cryptoOrder.AveragePrice = parsed
		}
	} else if val, ok := orderMap["average_price"].(float64); ok {
		cryptoOrder.AveragePrice = val
	}

	// Handle filled_asset_quantity (can be string or float)
	if val, ok := orderMap["filled_asset_quantity"].(string); ok {
		if parsed, err := strconv.ParseFloat(val, 64); err == nil {
			cryptoOrder.FilledAssetQuantity = parsed
		}
	} else if val, ok := orderMap["filled_asset_quantity"].(float64); ok {
		cryptoOrder.FilledAssetQuantity = val
	}

	return &cryptoOrder, nil
}

// CancelCryptoOrder cancels an existing crypto order
func (c *CryptoClient) CancelCryptoOrder(orderID string) error {
	endpoint := fmt.Sprintf("%s/orders/%s/cancel/", TradingURL, orderID)

	resp, err := c.makeRequest("POST", endpoint, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetTradingPairs retrieves list of supported trading pairs
func (c *CryptoClient) GetTradingPairs(symbols []string) ([]map[string]interface{}, error) {
	queryParams := ""
	for i, symbol := range symbols {
		if i == 0 {
			queryParams = "?symbol=" + symbol
		} else {
			queryParams += "&symbol=" + symbol
		}
	}

	endpoint := TradingURL + "/trading_pairs/" + queryParams

	resp, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var response PaginatedResponse[map[string]interface{}]
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Results, nil
}