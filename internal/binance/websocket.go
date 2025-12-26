package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

// WSHandler handles WebSocket messages
type WSHandler interface {
	OnKline(event KlineEvent)
	OnTrade(event TradeEvent)
	OnDepth(event DepthEvent)
	OnMiniTicker(event MiniTickerEvent)
	OnError(err error)
	OnDisconnect()
	OnReconnect()
}

// DefaultWSHandler provides default implementations
type DefaultWSHandler struct{}

func (h *DefaultWSHandler) OnKline(event KlineEvent)           {}
func (h *DefaultWSHandler) OnTrade(event TradeEvent)           {}
func (h *DefaultWSHandler) OnDepth(event DepthEvent)           {}
func (h *DefaultWSHandler) OnMiniTicker(event MiniTickerEvent) {}
func (h *DefaultWSHandler) OnError(err error)                  {}
func (h *DefaultWSHandler) OnDisconnect()                      {}
func (h *DefaultWSHandler) OnReconnect()                       {}

// WSClient is the Binance WebSocket client
type WSClient struct {
	baseURL       string
	conn          *websocket.Conn
	handler       WSHandler
	subscriptions []string
	mu            sync.RWMutex

	// Connection management
	connected     atomic.Bool
	reconnecting  atomic.Bool
	ctx           context.Context
	cancel        context.CancelFunc
	done          chan struct{}

	// Configuration
	pingInterval  time.Duration
	pongTimeout   time.Duration
	reconnectWait time.Duration
	maxReconnects int
	testnet       bool
}

// WSClientOption configures the WebSocket client
type WSClientOption func(*WSClient)

// WithWSTestnet enables testnet mode
func WithWSTestnet(enabled bool) WSClientOption {
	return func(c *WSClient) {
		c.testnet = enabled
		if enabled {
			c.baseURL = WSBaseURLTestnet
		}
	}
}

// WithPingInterval sets the ping interval
func WithPingInterval(d time.Duration) WSClientOption {
	return func(c *WSClient) {
		c.pingInterval = d
	}
}

// WithReconnectWait sets the reconnect wait time
func WithReconnectWait(d time.Duration) WSClientOption {
	return func(c *WSClient) {
		c.reconnectWait = d
	}
}

// WithMaxReconnects sets maximum reconnection attempts
func WithMaxReconnects(n int) WSClientOption {
	return func(c *WSClient) {
		c.maxReconnects = n
	}
}

// NewWSClient creates a new WebSocket client
func NewWSClient(handler WSHandler, opts ...WSClientOption) *WSClient {
	if handler == nil {
		handler = &DefaultWSHandler{}
	}

	c := &WSClient{
		baseURL:       WSBaseURLSpot,
		handler:       handler,
		subscriptions: make([]string, 0),
		done:          make(chan struct{}),
		pingInterval:  30 * time.Second,
		pongTimeout:   10 * time.Second,
		reconnectWait: 5 * time.Second,
		maxReconnects: 10,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Connect establishes WebSocket connection
func (c *WSClient) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected.Load() {
		return nil
	}

	c.ctx, c.cancel = context.WithCancel(ctx)

	if err := c.connect(); err != nil {
		return err
	}

	// Start message reader
	go c.readLoop()

	// Start ping/pong handler
	go c.pingLoop()

	c.connected.Store(true)
	log.Info().Str("url", c.baseURL).Msg("WebSocket connected")

	return nil
}

// connect performs the actual connection
func (c *WSClient) connect() error {
	var url string
	if len(c.subscriptions) > 0 {
		// Use combined streams endpoint for multiple subscriptions
		url = strings.Replace(c.baseURL, "/ws", "/stream?streams=", 1) + strings.Join(c.subscriptions, "/")
	} else {
		url = c.baseURL
	}

	log.Debug().Str("url", url).Msg("Connecting to Binance WebSocket")

	dialer := websocket.Dialer{
		HandshakeTimeout: 15 * time.Second,
	}

	conn, resp, err := dialer.DialContext(c.ctx, url, nil)
	if err != nil {
		if resp != nil {
			log.Error().Int("status", resp.StatusCode).Msg("WebSocket handshake failed")
		}
		return fmt.Errorf("failed to connect: %w", err)
	}

	c.conn = conn
	return nil
}

// Disconnect closes the WebSocket connection
func (c *WSClient) Disconnect() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected.Load() {
		return
	}

	c.cancel()
	c.connected.Store(false)

	if c.conn != nil {
		c.conn.Close()
	}

	close(c.done)
	log.Info().Msg("WebSocket disconnected")
}

// Subscribe adds subscriptions
func (c *WSClient) Subscribe(streams ...string) error {
	c.mu.Lock()
	c.subscriptions = append(c.subscriptions, streams...)
	c.mu.Unlock()

	if !c.connected.Load() {
		return nil
	}

	// Send subscription message
	msg := WSSubscription{
		Method: "SUBSCRIBE",
		Params: streams,
		ID:     int(time.Now().Unix()),
	}

	return c.sendJSON(msg)
}

