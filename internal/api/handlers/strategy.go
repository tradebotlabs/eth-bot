package handlers

import (
	"net/http"

	"github.com/eth-trading/internal/orchestrator"
	"github.com/labstack/echo/v4"
)

// StrategyHandler handles strategy endpoints
type StrategyHandler struct {
	orchestrator *orchestrator.Orchestrator
}

// NewStrategyHandler creates a new strategy handler
func NewStrategyHandler(orch *orchestrator.Orchestrator) *StrategyHandler {
	return &StrategyHandler{orchestrator: orch}
}

// StrategyInfo represents strategy information
type StrategyInfo struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Enabled     bool                   `json:"enabled"`
	Config      map[string]interface{} `json:"config"`
	Performance *StrategyPerformance   `json:"performance,omitempty"`
}

// StrategyPerformance represents strategy performance metrics
type StrategyPerformance struct {
	TotalTrades   int     `json:"totalTrades"`
	WinRate       float64 `json:"winRate"`
	ProfitFactor  float64 `json:"profitFactor"`
	NetProfit     float64 `json:"netProfit"`
	Contribution  float64 `json:"contribution"`
}

// GetStrategies returns all strategies
func (h *StrategyHandler) GetStrategies(c echo.Context) error {
	strategies := []StrategyInfo{
		{
			Name:        "TrendFollowing",
			Description: "Trades in direction of prevailing trend using ADX and MA crossovers",
			Enabled:     true,
			Config: map[string]interface{}{
				"adxThreshold":     25.0,
				"fastMAPeriod":     20,
				"slowMAPeriod":     50,
				"atrMultiplierSL":  2.0,
				"atrMultiplierTP":  3.0,
			},
		},
		{
			Name:        "MeanReversion",
			Description: "Trades price returns to mean using RSI and Bollinger Bands",
			Enabled:     true,
			Config: map[string]interface{}{
				"rsiOversold":     30.0,
				"rsiOverbought":   70.0,
				"bbPeriod":        20,
				"bbStdDev":        2.0,
			},
		},
		{
			Name:        "Breakout",
			Description: "Trades breakouts from consolidation using Bollinger and Donchian",
			Enabled:     true,
			Config: map[string]interface{}{
				"donchianPeriod":  20,
				"squeezeRequired": true,
				"volumeMultiplier": 1.5,
			},
		},
		{
			Name:        "Volatility",
			Description: "Trades volatility expansion using ATR and TTM Squeeze",
			Enabled:     true,
			Config: map[string]interface{}{
				"atrPeriod":       14,
				"atrMultiplier":   2.0,
				"squeezeRequired": true,
			},
		},
		{
			Name:        "StatArb",
			Description: "Statistical arbitrage using Z-score and mean reversion",
			Enabled:     true,
			Config: map[string]interface{}{
				"zScoreThreshold": 2.0,
				"lookbackPeriod":  20,
				"halfLife":        10,
			},
		},
	}

	return c.JSON(http.StatusOK, strategies)
}

// GetStrategy returns a specific strategy
func (h *StrategyHandler) GetStrategy(c echo.Context) error {
	name := c.Param("name")

	// In real implementation, would fetch from strategy manager
	strategy := StrategyInfo{
		Name:    name,
		Enabled: true,
		Config:  map[string]interface{}{},
	}

	return c.JSON(http.StatusOK, strategy)
}

// UpdateStrategyRequest represents strategy update request
type UpdateStrategyRequest struct {
	Config map[string]interface{} `json:"config"`
}

// UpdateStrategy updates a strategy's configuration
func (h *StrategyHandler) UpdateStrategy(c echo.Context) error {
	name := c.Param("name")

	var req UpdateStrategyRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	// In real implementation, would update strategy config
	return c.JSON(http.StatusOK, map[string]string{"status": "updated", "strategy": name})
}

// EnableStrategy enables a strategy
func (h *StrategyHandler) EnableStrategy(c echo.Context) error {
	name := c.Param("name")

	strategyMgr := h.orchestrator.GetStrategyManager()
	if strategyMgr == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "Strategy manager not available"})
	}

	strategies := strategyMgr.GetStrategies()
	strategy, ok := strategies[name]
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Strategy not found"})
	}

	strategy.SetEnabled(true)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":   "enabled",
		"strategy": name,
		"enabled":  true,
	})
}

// DisableStrategy disables a strategy
func (h *StrategyHandler) DisableStrategy(c echo.Context) error {
	name := c.Param("name")

	strategyMgr := h.orchestrator.GetStrategyManager()
	if strategyMgr == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "Strategy manager not available"})
	}

	strategies := strategyMgr.GetStrategies()
	strategy, ok := strategies[name]
	if !ok {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Strategy not found"})
	}

	strategy.SetEnabled(false)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":   "disabled",
		"strategy": name,
		"enabled":  false,
	})
}

// SignalInfo represents a trading signal
type SignalInfo struct {
	Timestamp  string  `json:"timestamp"`
	Strategy   string  `json:"strategy"`
	Symbol     string  `json:"symbol"`
	Direction  string  `json:"direction"`
	EntryPrice float64 `json:"entryPrice"`
	StopLoss   float64 `json:"stopLoss"`
	TakeProfit float64 `json:"takeProfit"`
	Confidence float64 `json:"confidence"`
	Reason     string  `json:"reason"`
	Approved   bool    `json:"approved"`
}

// GetSignals returns recent signals for a strategy
func (h *StrategyHandler) GetSignals(c echo.Context) error {
	// In real implementation, would fetch from signal history
	signals := []SignalInfo{}
	return c.JSON(http.StatusOK, signals)
}

// RegimeInfo represents market regime information
type RegimeInfo struct {
	Current     string  `json:"current"`
	Confidence  float64 `json:"confidence"`
	ADX         float64 `json:"adx"`
	RSI         float64 `json:"rsi"`
	Volatility  string  `json:"volatility"`
	Description string  `json:"description"`
}

// GetRegime returns current market regime
func (h *StrategyHandler) GetRegime(c echo.Context) error {
	if h.orchestrator == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "Orchestrator not available"})
	}

	state := h.orchestrator.GetState()

	regime := RegimeInfo{
		Current:    state.CurrentRegime,
		Confidence: 0.8,
	}

	// Add description based on regime
	switch state.CurrentRegime {
	case "TRENDING":
		regime.Description = "Market is in a strong trend"
	case "MEAN_REVERTING":
		regime.Description = "Market is range-bound and mean-reverting"
	case "BREAKOUT":
		regime.Description = "Market is breaking out of consolidation"
	case "HIGH_VOLATILITY":
		regime.Description = "Market has elevated volatility"
	default:
		regime.Description = "Market is consolidating"
	}

	return c.JSON(http.StatusOK, regime)
}
