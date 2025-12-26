package orchestrator

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/eth-trading/internal/binance"
	"github.com/eth-trading/internal/execution"
	"github.com/eth-trading/internal/indicators"
	"github.com/eth-trading/internal/risk"
	"github.com/eth-trading/internal/storage"
	"github.com/eth-trading/internal/strategy"
	"github.com/rs/zerolog/log"
)

// Orchestrator coordinates all trading components
type Orchestrator struct {
	config        *OrchestratorConfig

	// Components
	binanceClient *binance.Client
	wsClient      *binance.WSClient
	dataService   *storage.DataService
	executor      execution.Executor
	riskManager   *risk.Manager
	strategyMgr   *strategy.Manager
	indicatorMgr  *indicators.Manager

	// State
	state         *TradingState
	stateMu       sync.RWMutex

	// Signal history (recent signals for UI)
	signals       []SignalRecord
	signalsMu     sync.RWMutex

	// Broadcasting
	broadcaster   *Broadcaster
	subscribers   map[string]chan BroadcastMessage

	// Control
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	startTime     time.Time
}

// NewOrchestrator creates a new orchestrator
func NewOrchestrator(config *OrchestratorConfig) *Orchestrator {
	if config == nil {
		config = DefaultOrchestratorConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	o := &Orchestrator{
		config:      config,
		state:       &TradingState{},
		subscribers: make(map[string]chan BroadcastMessage),
		ctx:         ctx,
		cancel:      cancel,
	}

	o.broadcaster = NewBroadcaster(o)

	return o
}

// SetBinanceClient sets the Binance client
func (o *Orchestrator) SetBinanceClient(client *binance.Client) {
	o.binanceClient = client
}

// SetWebSocketClient sets the WebSocket client
func (o *Orchestrator) SetWebSocketClient(ws *binance.WSClient) {
	o.wsClient = ws
}

// SetDataService sets the data service
func (o *Orchestrator) SetDataService(ds *storage.DataService) {
	o.dataService = ds
}

// GetDataService returns the data service
func (o *Orchestrator) GetDataService() *storage.DataService {
	return o.dataService
}

// SetExecutor sets the executor
func (o *Orchestrator) SetExecutor(exec execution.Executor) {
	o.executor = exec
}

// SetRiskManager sets the risk manager
func (o *Orchestrator) SetRiskManager(rm *risk.Manager) {
	o.riskManager = rm

	// Set up risk event callback
	rm.SetOnRiskEvent(func(event risk.RiskEvent) {
		o.broadcastRiskEvent(event)
	})
}

// GetRiskManager returns the risk manager
func (o *Orchestrator) GetRiskManager() *risk.Manager {
	return o.riskManager
}

// GetStrategyManager returns the strategy manager
func (o *Orchestrator) GetStrategyManager() *strategy.Manager {
	return o.strategyMgr
}

// SetStrategyManager sets the strategy manager
func (o *Orchestrator) SetStrategyManager(sm *strategy.Manager) {
	o.strategyMgr = sm
}

// SetIndicatorManager sets the indicator manager
func (o *Orchestrator) SetIndicatorManager(im *indicators.Manager) {
	o.indicatorMgr = im
}

// Start starts the orchestrator
func (o *Orchestrator) Start() error {
	log.Info().
		Str("symbol", o.config.Symbol).
		Str("mode", o.config.Mode.String()).
		Msg("Starting orchestrator")

	// Validate components
	if o.binanceClient == nil {
		return fmt.Errorf("binance client not set")
	}
	if o.dataService == nil {
		return fmt.Errorf("data service not set")
	}
	if o.executor == nil {
		return fmt.Errorf("executor not set")
	}
	if o.strategyMgr == nil {
		return fmt.Errorf("strategy manager not set")
	}

	o.startTime = time.Now()

	// Initialize state
	o.stateMu.Lock()
	o.state = &TradingState{
		Mode:           o.config.Mode,
		IsRunning:      true,
		StartTime:      o.startTime,
		ActiveStrategies: o.config.EnabledStrategies,
	}
	o.stateMu.Unlock()

	// Load historical data
	if err := o.loadHistoricalData(); err != nil {
		log.Warn().Err(err).Msg("Failed to load historical data")
	}

	// Start WebSocket subscription
	if o.wsClient != nil {
		o.startWebSocketSubscription()
	}

	// Start broadcast loop
	if o.config.EnableWebSocket {
		o.wg.Add(1)
		go o.broadcastLoop()
	}

	// Initialize risk metrics before starting monitor loop
	o.updateRiskMetrics()

	// Start risk monitoring
	o.wg.Add(1)
	go o.riskMonitorLoop()

	// Set up executor callbacks
	o.setupExecutorCallbacks()

	log.Info().Msg("Orchestrator started")
	return nil
}

// Stop stops the orchestrator
func (o *Orchestrator) Stop() {
	log.Info().Msg("Stopping orchestrator")

	o.stateMu.Lock()
	o.state.IsRunning = false
	o.stateMu.Unlock()

	o.cancel()
	o.wg.Wait()

	if o.wsClient != nil {
		o.wsClient.Disconnect()
	}

	log.Info().Msg("Orchestrator stopped")
}

// loadHistoricalData loads historical klines
func (o *Orchestrator) loadHistoricalData() error {
	for _, tf := range o.config.Timeframes {
		// Fetch last 500 candles for each timeframe
		klines, err := o.binanceClient.GetKlines(o.config.Symbol, tf, 500, 0, 0)
		if err != nil {
			log.Warn().Str("timeframe", tf).Err(err).Msg("Failed to fetch klines")
			continue
		}

		// Store in data service
		for _, k := range klines {
			candle := convertKlineToCandle(k, o.config.Symbol, tf)
			o.dataService.AddCandle(*candle)
		}

		log.Debug().
			Str("timeframe", tf).
			Int("count", len(klines)).
			Msg("Loaded historical klines")
	}

	return nil
}

// startWebSocketSubscription starts WebSocket subscriptions
func (o *Orchestrator) startWebSocketSubscription() {
	// Subscribe to kline streams for each timeframe (must use lowercase symbol)
	symbol := strings.ToLower(o.config.Symbol)
	var streams []string
	for _, tf := range o.config.Timeframes {
		stream := fmt.Sprintf("%s@kline_%s", symbol, tf)
		streams = append(streams, stream)
	}
	// Add trade stream for real-time price updates (millisecond latency)
	streams = append(streams, fmt.Sprintf("%s@trade", symbol))
	o.wsClient.Subscribe(streams...)

	// Connect the WebSocket
	if err := o.wsClient.Connect(o.ctx); err != nil {
		log.Warn().Err(err).Msg("Binance WebSocket connection failed, using REST API polling")
		// Start polling fallback only if WebSocket fails
		o.wg.Add(1)
		go o.pollPriceFallback()
	} else {
		log.Info().Msg("Binance WebSocket connected - real-time data active")
		// Start WebSocket message handler
		o.wg.Add(1)
		go o.handleBinanceWebSocket()
	}
}

// handleBinanceWebSocket handles real-time WebSocket messages from Binance
func (o *Orchestrator) handleBinanceWebSocket() {
	defer o.wg.Done()

	log.Info().Msg("Started Binance WebSocket handler - real-time data active")

	// The wsClient is already connected, handler receives events
	// Just monitor connection status here
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-o.ctx.Done():
			return
		case <-ticker.C:
			if !o.wsClient.IsConnected() {
				log.Warn().Msg("Binance WebSocket disconnected, will auto-reconnect")
			}
		}
	}
}

