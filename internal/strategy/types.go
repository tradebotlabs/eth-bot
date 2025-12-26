package strategy

import (
	"time"

	"github.com/eth-trading/internal/indicators"
)

// Signal represents a trading signal
type Signal struct {
	Type        SignalType       `json:"type"`
	Direction   Direction        `json:"direction"`
	Strength    float64          `json:"strength"`    // 0-1 signal strength
	Price       float64          `json:"price"`       // Current price
	StopLoss    float64          `json:"stopLoss"`    // Suggested stop loss
	TakeProfit  float64          `json:"takeProfit"`  // Suggested take profit
	Confidence  float64          `json:"confidence"`  // 0-1 confidence level
	Reason      string           `json:"reason"`      // Signal reason
	Strategy    string           `json:"strategy"`    // Strategy that generated signal
	Timestamp   time.Time        `json:"timestamp"`
	Timeframe   string           `json:"timeframe"`
	Symbol      string           `json:"symbol"`
	Indicators  SignalIndicators `json:"indicators"`
}

// SignalType represents type of signal
type SignalType int

const (
	SignalTypeNone SignalType = iota
	SignalTypeEntry
	SignalTypeExit
	SignalTypeStopLoss
	SignalTypeTakeProfit
)

func (st SignalType) String() string {
	switch st {
	case SignalTypeEntry:
		return "ENTRY"
	case SignalTypeExit:
		return "EXIT"
	case SignalTypeStopLoss:
		return "STOP_LOSS"
	case SignalTypeTakeProfit:
		return "TAKE_PROFIT"
	default:
		return "NONE"
	}
}

// Direction represents trade direction
type Direction int

const (
	DirectionNone Direction = iota
	DirectionLong
	DirectionShort
)

func (d Direction) String() string {
	switch d {
	case DirectionLong:
		return "LONG"
	case DirectionShort:
		return "SHORT"
	default:
		return "NONE"
	}
}

// MarshalJSON implements json.Marshaler for Direction
func (d Direction) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.String() + `"`), nil
}

// SignalIndicators holds indicator values at signal time
type SignalIndicators struct {
	RSI        float64
	MACD       float64
	MACDSignal float64
	ADX        float64
	ATR        float64
	BBPercentB float64
	Volume     float64
}

// Strategy is the interface for all trading strategies
type Strategy interface {
	// Name returns the strategy name
	Name() string

	// Analyze analyzes market data and returns signals
	Analyze(data *MarketData) []Signal

	// ShouldEnter returns true if entry conditions are met
	ShouldEnter(data *MarketData) (bool, Direction, float64)

	// ShouldExit returns true if exit conditions are met for a position
	ShouldExit(data *MarketData, position *Position) (bool, string)

	// CalculateStopLoss calculates stop loss for a position
	CalculateStopLoss(data *MarketData, direction Direction, entryPrice float64) float64

	// CalculateTakeProfit calculates take profit for a position
	CalculateTakeProfit(data *MarketData, direction Direction, entryPrice float64) float64

	// GetMinDataPoints returns minimum candles needed for analysis
	GetMinDataPoints() int

	// IsEnabled returns whether strategy is enabled
	IsEnabled() bool

	// SetEnabled enables/disables the strategy
	SetEnabled(enabled bool)

	// GetConfig returns strategy configuration
	GetConfig() interface{}
}

// MarketData holds all data needed for strategy analysis
type MarketData struct {
	Symbol    string
	Timeframe string
	Timestamp time.Time

	// OHLCV data
	Opens   []float64
	Highs   []float64
	Lows    []float64
	Closes  []float64
	Volumes []float64

	// Pre-calculated indicators
	Analysis indicators.AnalysisResult

	// Regime
	Regime RegimeResult

	// Current price
	CurrentPrice float64
	Bid          float64
	Ask          float64

	// Additional context
	DailyHigh  float64
	DailyLow   float64
	DailyOpen  float64
}

