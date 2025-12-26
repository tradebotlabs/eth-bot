package execution

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// PaperExecutor simulates order execution for paper trading
type PaperExecutor struct {
	config *ExecutorConfig

	// Account state
	balance     map[string]float64 // asset -> balance
	positions   map[string]*Position // symbol -> position
	orders      map[string]*Order
	trades      []*Trade

	// Statistics
	stats       *TradeStats
	totalPnL    float64
	totalCommission float64

	// Current prices (updated externally)
	prices      map[string]float64

	// Callbacks
	onFill      func(FillEvent)
	onPosition  func(PositionEvent)

	mu sync.RWMutex
	nextPosID int64
}

// NewPaperExecutor creates a new paper executor
func NewPaperExecutor(config *ExecutorConfig) *PaperExecutor {
	if config == nil {
		config = DefaultExecutorConfig()
	}
	config.Mode = ModePaper

	pe := &PaperExecutor{
		config:    config,
		balance:   make(map[string]float64),
		positions: make(map[string]*Position),
		orders:    make(map[string]*Order),
		trades:    make([]*Trade, 0),
		prices:    make(map[string]float64),
		stats:     &TradeStats{},
		nextPosID: 1,
	}

	// Initialize balance
	pe.balance["USDT"] = config.InitialBalance

	log.Info().
		Float64("balance", config.InitialBalance).
		Float64("commission", config.Commission).
		Msg("Paper executor initialized")

	return pe
}

// GetMode returns execution mode
func (pe *PaperExecutor) GetMode() ExecutionMode {
	return ModePaper
}

// SetOnFill sets fill callback
func (pe *PaperExecutor) SetOnFill(fn func(FillEvent)) {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	pe.onFill = fn
}

// SetOnPosition sets position callback
func (pe *PaperExecutor) SetOnPosition(fn func(PositionEvent)) {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	pe.onPosition = fn
}

// UpdatePrice updates current price for a symbol
func (pe *PaperExecutor) UpdatePrice(symbol string, price float64) {
	pe.mu.Lock()
	pe.prices[symbol] = price

	// Update position P&L
	if pos, exists := pe.positions[symbol]; exists {
		pos.CurrentPrice = price
		pos.UpdatedAt = time.Now()

		if pos.Side == PositionSideLong {
			pos.UnrealizedPnL = (price - pos.EntryPrice) * pos.Quantity
		} else {
			pos.UnrealizedPnL = (pos.EntryPrice - price) * pos.Quantity
		}

		if pos.EntryPrice > 0 {
			pos.UnrealizedPnLPct = pos.UnrealizedPnL / (pos.EntryPrice * pos.Quantity)
		}

		// Check stop loss / take profit
		pe.checkStopTakeProfit(pos, price)
	}
	pe.mu.Unlock()
}

// checkStopTakeProfit checks and executes stop loss / take profit
func (pe *PaperExecutor) checkStopTakeProfit(pos *Position, price float64) {
	if pos.Side == PositionSideLong {
		// Stop loss
		if pos.StopLoss > 0 && price <= pos.StopLoss {
			pe.mu.Unlock() // Unlock before closing
			pe.closePositionInternal(pos.ID, price, PositionEventStopLossHit)
			pe.mu.Lock()
			return
		}
		// Take profit
		if pos.TakeProfit > 0 && price >= pos.TakeProfit {
			pe.mu.Unlock()
			pe.closePositionInternal(pos.ID, price, PositionEventTakeProfitHit)
			pe.mu.Lock()
			return
		}
	} else {
		// Short position
		// Stop loss
		if pos.StopLoss > 0 && price >= pos.StopLoss {
			pe.mu.Unlock()
			pe.closePositionInternal(pos.ID, price, PositionEventStopLossHit)
			pe.mu.Lock()
			return
		}
		// Take profit
		if pos.TakeProfit > 0 && price <= pos.TakeProfit {
			pe.mu.Unlock()
			pe.closePositionInternal(pos.ID, price, PositionEventTakeProfitHit)
			pe.mu.Lock()
			return
		}
	}
}