// OnKline handles kline events from Binance WebSocket
func (h *BinanceWSHandler) OnKline(event binance.KlineEvent) {
	if h.orchestrator == nil {
		return
	}
	h.orchestrator.processKlineUpdate(&event)
}

// OnTrade handles trade events from Binance WebSocket (real-time price)
func (h *BinanceWSHandler) OnTrade(event binance.TradeEvent) {
	if h.orchestrator == nil {
		return
	}

	price, err := strconv.ParseFloat(event.Price, 64)
	if err != nil {
		return
	}

	now := time.Now()
	h.orchestrator.stateMu.Lock()
	h.orchestrator.state.CurrentPrice = price
	h.orchestrator.state.LastUpdate = now
	h.orchestrator.stateMu.Unlock()

	// Update executor price cache (for paper trading)
	if paperExec, ok := h.orchestrator.executor.(*execution.PaperExecutor); ok {
		paperExec.UpdatePrice(event.Symbol, price)
	}

	// Broadcast price immediately for real-time updates
	h.orchestrator.broadcast(BroadcastMessage{
		Type:      MessageTypePrice,
		Timestamp: now,
		Data: PriceUpdate{
			Symbol:    event.Symbol,
			Price:     price,
			Timestamp: now,
		},
	})
}

// OnDepth handles depth events (not used for now)
func (h *BinanceWSHandler) OnDepth(event binance.DepthEvent) {}

