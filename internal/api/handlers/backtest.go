package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/eth-trading/internal/backtest"
	"github.com/eth-trading/internal/orchestrator"
	"github.com/eth-trading/internal/strategy"
	"github.com/labstack/echo/v4"
)

// BacktestHandler handles backtest endpoints
type BacktestHandler struct {
	orchestrator *orchestrator.Orchestrator
}

// NewBacktestHandler creates a new backtest handler
func NewBacktestHandler(orch *orchestrator.Orchestrator) *BacktestHandler {
	return &BacktestHandler{orchestrator: orch}
}

// BacktestRequest represents a backtest request
type BacktestRequest struct {
	Symbol         string   `json:"symbol"`
	Timeframe      string   `json:"timeframe"`
	StartDate      string   `json:"startDate"`      // ISO 8601 format
	EndDate        string   `json:"endDate"`        // ISO 8601 format
	InitialCapital float64  `json:"initialCapital"`
	Commission     float64  `json:"commission"`
	Slippage       float64  `json:"slippage"`
	Strategies     []string `json:"strategies"`
	RiskPerTrade   float64  `json:"riskPerTrade"`
}

// BacktestResponse represents a backtest response
type BacktestResponse struct {
	ID             string                `json:"id"`
	Status         string                `json:"status"`
	Config         BacktestConfigData    `json:"config"`
	Metrics        *BacktestMetricsData  `json:"metrics,omitempty"`
	EquityCurve    []EquityCurvePoint    `json:"equityCurve,omitempty"`
	Trades         []BacktestTradeData   `json:"trades,omitempty"`
	MonthlyReturns map[string]float64    `json:"monthlyReturns,omitempty"`
	StrategyStats  map[string]StrategyStatsData `json:"strategyStats,omitempty"`
	ExecutionTime  string                `json:"executionTime,omitempty"`
	Error          string                `json:"error,omitempty"`
}

// BacktestConfigData represents backtest config for API
type BacktestConfigData struct {
	Symbol         string   `json:"symbol"`
	Timeframe      string   `json:"timeframe"`
	StartDate      string   `json:"startDate"`
	EndDate        string   `json:"endDate"`
	InitialCapital float64  `json:"initialCapital"`
	Commission     float64  `json:"commission"`
	Slippage       float64  `json:"slippage"`
	Strategies     []string `json:"strategies"`
}

// BacktestMetricsData represents backtest metrics for API
type BacktestMetricsData struct {
	TotalReturn       float64 `json:"totalReturn"`
	AnnualizedReturn  float64 `json:"annualizedReturn"`
	SharpeRatio       float64 `json:"sharpeRatio"`
	SortinoRatio      float64 `json:"sortinoRatio"`
	CalmarRatio       float64 `json:"calmarRatio"`
	MaxDrawdown       float64 `json:"maxDrawdown"`
	TotalTrades       int     `json:"totalTrades"`
	WinningTrades     int     `json:"winningTrades"`
	LosingTrades      int     `json:"losingTrades"`
	WinRate           float64 `json:"winRate"`
	ProfitFactor      float64 `json:"profitFactor"`
	AvgWin            float64 `json:"avgWin"`
	AvgLoss           float64 `json:"avgLoss"`
	LargestWin        float64 `json:"largestWin"`
	LargestLoss       float64 `json:"largestLoss"`
	AvgHoldingTime    string  `json:"avgHoldingTime"`
	Expectancy        float64 `json:"expectancy"`
	RecoveryFactor    float64 `json:"recoveryFactor"`
	StartingCapital   float64 `json:"startingCapital"`
	EndingCapital     float64 `json:"endingCapital"`
	NetProfit         float64 `json:"netProfit"`
}

// BacktestTradeData represents a trade in backtest results
type BacktestTradeData struct {
	ID            int     `json:"id"`
	Symbol        string  `json:"symbol"`
	Strategy      string  `json:"strategy"`
	Direction     string  `json:"direction"`
	EntryTime     string  `json:"entryTime"`
	ExitTime      string  `json:"exitTime"`
	EntryPrice    float64 `json:"entryPrice"`
	ExitPrice     float64 `json:"exitPrice"`
	Quantity      float64 `json:"quantity"`
	NetProfit     float64 `json:"netProfit"`
	ReturnPercent float64 `json:"returnPercent"`
	ExitReason    string  `json:"exitReason"`
}

// StrategyStatsData represents per-strategy stats
type StrategyStatsData struct {
	Name         string  `json:"name"`
	TotalTrades  int     `json:"totalTrades"`
	WinRate      float64 `json:"winRate"`
	ProfitFactor float64 `json:"profitFactor"`
	NetProfit    float64 `json:"netProfit"`
	Contribution float64 `json:"contribution"`
}

