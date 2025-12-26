package binance

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// Client is the Binance REST API client
type Client struct {
	apiKey     string
	secretKey  string
	baseURL    string
	httpClient *http.Client
	testnet    bool
}

// ClientOption configures the client
type ClientOption func(*Client)

// WithTestnet enables testnet mode
func WithTestnet(enabled bool) ClientOption {
	return func(c *Client) {
		c.testnet = enabled
		if enabled {
			c.baseURL = BaseURLTestnet
		}
	}
}

// WithHTTPClient sets custom HTTP client
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = client
	}
}

// WithBaseURL sets custom base URL
func WithBaseURL(url string) ClientOption {
	return func(c *Client) {
		c.baseURL = url
	}
}

// Config holds client configuration
type Config struct {
	APIKey    string
	SecretKey string
	Testnet   bool
	Timeout   time.Duration
}

// NewClient creates a new Binance client
func NewClient(cfg *Config, opts ...ClientOption) *Client {
	apiKey := ""
	secretKey := ""
	baseURL := BaseURLSpot
	timeout := 30 * time.Second

	if cfg != nil {
		apiKey = cfg.APIKey
		secretKey = cfg.SecretKey
		if cfg.Testnet {
			baseURL = BaseURLTestnet
		}
		if cfg.Timeout > 0 {
			timeout = cfg.Timeout
		}
	}

	c := &Client{
		apiKey:    apiKey,
		secretKey: secretKey,
		baseURL:   baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// sign creates HMAC SHA256 signature
func (c *Client) sign(queryString string) string {
	h := hmac.New(sha256.New, []byte(c.secretKey))
	h.Write([]byte(queryString))
	return hex.EncodeToString(h.Sum(nil))
}

// doRequest performs HTTP request
func (c *Client) doRequest(method, endpoint string, params url.Values, signed bool) ([]byte, error) {
	var reqBody io.Reader
	fullURL := c.baseURL + endpoint

	if signed {
		if params == nil {
			params = url.Values{}
		}
		params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
		params.Set("signature", c.sign(params.Encode()))
	}

	if method == http.MethodGet && params != nil {
		fullURL += "?" + params.Encode()
	} else if params != nil {
		reqBody = strings.NewReader(params.Encode())
	}

	req, err := http.NewRequest(method, fullURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-MBX-APIKEY", c.apiKey)
	if method == http.MethodPost || method == http.MethodPut || method == http.MethodDelete {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var apiErr APIError
		if err := json.Unmarshal(body, &apiErr); err != nil {
			return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
		}
		return nil, &apiErr
	}

	return body, nil
}

// Ping tests connectivity
func (c *Client) Ping() error {
	_, err := c.doRequest(http.MethodGet, EndpointPing, nil, false)
	return err
}

// GetServerTime returns server time
func (c *Client) GetServerTime() (*ServerTime, error) {
	data, err := c.doRequest(http.MethodGet, EndpointTime, nil, false)
	if err != nil {
		return nil, err
	}

	var result ServerTime
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &result, nil
}

// GetExchangeInfo returns exchange information
func (c *Client) GetExchangeInfo() (*ExchangeInfo, error) {
	data, err := c.doRequest(http.MethodGet, EndpointExchangeInfo, nil, false)
	if err != nil {
		return nil, err
	}

	var result ExchangeInfo
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &result, nil
}

// GetSymbolInfo returns info for a specific symbol with parsed filters
func (c *Client) GetSymbolInfo(symbol string) (*SymbolInfo, error) {
	info, err := c.GetExchangeInfo()
	if err != nil {
		return nil, err
	}

	for _, s := range info.Symbols {
		if s.Symbol == symbol {
			// Parse filters
			for _, f := range s.Filters {
				switch f.FilterType {
				case "PRICE_FILTER":
					s.MinPrice, _ = strconv.ParseFloat(f.MinPrice, 64)
					s.MaxPrice, _ = strconv.ParseFloat(f.MaxPrice, 64)
					s.TickSize, _ = strconv.ParseFloat(f.TickSize, 64)
					// Calculate price precision from tick size
					if s.TickSize > 0 {
						s.PricePrecision = countDecimals(f.TickSize)
					}
				case "LOT_SIZE":
					s.MinQty, _ = strconv.ParseFloat(f.MinQty, 64)
					s.MaxQty, _ = strconv.ParseFloat(f.MaxQty, 64)
					s.StepSize, _ = strconv.ParseFloat(f.StepSize, 64)
					// Calculate quantity precision from step size
					if s.StepSize > 0 {
						s.QuantityPrecision = countDecimals(f.StepSize)
					}
				case "MIN_NOTIONAL", "NOTIONAL":
					s.MinNotional, _ = strconv.ParseFloat(f.MinNotional, 64)
				}
			}
			return &s, nil
		}
	}
	return nil, fmt.Errorf("symbol %s not found", symbol)
}

// countDecimals counts decimal places in a string number
func countDecimals(s string) int {
	idx := strings.Index(s, ".")
	if idx < 0 {
		return 0
	}
	// Count non-zero digits after decimal point
	decimals := s[idx+1:]
	count := len(decimals)
	// Trim trailing zeros
	for i := len(decimals) - 1; i >= 0; i-- {
		if decimals[i] != '0' {
			break
		}
		count--
	}
	return count
}

// GetKlines returns candlestick data
func (c *Client) GetKlines(symbol, interval string, limit int, startTime, endTime int64) ([]Kline, error) {
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("interval", interval)

	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	if startTime > 0 {
		params.Set("startTime", strconv.FormatInt(startTime, 10))
	}
	if endTime > 0 {
		params.Set("endTime", strconv.FormatInt(endTime, 10))
	}

	data, err := c.doRequest(http.MethodGet, EndpointKlines, params, false)
	if err != nil {
		return nil, err
	}

	var rawKlines [][]interface{}
	if err := json.Unmarshal(data, &rawKlines); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	klines := make([]Kline, len(rawKlines))
	for i, raw := range rawKlines {
		if len(raw) < 11 {
			continue
		}
		klines[i] = Kline{
			OpenTime:                 int64(raw[0].(float64)),
			Open:                     raw[1].(string),
			High:                     raw[2].(string),
			Low:                      raw[3].(string),
			Close:                    raw[4].(string),
			Volume:                   raw[5].(string),
			CloseTime:                int64(raw[6].(float64)),
			QuoteAssetVolume:         raw[7].(string),
			NumberOfTrades:           int64(raw[8].(float64)),
			TakerBuyBaseAssetVolume:  raw[9].(string),
			TakerBuyQuoteAssetVolume: raw[10].(string),
		}
	}

	return klines, nil
}

// GetTicker24hr returns 24hr price change statistics
func (c *Client) GetTicker24hr(symbol string) (*Ticker24hr, error) {
	params := url.Values{}
	params.Set("symbol", symbol)

	data, err := c.doRequest(http.MethodGet, EndpointTicker24hr, params, false)
	if err != nil {
		return nil, err
	}

	var result Ticker24hr
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &result, nil
}

// GetTickerPrice returns current price
func (c *Client) GetTickerPrice(symbol string) (*TickerPrice, error) {
	params := url.Values{}
	params.Set("symbol", symbol)

	data, err := c.doRequest(http.MethodGet, EndpointTickerPrice, params, false)
	if err != nil {
		return nil, err
	}

	var result TickerPrice
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &result, nil
}

// GetDepth returns order book depth
func (c *Client) GetDepth(symbol string, limit int) (*Depth, error) {
	params := url.Values{}
	params.Set("symbol", symbol)
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}

	data, err := c.doRequest(http.MethodGet, EndpointDepth, params, false)
	if err != nil {
		return nil, err
	}

	var result Depth
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &result, nil
}

// GetAccount returns account information (requires signature)
func (c *Client) GetAccount() (*Account, error) {
	data, err := c.doRequest(http.MethodGet, EndpointAccount, nil, true)
	if err != nil {
		return nil, err
	}

	var result Account
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Parse balance strings to floats
	for i := range result.Balances {
		result.Balances[i].Free, _ = strconv.ParseFloat(result.Balances[i].FreeStr, 64)
		result.Balances[i].Locked, _ = strconv.ParseFloat(result.Balances[i].LockedStr, 64)
	}

	return &result, nil
}

// GetBalance returns balance for a specific asset
func (c *Client) GetBalance(asset string) (*Balance, error) {
	account, err := c.GetAccount()
	if err != nil {
		return nil, err
	}

	for _, b := range account.Balances {
		if b.Asset == asset {
			return &b, nil
		}
	}
	return nil, fmt.Errorf("asset %s not found", asset)
}

// GetMyTrades returns account trades
func (c *Client) GetMyTrades(symbol string, limit int) ([]Trade, error) {
	params := url.Values{}
	params.Set("symbol", symbol)
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}

	data, err := c.doRequest(http.MethodGet, EndpointMyTrades, params, true)
	if err != nil {
		return nil, err
	}

	var result []Trade
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return result, nil
}

// CreateOrder creates a new order
func (c *Client) CreateOrder(req OrderRequest) (*Order, error) {
	params := url.Values{}
	params.Set("symbol", req.Symbol)
	params.Set("side", string(req.Side))
	params.Set("type", string(req.Type))

	if req.Quantity > 0 {
		params.Set("quantity", strconv.FormatFloat(req.Quantity, 'f', -1, 64))
	}

	switch req.Type {
	case OrderTypeLimit:
		params.Set("timeInForce", string(req.TimeInForce))
		params.Set("price", strconv.FormatFloat(req.Price, 'f', -1, 64))
	case OrderTypeStopLoss, OrderTypeTakeProfit:
		params.Set("stopPrice", strconv.FormatFloat(req.StopPrice, 'f', -1, 64))
	case OrderTypeStopLossLimit, OrderTypeTakeProfitLimit:
		params.Set("timeInForce", string(req.TimeInForce))
		params.Set("price", strconv.FormatFloat(req.Price, 'f', -1, 64))
		params.Set("stopPrice", strconv.FormatFloat(req.StopPrice, 'f', -1, 64))
	}

	if req.NewClientOrderID != "" {
		params.Set("newClientOrderId", req.NewClientOrderID)
	}

	data, err := c.doRequest(http.MethodPost, EndpointOrder, params, true)
	if err != nil {
		return nil, err
	}

	var result Order
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	log.Info().
		Str("symbol", req.Symbol).
		Str("side", string(req.Side)).
		Str("type", string(req.Type)).
		Float64("quantity", req.Quantity).
		Int64("orderID", result.OrderID).
		Msg("Order created")

	return &result, nil
}

// CreateMarketOrder creates a market order
func (c *Client) CreateMarketOrder(symbol string, side OrderSide, quantity float64) (*Order, error) {
	return c.CreateOrder(OrderRequest{
		Symbol:   symbol,
		Side:     side,
		Type:     OrderTypeMarket,
		Quantity: quantity,
	})
}

// CreateLimitOrder creates a limit order
func (c *Client) CreateLimitOrder(symbol string, side OrderSide, quantity, price float64) (*Order, error) {
	return c.CreateOrder(OrderRequest{
		Symbol:      symbol,
		Side:        side,
		Type:        OrderTypeLimit,
		TimeInForce: TimeInForceGTC,
		Quantity:    quantity,
		Price:       price,
	})
}

// GetOrder returns order status
func (c *Client) GetOrder(symbol string, orderID int64) (*Order, error) {
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("orderId", strconv.FormatInt(orderID, 10))

	data, err := c.doRequest(http.MethodGet, EndpointOrder, params, true)
	if err != nil {
		return nil, err
	}

	var result Order
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &result, nil
}

// CancelOrder cancels an order
func (c *Client) CancelOrder(symbol string, orderID int64) (*Order, error) {
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("orderId", strconv.FormatInt(orderID, 10))

	data, err := c.doRequest(http.MethodDelete, EndpointOrder, params, true)
	if err != nil {
		return nil, err
	}

	var result Order
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	log.Info().
		Str("symbol", symbol).
		Int64("orderID", orderID).
		Msg("Order canceled")

	return &result, nil
}

// GetOpenOrders returns all open orders
func (c *Client) GetOpenOrders(symbol string) ([]Order, error) {
	params := url.Values{}
	if symbol != "" {
		params.Set("symbol", symbol)
	}

	data, err := c.doRequest(http.MethodGet, EndpointOpenOrders, params, true)
	if err != nil {
		return nil, err
	}

	var result []Order
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return result, nil
}

// GetAllOrders returns all orders (requires symbol)
func (c *Client) GetAllOrders(symbol string, limit int) ([]Order, error) {
	params := url.Values{}
	params.Set("symbol", symbol)
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}

	data, err := c.doRequest(http.MethodGet, EndpointAllOrders, params, true)
	if err != nil {
		return nil, err
	}

	var result []Order
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return result, nil
}

