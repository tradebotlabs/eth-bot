package execution

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/eth-trading/internal/binance"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// LiveExecutor executes orders on real Binance exchange
type LiveExecutor struct {
	config    *ExecutorConfig
	client    *binance.Client
	wsClient  *binance.WSClient

	// State
	orders    map[string]*Order
	positions map[string]*Position
	balances  map[string]struct {
		Free   float64
		Locked float64
	}

	// Position ID counter
	nextPositionID int64

	// Symbol info cache
	symbolInfo map[string]*binance.SymbolInfo

	// Callbacks
	onFill     func(FillEvent)
	onPosition func(PositionEvent)

	// Sync
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
	syncTicker *time.Ticker
}

// NewLiveExecutor creates a new live executor
func NewLiveExecutor(config *ExecutorConfig) (*LiveExecutor, error) {
	if config == nil {
		return nil, fmt.Errorf("config is required")
	}

	// Create Binance client
	client := binance.NewClient(&binance.Config{
		APIKey:    config.APIKey,
		SecretKey: config.SecretKey,
		Testnet:   config.Testnet,
		Timeout:   30 * time.Second,
	})

	// Test connection
	if err := client.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to Binance: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	executor := &LiveExecutor{
		config:         config,
		client:         client,
		orders:         make(map[string]*Order),
		positions:      make(map[string]*Position),
		balances:       make(map[string]struct{ Free, Locked float64 }),
		symbolInfo:     make(map[string]*binance.SymbolInfo),
		nextPositionID: 1,
		ctx:            ctx,
		cancel:         cancel,
	}

	// Initial sync
	if err := executor.Sync(); err != nil {
		cancel()
		return nil, fmt.Errorf("initial sync failed: %w", err)
	}

	// Start periodic sync
	executor.syncTicker = time.NewTicker(30 * time.Second)
	go executor.periodicSync()

	log.Info().
		Bool("testnet", config.Testnet).
		Str("symbol", config.Symbol).
		Msg("Live executor initialized")

	return executor, nil
}

// GetMode returns execution mode
func (e *LiveExecutor) GetMode() ExecutionMode {
	return ModeLive
}

// SetOnFill sets fill event callback
func (e *LiveExecutor) SetOnFill(fn func(FillEvent)) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.onFill = fn
}

// SetOnPosition sets position event callback
func (e *LiveExecutor) SetOnPosition(fn func(PositionEvent)) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.onPosition = fn
}

