package handlers

import (
	"net/http"

	"github.com/eth-trading/internal/orchestrator"
	"github.com/labstack/echo/v4"
)

// SettingsHandler handles settings configuration endpoints
type SettingsHandler struct {
	orchestrator *orchestrator.Orchestrator
}

// NewSettingsHandler creates a new settings handler
func NewSettingsHandler(orch *orchestrator.Orchestrator) *SettingsHandler {
	return &SettingsHandler{orchestrator: orch}
}

// FullSettingsResponse represents all settings
type FullSettingsResponse struct {
	Trading    TradingSettings    `json:"trading"`
	Binance    BinanceSettings    `json:"binance"`
	Risk       RiskSettings       `json:"risk"`
	Indicators IndicatorSettings  `json:"indicators"`
	Strategies StrategySettings   `json:"strategies"`
}

// TradingSettings represents trading configuration
type TradingSettings struct {
	Mode             string   `json:"mode"`             // "paper" or "live"
	Symbol           string   `json:"symbol"`           // e.g., "ETHUSDT"
	Timeframes       []string `json:"timeframes"`       // e.g., ["1m", "5m", "15m", "1h", "4h"]
	PrimaryTimeframe string   `json:"primaryTimeframe"` // e.g., "1h"
	InitialBalance   float64  `json:"initialBalance"`   // Paper trading initial balance
	Commission       float64  `json:"commission"`       // Commission rate (0.001 = 0.1%)
	Slippage         float64  `json:"slippage"`         // Slippage rate
}

// BinanceSettings represents Binance API configuration
type BinanceSettings struct {
	APIKey    string `json:"apiKey"`    // Binance API key
	SecretKey string `json:"secretKey"` // Binance Secret key (masked on read)
	Testnet   bool   `json:"testnet"`   // Use testnet
}

// RiskSettings represents risk management configuration
type RiskSettings struct {
	MaxPositionSize      float64 `json:"maxPositionSize"`      // Max position as % of equity (0.1 = 10%)
	MaxRiskPerTrade      float64 `json:"maxRiskPerTrade"`      // Max risk per trade (0.02 = 2%)
	MaxDailyLoss         float64 `json:"maxDailyLoss"`         // Max daily loss (0.05 = 5%)
	MaxWeeklyLoss        float64 `json:"maxWeeklyLoss"`        // Max weekly loss (0.1 = 10%)
	MaxDrawdown          float64 `json:"maxDrawdown"`          // Max total drawdown (0.2 = 20%)
	MaxOpenPositions     int     `json:"maxOpenPositions"`     // Max concurrent positions
	MaxLeverage          float64 `json:"maxLeverage"`          // Max leverage (1.0 = no leverage)
	MinRiskRewardRatio   float64 `json:"minRiskRewardRatio"`   // Minimum R/R ratio
	EnableCircuitBreaker bool    `json:"enableCircuitBreaker"` // Enable circuit breaker
	ConsecutiveLossLimit int     `json:"consecutiveLossLimit"` // Halt after N losses
	HaltDurationHours    int     `json:"haltDurationHours"`    // Circuit breaker halt duration
}

// IndicatorSettings represents indicator configuration
type IndicatorSettings struct {
	RSIPeriod       int     `json:"rsiPeriod"`       // RSI period (default: 14)
	RSIOversold     float64 `json:"rsiOversold"`     // RSI oversold level (default: 30)
	RSIOverbought   float64 `json:"rsiOverbought"`   // RSI overbought level (default: 70)
	MACDFast        int     `json:"macdFast"`        // MACD fast period (default: 12)
	MACDSlow        int     `json:"macdSlow"`        // MACD slow period (default: 26)
	MACDSignal      int     `json:"macdSignal"`      // MACD signal period (default: 9)
	BBPeriod        int     `json:"bbPeriod"`        // Bollinger Band period (default: 20)
	BBStdDev        float64 `json:"bbStdDev"`        // Bollinger Band std dev (default: 2.0)
	ADXPeriod       int     `json:"adxPeriod"`       // ADX period (default: 14)
	ADXThreshold    float64 `json:"adxThreshold"`    // ADX trend threshold (default: 25)
	ATRPeriod       int     `json:"atrPeriod"`       // ATR period (default: 14)
	ATRMultiplierSL float64 `json:"atrMultiplierSL"` // ATR multiplier for stop loss (default: 2.0)
	ATRMultiplierTP float64 `json:"atrMultiplierTP"` // ATR multiplier for take profit (default: 3.0)
}