// PlaceOrder places a new order
func (pe *PaperExecutor) PlaceOrder(order *Order) (*ExecutionResult, error) {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	start := time.Now()

	// Generate order ID
	if order.ID == "" {
		order.ID = uuid.New().String()
	}
	if order.ClientID == "" {
		order.ClientID = fmt.Sprintf("paper_%d", time.Now().UnixNano())
	}

	order.CreatedAt = time.Now()
	order.Status = OrderStatusPending

	// Get current price
	price, ok := pe.prices[order.Symbol]
	if !ok {
		order.Status = OrderStatusRejected
		return &ExecutionResult{
			Success: false,
			Order:   order,
			Error:   fmt.Errorf("no price available for %s", order.Symbol),
			Message: "No price available",
			Latency: time.Since(start),
		}, fmt.Errorf("no price for symbol")
	}

	// Determine execution price
	execPrice := price
	if order.Type == OrderTypeLimit {
		execPrice = order.Price
	} else if order.Type == OrderTypeMarket {
		// Apply slippage
		if order.Side == OrderSideBuy {
			execPrice = price * (1 + pe.config.Slippage)
		} else {
			execPrice = price * (1 - pe.config.Slippage)
		}
	}

	// Calculate order value
	orderValue := order.Quantity * execPrice
	commission := orderValue * pe.config.Commission

	// Check balance
	if order.Side == OrderSideBuy {
		available := pe.balance["USDT"]
		required := orderValue + commission
		if available < required {
			order.Status = OrderStatusRejected
			return &ExecutionResult{
				Success: false,
				Order:   order,
				Error:   fmt.Errorf("insufficient balance: have %.2f, need %.2f", available, required),
				Message: "Insufficient balance",
				Latency: time.Since(start),
			}, nil
		}
	}

	// Execute order immediately (market orders)
	if order.Type == OrderTypeMarket {
		return pe.executeOrder(order, execPrice, commission, start)
	}

	// Store limit order
	pe.orders[order.ID] = order
	order.Status = OrderStatusOpen

	return &ExecutionResult{
		Success: true,
		Order:   order,
		Message: "Order placed",
		Latency: time.Since(start),
	}, nil
}

// executeOrder executes an order
func (pe *PaperExecutor) executeOrder(order *Order, execPrice, commission float64, start time.Time) (*ExecutionResult, error) {
	order.FilledQuantity = order.Quantity
	order.AvgFillPrice = execPrice
	order.Commission = commission
	order.CommissionAsset = "USDT"
	order.Status = OrderStatusFilled
	order.FilledAt = time.Now()
	order.UpdatedAt = time.Now()

	// Update balance
	orderValue := order.Quantity * execPrice

	if order.Side == OrderSideBuy {
		pe.balance["USDT"] -= (orderValue + commission)
	} else {
		pe.balance["USDT"] += (orderValue - commission)
	}

	pe.totalCommission += commission

	// Create trade record
	trade := &Trade{
		ID:              uuid.New().String(),
		OrderID:         order.ID,
		Symbol:          order.Symbol,
		Side:            order.Side,
		Quantity:        order.Quantity,
		Price:           execPrice,
		Commission:      commission,
		CommissionAsset: "USDT",
		Strategy:        order.Strategy,
		ExecutedAt:      time.Now(),
	}

	// Handle position
	var position *Position
	var posEvent PositionEventType

	existingPos, hasPosition := pe.positions[order.Symbol]

	if hasPosition {
		// Modify existing position
		position, posEvent = pe.handleExistingPosition(existingPos, order, trade, execPrice)
	} else {
		// Open new position
		position, posEvent = pe.openNewPosition(order, trade, execPrice)
	}

	if position != nil {
		trade.PositionID = position.ID
	}

	pe.trades = append(pe.trades, trade)
	pe.orders[order.ID] = order

	// Emit events
	if pe.onFill != nil {
		go pe.onFill(FillEvent{
			OrderID:    order.ID,
			TradeID:    trade.ID,
			Symbol:     order.Symbol,
			Side:       order.Side,
			Quantity:   order.Quantity,
			Price:      execPrice,
			Commission: commission,
			Timestamp:  time.Now(),
		})
	}

	if pe.onPosition != nil && position != nil {
		go pe.onPosition(PositionEvent{
			Type:      posEvent,
			Position:  position,
			Trade:     trade,
			Timestamp: time.Now(),
		})
	}

	log.Info().
		Str("orderID", order.ID).
		Str("symbol", order.Symbol).
		Str("side", string(order.Side)).
		Float64("quantity", order.Quantity).
		Float64("price", execPrice).
		Float64("commission", commission).
		Msg("Order executed (paper)")

	return &ExecutionResult{
		Success:  true,
		Order:    order,
		Trade:    trade,
		Position: position,
		Message:  "Order filled",
		Latency:  time.Since(start),
	}, nil
}