// Unsubscribe removes subscriptions
func (c *WSClient) Unsubscribe(streams ...string) error {
	c.mu.Lock()
	// Remove from subscriptions list
	newSubs := make([]string, 0)
	for _, s := range c.subscriptions {
		found := false
		for _, stream := range streams {
			if s == stream {
				found = true
				break
			}
		}
		if !found {
			newSubs = append(newSubs, s)
		}
	}
	c.subscriptions = newSubs
	c.mu.Unlock()

	if !c.connected.Load() {
		return nil
	}

	msg := WSSubscription{
		Method: "UNSUBSCRIBE",
		Params: streams,
		ID:     int(time.Now().Unix()),
	}

	return c.sendJSON(msg)
}

// SubscribeKline subscribes to kline stream
func (c *WSClient) SubscribeKline(symbol, interval string) error {
	stream := fmt.Sprintf("%s@kline_%s", strings.ToLower(symbol), interval)
	return c.Subscribe(stream)
}

// SubscribeTrade subscribes to trade stream
func (c *WSClient) SubscribeTrade(symbol string) error {
	stream := fmt.Sprintf("%s@trade", strings.ToLower(symbol))
	return c.Subscribe(stream)
}

// SubscribeDepth subscribes to order book stream
func (c *WSClient) SubscribeDepth(symbol string, levels int) error {
	stream := fmt.Sprintf("%s@depth%d@100ms", strings.ToLower(symbol), levels)
	return c.Subscribe(stream)
}

// SubscribeMiniTicker subscribes to mini ticker stream
func (c *WSClient) SubscribeMiniTicker(symbol string) error {
	stream := fmt.Sprintf("%s@miniTicker", strings.ToLower(symbol))
	return c.Subscribe(stream)
}

// SubscribeAllMiniTickers subscribes to all mini tickers
func (c *WSClient) SubscribeAllMiniTickers() error {
	return c.Subscribe("!miniTicker@arr")
}

// sendJSON sends JSON message
func (c *WSClient) sendJSON(v interface{}) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.conn == nil {
		return fmt.Errorf("not connected")
	}

	return c.conn.WriteJSON(v)
}

// readLoop reads messages from WebSocket
func (c *WSClient) readLoop() {
	defer func() {
		c.connected.Store(false)
		c.handler.OnDisconnect()
		c.reconnect()
	}()

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				log.Debug().Msg("WebSocket closed normally")
				return
			}
			if c.ctx.Err() == nil {
				c.handler.OnError(fmt.Errorf("read error: %w", err))
			}
			return
		}

		c.handleMessage(message)
	}
}

// handleMessage processes incoming WebSocket message
func (c *WSClient) handleMessage(data []byte) {
	// Try to detect message type
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		// Try array (for combined streams)
		var arr []interface{}
		if err := json.Unmarshal(data, &arr); err != nil {
			c.handler.OnError(fmt.Errorf("failed to parse message: %w", err))
			return
		}
		// Handle array messages
		for _, item := range arr {
			itemData, _ := json.Marshal(item)
			c.handleMessage(itemData)
		}
		return
	}

	// Check if it's a combined stream message (has "stream" and "data" fields)
	if stream, ok := raw["stream"].(string); ok {
		if dataObj, ok := raw["data"].(map[string]interface{}); ok {
			// Re-marshal the data object
			data, _ = json.Marshal(dataObj)
			// Re-parse into raw for event type detection
			json.Unmarshal(data, &raw)
		}
		_ = stream // Stream name is useful for debugging
	}

	eventType, ok := raw["e"].(string)
	if !ok {
		// Might be a subscription response
		if _, ok := raw["result"]; ok {
			log.Debug().Interface("response", raw).Msg("Subscription response")
			return
		}
		return
	}

	switch eventType {
	case "kline":
		var event KlineEvent
		if err := json.Unmarshal(data, &event); err != nil {
			c.handler.OnError(fmt.Errorf("failed to parse kline: %w", err))
			return
		}
		c.handler.OnKline(event)

	case "trade":
		var event TradeEvent
		if err := json.Unmarshal(data, &event); err != nil {
			c.handler.OnError(fmt.Errorf("failed to parse trade: %w", err))
			return
		}
		c.handler.OnTrade(event)

	case "depthUpdate":
		var event DepthEvent
		if err := json.Unmarshal(data, &event); err != nil {
			c.handler.OnError(fmt.Errorf("failed to parse depth: %w", err))
			return
		}
		c.handler.OnDepth(event)

	case "24hrMiniTicker":
		var event MiniTickerEvent
		if err := json.Unmarshal(data, &event); err != nil {
			c.handler.OnError(fmt.Errorf("failed to parse mini ticker: %w", err))
			return
		}
		c.handler.OnMiniTicker(event)

	default:
		log.Debug().Str("event", eventType).Msg("Unknown event type")
	}
}

