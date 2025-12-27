package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eth-trading/internal/api"
	"github.com/eth-trading/internal/auth"
	"github.com/eth-trading/internal/binance"
	"github.com/eth-trading/internal/config"
	"github.com/eth-trading/internal/execution"
	"github.com/eth-trading/internal/indicators"
	"github.com/eth-trading/internal/orchestrator"
	"github.com/eth-trading/internal/risk"
	"github.com/eth-trading/internal/storage"
	"github.com/eth-trading/internal/strategy"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Setup logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	log.Info().Msg("Starting ETH Trading Bot...")

	// Load configuration
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Warn().Err(err).Msg("Failed to load config, using defaults")
		cfg = config.DefaultConfig()
	}

	// Initialize PostgreSQL database for user/auth data
	pgCfg := &storage.PostgresConfig{
		Host:            cfg.Postgres.Host,
		Port:            cfg.Postgres.Port,
		User:            cfg.Postgres.User,
		Password:        cfg.Postgres.Password,
		DBName:          cfg.Postgres.DBName,
		SSLMode:         cfg.Postgres.SSLMode,
		MaxConns:        cfg.Postgres.MaxConns,
		MaxIdle:         cfg.Postgres.MaxIdle,
		ConnMaxLifetime: cfg.Postgres.ConnMaxLifetime,
	}

	pgDB, err := storage.NewPostgresDB(pgCfg)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to connect to PostgreSQL, authentication will not be available")
		pgDB = nil
	}
	if pgDB != nil {
		defer pgDB.Close()
	}

	// Initialize repositories
	var userRepo *storage.UserRepository
	var sessionRepo *storage.SessionRepository
	var tradingAccountRepo *storage.TradingAccountRepository
	var authService *auth.Service

	if pgDB != nil {
		userRepo = storage.NewUserRepository(pgDB)
		sessionRepo = storage.NewSessionRepository(pgDB)
		tradingAccountRepo = storage.NewTradingAccountRepository(pgDB)

		// Initialize auth service
		authCfg := &auth.Config{
			JWTSecret:          cfg.Auth.JWTSecret,
			TokenExpiry:        cfg.Auth.TokenExpiry,
			RefreshTokenExpiry: cfg.Auth.RefreshTokenExpiry,
		}
		authService = auth.NewService(authCfg, userRepo, sessionRepo, tradingAccountRepo)
		log.Info().Msg("Authentication service initialized")
	} else {
		log.Warn().Msg("Running without authentication - PostgreSQL not available")
	}

	// Initialize SQLite database for trading data (will migrate to PostgreSQL later)
	db, err := storage.NewSQLiteDB(cfg.Database.Path)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer db.Close()

	// Initialize data service
	dataService := storage.NewDataService(db, cfg.DataService.CacheExpiry, nil)

	// Initialize Binance client
	binanceClient := binance.NewClient(&binance.Config{
		APIKey:    cfg.Binance.APIKey,
		SecretKey: cfg.Binance.SecretKey,
		Testnet:   cfg.Binance.Testnet,
		Timeout:   30 * time.Second,
	})

	// Test Binance connection
	if err := binanceClient.Ping(); err != nil {
		log.Warn().Err(err).Msg("Binance connection test failed")
	} else {
		log.Info().Msg("Binance connection successful")
	}

	// Initialize orchestrator first (for handler creation)
	orchCfg := &orchestrator.OrchestratorConfig{
		Symbol:           cfg.Trading.Symbol,
		Timeframes:       cfg.Trading.Timeframes,
		PrimaryTimeframe: cfg.Trading.PrimaryTimeframe,
		Mode:             orchestrator.TradingModePaper, // Will be set properly later
		InitialCapital:   cfg.Trading.InitialBalance,
		EnabledStrategies: cfg.Strategies.Enabled,
		EnableWebSocket:   true,
		BroadcastInterval: time.Second,
	}
	orch := orchestrator.NewOrchestrator(orchCfg)

	// Create WebSocket handler that connects to orchestrator
	wsHandler := orch.CreateWSHandler()

	// Initialize WebSocket client with handler for real-time events
	wsOpts := []binance.WSClientOption{
		binance.WithWSTestnet(cfg.Binance.Testnet),
		binance.WithReconnectWait(3 * time.Second),
		binance.WithPingInterval(30 * time.Second),
		binance.WithMaxReconnects(20),
	}
	wsClient := binance.NewWSClient(wsHandler, wsOpts...)

	// Initialize indicator manager
	indicatorCfg := &indicators.IndicatorConfig{
		RSIPeriod:     cfg.Indicators.RSIPeriod,
		MACDFast:      cfg.Indicators.MACDFast,
		MACDSlow:      cfg.Indicators.MACDSlow,
		MACDSignal:    cfg.Indicators.MACDSignal,
		BBPeriod:      cfg.Indicators.BBPeriod,
		BBStdDev:      cfg.Indicators.BBStdDev,
		ADXPeriod:     cfg.Indicators.ADXPeriod,
		ATRPeriod:     cfg.Indicators.ATRPeriod,
	}
	indicatorMgr := indicators.NewManager(indicatorCfg)

	// Initialize risk manager
	riskCfg := &risk.RiskConfig{
		MaxPositionSize:         cfg.Risk.MaxPositionSize,
		MaxPositionValue:        10000, // $10,000 max position value
		DefaultPositionSize:     0.05,  // 5% of equity
		MaxRiskPerTrade:         cfg.Risk.MaxRiskPerTrade,
		MinRiskRewardRatio:      cfg.Risk.MinRiskRewardRatio,
		MaxDailyLoss:            cfg.Risk.MaxDailyLoss,
		MaxWeeklyLoss:           cfg.Risk.MaxWeeklyLoss,
		MaxTotalDrawdown:        cfg.Risk.MaxDrawdown,
		MaxOpenPositions:        cfg.Risk.MaxOpenPositions,
		MaxPositionsPerSymbol:   1,
		MaxLeverage:             cfg.Risk.MaxLeverage,
		EnableCircuitBreaker:    cfg.Risk.EnableCircuitBreaker,
		ConsecutiveLossLimit:    cfg.Risk.ConsecutiveLossLimit,
		HaltDuration:            time.Duration(cfg.Risk.HaltDurationHours) * time.Hour,
		AdjustForVolatility:     true,
		HighVolatilityReduction: 0.5,
		MaxCorrelation:          0.7,
		TradingHoursOnly:        false,
		TradingStartHour:        0,
		TradingEndHour:          24,
		AvoidWeekends:           false,
	}
	riskManager := risk.NewManager(riskCfg)

	// Initialize strategies
	strategyMgr := strategy.NewManager(nil, indicatorCfg)
	log.Info().Int("strategies", len(strategyMgr.GetStrategies())).Msg("Strategies initialized")

	// Initialize executor based on mode
	var executor execution.Executor
	mode := orchestrator.TradingModePaper
	if cfg.Trading.Mode == "live" {
		mode = orchestrator.TradingModeLive
		liveExec, err := execution.NewLiveExecutor(&execution.ExecutorConfig{
			Mode:      execution.ModeLive,
			Symbol:    cfg.Trading.Symbol,
			APIKey:    cfg.Binance.APIKey,
			SecretKey: cfg.Binance.SecretKey,
			Testnet:   cfg.Binance.Testnet,
		})
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize live executor")
		}
		executor = liveExec
		log.Info().Msg("Live trading mode enabled")
	} else {
		paperExec := execution.NewPaperExecutor(&execution.ExecutorConfig{
			Mode:           execution.ModePaper,
			Symbol:         cfg.Trading.Symbol,
			InitialBalance: cfg.Trading.InitialBalance,
			Commission:     cfg.Trading.Commission,
			Slippage:       cfg.Trading.Slippage,
		})
		executor = paperExec
		log.Info().Float64("balance", cfg.Trading.InitialBalance).Msg("Paper trading mode enabled")
	}

	// Set orchestrator components (orch was created earlier for handler)
	orchCfg.Mode = mode // Update mode based on config
	orch.SetBinanceClient(binanceClient)
	orch.SetWebSocketClient(wsClient)
	orch.SetDataService(dataService)
	orch.SetExecutor(executor)
	orch.SetRiskManager(riskManager)
	orch.SetStrategyManager(strategyMgr)
	orch.SetIndicatorManager(indicatorMgr)

	// Initialize API server
	apiCfg := &api.ServerConfig{
		Port:         cfg.API.Port,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		CORSOrigins:  cfg.API.CORSOrigins,
	}
	server := api.NewServer(apiCfg, orch, authService)

	// Start orchestrator
	if err := orch.Start(); err != nil {
		log.Fatal().Err(err).Msg("Failed to start orchestrator")
	}

	// Start API server in goroutine
	go func() {
		if err := server.Start(); err != nil {
			log.Error().Err(err).Msg("API server error")
		}
	}()

	log.Info().
		Str("symbol", cfg.Trading.Symbol).
		Str("mode", cfg.Trading.Mode).
		Str("apiPort", cfg.API.Port).
		Msg("ETH Trading Bot started")

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop orchestrator
	orch.Stop()

	// Stop API server
	if err := server.Shutdown(); err != nil {
		log.Error().Err(err).Msg("API server shutdown error")
	}

	// Close WebSocket client
	wsClient.Disconnect()

	_ = ctx // Used for shutdown timeout

	log.Info().Msg("ETH Trading Bot stopped")
}