// OnMiniTicker handles mini ticker events (not used for now)
func (h *BinanceWSHandler) OnMiniTicker(event binance.MiniTickerEvent) {}

// OnError handles WebSocket errors
func (h *BinanceWSHandler) OnError(err error) {
	log.Error().Err(err).Msg("Binance WebSocket error")
}

// OnDisconnect handles WebSocket disconnection
func (h *BinanceWSHandler) OnDisconnect() {
	log.Warn().Msg("Binance WebSocket disconnected")
}

// OnReconnect handles WebSocket reconnection
func (h *BinanceWSHandler) OnReconnect() {
	log.Info().Msg("Binance WebSocket reconnected")
}

// CreateWSHandler creates a WebSocket handler for this orchestrator
func (o *Orchestrator) CreateWSHandler() *BinanceWSHandler {
	return NewBinanceWSHandler(o)
}

// pollPriceFallback polls price using REST API as a fallback
func (o *Orchestrator) pollPriceFallback() {
	defer o.wg.Done()

	log.Info().Msg("Started REST API price polling (fallback mode)")

	priceTicker := time.NewTicker(2 * time.Second) // Poll price every 2s
	klineTicker := time.NewTicker(15 * time.Second) // Poll klines every 15s and run trading logic
	defer priceTicker.Stop()
	defer klineTicker.Stop()

	for {
		select {
		case <-o.ctx.Done():
			return
		case <-priceTicker.C:
			if o.binanceClient != nil {
				tickerPrice, err := o.binanceClient.GetTickerPrice(o.config.Symbol)
				if err != nil {
					log.Debug().Err(err).Msg("Failed to fetch price")
					continue
				}

				price, err := strconv.ParseFloat(tickerPrice.Price, 64)
				if err != nil {
					log.Debug().Err(err).Msg("Failed to parse price")
					continue
				}

				o.stateMu.Lock()
				o.state.CurrentPrice = price
				o.state.LastUpdate = time.Now()
				o.stateMu.Unlock()

				// Broadcast price update
				o.broadcast(BroadcastMessage{
					Type:      MessageTypePrice,
					Timestamp: time.Now(),
					Data: PriceUpdate{
						Symbol:    o.config.Symbol,
						Price:     price,
						Timestamp: time.Now(),
					},
				})
			}
		case <-klineTicker.C:
			// Fetch latest klines and run trading logic
			o.pollKlinesAndTrade()
		}
	}
}

// pollKlinesAndTrade fetches latest klines and runs trading logic
func (o *Orchestrator) pollKlinesAndTrade() {
	if o.binanceClient == nil || o.dataService == nil {
		return
	}

	// Fetch last 10 klines for primary timeframe
	klines, err := o.binanceClient.GetKlines(o.config.Symbol, o.config.PrimaryTimeframe, 10, 0, 0)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to fetch klines for trading")
		return
	}

	if len(klines) == 0 {
		return
	}

	// Get the most recent closed candle
	var latestClosed *binance.Kline
	now := time.Now().UnixMilli()
	for i := len(klines) - 1; i >= 0; i-- {
		if klines[i].CloseTime < now {
			latestClosed = &klines[i]
			break
		}
	}

	if latestClosed == nil {
		return
	}

	// Convert and store
	candle := convertKlineToCandle(*latestClosed, o.config.Symbol, o.config.PrimaryTimeframe)
	candle.IsClosed = true

	// Check if this candle is new (not already processed)
	existingCandles := o.dataService.GetLastCandles(o.config.Symbol, o.config.PrimaryTimeframe, 1)
	if len(existingCandles) > 0 {
		lastTime := existingCandles[0].CloseTime
		if !candle.CloseTime.After(lastTime) {
			// Already processed, but still run trading logic periodically
			o.processTradingLogic()
			return
		}
	}

	// Add new candle
	o.dataService.AddCandle(*candle)

	// Update state
	o.stateMu.Lock()
	o.state.CandleCount++
	o.state.LastCandleTime = candle.CloseTime
	closePrice := candle.Close
	o.state.CurrentPrice = closePrice
	o.stateMu.Unlock()

	// Broadcast candle
	o.broadcast(BroadcastMessage{
		Type:      MessageTypeCandle,
		Timestamp: time.Now(),
		Data: CandleUpdate{
			Symbol:    candle.Symbol,
			Timeframe: candle.Timeframe,
			Timestamp: candle.OpenTime,
			Open:      candle.Open,
			High:      candle.High,
			Low:       candle.Low,
			Close:     candle.Close,
			Volume:    candle.Volume,
			IsClosed:  true,
		},
	})

	log.Debug().
		Str("timeframe", o.config.PrimaryTimeframe).
		Float64("close", closePrice).
		Time("time", candle.CloseTime).
		Msg("Processed new kline via REST polling")

	// Run trading logic
	o.processTradingLogic()
}