// handleExistingPosition handles order for existing position
func (pe *PaperExecutor) handleExistingPosition(pos *Position, order *Order, trade *Trade, execPrice float64) (*Position, PositionEventType) {
	// Check if closing or adding to position
	isClosing := (pos.Side == PositionSideLong && order.Side == OrderSideSell) ||
		(pos.Side == PositionSideShort && order.Side == OrderSideBuy)

	if isClosing {
		// Close position
		var pnl float64
		if pos.Side == PositionSideLong {
			pnl = (execPrice - pos.EntryPrice) * order.Quantity
		} else {
			pnl = (pos.EntryPrice - execPrice) * order.Quantity
		}

		trade.RealizedPnL = pnl
		pos.RealizedPnL += pnl
		pe.totalPnL += pnl

		// Update stats
		pe.updateStats(pnl, pos.OpenTime)

		if order.Quantity >= pos.Quantity {
			// Full close
			delete(pe.positions, order.Symbol)
			return pos, PositionEventClosed
		} else {
			// Partial close
			pos.Quantity -= order.Quantity
			pos.UpdatedAt = time.Now()
			return pos, PositionEventUpdated
		}
	} else {
		// Add to position (average in)
		totalQty := pos.Quantity + order.Quantity
		pos.EntryPrice = (pos.EntryPrice*pos.Quantity + execPrice*order.Quantity) / totalQty
		pos.Quantity = totalQty
		pos.UpdatedAt = time.Now()
		pos.Orders = append(pos.Orders, order.ID)
		return pos, PositionEventUpdated
	}
}

// openNewPosition opens a new position
func (pe *PaperExecutor) openNewPosition(order *Order, trade *Trade, execPrice float64) (*Position, PositionEventType) {
	var side PositionSide
	if order.Side == OrderSideBuy {
		side = PositionSideLong
	} else {
		side = PositionSideShort
	}

	pos := &Position{
		ID:           pe.nextPosID,
		Symbol:       order.Symbol,
		Side:         side,
		Quantity:     order.Quantity,
		EntryPrice:   execPrice,
		CurrentPrice: execPrice,
		Strategy:     order.Strategy,
		OpenTime:     time.Now(),
		UpdatedAt:    time.Now(),
		Orders:       []string{order.ID},
	}

	// Set stop loss / take profit from signal
	if order.Signal != nil {
		pos.StopLoss = order.Signal.StopLoss
		pos.TakeProfit = order.Signal.TakeProfit
	}

	pe.nextPosID++
	pe.positions[order.Symbol] = pos

	return pos, PositionEventOpened
}

// updateStats updates trading statistics
func (pe *PaperExecutor) updateStats(pnl float64, openTime time.Time) {
	pe.stats.TotalTrades++

	if pnl > 0 {
		pe.stats.WinningTrades++
		pe.stats.GrossProfit += pnl
		if pnl > pe.stats.LargestWin {
			pe.stats.LargestWin = pnl
		}
	} else {
		pe.stats.LosingTrades++
		pe.stats.GrossLoss += pnl
		if pnl < pe.stats.LargestLoss {
			pe.stats.LargestLoss = pnl
		}
	}

	pe.stats.NetProfit = pe.stats.GrossProfit + pe.stats.GrossLoss

	if pe.stats.TotalTrades > 0 {
		pe.stats.WinRate = float64(pe.stats.WinningTrades) / float64(pe.stats.TotalTrades)
	}

	if pe.stats.WinningTrades > 0 {
		pe.stats.AvgWin = pe.stats.GrossProfit / float64(pe.stats.WinningTrades)
	}

	if pe.stats.LosingTrades > 0 {
		pe.stats.AvgLoss = pe.stats.GrossLoss / float64(pe.stats.LosingTrades)
	}

	if pe.stats.GrossLoss != 0 {
		pe.stats.ProfitFactor = pe.stats.GrossProfit / (-pe.stats.GrossLoss)
	}
}