// PlaceOrder places a new order on Binance
func (e *LiveExecutor) PlaceOrder(order *Order) (*ExecutionResult, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	startTime := time.Now()

	// Generate client order ID
	if order.ClientID == "" {
		order.ClientID = uuid.New().String()
	}

	// Get symbol info for precision
	info, err := e.getSymbolInfo(order.Symbol)
	if err != nil {
		return &ExecutionResult{
			Success: false,
			Error:   err,
			Message: "Failed to get symbol info",
			Latency: time.Since(startTime),
		}, err
	}

	// Round quantity to valid precision
	quantity := roundToStepSize(order.Quantity, info.StepSize, info.QuantityPrecision)

	// Check minimum notional
	notional := quantity * order.Price
	if order.Type == OrderTypeMarket {
		// For market orders, estimate with current price
		ticker, err := e.client.GetTicker(order.Symbol)
		if err == nil {
			notional = quantity * ticker.LastPrice
		}
	}
	if notional < info.MinNotional {
		err := fmt.Errorf("order value %.2f below minimum %.2f", notional, info.MinNotional)
		return &ExecutionResult{
			Success: false,
			Error:   err,
			Message: err.Error(),
			Latency: time.Since(startTime),
		}, err
	}

	// Build order request
	req := &binance.OrderRequest{
		Symbol:           order.Symbol,
		Side:             toBinanceSide(order.Side),
		Type:             toBinanceOrderType(order.Type),
		Quantity:         quantity,
		NewClientOrderID: order.ClientID,
	}

	// Set price for limit orders
	if order.Type == OrderTypeLimit {
		req.Price = roundToTickSize(order.Price, info.TickSize, info.PricePrecision)
		req.TimeInForce = binance.TimeInForceGTC
	}

	// Set stop price for stop orders
	if order.Type == OrderTypeStopLoss || order.Type == OrderTypeTakeProfit {
		req.StopPrice = roundToTickSize(order.StopPrice, info.TickSize, info.PricePrecision)
		req.Type = binance.OrderTypeStopLossLimit
		req.Price = roundToTickSize(order.Price, info.TickSize, info.PricePrecision)
		req.TimeInForce = binance.TimeInForceGTC
	}

	// Place order on Binance
	binanceOrder, err := e.client.PlaceOrder(req)
	if err != nil {
		return &ExecutionResult{
			Success: false,
			Error:   err,
			Message: fmt.Sprintf("Failed to place order: %v", err),
			Latency: time.Since(startTime),
		}, err
	}

	// Update order with exchange response
	order.ID = fmt.Sprintf("%d", binanceOrder.OrderID)
	order.Status = mapOrderStatus(binanceOrder.Status)
	order.FilledQuantity = binanceOrder.ExecutedQty
	order.CreatedAt = time.UnixMilli(binanceOrder.TransactTime)
	order.UpdatedAt = time.Now()

	// Calculate average fill price from fills
	if len(binanceOrder.Fills) > 0 {
		var totalValue, totalQty, totalCommission float64
		for _, fill := range binanceOrder.Fills {
			totalValue += fill.Price * fill.Qty
			totalQty += fill.Qty
			totalCommission += fill.Commission
		}
		if totalQty > 0 {
			order.AvgFillPrice = totalValue / totalQty
		}
		order.Commission = totalCommission
		if len(binanceOrder.Fills) > 0 {
			order.CommissionAsset = binanceOrder.Fills[0].CommissionAsset
		}
	}

	// Store order
	e.orders[order.ID] = order

	result := &ExecutionResult{
		Success: true,
		Order:   order,
		Message: fmt.Sprintf("Order %s placed successfully", order.ID),
		Latency: time.Since(startTime),
	}

	// Handle filled orders
	if order.Status == OrderStatusFilled {
		order.FilledAt = time.Now()
		result.Trade, result.Position = e.handleFill(order)
	}

	log.Info().
		Str("orderID", order.ID).
		Str("symbol", order.Symbol).
		Str("side", string(order.Side)).
		Str("type", string(order.Type)).
		Float64("quantity", order.Quantity).
		Str("status", string(order.Status)).
		Dur("latency", result.Latency).
		Msg("Order placed on Binance")

	return result, nil
}

// handleFill processes a filled order
func (e *LiveExecutor) handleFill(order *Order) (*Trade, *Position) {
	// Create trade record
	trade := &Trade{
		ID:              uuid.New().String(),
		OrderID:         order.ID,
		Symbol:          order.Symbol,
		Side:            order.Side,
		Quantity:        order.FilledQuantity,
		Price:           order.AvgFillPrice,
		Commission:      order.Commission,
		CommissionAsset: order.CommissionAsset,
		Strategy:        order.Strategy,
		ExecutedAt:      time.Now(),
	}

	// Check for existing position
	position, exists := e.positions[order.Symbol]

	if !exists {
		// Open new position
		side := PositionSideLong
		if order.Side == OrderSideSell {
			side = PositionSideShort
		}

		position = &Position{
			ID:           e.nextPositionID,
			Symbol:       order.Symbol,
			Side:         side,
			Quantity:     order.FilledQuantity,
			EntryPrice:   order.AvgFillPrice,
			CurrentPrice: order.AvgFillPrice,
			Commission:   order.Commission,
			Strategy:     order.Strategy,
			OpenTime:     time.Now(),
			UpdatedAt:    time.Now(),
			Orders:       []string{order.ID},
		}
		e.nextPositionID++
		e.positions[order.Symbol] = position
		trade.PositionID = position.ID

		e.emitPositionEvent(PositionEventOpened, position, trade)
	} else {
		// Update existing position
		trade.PositionID = position.ID

		isClosing := (position.Side == PositionSideLong && order.Side == OrderSideSell) ||
			(position.Side == PositionSideShort && order.Side == OrderSideBuy)

		if isClosing {
			// Calculate realized P&L
			var pnl float64
			if position.Side == PositionSideLong {
				pnl = (order.AvgFillPrice - position.EntryPrice) * order.FilledQuantity
			} else {
				pnl = (position.EntryPrice - order.AvgFillPrice) * order.FilledQuantity
			}
			pnl -= order.Commission
			trade.RealizedPnL = pnl
			position.RealizedPnL += pnl

			if order.FilledQuantity >= position.Quantity {
				// Fully closed
				delete(e.positions, order.Symbol)
				e.emitPositionEvent(PositionEventClosed, position, trade)
			} else {
				// Partial close
				position.Quantity -= order.FilledQuantity
				position.Commission += order.Commission
				position.UpdatedAt = time.Now()
				position.Orders = append(position.Orders, order.ID)
				e.emitPositionEvent(PositionEventUpdated, position, trade)
			}
		} else {
			// Adding to position (averaging)
			totalQty := position.Quantity + order.FilledQuantity
			position.EntryPrice = (position.EntryPrice*position.Quantity + order.AvgFillPrice*order.FilledQuantity) / totalQty
			position.Quantity = totalQty
			position.Commission += order.Commission
			position.UpdatedAt = time.Now()
			position.Orders = append(position.Orders, order.ID)
			e.emitPositionEvent(PositionEventUpdated, position, trade)
		}
	}

	// Emit fill event
	if e.onFill != nil {
		e.onFill(FillEvent{
			OrderID:    order.ID,
			TradeID:    trade.ID,
			Symbol:     order.Symbol,
			Side:       order.Side,
			Quantity:   order.FilledQuantity,
			Price:      order.AvgFillPrice,
			Commission: order.Commission,
			Timestamp:  time.Now(),
		})
	}

	return trade, position
}