// CancelAllOpenOrders cancels all open orders for a symbol
func (c *Client) CancelAllOpenOrders(symbol string) ([]Order, error) {
	orders, err := c.GetOpenOrders(symbol)
	if err != nil {
		return nil, err
	}

	var canceled []Order
	for _, order := range orders {
		result, err := c.CancelOrder(symbol, order.OrderID)
		if err != nil {
			log.Error().Err(err).Int64("orderID", order.OrderID).Msg("Failed to cancel order")
			continue
		}
		canceled = append(canceled, *result)
	}

	return canceled, nil
}

// GetHistoricalKlines fetches historical klines with pagination
func (c *Client) GetHistoricalKlines(symbol, interval string, startTime, endTime time.Time) ([]Kline, error) {
	var allKlines []Kline
	limit := 1000 // max limit per request
	start := startTime.UnixMilli()
	end := endTime.UnixMilli()

	for start < end {
		klines, err := c.GetKlines(symbol, interval, limit, start, end)
		if err != nil {
			return nil, err
		}

		if len(klines) == 0 {
			break
		}

		allKlines = append(allKlines, klines...)

		// Move start to after the last kline
		lastKline := klines[len(klines)-1]
		start = lastKline.CloseTime + 1

		// If we got less than limit, we've reached the end
		if len(klines) < limit {
			break
		}

		// Small delay to avoid rate limiting
		time.Sleep(100 * time.Millisecond)
	}

	log.Debug().
		Str("symbol", symbol).
		Str("interval", interval).
		Int("count", len(allKlines)).
		Msg("Fetched historical klines")

	return allKlines, nil
}