// closePositionInternal closes a position internally
func (pe *PaperExecutor) closePositionInternal(positionID int64, price float64, eventType PositionEventType) {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	var targetPos *Position
	var symbol string

	for sym, pos := range pe.positions {
		if pos.ID == positionID {
			targetPos = pos
			symbol = sym
			break
		}
	}

	if targetPos == nil {
		return
	}

	// Create closing order
	var side OrderSide
	if targetPos.Side == PositionSideLong {
		side = OrderSideSell
	} else {
		side = OrderSideBuy
	}

	order := &Order{
		ID:        uuid.New().String(),
		Symbol:    symbol,
		Side:      side,
		Type:      OrderTypeMarket,
		Quantity:  targetPos.Quantity,
		Strategy:  targetPos.Strategy,
		CreatedAt: time.Now(),
	}

	// Calculate P&L
	var pnl float64
	if targetPos.Side == PositionSideLong {
		pnl = (price - targetPos.EntryPrice) * targetPos.Quantity
	} else {
		pnl = (targetPos.EntryPrice - price) * targetPos.Quantity
	}

	commission := targetPos.Quantity * price * pe.config.Commission

	// Create trade
	trade := &Trade{
		ID:          uuid.New().String(),
		OrderID:     order.ID,
		PositionID:  positionID,
		Symbol:      symbol,
		Side:        side,
		Quantity:    targetPos.Quantity,
		Price:       price,
		Commission:  commission,
		RealizedPnL: pnl,
		Strategy:    targetPos.Strategy,
		ExecutedAt:  time.Now(),
	}

	// Update balance
	orderValue := targetPos.Quantity * price
	if side == OrderSideSell {
		pe.balance["USDT"] += (orderValue - commission)
	} else {
		pe.balance["USDT"] -= (orderValue + commission)
	}

	pe.totalCommission += commission
	pe.totalPnL += pnl

	// Update stats
	pe.updateStats(pnl, targetPos.OpenTime)

	// Remove position
	delete(pe.positions, symbol)

	// Store records
	order.Status = OrderStatusFilled
	order.FilledQuantity = order.Quantity
	order.AvgFillPrice = price
	order.Commission = commission
	order.FilledAt = time.Now()

	pe.orders[order.ID] = order
	pe.trades = append(pe.trades, trade)

	// Emit event
	if pe.onPosition != nil {
		go pe.onPosition(PositionEvent{
			Type:      eventType,
			Position:  targetPos,
			Trade:     trade,
			Timestamp: time.Now(),
		})
	}

	log.Info().
		Int64("positionID", positionID).
		Str("symbol", symbol).
		Float64("pnl", pnl).
		Str("reason", eventType.String()).
		Msg("Position closed (paper)")
}

// CancelOrder cancels an order
func (pe *PaperExecutor) CancelOrder(orderID string) error {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	order, exists := pe.orders[orderID]
	if !exists {
		return fmt.Errorf("order not found: %s", orderID)
	}

	if order.Status != OrderStatusOpen && order.Status != OrderStatusPending {
		return fmt.Errorf("order cannot be canceled: %s", order.Status)
	}

	order.Status = OrderStatusCanceled
	order.UpdatedAt = time.Now()

	return nil
}

// GetOrder returns order by ID
func (pe *PaperExecutor) GetOrder(orderID string) (*Order, error) {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	order, exists := pe.orders[orderID]
	if !exists {
		return nil, fmt.Errorf("order not found: %s", orderID)
	}

	return order, nil
}

// GetOpenOrders returns all open orders
func (pe *PaperExecutor) GetOpenOrders(symbol string) ([]*Order, error) {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	var orders []*Order
	for _, order := range pe.orders {
		if order.Status == OrderStatusOpen || order.Status == OrderStatusPending {
			if symbol == "" || order.Symbol == symbol {
				orders = append(orders, order)
			}
		}
	}

	return orders, nil
}

