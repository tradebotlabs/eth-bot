package storage

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// DataService coordinates between in-memory queues and SQLite database
type DataService struct {
	db              *SQLiteDB
	queueManager    *QueueManager
	candleRepo      *CandleRepository
	tradeRepo       *TradeRepository
	positionRepo    *PositionRepository
	accountRepo     *AccountRepository
	alertRepo       *AlertRepository
	backtestRepo    *BacktestRepository
	strategyPerfRepo *StrategyPerformanceRepository

	// Persistence settings
	persistInterval time.Duration
	pendingCandles  []Candle
	pendingMu       sync.Mutex

	// State
	running bool
	cancel  context.CancelFunc
}

// NewDataService creates a new data service
func NewDataService(db *SQLiteDB, persistInterval time.Duration, capacities map[string]int) *DataService {
	if persistInterval <= 0 {
		persistInterval = 10 * time.Second
	}

	defaultCapacity := 200
	if capacities == nil {
		capacities = DefaultCapacities
	}

	return &DataService{
		db:               db,
		queueManager:     NewQueueManager(defaultCapacity, capacities),
		candleRepo:       NewCandleRepository(db),
		tradeRepo:        NewTradeRepository(db),
		positionRepo:     NewPositionRepository(db),
		accountRepo:      NewAccountRepository(db),
		alertRepo:        NewAlertRepository(db),
		backtestRepo:     NewBacktestRepository(db),
		strategyPerfRepo: NewStrategyPerformanceRepository(db),
		persistInterval:  persistInterval,
		pendingCandles:   make([]Candle, 0, 100),
	}
}

// Start starts the background persistence goroutine
func (ds *DataService) Start(ctx context.Context) {
	if ds.running {
		return
	}

	ctx, ds.cancel = context.WithCancel(ctx)
	ds.running = true

	go ds.persistenceLoop(ctx)
	log.Info().Dur("interval", ds.persistInterval).Msg("Data service started")
}

// Stop stops the data service
func (ds *DataService) Stop() {
	if !ds.running {
		return
	}

	ds.cancel()
	ds.running = false

	// Final flush
	ds.flushPendingCandles()
	log.Info().Msg("Data service stopped")
}

// persistenceLoop runs the background persistence
func (ds *DataService) persistenceLoop(ctx context.Context) {
	ticker := time.NewTicker(ds.persistInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			ds.flushPendingCandles()
			return
		case <-ticker.C:
			ds.flushPendingCandles()
		}
	}
}

// flushPendingCandles writes pending candles to SQLite
func (ds *DataService) flushPendingCandles() {
	ds.pendingMu.Lock()
	if len(ds.pendingCandles) == 0 {
		ds.pendingMu.Unlock()
		return
	}
	candles := ds.pendingCandles
	ds.pendingCandles = make([]Candle, 0, 100)
	ds.pendingMu.Unlock()

	if err := ds.candleRepo.InsertBatch(candles); err != nil {
		log.Error().Err(err).Int("count", len(candles)).Msg("Failed to persist candles")
		// Re-queue failed candles
		ds.pendingMu.Lock()
		ds.pendingCandles = append(candles, ds.pendingCandles...)
		ds.pendingMu.Unlock()
	} else {
		log.Debug().Int("count", len(candles)).Msg("Persisted candles to SQLite")
	}
}

// AddCandle adds a candle to both in-memory queue and persistence queue
func (ds *DataService) AddCandle(candle Candle) {
	// Add to in-memory queue for fast access
	queue := ds.queueManager.GetOrCreate(candle.Symbol, candle.Timeframe)

	// If candle is not closed, update the latest candle
	if !candle.IsClosed {
		if latest, ok := queue.GetLatest(); ok && latest.OpenTime.Equal(candle.OpenTime) {
			queue.UpdateLatest(candle)
			return
		}
	}

	queue.Push(candle)

	// Queue for async persistence (only closed candles)
	if candle.IsClosed {
		ds.pendingMu.Lock()
		ds.pendingCandles = append(ds.pendingCandles, candle)
		ds.pendingMu.Unlock()
	}
}

// UpdateCandle updates the latest candle in the queue
func (ds *DataService) UpdateCandle(candle Candle) bool {
	return ds.queueManager.UpdateLatestCandle(candle)
}

// GetCandles returns candles from in-memory queue
func (ds *DataService) GetCandles(symbol, timeframe string) []Candle {
	return ds.queueManager.GetCandles(symbol, timeframe)
}

