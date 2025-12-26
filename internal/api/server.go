package api

import (
	"context"
	"net/http"
	"time"

	"github.com/eth-trading/internal/api/handlers"
	"github.com/eth-trading/internal/api/middleware"
	"github.com/eth-trading/internal/api/websocket"
	"github.com/eth-trading/internal/orchestrator"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
)

// ServerConfig holds server configuration
type ServerConfig struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	CORSOrigins     []string
	EnableSwagger   bool
}

// DefaultServerConfig returns default configuration
func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Port:            ":8080",
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		ShutdownTimeout: 10 * time.Second,
		CORSOrigins:     []string{"*"},
		EnableSwagger:   true,
	}
}

// Server is the API server
type Server struct {
	config       *ServerConfig
	echo         *echo.Echo
	orchestrator *orchestrator.Orchestrator
	wsHub        *websocket.Hub
}

// NewServer creates a new API server
func NewServer(config *ServerConfig, orch *orchestrator.Orchestrator) *Server {
	if config == nil {
		config = DefaultServerConfig()
	}

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	server := &Server{
		config:       config,
		echo:         e,
		orchestrator: orch,
		wsHub:        websocket.NewHub(),
	}

	server.setupMiddleware()
	server.setupRoutes()

	return server
}

// setupMiddleware configures middleware
func (s *Server) setupMiddleware() {
	// Recovery middleware
	s.echo.Use(echoMiddleware.Recover())

	// Logger middleware
	s.echo.Use(middleware.Logger())

	// CORS middleware
	s.echo.Use(echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
		AllowOrigins: s.config.CORSOrigins,
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	// Request ID middleware
	s.echo.Use(echoMiddleware.RequestID())

	// Gzip compression
	s.echo.Use(echoMiddleware.Gzip())
}

// setupRoutes configures API routes
func (s *Server) setupRoutes() {
	// Create handlers
	dashboardHandler := handlers.NewDashboardHandler(s.orchestrator)
	tradingHandler := handlers.NewTradingHandler(s.orchestrator)
	strategyHandler := handlers.NewStrategyHandler(s.orchestrator)
	riskHandler := handlers.NewRiskHandler(s.orchestrator)
	backtestHandler := handlers.NewBacktestHandler(s.orchestrator)
	positionHandler := handlers.NewPositionHandler(s.orchestrator)
	orderHandler := handlers.NewOrderHandler(s.orchestrator)
	candleHandler := handlers.NewCandleHandler(s.orchestrator)

	// Health check
	s.echo.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "healthy"})
	})

	// API v1 group
	v1 := s.echo.Group("/api/v1")

	// Dashboard routes
	v1.GET("/dashboard", dashboardHandler.GetDashboard)
	v1.GET("/dashboard/summary", dashboardHandler.GetSummary)
	v1.GET("/dashboard/equity-curve", dashboardHandler.GetEquityCurve)
	v1.GET("/dashboard/performance", dashboardHandler.GetPerformance)

	// Trading routes
	v1.GET("/trading/state", tradingHandler.GetState)
	v1.POST("/trading/start", tradingHandler.Start)
	v1.POST("/trading/stop", tradingHandler.Stop)
	v1.POST("/trading/pause", tradingHandler.Pause)
	v1.POST("/trading/resume", tradingHandler.Resume)
	v1.GET("/trading/mode", tradingHandler.GetMode)
	v1.POST("/trading/mode", tradingHandler.SetMode)

	// Strategy routes
	v1.GET("/strategies", strategyHandler.GetStrategies)
	v1.GET("/strategies/:name", strategyHandler.GetStrategy)
	v1.PUT("/strategies/:name", strategyHandler.UpdateStrategy)
	v1.POST("/strategies/:name/enable", strategyHandler.EnableStrategy)
	v1.POST("/strategies/:name/disable", strategyHandler.DisableStrategy)
	v1.GET("/strategies/:name/signals", strategyHandler.GetSignals)
	v1.GET("/regime", strategyHandler.GetRegime)

	// Risk routes
	v1.GET("/risk", riskHandler.GetRiskStatus)
	v1.GET("/risk/config", riskHandler.GetConfig)
	v1.PUT("/risk/config", riskHandler.UpdateConfig)
	v1.GET("/risk/limits", riskHandler.GetLimits)
	v1.GET("/risk/drawdown", riskHandler.GetDrawdown)
	v1.GET("/risk/events", riskHandler.GetEvents)
	v1.POST("/risk/circuit-breaker/reset", riskHandler.ResetCircuitBreaker)

	// Position routes
	v1.GET("/positions", positionHandler.GetPositions)
	v1.GET("/positions/:id", positionHandler.GetPosition)
	v1.POST("/positions/:id/close", positionHandler.ClosePosition)
	v1.PUT("/positions/:id/stop-loss", positionHandler.UpdateStopLoss)
	v1.PUT("/positions/:id/take-profit", positionHandler.UpdateTakeProfit)

	// Order routes
	v1.GET("/orders", orderHandler.GetOrders)
	v1.GET("/orders/open", orderHandler.GetOpenOrders)
	v1.POST("/orders", orderHandler.PlaceOrder)
	v1.DELETE("/orders/:id", orderHandler.CancelOrder)

	// Candle/Market Data routes
	v1.GET("/candles", candleHandler.GetCandles)
	v1.GET("/candles/:symbol/:timeframe", candleHandler.GetCandlesBySymbol)
	v1.GET("/ticker", candleHandler.GetTicker)
	v1.GET("/indicators", candleHandler.GetIndicators)

	// Backtest routes
	v1.POST("/backtest", backtestHandler.RunBacktest)
	v1.GET("/backtest/results", backtestHandler.GetResults)
	v1.GET("/backtest/results/:id", backtestHandler.GetResult)

	// Settings routes - for UI configuration
	settingsHandler := handlers.NewSettingsHandler(s.orchestrator)
	v1.GET("/settings", settingsHandler.GetSettings)
	v1.POST("/settings/reset", settingsHandler.ResetSettings)
	v1.GET("/settings/trading", settingsHandler.GetTradingSettings)
	v1.PUT("/settings/trading", settingsHandler.UpdateTradingSettings)
	v1.GET("/settings/binance", settingsHandler.GetBinanceSettings)
	v1.PUT("/settings/binance", settingsHandler.UpdateBinanceSettings)
	v1.GET("/settings/risk", settingsHandler.GetRiskSettings)
	v1.PUT("/settings/risk", settingsHandler.UpdateRiskSettings)
	v1.GET("/settings/indicators", settingsHandler.GetIndicatorSettings)
	v1.PUT("/settings/indicators", settingsHandler.UpdateIndicatorSettings)
	v1.GET("/settings/strategies", settingsHandler.GetStrategySettings)
	v1.PUT("/settings/strategies", settingsHandler.UpdateStrategySettings)

	// WebSocket
	s.echo.GET("/ws", s.handleWebSocket)
}