// CancelOrder cancels an existing order
func (e *LiveExecutor) CancelOrder(orderID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	order, exists := e.orders[orderID]
	if !exists {
		return fmt.Errorf("order not found: %s", orderID)
	}

	binanceOrderID, err := strconv.ParseInt(orderID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid order ID: %s", orderID)
	}

	// Cancel on Binance
	_, err = e.client.CancelOrder(order.Symbol, binanceOrderID)
	if err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	order.Status = OrderStatusCanceled
	order.UpdatedAt = time.Now()

	log.Info().
		Str("orderID", orderID).
		Str("symbol", order.Symbol).
		Msg("Order canceled")

	return nil
}

// GetOrder returns order by ID
func (e *LiveExecutor) GetOrder(orderID string) (*Order, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	order, exists := e.orders[orderID]
	if !exists {
		return nil, fmt.Errorf("order not found: %s", orderID)
	}

	return order, nil
}

// GetOpenOrders returns all open orders
func (e *LiveExecutor) GetOpenOrders(symbol string) ([]*Order, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Fetch from Binance for accurate state
	binanceOrders, err := e.client.GetOpenOrders(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch open orders: %w", err)
	}

	orders := make([]*Order, 0, len(binanceOrders))
	for _, bo := range binanceOrders {
		origQty, _ := strconv.ParseFloat(bo.OrigQty, 64)
		price, _ := strconv.ParseFloat(bo.Price, 64)
		stopPrice, _ := strconv.ParseFloat(bo.StopPrice, 64)
		executedQty, _ := strconv.ParseFloat(bo.ExecutedQty, 64)

		order := &Order{
			ID:             fmt.Sprintf("%d", bo.OrderID),
			ClientID:       bo.ClientOrderID,
			Symbol:         bo.Symbol,
			Side:           fromBinanceSide(bo.Side),
			Type:           fromBinanceOrderType(bo.Type),
			Quantity:       origQty,
			Price:          price,
			StopPrice:      stopPrice,
			Status:         mapOrderStatus(string(bo.Status)),
			FilledQuantity: executedQty,
			CreatedAt:      time.UnixMilli(bo.Time),
			UpdatedAt:      time.UnixMilli(bo.UpdateTime),
		}
		orders = append(orders, order)

		// Update local cache
		e.orders[order.ID] = order
	}

	return orders, nil
}