// GetPosition returns position by symbol
func (pe *PaperExecutor) GetPosition(symbol string) (*Position, error) {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	pos, exists := pe.positions[symbol]
	if !exists {
		return nil, nil
	}

	return pos, nil
}

// GetPositions returns all open positions
func (pe *PaperExecutor) GetPositions() ([]*Position, error) {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	positions := make([]*Position, 0, len(pe.positions))
	for _, pos := range pe.positions {
		positions = append(positions, pos)
	}

	return positions, nil
}

// ClosePosition closes a position
func (pe *PaperExecutor) ClosePosition(positionID int64) (*ExecutionResult, error) {
	pe.mu.RLock()
	var targetPos *Position
	var symbol string

	for sym, pos := range pe.positions {
		if pos.ID == positionID {
			targetPos = pos
			symbol = sym
			break
		}
	}
	pe.mu.RUnlock()

	if targetPos == nil {
		return nil, fmt.Errorf("position not found: %d", positionID)
	}

	price := pe.prices[symbol]
	pe.closePositionInternal(positionID, price, PositionEventClosed)

	return &ExecutionResult{
		Success: true,
		Message: "Position closed",
	}, nil
}

// UpdateStopLoss updates position stop loss
func (pe *PaperExecutor) UpdateStopLoss(positionID int64, stopLoss float64) error {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	for _, pos := range pe.positions {
		if pos.ID == positionID {
			pos.StopLoss = stopLoss
			pos.UpdatedAt = time.Now()
			return nil
		}
	}

	return fmt.Errorf("position not found: %d", positionID)
}

// UpdateTakeProfit updates position take profit
func (pe *PaperExecutor) UpdateTakeProfit(positionID int64, takeProfit float64) error {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	for _, pos := range pe.positions {
		if pos.ID == positionID {
			pos.TakeProfit = takeProfit
			pos.UpdatedAt = time.Now()
			return nil
		}
	}

	return fmt.Errorf("position not found: %d", positionID)
}

// GetBalance returns account balance
func (pe *PaperExecutor) GetBalance(asset string) (free, locked float64, err error) {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	free = pe.balance[asset]
	return free, 0, nil
}

// GetEquity returns total equity
func (pe *PaperExecutor) GetEquity() (float64, error) {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	equity := pe.balance["USDT"]

	// Add unrealized P&L
	for _, pos := range pe.positions {
		equity += pos.UnrealizedPnL
	}

	return equity, nil
}

// Sync is no-op for paper trading
func (pe *PaperExecutor) Sync() error {
	return nil
}

// GetStats returns trading statistics
func (pe *PaperExecutor) GetStats() *TradeStats {
	pe.mu.RLock()
	defer pe.mu.RUnlock()
	return pe.stats
}

// GetAccountSummary returns account summary
func (pe *PaperExecutor) GetAccountSummary() AccountSummary {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	equity, _ := pe.GetEquity()

	var unrealizedPnL float64
	for _, pos := range pe.positions {
		unrealizedPnL += pos.UnrealizedPnL
	}

	return AccountSummary{
		Mode:             ModePaper,
		Equity:           equity,
		AvailableBalance: pe.balance["USDT"],
		UnrealizedPnL:    unrealizedPnL,
		RealizedPnL:      pe.totalPnL,
		TotalCommission:  pe.totalCommission,
		OpenPositions:    len(pe.positions),
		TotalTrades:      pe.stats.TotalTrades,
		WinRate:          pe.stats.WinRate,
		ProfitFactor:     pe.stats.ProfitFactor,
	}
}

// GetTrades returns all trades
func (pe *PaperExecutor) GetTrades() []*Trade {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	trades := make([]*Trade, len(pe.trades))
	copy(trades, pe.trades)
	return trades
}

// Reset resets paper trading state
func (pe *PaperExecutor) Reset() {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	pe.balance = map[string]float64{"USDT": pe.config.InitialBalance}
	pe.positions = make(map[string]*Position)
	pe.orders = make(map[string]*Order)
	pe.trades = make([]*Trade, 0)
	pe.stats = &TradeStats{}
	pe.totalPnL = 0
	pe.totalCommission = 0
	pe.nextPosID = 1

	log.Info().Msg("Paper executor reset")
}