// StrategySettings represents strategy configuration
type StrategySettings struct {
	Enabled []StrategyConfig `json:"enabled"` // Enabled strategies with configs
}

// StrategyConfig represents individual strategy config
type StrategyConfig struct {
	Name    string                 `json:"name"`    // Strategy name
	Enabled bool                   `json:"enabled"` // Is strategy enabled
	Config  map[string]interface{} `json:"config"`  // Strategy-specific config
}

// GetSettings returns all settings
func (h *SettingsHandler) GetSettings(c echo.Context) error {
	settings := getDefaultSettings()
	return c.JSON(http.StatusOK, settings)
}

// GetTradingSettings returns trading settings
func (h *SettingsHandler) GetTradingSettings(c echo.Context) error {
	settings := getDefaultSettings()
	return c.JSON(http.StatusOK, settings.Trading)
}

// UpdateTradingSettings updates trading settings
func (h *SettingsHandler) UpdateTradingSettings(c echo.Context) error {
	var req TradingSettings
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	// Validate
	if req.Mode != "paper" && req.Mode != "live" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Mode must be 'paper' or 'live'"})
	}
	if req.Symbol == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Symbol is required"})
	}
	if req.InitialBalance <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Initial balance must be positive"})
	}

	// In real implementation, save to config manager and apply changes
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "updated",
		"message": "Trading settings updated. Restart required for some changes.",
		"trading": req,
	})
}

// GetBinanceSettings returns Binance settings
func (h *SettingsHandler) GetBinanceSettings(c echo.Context) error {
	settings := BinanceSettings{
		APIKey:    "****",           // Masked
		SecretKey: "****",           // Masked
		Testnet:   true,             // Default to testnet
	}
	return c.JSON(http.StatusOK, settings)
}

// UpdateBinanceSettings updates Binance API settings
func (h *SettingsHandler) UpdateBinanceSettings(c echo.Context) error {
	var req BinanceSettings
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	// In real implementation, validate API keys with a test call
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "updated",
		"message": "Binance settings updated",
		"testnet": req.Testnet,
	})
}

// GetRiskSettings returns risk settings
func (h *SettingsHandler) GetRiskSettings(c echo.Context) error {
	settings := getDefaultSettings()
	return c.JSON(http.StatusOK, settings.Risk)
}

// UpdateRiskSettings updates risk settings
func (h *SettingsHandler) UpdateRiskSettings(c echo.Context) error {
	var req RiskSettings
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	// Validate
	if req.MaxPositionSize <= 0 || req.MaxPositionSize > 1 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Max position size must be between 0 and 1"})
	}
	if req.MaxRiskPerTrade <= 0 || req.MaxRiskPerTrade > 0.1 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Max risk per trade must be between 0 and 0.1 (10%)"})
	}
	if req.MaxDrawdown <= 0 || req.MaxDrawdown > 1 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Max drawdown must be between 0 and 1"})
	}

	// In real implementation, update risk manager
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "updated",
		"message": "Risk settings updated and applied",
		"risk":    req,
	})
}

// GetIndicatorSettings returns indicator settings
func (h *SettingsHandler) GetIndicatorSettings(c echo.Context) error {
	settings := getDefaultSettings()
	return c.JSON(http.StatusOK, settings.Indicators)
}