// GetPosition returns position by symbol
func (e *LiveExecutor) GetPosition(symbol string) (*Position, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	position, exists := e.positions[symbol]
	if !exists {
		return nil, nil
	}

	// Update current price
	ticker, err := e.client.GetTicker(symbol)
	if err == nil {
		e.updatePositionPrice(position, ticker.LastPrice)
	}

	return position, nil
}

// GetPositions returns all open positions
func (e *LiveExecutor) GetPositions() ([]*Position, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	positions := make([]*Position, 0, len(e.positions))
	for _, pos := range e.positions {
		positions = append(positions, pos)
	}

	return positions, nil
}

// ClosePosition closes a position
func (e *LiveExecutor) ClosePosition(positionID int64) (*ExecutionResult, error) {
	e.mu.Lock()

	var position *Position
	var symbol string
	for s, p := range e.positions {
		if p.ID == positionID {
			position = p
			symbol = s
			break
		}
	}

	if position == nil {
		e.mu.Unlock()
		return nil, fmt.Errorf("position not found: %d", positionID)
	}

	// Determine closing side
	side := OrderSideSell
	if position.Side == PositionSideShort {
		side = OrderSideBuy
	}

	e.mu.Unlock()

	// Place market order to close
	closeOrder := &Order{
		Symbol:   symbol,
		Side:     side,
		Type:     OrderTypeMarket,
		Quantity: position.Quantity,
		Strategy: position.Strategy,
	}

	return e.PlaceOrder(closeOrder)
}

// UpdateStopLoss updates position stop loss
func (e *LiveExecutor) UpdateStopLoss(positionID int64, stopLoss float64) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	var position *Position
	for _, p := range e.positions {
		if p.ID == positionID {
			position = p
			break
		}
	}

	if position == nil {
		return fmt.Errorf("position not found: %d", positionID)
	}

	// Cancel existing stop loss orders
	for _, orderID := range position.Orders {
		order, exists := e.orders[orderID]
		if exists && order.Type == OrderTypeStopLoss && order.Status == OrderStatusOpen {
			binanceOrderID, _ := strconv.ParseInt(orderID, 10, 64)
			e.client.CancelOrder(position.Symbol, binanceOrderID)
		}
	}

	position.StopLoss = stopLoss
	position.UpdatedAt = time.Now()

	// Place new stop loss order
	if stopLoss > 0 {
		side := OrderSideSell
		if position.Side == PositionSideShort {
			side = OrderSideBuy
		}

		// Use stop-loss-limit with price slightly worse than stop
		price := stopLoss * 0.995 // 0.5% slippage allowance for longs
		if position.Side == PositionSideShort {
			price = stopLoss * 1.005
		}

		e.mu.Unlock()
		_, err := e.PlaceOrder(&Order{
			Symbol:    position.Symbol,
			Side:      side,
			Type:      OrderTypeStopLoss,
			Quantity:  position.Quantity,
			Price:     price,
			StopPrice: stopLoss,
			Strategy:  position.Strategy,
		})
		e.mu.Lock()

		if err != nil {
			log.Error().Err(err).Msg("Failed to place stop loss order")
		}
	}

	return nil
}

// UpdateTakeProfit updates position take profit
func (e *LiveExecutor) UpdateTakeProfit(positionID int64, takeProfit float64) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	var position *Position
	for _, p := range e.positions {
		if p.ID == positionID {
			position = p
			break
		}
	}

	if position == nil {
		return fmt.Errorf("position not found: %d", positionID)
	}

	// Cancel existing take profit orders
	for _, orderID := range position.Orders {
		order, exists := e.orders[orderID]
		if exists && order.Type == OrderTypeTakeProfit && order.Status == OrderStatusOpen {
			binanceOrderID, _ := strconv.ParseInt(orderID, 10, 64)
			e.client.CancelOrder(position.Symbol, binanceOrderID)
		}
	}

	position.TakeProfit = takeProfit
	position.UpdatedAt = time.Now()

	// Place new take profit order (as limit order)
	if takeProfit > 0 {
		side := OrderSideSell
		if position.Side == PositionSideShort {
			side = OrderSideBuy
		}

		e.mu.Unlock()
		_, err := e.PlaceOrder(&Order{
			Symbol:   position.Symbol,
			Side:     side,
			Type:     OrderTypeLimit,
			Quantity: position.Quantity,
			Price:    takeProfit,
			Strategy: position.Strategy,
		})
		e.mu.Lock()

		if err != nil {
			log.Error().Err(err).Msg("Failed to place take profit order")
		}
	}

	return nil
}

