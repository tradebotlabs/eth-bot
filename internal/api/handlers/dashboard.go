package handlers

import (
	"net/http"
	"time"

	"github.com/eth-trading/internal/execution"
	"github.com/eth-trading/internal/orchestrator"
	"github.com/labstack/echo/v4"
)

// DashboardHandler handles dashboard endpoints
type DashboardHandler struct {
	orchestrator *orchestrator.Orchestrator
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(orch *orchestrator.Orchestrator) *DashboardHandler {
	return &DashboardHandler{orchestrator: orch}
}

// DashboardResponse represents dashboard data
type DashboardResponse struct {
	State        *orchestrator.TradingState   `json:"state"`
	Summary      *orchestrator.AccountSummary `json:"summary"`
	Performance  *PerformanceData             `json:"performance"`
	RecentTrades []TradeData                  `json:"recentTrades"`
	Positions    []PositionData               `json:"positions"`
	Signals      []orchestrator.SignalRecord  `json:"signals"`
	Timestamp    time.Time                    `json:"timestamp"`
}

// PerformanceData represents performance metrics
type PerformanceData struct {
	TotalReturn      float64 `json:"totalReturn"`
	DailyReturn      float64 `json:"dailyReturn"`
	WeeklyReturn     float64 `json:"weeklyReturn"`
	MonthlyReturn    float64 `json:"monthlyReturn"`
	SharpeRatio      float64 `json:"sharpeRatio"`
	MaxDrawdown      float64 `json:"maxDrawdown"`
	WinRate          float64 `json:"winRate"`
	ProfitFactor     float64 `json:"profitFactor"`
	TotalTrades      int     `json:"totalTrades"`
	WinningTrades    int     `json:"winningTrades"`
	LosingTrades     int     `json:"losingTrades"`
}

// TradeData represents a trade for API response
type TradeData struct {
	ID          string    `json:"id"`
	Symbol      string    `json:"symbol"`
	Side        string    `json:"side"`
	Type        string    `json:"type"`
	Quantity    float64   `json:"quantity"`
	Price       float64   `json:"price"`
	RealizedPnL float64   `json:"realizedPnL"`
	Commission  float64   `json:"commission"`
	Strategy    string    `json:"strategy"`
	Timestamp   time.Time `json:"timestamp"`
}

// PositionData represents a position for API response
type PositionData struct {
	ID            int64     `json:"id"`
	Symbol        string    `json:"symbol"`
	Side          string    `json:"side"`
	Quantity      float64   `json:"quantity"`
	EntryPrice    float64   `json:"entryPrice"`
	CurrentPrice  float64   `json:"currentPrice"`
	StopLoss      float64   `json:"stopLoss"`
	TakeProfit    float64   `json:"takeProfit"`
	UnrealizedPnL float64   `json:"unrealizedPnL"`
	UnrealizedPct float64   `json:"unrealizedPct"`
	Strategy      string    `json:"strategy"`
	OpenTime      time.Time `json:"openTime"`
	Duration      string    `json:"duration"`
}

// GetDashboard returns full dashboard data
func (h *DashboardHandler) GetDashboard(c echo.Context) error {
	if h.orchestrator == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "Orchestrator not available"})
	}

	state := h.orchestrator.GetState()

	// Get positions
	var positions []PositionData
	// Would get from executor if available

	// Get recent signals
	signals := h.orchestrator.GetSignals(20)

	response := DashboardResponse{
		State:        state,
		Performance:  &PerformanceData{},
		RecentTrades: []TradeData{},
		Positions:    positions,
		Signals:      signals,
		Timestamp:    time.Now(),
	}

	return c.JSON(http.StatusOK, response)
}

// GetSummary returns account summary
func (h *DashboardHandler) GetSummary(c echo.Context) error {
	if h.orchestrator == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "Orchestrator not available"})
	}

	state := h.orchestrator.GetState()

	summary := &orchestrator.AccountSummary{
		Equity:           state.Equity,
		AvailableBalance: state.AvailableBalance,
		UnrealizedPnL:    state.UnrealizedPnL,
		RealizedPnL:      state.RealizedPnL,
		DailyPnL:         state.DailyPnL,
		OpenPositions:    state.OpenPositions,
		TotalTrades:      state.TotalTrades,
		WinRate:          state.WinRate,
	}

	return c.JSON(http.StatusOK, summary)
}

// EquityCurvePoint represents a point on the equity curve
type EquityCurvePoint struct {
	Time       time.Time `json:"time"`
	Equity     float64   `json:"equity"`
	Drawdown   float64   `json:"drawdown"`
	Return     float64   `json:"return"`
}

// GetEquityCurve returns the equity curve data
func (h *DashboardHandler) GetEquityCurve(c echo.Context) error {
	// In a real implementation, this would fetch from storage
	// For now, return empty array
	return c.JSON(http.StatusOK, []EquityCurvePoint{})
}

// GetPerformance returns performance metrics
func (h *DashboardHandler) GetPerformance(c echo.Context) error {
	if h.orchestrator == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "Orchestrator not available"})
	}

	state := h.orchestrator.GetState()

	performance := &PerformanceData{
		MaxDrawdown:   state.MaxDrawdown,
		WinRate:       state.WinRate,
		TotalTrades:   state.TotalTrades,
	}

	return c.JSON(http.StatusOK, performance)
}

// Helper to convert execution position to API position
func convertPosition(pos *execution.Position) PositionData {
	duration := time.Since(pos.OpenTime)
	durationStr := duration.Round(time.Minute).String()

	return PositionData{
		ID:            pos.ID,
		Symbol:        pos.Symbol,
		Side:          string(pos.Side),
		Quantity:      pos.Quantity,
		EntryPrice:    pos.EntryPrice,
		CurrentPrice:  pos.CurrentPrice,
		StopLoss:      pos.StopLoss,
		TakeProfit:    pos.TakeProfit,
		UnrealizedPnL: pos.UnrealizedPnL,
		UnrealizedPct: pos.UnrealizedPnLPct,
		Strategy:      pos.Strategy,
		OpenTime:      pos.OpenTime,
		Duration:      durationStr,
	}
}