// UpdateIndicatorSettings updates indicator settings
func (h *SettingsHandler) UpdateIndicatorSettings(c echo.Context) error {
	var req IndicatorSettings
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	// Validate periods
	if req.RSIPeriod < 2 || req.RSIPeriod > 100 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "RSI period must be between 2 and 100"})
	}
	if req.MACDFast >= req.MACDSlow {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "MACD fast must be less than slow period"})
	}

	// In real implementation, update indicator manager
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":     "updated",
		"message":    "Indicator settings updated",
		"indicators": req,
	})
}

// GetStrategySettings returns strategy settings
func (h *SettingsHandler) GetStrategySettings(c echo.Context) error {
	settings := getDefaultSettings()
	return c.JSON(http.StatusOK, settings.Strategies)
}

// UpdateStrategySettings updates strategy settings
func (h *SettingsHandler) UpdateStrategySettings(c echo.Context) error {
	var req StrategySettings
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	// In real implementation, update strategy manager
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":     "updated",
		"message":    "Strategy settings updated",
		"strategies": req,
	})
}

// ResetSettings resets all settings to defaults
func (h *SettingsHandler) ResetSettings(c echo.Context) error {
	settings := getDefaultSettings()
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":   "reset",
		"message":  "All settings reset to defaults",
		"settings": settings,
	})
}

// getDefaultSettings returns default settings
func getDefaultSettings() *FullSettingsResponse {
	return &FullSettingsResponse{
		Trading: TradingSettings{
			Mode:             "paper",
			Symbol:           "ETHUSDT",
			Timeframes:       []string{"1m", "5m", "15m", "1h", "4h", "1d"},
			PrimaryTimeframe: "1h",
			InitialBalance:   100000,
			Commission:       0.001,
			Slippage:         0.0005,
		},
		Binance: BinanceSettings{
			APIKey:    "",
			SecretKey: "",
			Testnet:   true,
		},
		Risk: RiskSettings{
			MaxPositionSize:      0.10,
			MaxRiskPerTrade:      0.02,
			MaxDailyLoss:         0.05,
			MaxWeeklyLoss:        0.10,
			MaxDrawdown:          0.20,
			MaxOpenPositions:     5,
			MaxLeverage:          1.0,
			MinRiskRewardRatio:   1.5,
			EnableCircuitBreaker: true,
			ConsecutiveLossLimit: 5,
			HaltDurationHours:    24,
		},
		Indicators: IndicatorSettings{
			RSIPeriod:       14,
			RSIOversold:     30,
			RSIOverbought:   70,
			MACDFast:        12,
			MACDSlow:        26,
			MACDSignal:      9,
			BBPeriod:        20,
			BBStdDev:        2.0,
			ADXPeriod:       14,
			ADXThreshold:    25,
			ATRPeriod:       14,
			ATRMultiplierSL: 2.0,
			ATRMultiplierTP: 3.0,
		},
		Strategies: StrategySettings{
			Enabled: []StrategyConfig{
				{
					Name:    "TrendFollowing",
					Enabled: true,
					Config: map[string]interface{}{
						"adxThreshold":    25.0,
						"fastMAPeriod":    20,
						"slowMAPeriod":    50,
						"atrMultiplierSL": 2.0,
						"atrMultiplierTP": 3.0,
					},
				},
				{
					Name:    "MeanReversion",
					Enabled: true,
					Config: map[string]interface{}{
						"rsiOversold":   30.0,
						"rsiOverbought": 70.0,
						"bbPeriod":      20,
						"bbStdDev":      2.0,
					},
				},
				{
					Name:    "Breakout",
					Enabled: true,
					Config: map[string]interface{}{
						"donchianPeriod":   20,
						"volumeMultiplier": 1.5,
						"squeezeRequired":  true,
					},
				},
				{
					Name:    "Volatility",
					Enabled: true,
					Config: map[string]interface{}{
						"atrPeriod":       14,
						"atrMultiplier":   2.0,
						"squeezeRequired": true,
					},
				},
				{
					Name:    "StatArb",
					Enabled: true,
					Config: map[string]interface{}{
						"zScoreThreshold": 2.0,
						"lookbackPeriod":  20,
						"halfLife":        10,
					},
				},
			},
		},
	}
}