// GetBalance returns account balance for an asset
func (e *LiveExecutor) GetBalance(asset string) (free, locked float64, err error) {
	e.mu.RLock()
	balance, exists := e.balances[asset]
	e.mu.RUnlock()

	if !exists {
		// Fetch from Binance
		account, err := e.client.GetAccount()
		if err != nil {
			return 0, 0, fmt.Errorf("failed to fetch account: %w", err)
		}

		e.mu.Lock()
		for _, bal := range account.Balances {
			e.balances[bal.Asset] = struct{ Free, Locked float64 }{
				Free:   bal.Free,
				Locked: bal.Locked,
			}
		}
		balance = e.balances[asset]
		e.mu.Unlock()
	}

	return balance.Free, balance.Locked, nil
}

// GetEquity returns total equity in USDT
func (e *LiveExecutor) GetEquity() (float64, error) {
	account, err := e.client.GetAccount()
	if err != nil {
		return 0, fmt.Errorf("failed to fetch account: %w", err)
	}

	// Sum up USDT and USDT-equivalent values
	var equity float64
	for _, bal := range account.Balances {
		total := bal.Free + bal.Locked
		if total == 0 {
			continue
		}

		if bal.Asset == "USDT" || bal.Asset == "BUSD" || bal.Asset == "USD" {
			equity += total
		} else if bal.Asset == "ETH" {
			ticker, err := e.client.GetTicker("ETHUSDT")
			if err == nil {
				equity += total * ticker.LastPrice
			}
		} else if bal.Asset == "BTC" {
			ticker, err := e.client.GetTicker("BTCUSDT")
			if err == nil {
				equity += total * ticker.LastPrice
			}
		}
	}

	// Add unrealized P&L from positions
	for _, pos := range e.positions {
		equity += pos.UnrealizedPnL
	}

	return equity, nil
}

// Sync synchronizes state with Binance
func (e *LiveExecutor) Sync() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Sync account balances
	account, err := e.client.GetAccount()
	if err != nil {
		return fmt.Errorf("failed to sync account: %w", err)
	}

	for _, bal := range account.Balances {
		e.balances[bal.Asset] = struct{ Free, Locked float64 }{
			Free:   bal.Free,
			Locked: bal.Locked,
		}
	}

	// Sync open orders
	if e.config.Symbol != "" {
		orders, err := e.client.GetOpenOrders(e.config.Symbol)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to sync open orders")
		} else {
			for _, bo := range orders {
				origQty, _ := strconv.ParseFloat(bo.OrigQty, 64)
				price, _ := strconv.ParseFloat(bo.Price, 64)
				stopPrice, _ := strconv.ParseFloat(bo.StopPrice, 64)
				executedQty, _ := strconv.ParseFloat(bo.ExecutedQty, 64)

				order := &Order{
					ID:             fmt.Sprintf("%d", bo.OrderID),
					ClientID:       bo.ClientOrderID,
					Symbol:         bo.Symbol,
					Side:           fromBinanceSide(bo.Side),
					Type:           fromBinanceOrderType(bo.Type),
					Quantity:       origQty,
					Price:          price,
					StopPrice:      stopPrice,
					Status:         mapOrderStatus(string(bo.Status)),
					FilledQuantity: executedQty,
					CreatedAt:      time.UnixMilli(bo.Time),
					UpdatedAt:      time.UnixMilli(bo.UpdateTime),
				}
				e.orders[order.ID] = order
			}
		}
	}

	// Update position prices
	for symbol, pos := range e.positions {
		ticker, err := e.client.GetTicker(symbol)
		if err == nil {
			e.updatePositionPrice(pos, ticker.LastPrice)
		}
	}

	log.Debug().Msg("State synchronized with Binance")
	return nil
}

