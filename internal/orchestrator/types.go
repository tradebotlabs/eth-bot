package orchestrator

import (
	"time"

	"github.com/eth-trading/internal/execution"
	"github.com/eth-trading/internal/risk"
	"github.com/eth-trading/internal/strategy"
)

// OrchestratorConfig holds orchestrator configuration
type OrchestratorConfig struct {
	// Trading
	Symbol          string
	Timeframes      []string // Timeframes to monitor
	PrimaryTimeframe string  // Main timeframe for signals

	// Mode
	Mode            TradingMode
	InitialCapital  float64

	// Strategy
	EnabledStrategies []string

	// WebSocket
	EnableWebSocket bool
	BroadcastInterval time.Duration
}

// TradingMode represents the trading mode
type TradingMode int

const (
	TradingModePaper TradingMode = iota
	TradingModeLive
)

func (m TradingMode) String() string {
	switch m {
	case TradingModePaper:
		return "PAPER"
	case TradingModeLive:
		return "LIVE"
	default:
		return "UNKNOWN"
	}
}

// DefaultOrchestratorConfig returns default configuration
func DefaultOrchestratorConfig() *OrchestratorConfig {
	return &OrchestratorConfig{
		Symbol:           "ETHUSDT",
		Timeframes:       []string{"1m", "5m", "15m", "1h", "4h", "1d"},
		PrimaryTimeframe: "1m", // Using 1m for faster signal generation
		Mode:             TradingModePaper,
		InitialCapital:   100000,
		EnabledStrategies: []string{
			"TrendFollowing",
			"MeanReversion",
			"Breakout",
			"Volatility",
			"StatArb",
		},
		EnableWebSocket:   true,
		BroadcastInterval: time.Second,
	}
}

// TradingState represents the current trading state
type TradingState struct {
	// General
	Mode           TradingMode
	IsRunning      bool
	IsPaused       bool
	StartTime      time.Time
	LastUpdate     time.Time

	// Market
	CurrentPrice   float64
	DailyChange    float64
	Volume24h      float64

	// Account
	Equity         float64
	AvailableBalance float64
	UnrealizedPnL  float64
	RealizedPnL    float64
	DailyPnL       float64

	// Positions
	OpenPositions  int
	TotalTrades    int
	WinRate        float64

	// Risk
	CurrentDrawdown float64
	MaxDrawdown    float64
	RiskLevel      risk.RiskLevel
	IsHalted       bool
	HaltReason     string

	// Strategy
	ActiveStrategies []string
	CurrentRegime   string
	LastSignal     *strategy.Signal

	// System
	CandleCount    int
	LastCandleTime time.Time
	Errors         []string
}