// Position represents an open position
type Position struct {
	ID          int64
	Symbol      string
	Direction   Direction
	EntryPrice  float64
	Quantity    float64
	CurrentPrice float64
	StopLoss    float64
	TakeProfit  float64
	Strategy    string
	OpenTime    time.Time
	UnrealizedPnL float64
	UnrealizedPnLPercent float64
}

// BaseStrategy provides common functionality
type BaseStrategy struct {
	name      string
	enabled   bool
	minData   int
	atrPeriod int
}

// NewBaseStrategy creates a new base strategy
func NewBaseStrategy(name string, minData, atrPeriod int) BaseStrategy {
	return BaseStrategy{
		name:      name,
		enabled:   true,
		minData:   minData,
		atrPeriod: atrPeriod,
	}
}

// Name returns strategy name
func (bs *BaseStrategy) Name() string {
	return bs.name
}

// GetMinDataPoints returns minimum data points
func (bs *BaseStrategy) GetMinDataPoints() int {
	return bs.minData
}

// IsEnabled returns if strategy is enabled
func (bs *BaseStrategy) IsEnabled() bool {
	return bs.enabled
}

// SetEnabled enables/disables strategy
func (bs *BaseStrategy) SetEnabled(enabled bool) {
	bs.enabled = enabled
}

// CalculateATRStop calculates ATR-based stop loss
func (bs *BaseStrategy) CalculateATRStop(data *MarketData, direction Direction, entryPrice float64, multiplier float64) float64 {
	atr := data.Analysis.ATR.ATR
	if atr == 0 {
		// Fallback: 2% stop
		if direction == DirectionLong {
			return entryPrice * 0.98
		}
		return entryPrice * 1.02
	}

	if direction == DirectionLong {
		return entryPrice - (atr * multiplier)
	}
	return entryPrice + (atr * multiplier)
}

// CalculateATRTarget calculates ATR-based take profit
func (bs *BaseStrategy) CalculateATRTarget(data *MarketData, direction Direction, entryPrice float64, multiplier float64) float64 {
	atr := data.Analysis.ATR.ATR
	if atr == 0 {
		// Fallback: 3% target
		if direction == DirectionLong {
			return entryPrice * 1.03
		}
		return entryPrice * 0.97
	}

	if direction == DirectionLong {
		return entryPrice + (atr * multiplier)
	}
	return entryPrice - (atr * multiplier)
}

// CreateSignal creates a trading signal
func (bs *BaseStrategy) CreateSignal(data *MarketData, signalType SignalType, direction Direction, strength float64, reason string) Signal {
	price := data.CurrentPrice
	if price == 0 && len(data.Closes) > 0 {
		price = data.Closes[len(data.Closes)-1]
	}

	signal := Signal{
		Type:       signalType,
		Direction:  direction,
		Strength:   strength,
		Price:      price,
		Confidence: strength,
		Reason:     reason,
		Strategy:   bs.name,
		Timestamp:  data.Timestamp,
		Timeframe:  data.Timeframe,
		Symbol:     data.Symbol,
		Indicators: SignalIndicators{
			RSI:        data.Analysis.RSI.Value,
			MACD:       data.Analysis.MACD.MACD,
			MACDSignal: data.Analysis.MACD.Signal,
			ADX:        data.Analysis.ADX.ADX,
			ATR:        data.Analysis.ATR.ATR,
			BBPercentB: data.Analysis.Bollinger.PercentB,
			Volume:     data.Analysis.Volume.Current,
		},
	}

	return signal
}

// StrategyStats holds strategy performance statistics
type StrategyStats struct {
	Name           string
	TotalSignals   int
	WinningSignals int
	LosingSignals  int
	WinRate        float64
	AvgWin         float64
	AvgLoss        float64
	ProfitFactor   float64
	Sharpe         float64
	MaxDrawdown    float64
}

// ScoreResult holds strategy scoring result
type ScoreResult struct {
	Strategy    string
	Score       float64
	Confidence  float64
	Direction   Direction
	Signals     []Signal
	Factors     map[string]float64
}

// StrategyWeight holds strategy weighting configuration
type StrategyWeight struct {
	Name       string
	Weight     float64
	MinScore   float64
	MaxSignals int
}