// handleWebSocketMessage handles incoming WebSocket messages
func (o *Orchestrator) handleWebSocketMessage(data []byte) {
	// Parse kline event
	var event binance.KlineEvent
	// This would use JSON unmarshaling in real implementation

	if event.EventType == "kline" {
		o.processKlineUpdate(&event)
	}
}

// processKlineUpdate processes a kline update
func (o *Orchestrator) processKlineUpdate(event *binance.KlineEvent) {
	kd := event.Kline

	// Convert to storage candle
	open, _ := strconv.ParseFloat(kd.Open, 64)
	high, _ := strconv.ParseFloat(kd.High, 64)
	low, _ := strconv.ParseFloat(kd.Low, 64)
	closePrice, _ := strconv.ParseFloat(kd.Close, 64)
	volume, _ := strconv.ParseFloat(kd.Volume, 64)

	candle := &storage.Candle{
		Symbol:    kd.Symbol,
		Timeframe: kd.Interval,
		OpenTime:  time.UnixMilli(kd.StartTime),
		CloseTime: time.UnixMilli(kd.CloseTime),
		Open:      open,
		High:      high,
		Low:       low,
		Close:     closePrice,
		Volume:    volume,
	}

	// Update current price
	o.stateMu.Lock()
	o.state.CurrentPrice = closePrice
	o.state.LastUpdate = time.Now()
	o.stateMu.Unlock()

	// Broadcast candle update
	o.broadcast(BroadcastMessage{
		Type:      MessageTypeCandle,
		Timestamp: time.Now(),
		Data: CandleUpdate{
			Symbol:    candle.Symbol,
			Timeframe: candle.Timeframe,
			Timestamp: candle.OpenTime,
			Open:      candle.Open,
			High:      candle.High,
			Low:       candle.Low,
			Close:     candle.Close,
			Volume:    candle.Volume,
			IsClosed:  kd.IsClosed,
		},
	})

	// If candle is closed
	if kd.IsClosed {
		candle.IsClosed = true
		// Add to data service
		o.dataService.AddCandle(*candle)

		// Update state
		o.stateMu.Lock()
		o.state.CandleCount++
		o.state.LastCandleTime = candle.CloseTime
		o.stateMu.Unlock()

		// Process trading logic on primary timeframe
		if kd.Interval == o.config.PrimaryTimeframe {
			o.processTradingLogic()
		}
	}
}

// processTradingLogic runs the main trading logic
func (o *Orchestrator) processTradingLogic() {
	// Get market data
	marketData := o.buildMarketData()
	if marketData == nil {
		return
	}

	// Check if trading is halted
	if o.riskManager != nil && o.riskManager.IsHalted() {
		return
	}

	// Run analysis through strategy manager
	if o.strategyMgr == nil {
		return
	}

	opens, highs, lows, closes, volumes := o.dataService.GetOHLCV(o.config.Symbol, o.config.PrimaryTimeframe)
	if len(closes) < 50 {
		return
	}

	currentPrice := closes[len(closes)-1]
	analysis := o.strategyMgr.Analyze(o.config.Symbol, o.config.PrimaryTimeframe, opens, highs, lows, closes, volumes, currentPrice)
	if analysis == nil {
		return
	}

	// Update regime in state
	o.stateMu.Lock()
	o.state.CurrentRegime = analysis.Regime.Regime.String()
	o.stateMu.Unlock()

	// Check if we have a trade recommendation
	rec := analysis.Recommendation
	if rec.Action == strategy.ActionNone {
		return
	}

	// Create signal from recommendation
	bestSignal := strategy.Signal{
		Type:       strategy.SignalTypeEntry,
		Direction:  rec.Direction,
		Price:      rec.Price,
		StopLoss:   rec.StopLoss,
		TakeProfit: rec.TakeProfit,
		Confidence: rec.Confidence,
		Reason:     rec.Reason,
		Strategy:   rec.Strategy,
		Symbol:     o.config.Symbol,
		Timeframe:  o.config.PrimaryTimeframe,
	}

	log.Info().
		Str("direction", rec.Direction.String()).
		Str("strategy", rec.Strategy).
		Float64("price", rec.Price).
		Float64("confidence", rec.Confidence).
		Msg("Signal generated")

	// Risk assessment
	var approved bool
	var rejectReason string
	if o.riskManager != nil {
		assessment := o.riskManager.AssessTrade(risk.TradeParams{
			Symbol:     bestSignal.Symbol,
			Direction:  bestSignal.Direction.String(),
			EntryPrice: bestSignal.Price,
			StopLoss:   bestSignal.StopLoss,
			TakeProfit: bestSignal.TakeProfit,
		})
		approved = assessment.Approved
		if !approved && len(assessment.Reasons) > 0 {
			rejectReason = assessment.Reasons[0]
			log.Warn().
				Str("strategy", rec.Strategy).
				Str("reason", rejectReason).
				Msg("Signal rejected by risk manager")
		} else {
			log.Debug().
				Str("strategy", rec.Strategy).
				Bool("approved", approved).
				Msg("Signal approved by risk manager")
		}
	} else {
		approved = true
	}

	// Broadcast signal
	o.broadcast(BroadcastMessage{
		Type:      MessageTypeSignal,
		Timestamp: time.Now(),
		Data: SignalUpdate{
			Signal:     &bestSignal,
			Approved:   approved,
			RejectedBy: "RiskManager",
			Reason:     rejectReason,
		},
	})

	o.stateMu.Lock()
	o.state.LastSignal = &bestSignal
	o.stateMu.Unlock()

	// Store signal in history
	o.addSignal(&bestSignal, approved, rejectReason)

	// Execute if approved
	if approved {
		o.executeSignal(bestSignal)
	}
}