// BroadcastMessage represents a WebSocket message
type BroadcastMessage struct {
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// MessageType constants
const (
	MessageTypeState      = "state"
	MessageTypeCandle     = "candle"
	MessageTypeSignal     = "signal"
	MessageTypeTrade      = "trade"
	MessageTypePosition   = "position"
	MessageTypeRisk       = "risk"
	MessageTypeError      = "error"
	MessageTypeIndicators = "indicators"
	MessageTypePrice      = "price" // Real-time price updates
)

// StateUpdate represents a state update message
type StateUpdate struct {
	State    *TradingState          `json:"state"`
	Summary  *AccountSummary        `json:"summary"`
}

// AccountSummary represents account summary for API
type AccountSummary struct {
	Equity           float64 `json:"equity"`
	AvailableBalance float64 `json:"availableBalance"`
	UsedMargin       float64 `json:"usedMargin"`
	UnrealizedPnL    float64 `json:"unrealizedPnL"`
	RealizedPnL      float64 `json:"realizedPnL"`
	DailyPnL         float64 `json:"dailyPnL"`
	WeeklyPnL        float64 `json:"weeklyPnL"`
	TotalReturn      float64 `json:"totalReturn"`
	OpenPositions    int     `json:"openPositions"`
	TotalTrades      int     `json:"totalTrades"`
	WinningTrades    int     `json:"winningTrades"`
	LosingTrades     int     `json:"losingTrades"`
	WinRate          float64 `json:"winRate"`
	ProfitFactor     float64 `json:"profitFactor"`
}

// CandleUpdate represents a candle update message
type CandleUpdate struct {
	Symbol    string    `json:"symbol"`
	Timeframe string    `json:"timeframe"`
	Timestamp time.Time `json:"timestamp"`
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	Volume    float64   `json:"volume"`
	IsClosed  bool      `json:"isClosed"`
}

// SignalUpdate represents a signal update message
type SignalUpdate struct {
	Signal      *strategy.Signal `json:"signal"`
	Approved    bool             `json:"approved"`
	RejectedBy  string           `json:"rejectedBy,omitempty"`
	Reason      string           `json:"reason,omitempty"`
}

// SignalRecord stores a signal with its approval status for history
type SignalRecord struct {
	Signal     *strategy.Signal `json:"signal"`
	Approved   bool             `json:"approved"`
	Reason     string           `json:"reason,omitempty"`
	ReceivedAt time.Time        `json:"receivedAt"`
}

// TradeUpdate represents a trade update message
type TradeUpdate struct {
	TradeID    string              `json:"tradeId"`
	OrderID    string              `json:"orderId"`
	Symbol     string              `json:"symbol"`
	Side       execution.OrderSide `json:"side"`
	Type       string              `json:"type"`
	Quantity   float64             `json:"quantity"`
	Price      float64             `json:"price"`
	Commission float64             `json:"commission"`
	RealizedPnL float64            `json:"realizedPnL"`
	Strategy   string              `json:"strategy"`
	Timestamp  time.Time           `json:"timestamp"`
}

// PositionUpdate represents a position update message
type PositionUpdate struct {
	PositionID    int64                   `json:"positionId"`
	Symbol        string                  `json:"symbol"`
	Side          execution.PositionSide  `json:"side"`
	Quantity      float64                 `json:"quantity"`
	EntryPrice    float64                 `json:"entryPrice"`
	CurrentPrice  float64                 `json:"currentPrice"`
	StopLoss      float64                 `json:"stopLoss"`
	TakeProfit    float64                 `json:"takeProfit"`
	UnrealizedPnL float64                 `json:"unrealizedPnL"`
	RealizedPnL   float64                 `json:"realizedPnL"`
	Strategy      string                  `json:"strategy"`
	OpenTime      time.Time               `json:"openTime"`
	EventType     string                  `json:"eventType"` // opened, closed, updated
}

// RiskUpdate represents a risk update message
type RiskUpdate struct {
	Level           risk.RiskLevel `json:"level"`
	Drawdown        float64        `json:"drawdown"`
	MaxDrawdown     float64        `json:"maxDrawdown"`
	DailyLossUsed   float64        `json:"dailyLossUsed"`
	DailyLossLimit  float64        `json:"dailyLossLimit"`
	WeeklyLossUsed  float64        `json:"weeklyLossUsed"`
	WeeklyLossLimit float64        `json:"weeklyLossLimit"`
	IsHalted        bool           `json:"isHalted"`
	HaltReason      string         `json:"haltReason,omitempty"`
	Events          []risk.RiskEvent `json:"events,omitempty"`
}

// IndicatorsUpdate represents indicators update message
type IndicatorsUpdate struct {
	Symbol    string             `json:"symbol"`
	Timeframe string             `json:"timeframe"`
	Timestamp time.Time          `json:"timestamp"`
	RSI       float64            `json:"rsi"`
	MACD      *MACDValue         `json:"macd"`
	BB        *BollingerValue    `json:"bb"`
	ADX       *ADXValue          `json:"adx"`
	ATR       float64            `json:"atr"`
	Regime    string             `json:"regime"`
}

// MACDValue represents MACD values
type MACDValue struct {
	MACD      float64 `json:"macd"`
	Signal    float64 `json:"signal"`
	Histogram float64 `json:"histogram"`
}

// BollingerValue represents Bollinger Band values
type BollingerValue struct {
	Upper  float64 `json:"upper"`
	Middle float64 `json:"middle"`
	Lower  float64 `json:"lower"`
	Width  float64 `json:"width"`
}

// ADXValue represents ADX values
type ADXValue struct {
	ADX     float64 `json:"adx"`
	PlusDI  float64 `json:"plusDI"`
	MinusDI float64 `json:"minusDI"`
}

// ErrorUpdate represents an error message
type ErrorUpdate struct {
	Code    string    `json:"code"`
	Message string    `json:"message"`
	Details string    `json:"details,omitempty"`
	Time    time.Time `json:"time"`
}

// PriceUpdate represents a real-time price update (lightweight for high frequency)
type PriceUpdate struct {
	Symbol    string    `json:"symbol"`
	Price     float64   `json:"price"`
	Timestamp time.Time `json:"timestamp"`
}

// BinanceWSHandler handles Binance WebSocket events for the orchestrator
type BinanceWSHandler struct {
	orchestrator *Orchestrator
}

// NewBinanceWSHandler creates a new WebSocket handler
func NewBinanceWSHandler(orch *Orchestrator) *BinanceWSHandler {
	return &BinanceWSHandler{orchestrator: orch}
}
