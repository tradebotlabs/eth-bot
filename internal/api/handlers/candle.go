package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/eth-trading/internal/orchestrator"
	"github.com/labstack/echo/v4"
)

// Note: time is still used in GetIndicators

// CandleHandler handles candle/market data endpoints
type CandleHandler struct {
	orchestrator *orchestrator.Orchestrator
}

// NewCandleHandler creates a new candle handler
func NewCandleHandler(orch *orchestrator.Orchestrator) *CandleHandler {
	return &CandleHandler{orchestrator: orch}
}

// CandleData represents candle data for API
type CandleData struct {
	Time   int64   `json:"time"`   // Unix timestamp in seconds
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume float64 `json:"volume"`
}

// GetCandles returns candle data
func (h *CandleHandler) GetCandles(c echo.Context) error {
	symbol := c.QueryParam("symbol")
	if symbol == "" {
		symbol = "ETHUSDT"
	}

	timeframe := c.QueryParam("timeframe")
	if timeframe == "" {
		timeframe = "15m"
	}

	limitStr := c.QueryParam("limit")
	limit := 500
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	if h.orchestrator == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "Orchestrator not available"})
	}

	// Fetch candles from orchestrator
	storageCandles := h.orchestrator.GetCandles(symbol, timeframe, limit)

	// Convert to API format
	candles := make([]CandleData, len(storageCandles))
	for i, sc := range storageCandles {
		candles[i] = CandleData{
			Time:   sc.OpenTime.UnixMilli(),
			Open:   sc.Open,
			High:   sc.High,
			Low:    sc.Low,
			Close:  sc.Close,
			Volume: sc.Volume,
		}
	}

	return c.JSON(http.StatusOK, candles)
}

// GetCandlesBySymbol returns candles for a specific symbol and timeframe
func (h *CandleHandler) GetCandlesBySymbol(c echo.Context) error {
	symbol := c.Param("symbol")
	timeframe := c.Param("timeframe")

	limitStr := c.QueryParam("limit")
	limit := 500
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	if h.orchestrator == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "Orchestrator not available"})
	}

	// Fetch candles from orchestrator
	storageCandles := h.orchestrator.GetCandles(symbol, timeframe, limit)

	// Convert to API format
	candles := make([]CandleData, len(storageCandles))
	for i, sc := range storageCandles {
		candles[i] = CandleData{
			Time:   sc.OpenTime.UnixMilli(),
			Open:   sc.Open,
			High:   sc.High,
			Low:    sc.Low,
			Close:  sc.Close,
			Volume: sc.Volume,
		}
	}

	return c.JSON(http.StatusOK, candles)
}

// TickerData represents ticker data
type TickerData struct {
	Symbol        string  `json:"symbol"`
	Price         float64 `json:"price"`
	PriceChange   float64 `json:"priceChange"`
	PercentChange float64 `json:"percentChange"`
	High24h       float64 `json:"high24h"`
	Low24h        float64 `json:"low24h"`
	Volume24h     float64 `json:"volume24h"`
	Timestamp     int64   `json:"timestamp"`
}

// GetTicker returns current ticker data
func (h *CandleHandler) GetTicker(c echo.Context) error {
	symbol := c.QueryParam("symbol")
	if symbol == "" {
		symbol = "ETHUSDT"
	}

	if h.orchestrator == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "Orchestrator not available"})
	}

	state := h.orchestrator.GetState()

	ticker := TickerData{
		Symbol:    symbol,
		Price:     state.CurrentPrice,
		Timestamp: time.Now().UnixMilli(),
	}

	return c.JSON(http.StatusOK, ticker)
}

// IndicatorData represents indicator values
type IndicatorData struct {
	RSI       *float64        `json:"rsi,omitempty"`
	MACD      *MACDData       `json:"macd,omitempty"`
	BB        *BollingerData  `json:"bb,omitempty"`
	ADX       *ADXData        `json:"adx,omitempty"`
	ATR       *float64        `json:"atr,omitempty"`
	Volume    *VolumeData     `json:"volume,omitempty"`
	Regime    string          `json:"regime"`
	Timestamp int64           `json:"timestamp"`
}

// MACDData represents MACD indicator values
type MACDData struct {
	MACD      float64 `json:"macd"`
	Signal    float64 `json:"signal"`
	Histogram float64 `json:"histogram"`
}

// BollingerData represents Bollinger Band values
type BollingerData struct {
	Upper   float64 `json:"upper"`
	Middle  float64 `json:"middle"`
	Lower   float64 `json:"lower"`
	Width   float64 `json:"width"`
	Percent float64 `json:"percent"`
}

// ADXData represents ADX indicator values
type ADXData struct {
	ADX     float64 `json:"adx"`
	PlusDI  float64 `json:"plusDI"`
	MinusDI float64 `json:"minusDI"`
	Trend   string  `json:"trend"`
}

// VolumeData represents volume indicator values
type VolumeData struct {
	Volume    float64 `json:"volume"`
	VolumeSMA float64 `json:"volumeSma"`
	OBV       float64 `json:"obv"`
	VWAP      float64 `json:"vwap"`
}

// GetIndicators returns current indicator values
func (h *CandleHandler) GetIndicators(c echo.Context) error {
	symbol := c.QueryParam("symbol")
	if symbol == "" {
		symbol = "ETHUSDT"
	}

	timeframe := c.QueryParam("timeframe")
	if timeframe == "" {
		timeframe = "1h"
	}

	if h.orchestrator == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "Orchestrator not available"})
	}

	state := h.orchestrator.GetState()

	// In real implementation, would calculate from data service
	indicators := IndicatorData{
		Regime:    state.CurrentRegime,
		Timestamp: time.Now().UnixMilli(),
	}

	return c.JSON(http.StatusOK, indicators)
}