// buildMarketData builds market data for strategies
func (o *Orchestrator) buildMarketData() *strategy.MarketData {
	// Get recent candles from data service
	candles := o.dataService.GetLastCandles(o.config.Symbol, o.config.PrimaryTimeframe, 200)
	if len(candles) < 50 {
		return nil
	}

	// Build price arrays
	opens := make([]float64, len(candles))
	highs := make([]float64, len(candles))
	lows := make([]float64, len(candles))
	closes := make([]float64, len(candles))
	volumes := make([]float64, len(candles))

	for i, c := range candles {
		opens[i] = c.Open
		highs[i] = c.High
		lows[i] = c.Low
		closes[i] = c.Close
		volumes[i] = c.Volume
	}

	lastCandle := candles[len(candles)-1]

	// Calculate indicators
	var analysisResult indicators.AnalysisResult
	if o.indicatorMgr != nil {
		analysisResult = o.indicatorMgr.Analyze(opens, highs, lows, closes, volumes)

		// Broadcast indicators
		o.broadcastIndicators(&analysisResult, lastCandle.CloseTime)
	}

	return &strategy.MarketData{
		Symbol:       o.config.Symbol,
		Timeframe:    o.config.PrimaryTimeframe,
		Timestamp:    lastCandle.CloseTime,
		Opens:        opens,
		Highs:        highs,
		Lows:         lows,
		Closes:       closes,
		Volumes:      volumes,
		CurrentPrice: lastCandle.Close,
		Analysis:     analysisResult,
	}
}

// executeSignal executes a trading signal
func (o *Orchestrator) executeSignal(signal strategy.Signal) {
	// Determine order side
	side := execution.OrderSideBuy
	if signal.Direction == strategy.DirectionShort {
		side = execution.OrderSideSell
	}

	// Calculate position size from risk manager
	var quantity float64
	if o.riskManager != nil {
		sizer := o.riskManager.GetPositionSizer()
		equity, _ := o.executor.GetEquity()
		result := sizer.CalculateSize(risk.PositionSizeParams{
			Equity:     equity,
			EntryPrice: signal.Price,
			StopLoss:   signal.StopLoss,
			TakeProfit: signal.TakeProfit,
			Direction:  signal.Direction.String(),
		})
		quantity = result.Size

		log.Debug().
			Float64("equity", equity).
			Float64("entryPrice", signal.Price).
			Float64("stopLoss", signal.StopLoss).
			Float64("quantity", quantity).
			Float64("riskPercent", result.RiskPercent).
			Msg("Position size calculated")
	} else {
		// Default sizing
		equity, _ := o.executor.GetEquity()
		quantity = (equity * 0.1) / signal.Price
	}

	if quantity <= 0 {
		log.Warn().
			Str("strategy", signal.Strategy).
			Float64("quantity", quantity).
			Float64("stopLoss", signal.StopLoss).
			Msg("Order skipped: Invalid position size")
		return
	}

	// Create order
	order := &execution.Order{
		Symbol:   signal.Symbol,
		Side:     side,
		Type:     execution.OrderTypeMarket,
		Quantity: quantity,
		Strategy: signal.Strategy,
		Signal:   &signal,
	}

	// Execute
	result, err := o.executor.PlaceOrder(order)
	if err != nil {
		log.Error().Err(err).Msg("Failed to execute order")
		o.broadcastError("ORDER_FAILED", "Failed to execute order", err.Error())
		return
	}

	if result.Success {
		log.Info().
			Str("orderID", result.Order.ID).
			Str("strategy", signal.Strategy).
			Float64("quantity", quantity).
			Msg("Order executed")

		// Set stop loss and take profit
		if result.Position != nil {
			if signal.StopLoss > 0 {
				o.executor.UpdateStopLoss(result.Position.ID, signal.StopLoss)
			}
			if signal.TakeProfit > 0 {
				o.executor.UpdateTakeProfit(result.Position.ID, signal.TakeProfit)
			}
		}
	}
}