// GetLastCandles returns the last N candles from in-memory queue
func (ds *DataService) GetLastCandles(symbol, timeframe string, n int) []Candle {
	return ds.queueManager.GetLastCandles(symbol, timeframe, n)
}

// GetLatestCandle returns the most recent candle
func (ds *DataService) GetLatestCandle(symbol, timeframe string) (Candle, bool) {
	return ds.queueManager.GetLatestCandle(symbol, timeframe)
}

// GetOHLCV returns OHLCV data for indicator calculations
func (ds *DataService) GetOHLCV(symbol, timeframe string) (opens, highs, lows, closes, volumes []float64) {
	return ds.queueManager.GetOHLCV(symbol, timeframe)
}

// GetCloses returns close prices for indicator calculations
func (ds *DataService) GetCloses(symbol, timeframe string) []float64 {
	queue, exists := ds.queueManager.Get(symbol, timeframe)
	if !exists {
		return nil
	}
	return queue.GetCloses()
}

// GetHighs returns high prices
func (ds *DataService) GetHighs(symbol, timeframe string) []float64 {
	queue, exists := ds.queueManager.Get(symbol, timeframe)
	if !exists {
		return nil
	}
	return queue.GetHighs()
}

// GetLows returns low prices
func (ds *DataService) GetLows(symbol, timeframe string) []float64 {
	queue, exists := ds.queueManager.Get(symbol, timeframe)
	if !exists {
		return nil
	}
	return queue.GetLows()
}

// GetVolumes returns volumes
func (ds *DataService) GetVolumes(symbol, timeframe string) []float64 {
	queue, exists := ds.queueManager.Get(symbol, timeframe)
	if !exists {
		return nil
	}
	return queue.GetVolumes()
}

// HasEnoughData checks if there's enough data for indicators
func (ds *DataService) HasEnoughData(symbol, timeframe string, n int) bool {
	return ds.queueManager.HasEnoughData(symbol, timeframe, n)
}

// LoadHistoricalCandles loads candles from SQLite into memory queues
func (ds *DataService) LoadHistoricalCandles(symbol, timeframe string) error {
	capacity := ds.queueManager.GetCapacity(timeframe)

	// Load from SQLite
	candles, err := ds.candleRepo.GetLast(symbol, timeframe, capacity)
	if err != nil {
		return err
	}

	if len(candles) == 0 {
		log.Debug().Str("symbol", symbol).Str("timeframe", timeframe).Msg("No historical candles found")
		return nil
	}

	// Create queue and populate
	queue := ds.queueManager.GetOrCreate(symbol, timeframe)
	for _, candle := range candles {
		queue.Push(candle)
	}

	log.Info().
		Str("symbol", symbol).
		Str("timeframe", timeframe).
		Int("count", len(candles)).
		Msg("Loaded historical candles")

	return nil
}

// LoadAllHistoricalCandles loads candles for all timeframes
func (ds *DataService) LoadAllHistoricalCandles(symbol string, timeframes []string) error {
	for _, tf := range timeframes {
		if err := ds.LoadHistoricalCandles(symbol, tf); err != nil {
			log.Error().Err(err).Str("timeframe", tf).Msg("Failed to load historical candles")
		}
	}
	return nil
}

// GetHistoricalCandles retrieves candles from SQLite for a date range
func (ds *DataService) GetHistoricalCandles(symbol, timeframe string, from, to time.Time) ([]Candle, error) {
	return ds.candleRepo.GetRange(symbol, timeframe, from, to)
}

// Trade methods

// AddTrade persists a trade
func (ds *DataService) AddTrade(trade Trade) error {
	return ds.tradeRepo.Insert(trade)
}

// GetTrades retrieves trades for a symbol
func (ds *DataService) GetTrades(symbol string, limit int) ([]Trade, error) {
	return ds.tradeRepo.GetBySymbol(symbol, limit)
}

// GetTradesByStrategy retrieves trades for a strategy
func (ds *DataService) GetTradesByStrategy(strategy string, limit int) ([]Trade, error) {
	return ds.tradeRepo.GetByStrategy(strategy, limit)
}

// GetTradesByDateRange retrieves trades within a date range
func (ds *DataService) GetTradesByDateRange(from, to time.Time) ([]Trade, error) {
	return ds.tradeRepo.GetByDateRange(from, to)
}

// Position methods