// handleWebSocket handles WebSocket connections
func (s *Server) handleWebSocket(c echo.Context) error {
	return websocket.HandleConnection(c, s.wsHub, s.orchestrator)
}

// Start starts the server
func (s *Server) Start() error {
	// Start WebSocket hub
	go s.wsHub.Run()

	// Connect orchestrator broadcasts to WebSocket hub
	go s.forwardBroadcasts()

	log.Info().Str("port", s.config.Port).Msg("Starting API server")

	return s.echo.Start(s.config.Port)
}

// forwardBroadcasts forwards orchestrator broadcasts to WebSocket hub
func (s *Server) forwardBroadcasts() {
	if s.orchestrator == nil {
		return
	}

	ch := s.orchestrator.Subscribe("api-server")
	if ch == nil {
		return
	}

	for msg := range ch {
		s.wsHub.Broadcast(msg)
	}
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
	defer cancel()

	// Close WebSocket hub
	s.wsHub.Close()

	// Unsubscribe from orchestrator
	if s.orchestrator != nil {
		s.orchestrator.Unsubscribe("api-server")
	}

	log.Info().Msg("Shutting down API server")
	return s.echo.Shutdown(ctx)
}

// GetEcho returns the Echo instance
func (s *Server) GetEcho() *echo.Echo {
	return s.echo
}