// setupExecutorCallbacks sets up callbacks for executor events
func (o *Orchestrator) setupExecutorCallbacks() {
	// Set fill callback for paper executor
	if paperExec, ok := o.executor.(*execution.PaperExecutor); ok {
		paperExec.SetOnFill(func(event execution.FillEvent) {
			o.broadcast(BroadcastMessage{
				Type:      MessageTypeTrade,
				Timestamp: time.Now(),
				Data: TradeUpdate{
					TradeID:    event.TradeID,
					OrderID:    event.OrderID,
					Symbol:     event.Symbol,
					Side:       event.Side,
					Quantity:   event.Quantity,
					Price:      event.Price,
					Commission: event.Commission,
					Timestamp:  event.Timestamp,
				},
			})

			// Update trade stats in state
			o.updateTradeStats()
		})

		paperExec.SetOnPosition(func(event execution.PositionEvent) {
			o.broadcast(BroadcastMessage{
				Type:      MessageTypePosition,
				Timestamp: time.Now(),
				Data: PositionUpdate{
					PositionID:    event.Position.ID,
					Symbol:        event.Position.Symbol,
					Side:          event.Position.Side,
					Quantity:      event.Position.Quantity,
					EntryPrice:    event.Position.EntryPrice,
					CurrentPrice:  event.Position.CurrentPrice,
					StopLoss:      event.Position.StopLoss,
					TakeProfit:    event.Position.TakeProfit,
					UnrealizedPnL: event.Position.UnrealizedPnL,
					RealizedPnL:   event.Position.RealizedPnL,
					Strategy:      event.Position.Strategy,
					OpenTime:      event.Position.OpenTime,
					EventType:     event.Type.String(),
				},
			})
		})
	}
}

// updateTradeStats updates trading statistics in state
func (o *Orchestrator) updateTradeStats() {
	if paperExec, ok := o.executor.(*execution.PaperExecutor); ok {
		stats := paperExec.GetStats()

		o.stateMu.Lock()
		o.state.TotalTrades = stats.TotalTrades
		o.state.WinRate = stats.WinRate
		o.stateMu.Unlock()
	}
}

// broadcastLoop sends periodic state updates
func (o *Orchestrator) broadcastLoop() {
	defer o.wg.Done()

	ticker := time.NewTicker(o.config.BroadcastInterval)
	defer ticker.Stop()

	for {
		select {
		case <-o.ctx.Done():
			return
		case <-ticker.C:
			o.broadcastState()
		}
	}
}

// riskMonitorLoop monitors risk metrics
func (o *Orchestrator) riskMonitorLoop() {
	defer o.wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-o.ctx.Done():
			return
		case <-ticker.C:
			o.updateRiskMetrics()
		}
	}
}