// updatePositionPrice updates position with current price
func (e *LiveExecutor) updatePositionPrice(pos *Position, price float64) {
	pos.CurrentPrice = price
	pos.UpdatedAt = time.Now()

	// Calculate unrealized P&L
	if pos.Side == PositionSideLong {
		pos.UnrealizedPnL = (price - pos.EntryPrice) * pos.Quantity
	} else {
		pos.UnrealizedPnL = (pos.EntryPrice - price) * pos.Quantity
	}

	if pos.EntryPrice > 0 {
		pos.UnrealizedPnLPct = pos.UnrealizedPnL / (pos.EntryPrice * pos.Quantity)
	}
}

// getSymbolInfo gets symbol trading rules
func (e *LiveExecutor) getSymbolInfo(symbol string) (*binance.SymbolInfo, error) {
	if info, exists := e.symbolInfo[symbol]; exists {
		return info, nil
	}

	info, err := e.client.GetSymbolInfo(symbol)
	if err != nil {
		return nil, err
	}

	e.symbolInfo[symbol] = info
	return info, nil
}

// periodicSync runs periodic synchronization
func (e *LiveExecutor) periodicSync() {
	for {
		select {
		case <-e.ctx.Done():
			return
		case <-e.syncTicker.C:
			if err := e.Sync(); err != nil {
				log.Error().Err(err).Msg("Periodic sync failed")
			}
		}
	}
}

// emitPositionEvent emits a position event
func (e *LiveExecutor) emitPositionEvent(eventType PositionEventType, position *Position, trade *Trade) {
	if e.onPosition == nil {
		return
	}

	e.onPosition(PositionEvent{
		Type:      eventType,
		Position:  position,
		Trade:     trade,
		Timestamp: time.Now(),
	})
}

// userDataHandler handles user data stream events
type userDataHandler struct {
	executor *LiveExecutor
}

func (h *userDataHandler) OnKline(event binance.KlineEvent)           {}
func (h *userDataHandler) OnTrade(event binance.TradeEvent)           {}
func (h *userDataHandler) OnDepth(event binance.DepthEvent)           {}
func (h *userDataHandler) OnMiniTicker(event binance.MiniTickerEvent) {}
func (h *userDataHandler) OnError(err error) {
	log.Error().Err(err).Msg("User data stream error")
}
func (h *userDataHandler) OnDisconnect() {
	log.Warn().Msg("User data stream disconnected")
}
func (h *userDataHandler) OnReconnect() {
	log.Info().Msg("User data stream reconnected")
}

// StartUserDataStream starts the user data stream for real-time updates
func (e *LiveExecutor) StartUserDataStream() error {
	// Get listen key
	listenKey, err := e.client.GetListenKey()
	if err != nil {
		return fmt.Errorf("failed to get listen key: %w", err)
	}

	// Create WebSocket client for user data
	handler := &userDataHandler{executor: e}
	opts := []binance.WSClientOption{
		binance.WithWSTestnet(e.config.Testnet),
		binance.WithPingInterval(30 * time.Second),
		binance.WithReconnectWait(5 * time.Second),
		binance.WithMaxReconnects(10),
	}
	e.wsClient = binance.NewWSClient(handler, opts...)

	// Subscribe to user data stream
	if err := e.wsClient.Subscribe(listenKey); err != nil {
		return fmt.Errorf("failed to subscribe to user data: %w", err)
	}

	if err := e.wsClient.Connect(e.ctx); err != nil {
		return fmt.Errorf("failed to connect to user data stream: %w", err)
	}

	// Keep listen key alive
	go e.keepAliveListenKey(listenKey)

	log.Info().Msg("User data stream started")
	return nil
}

// handleUserDataEvent handles user data stream events
func (e *LiveExecutor) handleUserDataEvent(msg []byte) {
	// Parse event type and handle accordingly
	// This would parse order updates, balance updates, etc.
	// Implementation depends on Binance user data stream format
	log.Debug().Str("event", string(msg)).Msg("User data event received")
}

// keepAliveListenKey keeps the listen key alive
func (e *LiveExecutor) keepAliveListenKey(listenKey string) {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
			if err := e.client.KeepAliveListenKey(listenKey); err != nil {
				log.Error().Err(err).Msg("Failed to keep listen key alive")
			}
		}
	}
}