// RunBacktest runs a backtest
func (h *BacktestHandler) RunBacktest(c echo.Context) error {
	var req BacktestRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	// Validate request
	if req.Symbol == "" {
		req.Symbol = "ETHUSDT"
	}
	if req.Timeframe == "" {
		req.Timeframe = "1h"
	}
	if req.InitialCapital <= 0 {
		req.InitialCapital = 100000
	}
	if req.Commission <= 0 {
		req.Commission = 0.001
	}
	if req.RiskPerTrade <= 0 {
		req.RiskPerTrade = 0.02
	}

	// Parse dates
	var startDate, endDate time.Time
	if req.StartDate != "" {
		startDate, _ = time.Parse("2006-01-02", req.StartDate)
	} else {
		startDate = time.Now().AddDate(0, -3, 0) // 3 months ago
	}
	if req.EndDate != "" {
		endDate, _ = time.Parse("2006-01-02", req.EndDate)
	} else {
		endDate = time.Now()
	}

	// Get data service
	dataService := h.orchestrator.GetDataService()
	if dataService == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "Data service not available"})
	}

	// Get historical candles
	storageCandles, err := dataService.GetHistoricalCandles(req.Symbol, req.Timeframe, startDate, endDate)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Failed to fetch historical data: %v", err)})
	}

	if len(storageCandles) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "No historical data available for the specified date range"})
	}

	// Convert storage candles to backtest candles
	backtestCandles := make([]backtest.Candle, len(storageCandles))
	for i, sc := range storageCandles {
		backtestCandles[i] = backtest.Candle{
			Timestamp: sc.OpenTime,
			Open:      sc.Open,
			High:      sc.High,
			Low:       sc.Low,
			Close:     sc.Close,
			Volume:    sc.Volume,
		}
	}

	historicalData := &backtest.HistoricalData{
		Symbol:    req.Symbol,
		Timeframe: req.Timeframe,
		Candles:   backtestCandles,
	}

	// Get strategy manager and selected strategies
	strategyMgr := h.orchestrator.GetStrategyManager()
	if strategyMgr == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "Strategy manager not available"})
	}

	allStrategies := strategyMgr.GetStrategies()
	var selectedStrategies []strategy.Strategy

	// If no strategies specified, use all enabled ones
	if len(req.Strategies) == 0 {
		for _, strat := range allStrategies {
			if strat.IsEnabled() {
				selectedStrategies = append(selectedStrategies, strat)
			}
		}
	} else {
		// Use requested strategies
		for _, name := range req.Strategies {
			if strat, ok := allStrategies[name]; ok {
				selectedStrategies = append(selectedStrategies, strat)
			}
		}
	}

	if len(selectedStrategies) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "No valid strategies selected"})
	}

	// Create backtest config
	btConfig := &backtest.Config{
		Symbol:         req.Symbol,
		Timeframe:      req.Timeframe,
		StartDate:      startDate,
		EndDate:        endDate,
		InitialCapital: req.InitialCapital,
		Commission:     req.Commission,
		Slippage:       req.Slippage,
		RiskPerTrade:   req.RiskPerTrade,
		Strategies:     selectedStrategies,
	}

	// Create and run backtest engine
	engine := backtest.NewEngine(btConfig)
	result, err := engine.Run(historicalData)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("Backtest failed: %v", err)})
	}

	// Convert result to API response
	response := h.convertBacktestResult(result)
	return c.JSON(http.StatusOK, response)
}