// updateRiskMetrics updates risk metrics
func (o *Orchestrator) updateRiskMetrics() {
	if o.riskManager == nil || o.executor == nil {
		return
	}

	// Get current equity
	equity, err := o.executor.GetEquity()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get equity")
		return
	}

	log.Debug().
		Float64("equity", equity).
		Msg("Updating risk metrics")

	// Get positions
	positions, _ := o.executor.GetPositions()
	openPositions := len(positions)

	// Calculate unrealized P&L
	var unrealizedPnL float64
	for _, pos := range positions {
		unrealizedPnL += pos.UnrealizedPnL
	}

	// Get account state from paper executor
	dailyPnL := 0.0
	weeklyPnL := 0.0
	if paperExec, ok := o.executor.(*execution.PaperExecutor); ok {
		summary := paperExec.GetAccountSummary()
		dailyPnL = summary.RealizedPnL // Simplified
	}

	// Update risk manager
	o.riskManager.UpdateAccountState(equity, equity, unrealizedPnL, dailyPnL, weeklyPnL, openPositions)

	// Check circuit breaker
	o.riskManager.CheckCircuitBreaker()

	// Update state
	state := o.riskManager.GetAccountState()
	limits := o.riskManager.GetRiskLimits()

	o.stateMu.Lock()
	o.state.Equity = equity
	o.state.AvailableBalance = equity - unrealizedPnL
	o.state.UnrealizedPnL = unrealizedPnL
	o.state.CurrentDrawdown = state.CurrentDrawdown
	o.state.MaxDrawdown = o.config.InitialCapital * 0.2 // From config
	o.state.OpenPositions = openPositions
	o.state.IsHalted = state.IsHalted
	o.state.HaltReason = state.HaltReason
	o.stateMu.Unlock()

	// Broadcast risk update
	o.broadcast(BroadcastMessage{
		Type:      MessageTypeRisk,
		Timestamp: time.Now(),
		Data: RiskUpdate{
			Level:           o.determineRiskLevel(state.CurrentDrawdown),
			Drawdown:        state.CurrentDrawdown,
			MaxDrawdown:     o.config.InitialCapital * 0.2,
			DailyLossUsed:   limits.DailyLossUsed,
			DailyLossLimit:  limits.DailyLossLimit,
			WeeklyLossUsed:  limits.WeeklyLossUsed,
			WeeklyLossLimit: limits.WeeklyLossLimit,
			IsHalted:        state.IsHalted,
			HaltReason:      state.HaltReason,
		},
	})
}

// determineRiskLevel determines risk level from drawdown
func (o *Orchestrator) determineRiskLevel(drawdown float64) risk.RiskLevel {
	switch {
	case drawdown >= 0.15:
		return risk.RiskCritical
	case drawdown >= 0.10:
		return risk.RiskHigh
	case drawdown >= 0.05:
		return risk.RiskMedium
	default:
		return risk.RiskLow
	}
}

// broadcastState broadcasts current state
func (o *Orchestrator) broadcastState() {
	o.stateMu.RLock()
	state := *o.state
	o.stateMu.RUnlock()

	summary := o.getAccountSummary()

	o.broadcast(BroadcastMessage{
		Type:      MessageTypeState,
		Timestamp: time.Now(),
		Data: StateUpdate{
			State:   &state,
			Summary: summary,
		},
	})
}

// getAccountSummary gets account summary
func (o *Orchestrator) getAccountSummary() *AccountSummary {
	summary := &AccountSummary{}

	if o.executor == nil {
		return summary
	}

	equity, _ := o.executor.GetEquity()
	summary.Equity = equity

	positions, _ := o.executor.GetPositions()
	summary.OpenPositions = len(positions)

	for _, pos := range positions {
		summary.UnrealizedPnL += pos.UnrealizedPnL
	}

	if paperExec, ok := o.executor.(*execution.PaperExecutor); ok {
		stats := paperExec.GetStats()
		summary.TotalTrades = stats.TotalTrades
		summary.WinningTrades = stats.WinningTrades
		summary.LosingTrades = stats.LosingTrades
		summary.WinRate = stats.WinRate
		summary.ProfitFactor = stats.ProfitFactor
		summary.RealizedPnL = stats.NetProfit

		accSummary := paperExec.GetAccountSummary()
		summary.AvailableBalance = accSummary.AvailableBalance
	}

	if o.config.InitialCapital > 0 {
		summary.TotalReturn = (equity - o.config.InitialCapital) / o.config.InitialCapital
	}

	return summary
}

// broadcastIndicators broadcasts indicator values
func (o *Orchestrator) broadcastIndicators(result *indicators.AnalysisResult, timestamp time.Time) {
	if result == nil {
		return
	}

	update := IndicatorsUpdate{
		Symbol:    o.config.Symbol,
		Timeframe: o.config.PrimaryTimeframe,
		Timestamp: timestamp,
		RSI:       result.RSI.Value,
		MACD: &MACDValue{
			MACD:      result.MACD.MACD,
			Signal:    result.MACD.Signal,
			Histogram: result.MACD.Histogram,
		},
		BB: &BollingerValue{
			Upper:  result.Bollinger.Upper,
			Middle: result.Bollinger.Middle,
			Lower:  result.Bollinger.Lower,
			Width:  result.Bollinger.Width,
		},
		ADX: &ADXValue{
			ADX:     result.ADX.ADX,
			PlusDI:  result.ADX.PlusDI,
			MinusDI: result.ADX.MinusDI,
		},
		ATR: result.ATR.ATR,
	}

	o.broadcast(BroadcastMessage{
		Type:      MessageTypeIndicators,
		Timestamp: time.Now(),
		Data:      update,
	})
}