// AddPosition creates a new position
func (ds *DataService) AddPosition(pos Position) (int64, error) {
	return ds.positionRepo.Insert(pos)
}

// UpdatePosition updates a position
func (ds *DataService) UpdatePosition(pos Position) error {
	return ds.positionRepo.Update(pos)
}

// GetOpenPositions retrieves all open positions
func (ds *DataService) GetOpenPositions() ([]Position, error) {
	return ds.positionRepo.GetOpen()
}

// GetPosition retrieves a position by ID
func (ds *DataService) GetPosition(id int64) (*Position, error) {
	return ds.positionRepo.GetByID(id)
}

// GetClosedPositions retrieves closed positions
func (ds *DataService) GetClosedPositions(limit int) ([]Position, error) {
	return ds.positionRepo.GetClosed(limit)
}

// Account methods

// AddAccountSnapshot persists an account snapshot
func (ds *DataService) AddAccountSnapshot(snapshot AccountSnapshot) error {
	return ds.accountRepo.InsertSnapshot(snapshot)
}

// GetLatestSnapshot retrieves the most recent account snapshot
func (ds *DataService) GetLatestSnapshot() (*AccountSnapshot, error) {
	return ds.accountRepo.GetLatestSnapshot()
}

// GetAccountHistory retrieves account snapshots for a date range
func (ds *DataService) GetAccountHistory(from, to time.Time) ([]AccountSnapshot, error) {
	return ds.accountRepo.GetSnapshotsRange(from, to)
}

// Alert methods

// AddAlert creates a new alert
func (ds *DataService) AddAlert(alert Alert) (int64, error) {
	return ds.alertRepo.Insert(alert)
}

// GetUnacknowledgedAlerts retrieves unacknowledged alerts
func (ds *DataService) GetUnacknowledgedAlerts(limit int) ([]Alert, error) {
	return ds.alertRepo.GetUnacknowledged(limit)
}

// AcknowledgeAlert marks an alert as acknowledged
func (ds *DataService) AcknowledgeAlert(id int64) error {
	return ds.alertRepo.Acknowledge(id)
}

// Strategy Performance methods

// UpdateStrategyPerformance updates strategy performance metrics
func (ds *DataService) UpdateStrategyPerformance(perf StrategyPerformance) error {
	return ds.strategyPerfRepo.Upsert(perf)
}

// GetStrategyPerformance retrieves performance records for a strategy
func (ds *DataService) GetStrategyPerformance(strategy string, limit int) ([]StrategyPerformance, error) {
	return ds.strategyPerfRepo.GetByStrategy(strategy, limit)
}

// Backtest methods

// CreateBacktestRun creates a new backtest run
func (ds *DataService) CreateBacktestRun(run BacktestRun) (int64, error) {
	return ds.backtestRepo.InsertRun(run)
}

// UpdateBacktestRun updates a backtest run with results
func (ds *DataService) UpdateBacktestRun(run BacktestRun) error {
	return ds.backtestRepo.UpdateRun(run)
}

// GetBacktestRun retrieves a backtest run by ID
func (ds *DataService) GetBacktestRun(id int64) (*BacktestRun, error) {
	return ds.backtestRepo.GetRun(id)
}

// GetBacktestRuns retrieves backtest runs
func (ds *DataService) GetBacktestRuns(limit int) ([]BacktestRun, error) {
	return ds.backtestRepo.GetRuns(limit)
}

// DeleteBacktestRun deletes a backtest run
func (ds *DataService) DeleteBacktestRun(id int64) error {
	return ds.backtestRepo.DeleteRun(id)
}

// Database methods

// GetDB returns the underlying database
func (ds *DataService) GetDB() *SQLiteDB {
	return ds.db
}

// GetQueueManager returns the queue manager
func (ds *DataService) GetQueueManager() *QueueManager {
	return ds.queueManager
}

// GetQueueStats returns statistics for all queues
func (ds *DataService) GetQueueStats() map[string]QueueStats {
	return ds.queueManager.GetStats()
}

// GetDBStats returns database statistics
func (ds *DataService) GetDBStats() (*DBStats, error) {
	return ds.db.GetStats()
}

// Cleanup removes old data
func (ds *DataService) Cleanup(candleRetentionDays, snapshotRetentionDays int) error {
	return ds.db.Cleanup(candleRetentionDays, snapshotRetentionDays)
}

// Close closes the data service and database
func (ds *DataService) Close() error {
	ds.Stop()
	return ds.db.Close()
}