// detectEventType detects event type from stream name
func (c *WSClient) detectEventType(stream string) string {
	parts := strings.Split(stream, "@")
	if len(parts) < 2 {
		return ""
	}

	streamType := parts[1]
	if strings.HasPrefix(streamType, "kline") {
		return "kline"
	}
	if streamType == "trade" {
		return "trade"
	}
	if strings.HasPrefix(streamType, "depth") {
		return "depthUpdate"
	}
	if streamType == "miniTicker" || streamType == "miniTicker@arr" {
		return "24hrMiniTicker"
	}
	return streamType
}

// pingLoop sends periodic pings
func (c *WSClient) pingLoop() {
	ticker := time.NewTicker(c.pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.mu.RLock()
			if c.conn == nil {
				c.mu.RUnlock()
				continue
			}

			if err := c.conn.WriteControl(
				websocket.PingMessage,
				[]byte{},
				time.Now().Add(c.pongTimeout),
			); err != nil {
				c.mu.RUnlock()
				c.handler.OnError(fmt.Errorf("ping failed: %w", err))
				continue
			}
			c.mu.RUnlock()
		}
	}
}

// reconnect attempts to reconnect
func (c *WSClient) reconnect() {
	if c.reconnecting.Load() || c.ctx.Err() != nil {
		return
	}

	c.reconnecting.Store(true)
	defer c.reconnecting.Store(false)

	for i := 0; i < c.maxReconnects; i++ {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		log.Info().Int("attempt", i+1).Msg("Attempting to reconnect")

		time.Sleep(c.reconnectWait)

		c.mu.Lock()
		if err := c.connect(); err != nil {
			c.mu.Unlock()
			log.Error().Err(err).Msg("Reconnection failed")
			continue
		}

		c.connected.Store(true)
		c.mu.Unlock()

		// Resubscribe
		c.mu.RLock()
		subs := make([]string, len(c.subscriptions))
		copy(subs, c.subscriptions)
		c.mu.RUnlock()

		if len(subs) > 0 {
			msg := WSSubscription{
				Method: "SUBSCRIBE",
				Params: subs,
				ID:     int(time.Now().Unix()),
			}
			if err := c.sendJSON(msg); err != nil {
				log.Error().Err(err).Msg("Failed to resubscribe")
			}
		}

		c.handler.OnReconnect()
		log.Info().Msg("Reconnected successfully")

		// Start new read loop
		go c.readLoop()
		return
	}

	log.Error().Int("attempts", c.maxReconnects).Msg("Failed to reconnect after max attempts")
}

// IsConnected returns connection status
func (c *WSClient) IsConnected() bool {
	return c.connected.Load()
}

// GetSubscriptions returns current subscriptions
func (c *WSClient) GetSubscriptions() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	subs := make([]string, len(c.subscriptions))
	copy(subs, c.subscriptions)
	return subs
}

// KlineHandler is a simplified handler for kline-only subscriptions
type KlineHandler struct {
	DefaultWSHandler
	OnKlineFunc func(event KlineEvent)
}

func (h *KlineHandler) OnKline(event KlineEvent) {
	if h.OnKlineFunc != nil {
		h.OnKlineFunc(event)
	}
}

// NewKlineWSClient creates a WebSocket client configured for kline streams
func NewKlineWSClient(symbol string, intervals []string, onKline func(KlineEvent), opts ...WSClientOption) *WSClient {
	handler := &KlineHandler{
		OnKlineFunc: onKline,
	}

	client := NewWSClient(handler, opts...)

	// Prepare subscriptions
	for _, interval := range intervals {
		stream := fmt.Sprintf("%s@kline_%s", strings.ToLower(symbol), interval)
		client.subscriptions = append(client.subscriptions, stream)
	}

	return client
}

// MultiSymbolKlineHandler handles klines for multiple symbols
type MultiSymbolKlineHandler struct {
	DefaultWSHandler
	OnKlineFunc func(symbol string, event KlineEvent)
}

func (h *MultiSymbolKlineHandler) OnKline(event KlineEvent) {
	if h.OnKlineFunc != nil {
		h.OnKlineFunc(event.Symbol, event)
	}
}

// NewMultiSymbolKlineWSClient creates a WebSocket client for multiple symbols
func NewMultiSymbolKlineWSClient(symbols []string, intervals []string, onKline func(string, KlineEvent), opts ...WSClientOption) *WSClient {
	handler := &MultiSymbolKlineHandler{
		OnKlineFunc: onKline,
	}

	client := NewWSClient(handler, opts...)

	// Prepare subscriptions
	for _, symbol := range symbols {
		for _, interval := range intervals {
			stream := fmt.Sprintf("%s@kline_%s", strings.ToLower(symbol), interval)
			client.subscriptions = append(client.subscriptions, stream)
		}
	}

	return client
}