// broadcastRiskEvent broadcasts a risk event
func (o *Orchestrator) broadcastRiskEvent(event risk.RiskEvent) {
	o.broadcast(BroadcastMessage{
		Type:      MessageTypeRisk,
		Timestamp: time.Now(),
		Data: RiskUpdate{
			Level:      event.Level,
			IsHalted:   event.Type == risk.RiskEventCircuitBreaker,
			HaltReason: event.Message,
			Events:     []risk.RiskEvent{event},
		},
	})
}

// broadcastError broadcasts an error
func (o *Orchestrator) broadcastError(code, message, details string) {
	o.broadcast(BroadcastMessage{
		Type:      MessageTypeError,
		Timestamp: time.Now(),
		Data: ErrorUpdate{
			Code:    code,
			Message: message,
			Details: details,
			Time:    time.Now(),
		},
	})
}

// broadcast sends a message to all subscribers
func (o *Orchestrator) broadcast(msg BroadcastMessage) {
	if o.broadcaster != nil {
		o.broadcaster.Broadcast(msg)
	}
}

// Subscribe adds a subscriber
func (o *Orchestrator) Subscribe(id string) chan BroadcastMessage {
	if o.broadcaster != nil {
		return o.broadcaster.Subscribe(id)
	}
	return nil
}

// Unsubscribe removes a subscriber
func (o *Orchestrator) Unsubscribe(id string) {
	if o.broadcaster != nil {
		o.broadcaster.Unsubscribe(id)
	}
}

// GetState returns current state
func (o *Orchestrator) GetState() *TradingState {
	o.stateMu.RLock()
	defer o.stateMu.RUnlock()
	state := *o.state
	return &state
}

// GetSignals returns recent signals (up to limit)
func (o *Orchestrator) GetSignals(limit int) []SignalRecord {
	o.signalsMu.RLock()
	defer o.signalsMu.RUnlock()

	if limit <= 0 || limit > len(o.signals) {
		limit = len(o.signals)
	}

	// Return most recent signals first
	result := make([]SignalRecord, limit)
	for i := 0; i < limit; i++ {
		result[i] = o.signals[len(o.signals)-1-i]
	}
	return result
}

// addSignal adds a signal to history (keeps last 50)
func (o *Orchestrator) addSignal(signal *strategy.Signal, approved bool, reason string) {
	o.signalsMu.Lock()
	defer o.signalsMu.Unlock()

	record := SignalRecord{
		Signal:     signal,
		Approved:   approved,
		Reason:     reason,
		ReceivedAt: time.Now(),
	}

	o.signals = append(o.signals, record)

	// Keep only last 50 signals
	if len(o.signals) > 50 {
		o.signals = o.signals[len(o.signals)-50:]
	}
}

// GetCandles returns candles from the data service
func (o *Orchestrator) GetCandles(symbol, timeframe string, limit int) []storage.Candle {
	if o.dataService == nil {
		return nil
	}
	candles := o.dataService.GetCandles(symbol, timeframe)
	if limit > 0 && len(candles) > limit {
		return candles[len(candles)-limit:]
	}
	return candles
}

// Pause pauses trading
func (o *Orchestrator) Pause() {
	o.stateMu.Lock()
	o.state.IsPaused = true
	o.stateMu.Unlock()
	log.Info().Msg("Trading paused")
}

// Resume resumes trading
func (o *Orchestrator) Resume() {
	o.stateMu.Lock()
	o.state.IsPaused = false
	o.stateMu.Unlock()
	log.Info().Msg("Trading resumed")
}

// convertKlineToCandle converts a Binance kline to storage candle
func convertKlineToCandle(k binance.Kline, symbol, timeframe string) *storage.Candle {
	open, _ := strconv.ParseFloat(k.Open, 64)
	high, _ := strconv.ParseFloat(k.High, 64)
	low, _ := strconv.ParseFloat(k.Low, 64)
	closePrice, _ := strconv.ParseFloat(k.Close, 64)
	volume, _ := strconv.ParseFloat(k.Volume, 64)

	return &storage.Candle{
		Symbol:    symbol,
		Timeframe: timeframe,
		OpenTime:  time.UnixMilli(k.OpenTime),
		CloseTime: time.UnixMilli(k.CloseTime),
		Open:      open,
		High:      high,
		Low:       low,
		Close:     closePrice,
		Volume:    volume,
	}
}