// GetTicker returns simple ticker with last price
func (c *Client) GetTicker(symbol string) (*SimpleTicker, error) {
	params := url.Values{}
	params.Set("symbol", symbol)

	data, err := c.doRequest(http.MethodGet, EndpointTickerPrice, params, false)
	if err != nil {
		return nil, err
	}

	var rawResult struct {
		Symbol string `json:"symbol"`
		Price  string `json:"price"`
	}
	if err := json.Unmarshal(data, &rawResult); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	price, _ := strconv.ParseFloat(rawResult.Price, 64)
	return &SimpleTicker{
		Symbol:    rawResult.Symbol,
		LastPrice: price,
	}, nil
}

// PlaceOrder places an order and returns full response with fills
func (c *Client) PlaceOrder(req *OrderRequest) (*OrderResponse, error) {
	params := url.Values{}
	params.Set("symbol", req.Symbol)
	params.Set("side", string(req.Side))
	params.Set("type", string(req.Type))

	if req.Quantity > 0 {
		params.Set("quantity", strconv.FormatFloat(req.Quantity, 'f', -1, 64))
	}

	if req.Price > 0 {
		params.Set("price", strconv.FormatFloat(req.Price, 'f', -1, 64))
	}

	if req.StopPrice > 0 {
		params.Set("stopPrice", strconv.FormatFloat(req.StopPrice, 'f', -1, 64))
	}

	if req.TimeInForce != "" {
		params.Set("timeInForce", string(req.TimeInForce))
	}

	if req.NewClientOrderID != "" {
		params.Set("newClientOrderId", req.NewClientOrderID)
	}

	// Request full response with fills
	params.Set("newOrderRespType", "FULL")

	data, err := c.doRequest(http.MethodPost, EndpointOrder, params, true)
	if err != nil {
		return nil, err
	}

	var result OrderResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	log.Info().
		Str("symbol", req.Symbol).
		Str("side", string(req.Side)).
		Str("type", string(req.Type)).
		Float64("quantity", req.Quantity).
		Int64("orderID", result.OrderID).
		Str("status", result.Status).
		Msg("Order placed")

	return &result, nil
}

// GetListenKey creates a new user data stream listen key
func (c *Client) GetListenKey() (string, error) {
	data, err := c.doRequest(http.MethodPost, EndpointUserDataStream, nil, false)
	if err != nil {
		return "", err
	}

	var result ListenKeyResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return result.ListenKey, nil
}

// KeepAliveListenKey keeps a listen key alive
func (c *Client) KeepAliveListenKey(listenKey string) error {
	params := url.Values{}
	params.Set("listenKey", listenKey)

	_, err := c.doRequest(http.MethodPut, EndpointUserDataStream, params, false)
	return err
}

// CloseListenKey closes a user data stream
func (c *Client) CloseListenKey(listenKey string) error {
	params := url.Values{}
	params.Set("listenKey", listenKey)

	_, err := c.doRequest(http.MethodDelete, EndpointUserDataStream, params, false)
	return err
}

