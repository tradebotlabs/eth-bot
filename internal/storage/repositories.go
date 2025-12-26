package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// CandleRepository handles candle persistence
type CandleRepository struct {
	db *SQLiteDB
}

// NewCandleRepository creates a new candle repository
func NewCandleRepository(db *SQLiteDB) *CandleRepository {
	return &CandleRepository{db: db}
}

// Insert adds a new candle (upsert)
func (r *CandleRepository) Insert(candle Candle) error {
	query := `
		INSERT INTO candles (symbol, timeframe, open_time, close_time, open, high, low, close, volume, trades)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(symbol, timeframe, open_time) DO UPDATE SET
			high = MAX(excluded.high, candles.high),
			low = MIN(excluded.low, candles.low),
			close = excluded.close,
			volume = excluded.volume,
			trades = excluded.trades
	`
	_, err := r.db.Exec(query,
		candle.Symbol, candle.Timeframe, candle.OpenTime, candle.CloseTime,
		candle.Open, candle.High, candle.Low, candle.Close, candle.Volume, candle.Trades,
	)
	return err
}

// InsertBatch inserts multiple candles efficiently
func (r *CandleRepository) InsertBatch(candles []Candle) error {
	if len(candles) == 0 {
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO candles (symbol, timeframe, open_time, close_time, open, high, low, close, volume, trades)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(symbol, timeframe, open_time) DO UPDATE SET
			high = MAX(excluded.high, candles.high),
			low = MIN(excluded.low, candles.low),
			close = excluded.close,
			volume = excluded.volume,
			trades = excluded.trades
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, candle := range candles {
		_, err := stmt.Exec(
			candle.Symbol, candle.Timeframe, candle.OpenTime, candle.CloseTime,
			candle.Open, candle.High, candle.Low, candle.Close, candle.Volume, candle.Trades,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetRange retrieves candles within a time range
func (r *CandleRepository) GetRange(symbol, timeframe string, from, to time.Time) ([]Candle, error) {
	query := `
		SELECT id, symbol, timeframe, open_time, close_time, open, high, low, close, volume, trades
		FROM candles
		WHERE symbol = ? AND timeframe = ? AND open_time >= ? AND open_time <= ?
		ORDER BY open_time ASC
	`
	rows, err := r.db.Query(query, symbol, timeframe, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanCandles(rows)
}

// GetLast retrieves the last N candles
func (r *CandleRepository) GetLast(symbol, timeframe string, limit int) ([]Candle, error) {
	query := `
		SELECT id, symbol, timeframe, open_time, close_time, open, high, low, close, volume, trades
		FROM candles
		WHERE symbol = ? AND timeframe = ?
		ORDER BY open_time DESC
		LIMIT ?
	`
	rows, err := r.db.Query(query, symbol, timeframe, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	candles, err := scanCandles(rows)
	if err != nil {
		return nil, err
	}

	// Reverse to get oldest first
	for i, j := 0, len(candles)-1; i < j; i, j = i+1, j-1 {
		candles[i], candles[j] = candles[j], candles[i]
	}
	return candles, nil
}

// GetLatest retrieves the most recent candle
func (r *CandleRepository) GetLatest(symbol, timeframe string) (*Candle, error) {
	query := `
		SELECT id, symbol, timeframe, open_time, close_time, open, high, low, close, volume, trades
		FROM candles
		WHERE symbol = ? AND timeframe = ?
		ORDER BY open_time DESC
		LIMIT 1
	`
	var c Candle
	err := r.db.QueryRow(query, symbol, timeframe).Scan(
		&c.ID, &c.Symbol, &c.Timeframe, &c.OpenTime, &c.CloseTime,
		&c.Open, &c.High, &c.Low, &c.Close, &c.Volume, &c.Trades,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// Count returns the number of candles
func (r *CandleRepository) Count(symbol, timeframe string) (int64, error) {
	var count int64
	err := r.db.QueryRow(
		"SELECT COUNT(*) FROM candles WHERE symbol = ? AND timeframe = ?",
		symbol, timeframe,
	).Scan(&count)
	return count, err
}

// DeleteOlderThan removes candles older than the given date
func (r *CandleRepository) DeleteOlderThan(cutoff time.Time) (int64, error) {
	result, err := r.db.Exec("DELETE FROM candles WHERE open_time < ?", cutoff)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func scanCandles(rows *sql.Rows) ([]Candle, error) {
	var candles []Candle
	for rows.Next() {
		var c Candle
		err := rows.Scan(
			&c.ID, &c.Symbol, &c.Timeframe, &c.OpenTime, &c.CloseTime,
			&c.Open, &c.High, &c.Low, &c.Close, &c.Volume, &c.Trades,
		)
		if err != nil {
			return nil, err
		}
		c.IsClosed = true
		candles = append(candles, c)
	}
	return candles, rows.Err()
}

// TradeRepository handles trade persistence
type TradeRepository struct {
	db *SQLiteDB
}

// NewTradeRepository creates a new trade repository
func NewTradeRepository(db *SQLiteDB) *TradeRepository {
	return &TradeRepository{db: db}
}

// Insert adds a new trade
func (r *TradeRepository) Insert(trade Trade) error {
	query := `
		INSERT INTO trades (order_id, symbol, side, type, quantity, price, commission, commission_asset, executed_at, strategy, signal_strength)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.Exec(query,
		trade.OrderID, trade.Symbol, trade.Side, trade.Type,
		trade.Quantity, trade.Price, trade.Commission, trade.CommissionAsset,
		trade.ExecutedAt, trade.Strategy, trade.SignalStrength,
	)
	return err
}

// GetBySymbol retrieves trades for a symbol
func (r *TradeRepository) GetBySymbol(symbol string, limit int) ([]Trade, error) {
	query := `
		SELECT id, order_id, symbol, side, type, quantity, price, commission, commission_asset, executed_at, strategy, signal_strength, created_at
		FROM trades
		WHERE symbol = ?
		ORDER BY executed_at DESC
		LIMIT ?
	`
	rows, err := r.db.Query(query, symbol, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanTrades(rows)
}

// GetByStrategy retrieves trades for a strategy
func (r *TradeRepository) GetByStrategy(strategy string, limit int) ([]Trade, error) {
	query := `
		SELECT id, order_id, symbol, side, type, quantity, price, commission, commission_asset, executed_at, strategy, signal_strength, created_at
		FROM trades
		WHERE strategy = ?
		ORDER BY executed_at DESC
		LIMIT ?
	`
	rows, err := r.db.Query(query, strategy, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanTrades(rows)
}

// GetByDateRange retrieves trades within a date range
func (r *TradeRepository) GetByDateRange(from, to time.Time) ([]Trade, error) {
	query := `
		SELECT id, order_id, symbol, side, type, quantity, price, commission, commission_asset, executed_at, strategy, signal_strength, created_at
		FROM trades
		WHERE executed_at >= ? AND executed_at <= ?
		ORDER BY executed_at ASC
	`
	rows, err := r.db.Query(query, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanTrades(rows)
}

func scanTrades(rows *sql.Rows) ([]Trade, error) {
	var trades []Trade
	for rows.Next() {
		var t Trade
		var commissionAsset sql.NullString
		err := rows.Scan(
			&t.ID, &t.OrderID, &t.Symbol, &t.Side, &t.Type,
			&t.Quantity, &t.Price, &t.Commission, &commissionAsset,
			&t.ExecutedAt, &t.Strategy, &t.SignalStrength, &t.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		if commissionAsset.Valid {
			t.CommissionAsset = commissionAsset.String
		}
		trades = append(trades, t)
	}
	return trades, rows.Err()
}

// PositionRepository handles position persistence
type PositionRepository struct {
	db *SQLiteDB
}

// NewPositionRepository creates a new position repository
func NewPositionRepository(db *SQLiteDB) *PositionRepository {
	return &PositionRepository{db: db}
}

// Insert adds a new position
func (r *PositionRepository) Insert(pos Position) (int64, error) {
	query := `
		INSERT INTO positions (symbol, side, entry_price, quantity, current_price, unrealized_pnl, stop_loss, take_profit, strategy, status, opened_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(query,
		pos.Symbol, pos.Side, pos.EntryPrice, pos.Quantity, pos.CurrentPrice,
		pos.UnrealizedPnL, pos.StopLoss, pos.TakeProfit, pos.Strategy, pos.Status, pos.OpenedAt,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// Update updates a position
func (r *PositionRepository) Update(pos Position) error {
	query := `
		UPDATE positions SET
			current_price = ?, unrealized_pnl = ?, realized_pnl = ?,
			stop_loss = ?, take_profit = ?, status = ?, closed_at = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	_, err := r.db.Exec(query,
		pos.CurrentPrice, pos.UnrealizedPnL, pos.RealizedPnL,
		pos.StopLoss, pos.TakeProfit, pos.Status, pos.ClosedAt, pos.ID,
	)
	return err
}

// GetOpen retrieves all open positions
func (r *PositionRepository) GetOpen() ([]Position, error) {
	query := `
		SELECT id, symbol, side, entry_price, quantity, current_price, unrealized_pnl, realized_pnl,
		       stop_loss, take_profit, strategy, status, opened_at, closed_at, created_at, updated_at
		FROM positions
		WHERE status = 'open'
		ORDER BY opened_at DESC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanPositions(rows)
}

// GetByID retrieves a position by ID
func (r *PositionRepository) GetByID(id int64) (*Position, error) {
	query := `
		SELECT id, symbol, side, entry_price, quantity, current_price, unrealized_pnl, realized_pnl,
		       stop_loss, take_profit, strategy, status, opened_at, closed_at, created_at, updated_at
		FROM positions
		WHERE id = ?
	`
	var pos Position
	var closedAt sql.NullTime
	err := r.db.QueryRow(query, id).Scan(
		&pos.ID, &pos.Symbol, &pos.Side, &pos.EntryPrice, &pos.Quantity,
		&pos.CurrentPrice, &pos.UnrealizedPnL, &pos.RealizedPnL,
		&pos.StopLoss, &pos.TakeProfit, &pos.Strategy, &pos.Status,
		&pos.OpenedAt, &closedAt, &pos.CreatedAt, &pos.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if closedAt.Valid {
		pos.ClosedAt = &closedAt.Time
	}
	return &pos, nil
}

// GetClosed retrieves closed positions
func (r *PositionRepository) GetClosed(limit int) ([]Position, error) {
	query := `
		SELECT id, symbol, side, entry_price, quantity, current_price, unrealized_pnl, realized_pnl,
		       stop_loss, take_profit, strategy, status, opened_at, closed_at, created_at, updated_at
		FROM positions
		WHERE status = 'closed'
		ORDER BY closed_at DESC
		LIMIT ?
	`
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanPositions(rows)
}

func scanPositions(rows *sql.Rows) ([]Position, error) {
	var positions []Position
	for rows.Next() {
		var pos Position
		var closedAt sql.NullTime
		err := rows.Scan(
			&pos.ID, &pos.Symbol, &pos.Side, &pos.EntryPrice, &pos.Quantity,
			&pos.CurrentPrice, &pos.UnrealizedPnL, &pos.RealizedPnL,
			&pos.StopLoss, &pos.TakeProfit, &pos.Strategy, &pos.Status,
			&pos.OpenedAt, &closedAt, &pos.CreatedAt, &pos.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if closedAt.Valid {
			pos.ClosedAt = &closedAt.Time
		}
		positions = append(positions, pos)
	}
	return positions, rows.Err()
}

// AccountRepository handles account snapshot persistence
type AccountRepository struct {
	db *SQLiteDB
}

// NewAccountRepository creates a new account repository
func NewAccountRepository(db *SQLiteDB) *AccountRepository {
	return &AccountRepository{db: db}
}

// InsertSnapshot adds a new account snapshot
func (r *AccountRepository) InsertSnapshot(snapshot AccountSnapshot) error {
	query := `
		INSERT INTO account_snapshots (total_equity, available_balance, unrealized_pnl, daily_pnl, open_positions, snapshot_time)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.Exec(query,
		snapshot.TotalEquity, snapshot.AvailableBalance, snapshot.UnrealizedPnL,
		snapshot.DailyPnL, snapshot.OpenPositions, snapshot.SnapshotTime,
	)
	return err
}

// GetLatestSnapshot retrieves the most recent snapshot
func (r *AccountRepository) GetLatestSnapshot() (*AccountSnapshot, error) {
	query := `
		SELECT id, total_equity, available_balance, unrealized_pnl, daily_pnl, open_positions, snapshot_time, created_at
		FROM account_snapshots
		ORDER BY snapshot_time DESC
		LIMIT 1
	`
	var s AccountSnapshot
	err := r.db.QueryRow(query).Scan(
		&s.ID, &s.TotalEquity, &s.AvailableBalance, &s.UnrealizedPnL,
		&s.DailyPnL, &s.OpenPositions, &s.SnapshotTime, &s.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// GetSnapshotsRange retrieves snapshots within a time range
func (r *AccountRepository) GetSnapshotsRange(from, to time.Time) ([]AccountSnapshot, error) {
	query := `
		SELECT id, total_equity, available_balance, unrealized_pnl, daily_pnl, open_positions, snapshot_time, created_at
		FROM account_snapshots
		WHERE snapshot_time >= ? AND snapshot_time <= ?
		ORDER BY snapshot_time ASC
	`
	rows, err := r.db.Query(query, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snapshots []AccountSnapshot
	for rows.Next() {
		var s AccountSnapshot
		err := rows.Scan(
			&s.ID, &s.TotalEquity, &s.AvailableBalance, &s.UnrealizedPnL,
			&s.DailyPnL, &s.OpenPositions, &s.SnapshotTime, &s.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, s)
	}
	return snapshots, rows.Err()
}

// StrategyPerformanceRepository handles strategy performance persistence
type StrategyPerformanceRepository struct {
	db *SQLiteDB
}

// NewStrategyPerformanceRepository creates a new strategy performance repository
func NewStrategyPerformanceRepository(db *SQLiteDB) *StrategyPerformanceRepository {
	return &StrategyPerformanceRepository{db: db}
}

// Upsert inserts or updates strategy performance for a date
func (r *StrategyPerformanceRepository) Upsert(perf StrategyPerformance) error {
	query := `
		INSERT INTO strategy_performance (strategy, date, trades, wins, losses, gross_profit, gross_loss, net_pnl, max_drawdown)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(strategy, date) DO UPDATE SET
			trades = excluded.trades,
			wins = excluded.wins,
			losses = excluded.losses,
			gross_profit = excluded.gross_profit,
			gross_loss = excluded.gross_loss,
			net_pnl = excluded.net_pnl,
			max_drawdown = excluded.max_drawdown
	`
	_, err := r.db.Exec(query,
		perf.Strategy, perf.Date, perf.Trades, perf.Wins, perf.Losses,
		perf.GrossProfit, perf.GrossLoss, perf.NetPnL, perf.MaxDrawdown,
	)
	return err
}

// GetByStrategy retrieves performance records for a strategy
func (r *StrategyPerformanceRepository) GetByStrategy(strategy string, limit int) ([]StrategyPerformance, error) {
	query := `
		SELECT id, strategy, date, trades, wins, losses, gross_profit, gross_loss, net_pnl, max_drawdown, created_at
		FROM strategy_performance
		WHERE strategy = ?
		ORDER BY date DESC
		LIMIT ?
	`
	rows, err := r.db.Query(query, strategy, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perfs []StrategyPerformance
	for rows.Next() {
		var p StrategyPerformance
		err := rows.Scan(
			&p.ID, &p.Strategy, &p.Date, &p.Trades, &p.Wins, &p.Losses,
			&p.GrossProfit, &p.GrossLoss, &p.NetPnL, &p.MaxDrawdown, &p.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		perfs = append(perfs, p)
	}
	return perfs, rows.Err()
}

// AlertRepository handles alert persistence
type AlertRepository struct {
	db *SQLiteDB
}

// NewAlertRepository creates a new alert repository
func NewAlertRepository(db *SQLiteDB) *AlertRepository {
	return &AlertRepository{db: db}
}

// Insert adds a new alert
func (r *AlertRepository) Insert(alert Alert) (int64, error) {
	query := `
		INSERT INTO alerts (type, severity, message, data)
		VALUES (?, ?, ?, ?)
	`
	var data string
	if alert.Data != "" {
		data = alert.Data
	}
	result, err := r.db.Exec(query, alert.Type, alert.Severity, alert.Message, data)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetUnacknowledged retrieves unacknowledged alerts
func (r *AlertRepository) GetUnacknowledged(limit int) ([]Alert, error) {
	query := `
		SELECT id, type, severity, message, data, acknowledged, created_at
		FROM alerts
		WHERE acknowledged = FALSE
		ORDER BY created_at DESC
		LIMIT ?
	`
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAlerts(rows)
}

// Acknowledge marks an alert as acknowledged
func (r *AlertRepository) Acknowledge(id int64) error {
	_, err := r.db.Exec("UPDATE alerts SET acknowledged = TRUE WHERE id = ?", id)
	return err
}

func scanAlerts(rows *sql.Rows) ([]Alert, error) {
	var alerts []Alert
	for rows.Next() {
		var a Alert
		var data sql.NullString
		err := rows.Scan(&a.ID, &a.Type, &a.Severity, &a.Message, &data, &a.Acknowledged, &a.CreatedAt)
		if err != nil {
			return nil, err
		}
		if data.Valid {
			a.Data = data.String
		}
		alerts = append(alerts, a)
	}
	return alerts, rows.Err()
}

// BacktestRepository handles backtest persistence
type BacktestRepository struct {
	db *SQLiteDB
}

// NewBacktestRepository creates a new backtest repository
func NewBacktestRepository(db *SQLiteDB) *BacktestRepository {
	return &BacktestRepository{db: db}
}

// BacktestRun represents a backtest run record
type BacktestRun struct {
	ID             int64           `json:"id"`
	Name           string          `json:"name"`
	Symbol         string          `json:"symbol"`
	Timeframe      string          `json:"timeframe"`
	StartDate      time.Time       `json:"start_date"`
	EndDate        time.Time       `json:"end_date"`
	InitialBalance float64         `json:"initial_balance"`
	FinalBalance   float64         `json:"final_balance"`
	TotalTrades    int             `json:"total_trades"`
	WinningTrades  int             `json:"winning_trades"`
	LosingTrades   int             `json:"losing_trades"`
	GrossProfit    float64         `json:"gross_profit"`
	GrossLoss      float64         `json:"gross_loss"`
	NetProfit      float64         `json:"net_profit"`
	MaxDrawdown    float64         `json:"max_drawdown"`
	MaxDrawdownPct float64         `json:"max_drawdown_pct"`
	WinRate        float64         `json:"win_rate"`
	ProfitFactor   float64         `json:"profit_factor"`
	SharpeRatio    float64         `json:"sharpe_ratio"`
	SortinoRatio   float64         `json:"sortino_ratio"`
	CalmarRatio    float64         `json:"calmar_ratio"`
	Strategies     []string        `json:"strategies"`
	Config         json.RawMessage `json:"config"`
	Status         string          `json:"status"`
	StartedAt      time.Time       `json:"started_at"`
	CompletedAt    *time.Time      `json:"completed_at,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}

// InsertRun inserts a new backtest run
func (r *BacktestRepository) InsertRun(run BacktestRun) (int64, error) {
	strategies, _ := json.Marshal(run.Strategies)
	config := string(run.Config)
	if config == "" {
		config = "{}"
	}

	query := `
		INSERT INTO backtest_runs (name, symbol, timeframe, start_date, end_date, initial_balance, strategies, config, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(query,
		run.Name, run.Symbol, run.Timeframe, run.StartDate, run.EndDate,
		run.InitialBalance, string(strategies), config, "running",
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// UpdateRun updates a backtest run with results
func (r *BacktestRepository) UpdateRun(run BacktestRun) error {
	query := `
		UPDATE backtest_runs SET
			final_balance = ?, total_trades = ?, winning_trades = ?, losing_trades = ?,
			gross_profit = ?, gross_loss = ?, net_profit = ?,
			max_drawdown = ?, max_drawdown_pct = ?,
			win_rate = ?, profit_factor = ?,
			sharpe_ratio = ?, sortino_ratio = ?, calmar_ratio = ?,
			status = ?, completed_at = ?
		WHERE id = ?
	`
	_, err := r.db.Exec(query,
		run.FinalBalance, run.TotalTrades, run.WinningTrades, run.LosingTrades,
		run.GrossProfit, run.GrossLoss, run.NetProfit,
		run.MaxDrawdown, run.MaxDrawdownPct,
		run.WinRate, run.ProfitFactor,
		run.SharpeRatio, run.SortinoRatio, run.CalmarRatio,
		run.Status, run.CompletedAt, run.ID,
	)
	return err
}

// GetRun retrieves a backtest run by ID
func (r *BacktestRepository) GetRun(id int64) (*BacktestRun, error) {
	query := `
		SELECT id, name, symbol, timeframe, start_date, end_date, initial_balance, final_balance,
		       total_trades, winning_trades, losing_trades, gross_profit, gross_loss, net_profit,
		       max_drawdown, max_drawdown_pct, win_rate, profit_factor,
		       sharpe_ratio, sortino_ratio, calmar_ratio,
		       strategies, config, status, started_at, completed_at, created_at
		FROM backtest_runs
		WHERE id = ?
	`
	var run BacktestRun
	var strategies, config string
	var completedAt sql.NullTime
	var name sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&run.ID, &name, &run.Symbol, &run.Timeframe, &run.StartDate, &run.EndDate,
		&run.InitialBalance, &run.FinalBalance,
		&run.TotalTrades, &run.WinningTrades, &run.LosingTrades,
		&run.GrossProfit, &run.GrossLoss, &run.NetProfit,
		&run.MaxDrawdown, &run.MaxDrawdownPct, &run.WinRate, &run.ProfitFactor,
		&run.SharpeRatio, &run.SortinoRatio, &run.CalmarRatio,
		&strategies, &config, &run.Status, &run.StartedAt, &completedAt, &run.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if name.Valid {
		run.Name = name.String
	}
	if completedAt.Valid {
		run.CompletedAt = &completedAt.Time
	}
	json.Unmarshal([]byte(strategies), &run.Strategies)
	run.Config = json.RawMessage(config)

	return &run, nil
}

// GetRuns retrieves backtest runs
func (r *BacktestRepository) GetRuns(limit int) ([]BacktestRun, error) {
	query := `
		SELECT id, name, symbol, timeframe, start_date, end_date, initial_balance, final_balance,
		       total_trades, winning_trades, losing_trades, gross_profit, gross_loss, net_profit,
		       max_drawdown, max_drawdown_pct, win_rate, profit_factor,
		       sharpe_ratio, sortino_ratio, calmar_ratio,
		       strategies, config, status, started_at, completed_at, created_at
		FROM backtest_runs
		ORDER BY created_at DESC
		LIMIT ?
	`
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []BacktestRun
	for rows.Next() {
		var run BacktestRun
		var strategies, config string
		var completedAt sql.NullTime
		var name sql.NullString

		err := rows.Scan(
			&run.ID, &name, &run.Symbol, &run.Timeframe, &run.StartDate, &run.EndDate,
			&run.InitialBalance, &run.FinalBalance,
			&run.TotalTrades, &run.WinningTrades, &run.LosingTrades,
			&run.GrossProfit, &run.GrossLoss, &run.NetProfit,
			&run.MaxDrawdown, &run.MaxDrawdownPct, &run.WinRate, &run.ProfitFactor,
			&run.SharpeRatio, &run.SortinoRatio, &run.CalmarRatio,
			&strategies, &config, &run.Status, &run.StartedAt, &completedAt, &run.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if name.Valid {
			run.Name = name.String
		}
		if completedAt.Valid {
			run.CompletedAt = &completedAt.Time
		}
		json.Unmarshal([]byte(strategies), &run.Strategies)
		run.Config = json.RawMessage(config)

		runs = append(runs, run)
	}
	return runs, rows.Err()
}

// DeleteRun deletes a backtest run and related data
func (r *BacktestRepository) DeleteRun(id int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete equity curve
	if _, err := tx.Exec("DELETE FROM backtest_equity WHERE backtest_id = ?", id); err != nil {
		return fmt.Errorf("failed to delete equity curve: %w", err)
	}

	// Delete trades
	if _, err := tx.Exec("DELETE FROM backtest_trades WHERE backtest_id = ?", id); err != nil {
		return fmt.Errorf("failed to delete trades: %w", err)
	}

	// Delete run
	if _, err := tx.Exec("DELETE FROM backtest_runs WHERE id = ?", id); err != nil {
		return fmt.Errorf("failed to delete run: %w", err)
	}

	return tx.Commit()
}