// Stop stops the live executor
func (e *LiveExecutor) Stop() {
	e.cancel()
	if e.syncTicker != nil {
		e.syncTicker.Stop()
	}
	if e.wsClient != nil {
		e.wsClient.Disconnect()
	}
	log.Info().Msg("Live executor stopped")
}

// GetAccountSummary returns account summary
func (e *LiveExecutor) GetAccountSummary() (*AccountSummary, error) {
	equity, err := e.GetEquity()
	if err != nil {
		return nil, err
	}

	e.mu.RLock()
	defer e.mu.RUnlock()

	usdtFree := e.balances["USDT"].Free
	usdtLocked := e.balances["USDT"].Locked

	var unrealizedPnL float64
	for _, pos := range e.positions {
		unrealizedPnL += pos.UnrealizedPnL
	}

	return &AccountSummary{
		Mode:             ModeLive,
		Equity:           equity,
		AvailableBalance: usdtFree,
		UsedMargin:       usdtLocked,
		UnrealizedPnL:    unrealizedPnL,
		OpenPositions:    len(e.positions),
	}, nil
}

// mapOrderStatus maps Binance order status to internal status
func mapOrderStatus(status string) OrderStatus {
	switch status {
	case "NEW":
		return OrderStatusOpen
	case "PARTIALLY_FILLED":
		return OrderStatusPartial
	case "FILLED":
		return OrderStatusFilled
	case "CANCELED":
		return OrderStatusCanceled
	case "PENDING_CANCEL":
		return OrderStatusCanceled
	case "REJECTED":
		return OrderStatusRejected
	case "EXPIRED":
		return OrderStatusExpired
	default:
		return OrderStatusPending
	}
}

// roundToStepSize rounds quantity to valid step size
func roundToStepSize(qty, stepSize float64, precision int) float64 {
	if stepSize <= 0 {
		return qty
	}
	steps := int(qty / stepSize)
	rounded := float64(steps) * stepSize

	// Apply precision
	multiplier := 1.0
	for i := 0; i < precision; i++ {
		multiplier *= 10
	}
	return float64(int(rounded*multiplier)) / multiplier
}

// roundToTickSize rounds price to valid tick size
func roundToTickSize(price, tickSize float64, precision int) float64 {
	if tickSize <= 0 {
		return price
	}
	ticks := int(price / tickSize)
	rounded := float64(ticks) * tickSize

	// Apply precision
	multiplier := 1.0
	for i := 0; i < precision; i++ {
		multiplier *= 10
	}
	return float64(int(rounded*multiplier)) / multiplier
}

// toBinanceSide converts internal OrderSide to Binance OrderSide
func toBinanceSide(side OrderSide) binance.OrderSide {
	switch side {
	case OrderSideBuy:
		return binance.SideBuy
	case OrderSideSell:
		return binance.SideSell
	default:
		return binance.SideBuy
	}
}

// fromBinanceSide converts Binance OrderSide to internal OrderSide
func fromBinanceSide(side binance.OrderSide) OrderSide {
	switch side {
	case binance.SideBuy:
		return OrderSideBuy
	case binance.SideSell:
		return OrderSideSell
	default:
		return OrderSideBuy
	}
}

// toBinanceOrderType converts internal OrderType to Binance OrderType
func toBinanceOrderType(t OrderType) binance.OrderType {
	switch t {
	case OrderTypeMarket:
		return binance.OrderTypeMarket
	case OrderTypeLimit:
		return binance.OrderTypeLimit
	case OrderTypeStopLoss:
		return binance.OrderTypeStopLoss
	case OrderTypeTakeProfit:
		return binance.OrderTypeTakeProfit
	default:
		return binance.OrderTypeMarket
	}
}

// fromBinanceOrderType converts Binance OrderType to internal OrderType
func fromBinanceOrderType(t binance.OrderType) OrderType {
	switch t {
	case binance.OrderTypeMarket:
		return OrderTypeMarket
	case binance.OrderTypeLimit:
		return OrderTypeLimit
	case binance.OrderTypeStopLoss, binance.OrderTypeStopLossLimit:
		return OrderTypeStopLoss
	case binance.OrderTypeTakeProfit, binance.OrderTypeTakeProfitLimit:
		return OrderTypeTakeProfit
	default:
		return OrderTypeMarket
	}
}
