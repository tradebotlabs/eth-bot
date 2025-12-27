package api

import (
	"context"
	"net/http"
	"time"

	"github.com/eth-trading/internal/api/handlers"
	"github.com/eth-trading/internal/api/middleware"
	"github.com/eth-trading/internal/api/websocket"
	"github.com/eth-trading/internal/auth"
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
	authService  *auth.Service
	wsHub        *websocket.Hub
}

// NewServer creates a new API server
func NewServer(config *ServerConfig, orch *orchestrator.Orchestrator, authService *auth.Service) *Server {
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
		authService:  authService,
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
	// Create auth middleware
	authMiddleware := middleware.NewAuthMiddleware(s.authService)

	// Create handlers
	authHandler := handlers.NewAuthHandler(s.authService)
	dashboardHandler := handlers.NewDashboardHandler(s.orchestrator)
	tradingHandler := handlers.NewTradingHandler(s.orchestrator)
	strategyHandler := handlers.NewStrategyHandler(s.orchestrator)
	riskHandler := handlers.NewRiskHandler(s.orchestrator)
	backtestHandler := handlers.NewBacktestHandler(s.orchestrator)
	positionHandler := handlers.NewPositionHandler(s.orchestrator)
	orderHandler := handlers.NewOrderHandler(s.orchestrator)
	candleHandler := handlers.NewCandleHandler(s.orchestrator)

	// Health check (public)
	s.echo.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "healthy"})
	})

	// API v1 group
	v1 := s.echo.Group("/api/v1")

	// Auth routes (public - no authentication required)
	authGroup := v1.Group("/auth")
	authGroup.POST("/register", authHandler.Register)
	authGroup.POST("/login", authHandler.Login)
	authGroup.POST("/refresh", authHandler.RefreshToken)
	authGroup.POST("/password-reset", authHandler.RequestPasswordReset)
	authGroup.POST("/password-reset/confirm", authHandler.ConfirmPasswordReset)

	// Protected auth routes (require authentication)
	authProtected := authGroup.Group("", authMiddleware.Authenticate)
	authProtected.POST("/logout", authHandler.Logout)
	authProtected.GET("/me", authHandler.GetMe)
	authProtected.POST("/change-password", authHandler.ChangePassword)

	// Protected routes (require authentication)
	protected := v1.Group("", authMiddleware.Authenticate)

	// Dashboard routes
	protected.GET("/dashboard", dashboardHandler.GetDashboard)
	protected.GET("/dashboard/summary", dashboardHandler.GetSummary)
	protected.GET("/dashboard/equity-curve", dashboardHandler.GetEquityCurve)
	protected.GET("/dashboard/performance", dashboardHandler.GetPerformance)

	// Trading routes
	protected.GET("/trading/state", tradingHandler.GetState)
	protected.POST("/trading/start", tradingHandler.Start)
	protected.POST("/trading/stop", tradingHandler.Stop)
	protected.POST("/trading/pause", tradingHandler.Pause)
	protected.POST("/trading/resume", tradingHandler.Resume)
	protected.GET("/trading/mode", tradingHandler.GetMode)
	protected.POST("/trading/mode", tradingHandler.SetMode)

	// Strategy routes
	protected.GET("/strategies", strategyHandler.GetStrategies)
	protected.GET("/strategies/:name", strategyHandler.GetStrategy)
	protected.PUT("/strategies/:name", strategyHandler.UpdateStrategy)
	protected.POST("/strategies/:name/enable", strategyHandler.EnableStrategy)
	protected.POST("/strategies/:name/disable", strategyHandler.DisableStrategy)
	protected.GET("/strategies/:name/signals", strategyHandler.GetSignals)
	protected.GET("/regime", strategyHandler.GetRegime)

	// Risk routes
	protected.GET("/risk", riskHandler.GetRiskStatus)
	protected.GET("/risk/config", riskHandler.GetConfig)
	protected.PUT("/risk/config", riskHandler.UpdateConfig)
	protected.GET("/risk/limits", riskHandler.GetLimits)
	protected.GET("/risk/drawdown", riskHandler.GetDrawdown)
	protected.GET("/risk/events", riskHandler.GetEvents)
	protected.POST("/risk/circuit-breaker/reset", riskHandler.ResetCircuitBreaker)

	// Position routes
	protected.GET("/positions", positionHandler.GetPositions)
	protected.GET("/positions/:id", positionHandler.GetPosition)
	protected.POST("/positions/:id/close", positionHandler.ClosePosition)
	protected.PUT("/positions/:id/stop-loss", positionHandler.UpdateStopLoss)
	protected.PUT("/positions/:id/take-profit", positionHandler.UpdateTakeProfit)

	// Order routes
	protected.GET("/orders", orderHandler.GetOrders)
	protected.GET("/orders/open", orderHandler.GetOpenOrders)
	protected.POST("/orders", orderHandler.PlaceOrder)
	protected.DELETE("/orders/:id", orderHandler.CancelOrder)

	// Candle/Market Data routes (public - no auth needed for market data)
	v1.GET("/candles", candleHandler.GetCandles)
	v1.GET("/candles/:symbol/:timeframe", candleHandler.GetCandlesBySymbol)
	v1.GET("/ticker", candleHandler.GetTicker)
	v1.GET("/indicators", candleHandler.GetIndicators)

	// Backtest routes
	protected.POST("/backtest", backtestHandler.RunBacktest)
	protected.GET("/backtest/results", backtestHandler.GetResults)
	protected.GET("/backtest/results/:id", backtestHandler.GetResult)

	// Settings routes - for UI configuration
	settingsHandler := handlers.NewSettingsHandler(s.orchestrator)
	protected.GET("/settings", settingsHandler.GetSettings)
	protected.POST("/settings/reset", settingsHandler.ResetSettings)
	protected.GET("/settings/trading", settingsHandler.GetTradingSettings)
	protected.PUT("/settings/trading", settingsHandler.UpdateTradingSettings)
	protected.GET("/settings/binance", settingsHandler.GetBinanceSettings)
	protected.PUT("/settings/binance", settingsHandler.UpdateBinanceSettings)
	protected.GET("/settings/risk", settingsHandler.GetRiskSettings)
	protected.PUT("/settings/risk", settingsHandler.UpdateRiskSettings)
	protected.GET("/settings/indicators", settingsHandler.GetIndicatorSettings)
	protected.PUT("/settings/indicators", settingsHandler.UpdateIndicatorSettings)
	protected.GET("/settings/strategies", settingsHandler.GetStrategySettings)
	protected.PUT("/settings/strategies", settingsHandler.UpdateStrategySettings)

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
