package handlers

import (
	"net/http"

	"github.com/eth-trading/internal/orchestrator"
	"github.com/eth-trading/internal/risk"
	"github.com/labstack/echo/v4"
)

// RiskHandler handles risk management endpoints
type RiskHandler struct {
	orchestrator *orchestrator.Orchestrator
	riskManager  *risk.Manager
}

// NewRiskHandler creates a new risk handler
func NewRiskHandler(orch *orchestrator.Orchestrator) *RiskHandler {
	rh := &RiskHandler{
		orchestrator: orch,
	}
	// Get risk manager from orchestrator
	if orch != nil {
		rh.riskManager = orch.GetRiskManager()
	}
	return rh
}

// SetRiskManager sets the risk manager
func (h *RiskHandler) SetRiskManager(rm *risk.Manager) {
	h.riskManager = rm
}

// RiskStatusResponse represents risk status
type RiskStatusResponse struct {
	Level            string  `json:"level"`
	CurrentDrawdown  float64 `json:"currentDrawdown"`
	MaxDrawdown      float64 `json:"maxDrawdown"`
	DailyLossUsed    float64 `json:"dailyLossUsed"`
	DailyLossLimit   float64 `json:"dailyLossLimit"`
	DailyLossPercent float64 `json:"dailyLossPercent"`
	WeeklyLossUsed   float64 `json:"weeklyLossUsed"`
	WeeklyLossLimit  float64 `json:"weeklyLossLimit"`
	OpenPositions    int     `json:"openPositions"`
	MaxPositions     int     `json:"maxPositions"`
	IsHalted         bool    `json:"isHalted"`
	HaltReason       string  `json:"haltReason,omitempty"`
	IsWithinLimits   bool    `json:"isWithinLimits"`
	Warnings         []string `json:"warnings,omitempty"`
}

// GetRiskStatus returns current risk status
func (h *RiskHandler) GetRiskStatus(c echo.Context) error {
	if h.riskManager == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "Risk manager not available"})
	}

	state := h.riskManager.GetAccountState()
	limits := h.riskManager.GetRiskLimits()

	response := RiskStatusResponse{
		Level:            determineRiskLevel(state.CurrentDrawdown),
		CurrentDrawdown:  state.CurrentDrawdown,
		MaxDrawdown:      limits.DrawdownLimit,
		DailyLossUsed:    limits.DailyLossUsed,
		DailyLossLimit:   limits.DailyLossLimit,
		DailyLossPercent: limits.DailyLossPercent,
		WeeklyLossUsed:   limits.WeeklyLossUsed,
		WeeklyLossLimit:  limits.WeeklyLossLimit,
		OpenPositions:    state.OpenPositions,
		MaxPositions:     limits.PositionsLimit,
		IsHalted:         state.IsHalted,
		HaltReason:       state.HaltReason,
		IsWithinLimits:   limits.IsWithinLimits,
		Warnings:         limits.LimitBreaches,
	}

	return c.JSON(http.StatusOK, response)
}

// RiskConfigResponse represents risk configuration
type RiskConfigResponse struct {
	MaxPositionSize       float64 `json:"maxPositionSize"`
	MaxPositionValue      float64 `json:"maxPositionValue"`
	MaxRiskPerTrade       float64 `json:"maxRiskPerTrade"`
	MinRiskRewardRatio    float64 `json:"minRiskRewardRatio"`
	MaxDailyLoss          float64 `json:"maxDailyLoss"`
	MaxWeeklyLoss         float64 `json:"maxWeeklyLoss"`
	MaxTotalDrawdown      float64 `json:"maxTotalDrawdown"`
	MaxOpenPositions      int     `json:"maxOpenPositions"`
	MaxLeverage           float64 `json:"maxLeverage"`
	EnableCircuitBreaker  bool    `json:"enableCircuitBreaker"`
	ConsecutiveLossLimit  int     `json:"consecutiveLossLimit"`
	AdjustForVolatility   bool    `json:"adjustForVolatility"`
	TradingHoursOnly      bool    `json:"tradingHoursOnly"`
}