// convertBacktestResult converts backtest result to API response
func (h *BacktestHandler) convertBacktestResult(result *backtest.Result) BacktestResponse {
	// Convert equity curve
	equityCurve := make([]EquityCurvePoint, len(result.EquityCurve))
	for i, point := range result.EquityCurve {
		// Calculate return from initial capital
		returnPercent := 0.0
		if result.Config.InitialCapital > 0 {
			returnPercent = (point.Equity - result.Config.InitialCapital) / result.Config.InitialCapital
		}

		equityCurve[i] = EquityCurvePoint{
			Time:     point.Timestamp,
			Equity:   point.Equity,
			Drawdown: point.Drawdown,
			Return:   returnPercent,
		}
	}

	// Convert trades
	trades := make([]BacktestTradeData, len(result.Trades))
	for i, trade := range result.Trades {
		trades[i] = BacktestTradeData{
			ID:            int(trade.ID),
			Symbol:        trade.Symbol,
			Strategy:      trade.Strategy,
			Direction:     trade.Direction,
			EntryTime:     trade.EntryTime.Format("2006-01-02T15:04:05Z"),
			ExitTime:      trade.ExitTime.Format("2006-01-02T15:04:05Z"),
			EntryPrice:    trade.EntryPrice,
			ExitPrice:     trade.ExitPrice,
			Quantity:      trade.Quantity,
			NetProfit:     trade.NetProfit,
			ReturnPercent: trade.ReturnPercent,
			ExitReason:    trade.ExitReason,
		}
	}

	// Convert strategy stats
	strategyStats := make(map[string]StrategyStatsData)
	for name, stats := range result.StrategyStats {
		strategyStats[name] = StrategyStatsData{
			Name:         stats.Name,
			TotalTrades:  stats.TotalTrades,
			WinRate:      stats.WinRate,
			ProfitFactor: stats.ProfitFactor,
			NetProfit:    stats.NetProfit,
			Contribution: stats.Contribution,
		}
	}

	return BacktestResponse{
		ID:     "bt-" + time.Now().Format("20060102150405"),
		Status: "completed",
		Config: BacktestConfigData{
			Symbol:         result.Config.Symbol,
			Timeframe:      result.Config.Timeframe,
			StartDate:      result.Config.StartDate.Format("2006-01-02"),
			EndDate:        result.Config.EndDate.Format("2006-01-02"),
			InitialCapital: result.Config.InitialCapital,
			Commission:     result.Config.Commission,
			Slippage:       result.Config.Slippage,
			Strategies:     h.getStrategyNames(result.Config.Strategies),
		},
		Metrics: &BacktestMetricsData{
			TotalReturn:      result.Metrics.TotalReturn,
			AnnualizedReturn: result.Metrics.AnnualizedReturn,
			SharpeRatio:      result.Metrics.SharpeRatio,
			SortinoRatio:     result.Metrics.SortinoRatio,
			CalmarRatio:      result.Metrics.CalmarRatio,
			MaxDrawdown:      result.Metrics.MaxDrawdown,
			TotalTrades:      result.Metrics.TotalTrades,
			WinningTrades:    result.Metrics.WinningTrades,
			LosingTrades:     result.Metrics.LosingTrades,
			WinRate:          result.Metrics.WinRate,
			ProfitFactor:     result.Metrics.ProfitFactor,
			AvgWin:           result.Metrics.AvgWin,
			AvgLoss:          result.Metrics.AvgLoss,
			LargestWin:       result.Metrics.LargestWin,
			LargestLoss:      result.Metrics.LargestLoss,
			AvgHoldingTime:   result.Metrics.AvgHoldingTime,
			Expectancy:       result.Metrics.Expectancy,
			RecoveryFactor:   result.Metrics.RecoveryFactor,
			StartingCapital:  result.Metrics.StartingCapital,
			EndingCapital:    result.Metrics.EndingCapital,
			NetProfit:        result.Metrics.NetProfit,
		},
		EquityCurve:    equityCurve,
		Trades:         trades,
		MonthlyReturns: result.MonthlyReturns,
		StrategyStats:  strategyStats,
		ExecutionTime:  result.ExecutionTime.String(),
	}
}

// getStrategyNames extracts strategy names
func (h *BacktestHandler) getStrategyNames(strategies []strategy.Strategy) []string {
	names := make([]string, len(strategies))
	for i, strat := range strategies {
		names[i] = strat.Name()
	}
	return names
}

// BacktestResultSummary represents a backtest result summary
type BacktestResultSummary struct {
	ID        string    `json:"id"`
	Symbol    string    `json:"symbol"`
	Timeframe string    `json:"timeframe"`
	StartDate string    `json:"startDate"`
	EndDate   string    `json:"endDate"`
	Return    float64   `json:"return"`
	Sharpe    float64   `json:"sharpe"`
	MaxDD     float64   `json:"maxDD"`
	Trades    int       `json:"trades"`
	CreatedAt time.Time `json:"createdAt"`
}

// GetResults returns all backtest results
func (h *BacktestHandler) GetResults(c echo.Context) error {
	// In real implementation, would fetch from storage
	results := []BacktestResultSummary{}
	return c.JSON(http.StatusOK, results)
}

// GetResult returns a specific backtest result
func (h *BacktestHandler) GetResult(c echo.Context) error {
	id := c.Param("id")

	// In real implementation, would fetch from storage
	_ = id

	return c.JSON(http.StatusNotFound, map[string]string{"error": "Backtest result not found"})
}
