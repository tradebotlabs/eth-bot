package storage

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
)

// SQLiteDB wraps the database connection
type SQLiteDB struct {
	db   *sql.DB
	path string
}

// NewSQLiteDB creates a new SQLite database connection
func NewSQLiteDB(dbPath string) (*SQLiteDB, error) {
	// Connection string with WAL mode and normal synchronous
	connStr := fmt.Sprintf("%s?_journal_mode=WAL&_synchronous=NORMAL&_busy_timeout=5000", dbPath)

	db, err := sql.Open("sqlite3", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(1) // SQLite only supports one writer
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	sqliteDB := &SQLiteDB{
		db:   db,
		path: dbPath,
	}

	// Run migrations
	if err := sqliteDB.migrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Info().Str("path", dbPath).Msg("SQLite database initialized")
	return sqliteDB, nil
}

// DB returns the underlying sql.DB
func (s *SQLiteDB) DB() *sql.DB {
	return s.db
}

// Close closes the database connection
func (s *SQLiteDB) Close() error {
	return s.db.Close()
}

// migrate runs database migrations
func (s *SQLiteDB) migrate() error {
	migrations := []string{
		// Candles table
		`CREATE TABLE IF NOT EXISTS candles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			symbol TEXT NOT NULL,
			timeframe TEXT NOT NULL,
			open_time DATETIME NOT NULL,
			close_time DATETIME NOT NULL,
			open REAL NOT NULL,
			high REAL NOT NULL,
			low REAL NOT NULL,
			close REAL NOT NULL,
			volume REAL NOT NULL,
			trades INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(symbol, timeframe, open_time)
		)`,

		// Index for fast candle queries
		`CREATE INDEX IF NOT EXISTS idx_candles_symbol_timeframe_time
		 ON candles(symbol, timeframe, open_time DESC)`,

		// Trades/Executions table
		`CREATE TABLE IF NOT EXISTS trades (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			order_id TEXT UNIQUE NOT NULL,
			symbol TEXT NOT NULL,
			side TEXT NOT NULL,
			type TEXT NOT NULL,
			quantity REAL NOT NULL,
			price REAL NOT NULL,
			commission REAL DEFAULT 0,
			commission_asset TEXT,
			executed_at DATETIME NOT NULL,
			strategy TEXT,
			signal_strength REAL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Index for trade queries
		`CREATE INDEX IF NOT EXISTS idx_trades_symbol_time
		 ON trades(symbol, executed_at DESC)`,

		`CREATE INDEX IF NOT EXISTS idx_trades_strategy
		 ON trades(strategy, executed_at DESC)`,

		// Positions table
		`CREATE TABLE IF NOT EXISTS positions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			symbol TEXT NOT NULL,
			side TEXT NOT NULL,
			entry_price REAL NOT NULL,
			quantity REAL NOT NULL,
			current_price REAL,
			unrealized_pnl REAL DEFAULT 0,
			realized_pnl REAL DEFAULT 0,
			stop_loss REAL,
			take_profit REAL,
			strategy TEXT,
			status TEXT DEFAULT 'open',
			opened_at DATETIME NOT NULL,
			closed_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Index for position queries
		`CREATE INDEX IF NOT EXISTS idx_positions_symbol_status
		 ON positions(symbol, status)`,

		`CREATE INDEX IF NOT EXISTS idx_positions_strategy
		 ON positions(strategy, status)`,

		// Orders table
		`CREATE TABLE IF NOT EXISTS orders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			order_id TEXT UNIQUE NOT NULL,
			client_order_id TEXT,
			symbol TEXT NOT NULL,
			side TEXT NOT NULL,
			type TEXT NOT NULL,
			quantity REAL NOT NULL,
			price REAL,
			stop_price REAL,
			status TEXT DEFAULT 'pending',
			filled_quantity REAL DEFAULT 0,
			avg_fill_price REAL,
			strategy TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Index for order queries
		`CREATE INDEX IF NOT EXISTS idx_orders_symbol_status
		 ON orders(symbol, status)`,

		// Account snapshots for equity tracking
		`CREATE TABLE IF NOT EXISTS account_snapshots (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			total_equity REAL NOT NULL,
			available_balance REAL NOT NULL,
			unrealized_pnl REAL DEFAULT 0,
			daily_pnl REAL DEFAULT 0,
			open_positions INTEGER DEFAULT 0,
			snapshot_time DATETIME NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Index for snapshot queries
		`CREATE INDEX IF NOT EXISTS idx_snapshots_time
		 ON account_snapshots(snapshot_time DESC)`,

		// Strategy performance tracking
		`CREATE TABLE IF NOT EXISTS strategy_performance (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			strategy TEXT NOT NULL,
			date DATE NOT NULL,
			trades INTEGER DEFAULT 0,
			wins INTEGER DEFAULT 0,
			losses INTEGER DEFAULT 0,
			gross_profit REAL DEFAULT 0,
			gross_loss REAL DEFAULT 0,
			net_pnl REAL DEFAULT 0,
			max_drawdown REAL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(strategy, date)
		)`,

		// Index for performance queries
		`CREATE INDEX IF NOT EXISTS idx_strategy_perf_date
		 ON strategy_performance(date DESC)`,

		// Configuration table
		`CREATE TABLE IF NOT EXISTS config (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Alerts/Notifications log
		`CREATE TABLE IF NOT EXISTS alerts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			type TEXT NOT NULL,
			severity TEXT NOT NULL,
			message TEXT NOT NULL,
			data TEXT,
			acknowledged BOOLEAN DEFAULT FALSE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Index for alert queries
		`CREATE INDEX IF NOT EXISTS idx_alerts_type_time
		 ON alerts(type, created_at DESC)`,

		// Backtest runs table
		`CREATE TABLE IF NOT EXISTS backtest_runs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			symbol TEXT NOT NULL,
			timeframe TEXT NOT NULL,
			start_date DATETIME NOT NULL,
			end_date DATETIME NOT NULL,
			initial_balance REAL NOT NULL,
			final_balance REAL,
			total_trades INTEGER DEFAULT 0,
			winning_trades INTEGER DEFAULT 0,
			losing_trades INTEGER DEFAULT 0,
			gross_profit REAL DEFAULT 0,
			gross_loss REAL DEFAULT 0,
			net_profit REAL DEFAULT 0,
			max_drawdown REAL DEFAULT 0,
			max_drawdown_pct REAL DEFAULT 0,
			win_rate REAL DEFAULT 0,
			profit_factor REAL DEFAULT 0,
			sharpe_ratio REAL DEFAULT 0,
			sortino_ratio REAL DEFAULT 0,
			calmar_ratio REAL DEFAULT 0,
			strategies TEXT,
			config TEXT,
			status TEXT DEFAULT 'running',
			started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Backtest trades table
		`CREATE TABLE IF NOT EXISTS backtest_trades (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			backtest_id INTEGER NOT NULL,
			symbol TEXT NOT NULL,
			side TEXT NOT NULL,
			entry_price REAL NOT NULL,
			exit_price REAL,
			quantity REAL NOT NULL,
			entry_time DATETIME NOT NULL,
			exit_time DATETIME,
			pnl REAL DEFAULT 0,
			pnl_pct REAL DEFAULT 0,
			strategy TEXT,
			entry_reason TEXT,
			exit_reason TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (backtest_id) REFERENCES backtest_runs(id)
		)`,

		// Index for backtest trade queries
		`CREATE INDEX IF NOT EXISTS idx_backtest_trades_run
		 ON backtest_trades(backtest_id, entry_time)`,

		// Equity curve table for backtests
		`CREATE TABLE IF NOT EXISTS backtest_equity (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			backtest_id INTEGER NOT NULL,
			timestamp DATETIME NOT NULL,
			equity REAL NOT NULL,
			drawdown REAL DEFAULT 0,
			drawdown_pct REAL DEFAULT 0,
			FOREIGN KEY (backtest_id) REFERENCES backtest_runs(id)
		)`,

		`CREATE INDEX IF NOT EXISTS idx_backtest_equity_run
		 ON backtest_equity(backtest_id, timestamp)`,
	}

	for _, migration := range migrations {
		if _, err := s.db.Exec(migration); err != nil {
			return fmt.Errorf("migration failed: %w\nSQL: %s", err, migration)
		}
	}

	log.Debug().Msg("Database migrations completed")
	return nil
}

// Exec executes a query without returning rows
func (s *SQLiteDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return s.db.Exec(query, args...)
}

// Query executes a query that returns rows
func (s *SQLiteDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return s.db.Query(query, args...)
}

// QueryRow executes a query that returns a single row
func (s *SQLiteDB) QueryRow(query string, args ...interface{}) *sql.Row {
	return s.db.QueryRow(query, args...)
}

// Begin starts a transaction
func (s *SQLiteDB) Begin() (*sql.Tx, error) {
	return s.db.Begin()
}

// Vacuum runs VACUUM to optimize the database
func (s *SQLiteDB) Vacuum() error {
	_, err := s.db.Exec("VACUUM")
	return err
}

// Checkpoint forces a WAL checkpoint
func (s *SQLiteDB) Checkpoint() error {
	_, err := s.db.Exec("PRAGMA wal_checkpoint(TRUNCATE)")
	return err
}

// GetConfig retrieves a config value
func (s *SQLiteDB) GetConfig(key string) (string, error) {
	var value string
	err := s.db.QueryRow("SELECT value FROM config WHERE key = ?", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

// SetConfig sets a config value
func (s *SQLiteDB) SetConfig(key, value string) error {
	_, err := s.db.Exec(`
		INSERT INTO config (key, value, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = CURRENT_TIMESTAMP
	`, key, value)
	return err
}

// Cleanup removes old data based on retention settings
func (s *SQLiteDB) Cleanup(candleRetentionDays, snapshotRetentionDays int) error {
	// Clean old candles
	candleCutoff := time.Now().AddDate(0, 0, -candleRetentionDays)
	if _, err := s.db.Exec("DELETE FROM candles WHERE open_time < ?", candleCutoff); err != nil {
		return fmt.Errorf("failed to cleanup candles: %w", err)
	}

	// Clean old snapshots
	snapshotCutoff := time.Now().AddDate(0, 0, -snapshotRetentionDays)
	if _, err := s.db.Exec("DELETE FROM account_snapshots WHERE snapshot_time < ?", snapshotCutoff); err != nil {
		return fmt.Errorf("failed to cleanup snapshots: %w", err)
	}

	// Clean old alerts (keep last 30 days)
	alertCutoff := time.Now().AddDate(0, 0, -30)
	if _, err := s.db.Exec("DELETE FROM alerts WHERE created_at < ? AND acknowledged = TRUE", alertCutoff); err != nil {
		return fmt.Errorf("failed to cleanup alerts: %w", err)
	}

	log.Debug().Msg("Database cleanup completed")
	return nil
}

// Stats returns database statistics
type DBStats struct {
	CandleCount     int64
	TradeCount      int64
	PositionCount   int64
	OrderCount      int64
	SnapshotCount   int64
	BacktestCount   int64
	AlertCount      int64
	DatabaseSize    int64
	WALSize         int64
}

// GetStats returns database statistics
func (s *SQLiteDB) GetStats() (*DBStats, error) {
	stats := &DBStats{}

	queries := []struct {
		query string
		dest  *int64
	}{
		{"SELECT COUNT(*) FROM candles", &stats.CandleCount},
		{"SELECT COUNT(*) FROM trades", &stats.TradeCount},
		{"SELECT COUNT(*) FROM positions", &stats.PositionCount},
		{"SELECT COUNT(*) FROM orders", &stats.OrderCount},
		{"SELECT COUNT(*) FROM account_snapshots", &stats.SnapshotCount},
		{"SELECT COUNT(*) FROM backtest_runs", &stats.BacktestCount},
		{"SELECT COUNT(*) FROM alerts", &stats.AlertCount},
	}

	for _, q := range queries {
		if err := s.db.QueryRow(q.query).Scan(q.dest); err != nil {
			return nil, err
		}
	}

	// Get database size
	var pageCount, pageSize int64
	if err := s.db.QueryRow("PRAGMA page_count").Scan(&pageCount); err != nil {
		return nil, err
	}
	if err := s.db.QueryRow("PRAGMA page_size").Scan(&pageSize); err != nil {
		return nil, err
	}
	stats.DatabaseSize = pageCount * pageSize

	return stats, nil
}