// GetConfig returns risk configuration
func (h *RiskHandler) GetConfig(c echo.Context) error {
	if h.riskManager == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "Risk manager not available"})
	}

	config := h.riskManager.GetConfig()

	response := RiskConfigResponse{
		MaxPositionSize:      config.MaxPositionSize,
		MaxPositionValue:     config.MaxPositionValue,
		MaxRiskPerTrade:      config.MaxRiskPerTrade,
		MinRiskRewardRatio:   config.MinRiskRewardRatio,
		MaxDailyLoss:         config.MaxDailyLoss,
		MaxWeeklyLoss:        config.MaxWeeklyLoss,
		MaxTotalDrawdown:     config.MaxTotalDrawdown,
		MaxOpenPositions:     config.MaxOpenPositions,
		MaxLeverage:          config.MaxLeverage,
		EnableCircuitBreaker: config.EnableCircuitBreaker,
		ConsecutiveLossLimit: config.ConsecutiveLossLimit,
		AdjustForVolatility:  config.AdjustForVolatility,
		TradingHoursOnly:     config.TradingHoursOnly,
	}

	return c.JSON(http.StatusOK, response)
}

// UpdateConfigRequest represents risk config update request
type UpdateConfigRequest struct {
	MaxRiskPerTrade      *float64 `json:"maxRiskPerTrade,omitempty"`
	MaxDailyLoss         *float64 `json:"maxDailyLoss,omitempty"`
	MaxTotalDrawdown     *float64 `json:"maxTotalDrawdown,omitempty"`
	MaxOpenPositions     *int     `json:"maxOpenPositions,omitempty"`
	EnableCircuitBreaker *bool    `json:"enableCircuitBreaker,omitempty"`
}

// UpdateConfig updates risk configuration
func (h *RiskHandler) UpdateConfig(c echo.Context) error {
	if h.riskManager == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "Risk manager not available"})
	}

	var req UpdateConfigRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	config := h.riskManager.GetConfig()

	// Apply updates
	if req.MaxRiskPerTrade != nil {
		config.MaxRiskPerTrade = *req.MaxRiskPerTrade
	}
	if req.MaxDailyLoss != nil {
		config.MaxDailyLoss = *req.MaxDailyLoss
	}
	if req.MaxTotalDrawdown != nil {
		config.MaxTotalDrawdown = *req.MaxTotalDrawdown
	}
	if req.MaxOpenPositions != nil {
		config.MaxOpenPositions = *req.MaxOpenPositions
	}
	if req.EnableCircuitBreaker != nil {
		config.EnableCircuitBreaker = *req.EnableCircuitBreaker
	}

	h.riskManager.UpdateConfig(config)

	return c.JSON(http.StatusOK, map[string]string{"status": "updated"})
}

// GetLimits returns current risk limits
func (h *RiskHandler) GetLimits(c echo.Context) error {
	if h.riskManager == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "Risk manager not available"})
	}

	limits := h.riskManager.GetRiskLimits()
	return c.JSON(http.StatusOK, limits)
}

// DrawdownResponse represents drawdown information
type DrawdownResponse struct {
	Current          float64 `json:"current"`
	Max              float64 `json:"max"`
	RecoveryRequired float64 `json:"recoveryRequired"`
	IsAtPeak         bool    `json:"isAtPeak"`
}

// GetDrawdown returns drawdown information
func (h *RiskHandler) GetDrawdown(c echo.Context) error {
	if h.riskManager == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "Risk manager not available"})
	}

	info := h.riskManager.GetDrawdownInfo()

	response := DrawdownResponse{
		Current:          info.CurrentDrawdown,
		Max:              info.MaxDrawdown,
		RecoveryRequired: info.RecoveryRequired,
		IsAtPeak:         info.CurrentDrawdown == 0,
	}

	return c.JSON(http.StatusOK, response)
}

// RiskEventResponse represents a risk event
type RiskEventResponse struct {
	Type      string `json:"type"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	Handled   bool   `json:"handled"`
}

// GetEvents returns recent risk events
func (h *RiskHandler) GetEvents(c echo.Context) error {
	if h.riskManager == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "Risk manager not available"})
	}

	events := h.riskManager.GetRecentEvents(20)

	response := make([]RiskEventResponse, len(events))
	for i, e := range events {
		response[i] = RiskEventResponse{
			Type:      e.Type.String(),
			Level:     e.Level.String(),
			Message:   e.Message,
			Timestamp: e.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
			Handled:   e.Handled,
		}
	}

	return c.JSON(http.StatusOK, response)
}

// ResetCircuitBreaker resets the circuit breaker
func (h *RiskHandler) ResetCircuitBreaker(c echo.Context) error {
	if h.riskManager == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "Risk manager not available"})
	}

	h.riskManager.ResetCircuitBreaker()
	return c.JSON(http.StatusOK, map[string]string{"status": "reset"})
}

// Helper function to determine risk level string
func determineRiskLevel(drawdown float64) string {
	switch {
	case drawdown >= 0.15:
		return "CRITICAL"
	case drawdown >= 0.10:
		return "HIGH"
	case drawdown >= 0.05:
		return "MEDIUM"
	default:
		return "LOW"
	}
}
