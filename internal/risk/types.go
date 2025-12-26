package risk

import (
	"time"
)

// RiskLevel represents risk severity
type RiskLevel int

const (
	RiskLow RiskLevel = iota
	RiskMedium
	RiskHigh
	RiskCritical
)

func (r RiskLevel) String() string {
	switch r {
	case RiskLow:
		return "LOW"
	case RiskMedium:
		return "MEDIUM"
	case RiskHigh:
		return "HIGH"
	case RiskCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// RiskConfig holds risk management configuration
type RiskConfig struct {
	// Position sizing
	MaxPositionSize        float64 // Max position as % of equity
	MaxPositionValue       float64 // Max position value in quote currency
	DefaultPositionSize    float64 // Default position as % of equity

	// Per-trade risk
	MaxRiskPerTrade        float64 // Max risk per trade as % of equity
	MinRiskRewardRatio     float64 // Minimum risk/reward ratio

	// Account limits
	MaxDailyLoss           float64 // Max daily loss as % of equity
	MaxWeeklyLoss          float64 // Max weekly loss as % of equity
	MaxTotalDrawdown       float64 // Max total drawdown as % of peak equity

	// Position limits
	MaxOpenPositions       int     // Maximum concurrent positions
	MaxPositionsPerSymbol  int     // Max positions per symbol

	// Leverage
	MaxLeverage            float64 // Maximum leverage allowed

	// Circuit breaker
	EnableCircuitBreaker   bool
	ConsecutiveLossLimit   int     // Halt after N consecutive losses
	HaltDuration           time.Duration // How long to halt trading

	// Volatility adjustment
	AdjustForVolatility    bool
	HighVolatilityReduction float64 // Reduce position size by this factor in high vol

	// Correlation
	MaxCorrelation         float64 // Max correlation between positions

	// Time-based
	TradingHoursOnly       bool
	TradingStartHour       int
	TradingEndHour         int
	AvoidWeekends          bool
}

// DefaultRiskConfig returns default risk configuration
func DefaultRiskConfig() *RiskConfig {
	return &RiskConfig{
		MaxPositionSize:         0.10,   // 10% of equity
		MaxPositionValue:        10000,  // $10,000 max
		DefaultPositionSize:     0.05,   // 5% of equity
		MaxRiskPerTrade:         0.02,   // 2% risk per trade
		MinRiskRewardRatio:      1.5,    // 1.5:1 min R/R
		MaxDailyLoss:            0.05,   // 5% max daily loss
		MaxWeeklyLoss:           0.10,   // 10% max weekly loss
		MaxTotalDrawdown:        0.20,   // 20% max drawdown
		MaxOpenPositions:        5,
		MaxPositionsPerSymbol:   1,
		MaxLeverage:             1.0,    // No leverage by default
		EnableCircuitBreaker:    true,
		ConsecutiveLossLimit:    5,
		HaltDuration:            24 * time.Hour,
		AdjustForVolatility:     true,
		HighVolatilityReduction: 0.5,
		MaxCorrelation:          0.7,
		TradingHoursOnly:        false,
		TradingStartHour:        0,
		TradingEndHour:          24,
		AvoidWeekends:           false,
	}
}

// RiskAssessment holds risk assessment for a trade
type RiskAssessment struct {
	Approved       bool
	RiskLevel      RiskLevel
	Reasons        []string
	Warnings       []string
	AdjustedSize   float64
	StopLoss       float64
	TakeProfit     float64
	RiskAmount     float64
	RewardAmount   float64
	RiskRewardRatio float64
}

// PositionSizeResult holds position sizing calculation
type PositionSizeResult struct {
	Size           float64 // Position size in base currency
	Value          float64 // Position value in quote currency
	RiskAmount     float64 // Amount at risk
	RiskPercent    float64 // Risk as % of equity
	StopDistance   float64 // Distance to stop loss
	Leverage       float64 // Effective leverage
}

// AccountState holds current account state for risk calculations
type AccountState struct {
	Equity              float64
	AvailableBalance    float64
	UsedMargin          float64
	UnrealizedPnL       float64
	DailyPnL            float64
	WeeklyPnL           float64
	PeakEquity          float64
	CurrentDrawdown     float64
	OpenPositions       int
	ConsecutiveLosses   int
	LastTradeTime       time.Time
	IsHalted            bool
	HaltReason          string
	HaltUntil           time.Time
}

// TradeMetrics holds metrics for a trade
type TradeMetrics struct {
	EntryPrice     float64
	ExitPrice      float64
	Quantity       float64
	Direction      string
	PnL            float64
	PnLPercent     float64
	RiskAmount     float64
	RewardAmount   float64
	Duration       time.Duration
	IsWin          bool
	MaxDrawdown    float64
	MaxProfit      float64
}

// PortfolioRisk holds portfolio-level risk metrics
type PortfolioRisk struct {
	TotalExposure      float64
	NetExposure        float64 // Long - Short
	LongExposure       float64
	ShortExposure      float64
	VaR                float64 // Value at Risk
	ExpectedShortfall  float64
	Beta               float64
	Correlation        float64
}

// RiskEvent represents a risk-related event
type RiskEvent struct {
	Type       RiskEventType
	Level      RiskLevel
	Message    string
	Details    map[string]interface{}
	Timestamp  time.Time
	Handled    bool
}

// RiskEventType represents types of risk events
type RiskEventType int

const (
	RiskEventDrawdown RiskEventType = iota
	RiskEventDailyLoss
	RiskEventConsecutiveLoss
	RiskEventCircuitBreaker
	RiskEventPositionLimit
	RiskEventVolatilitySpike
	RiskEventLiquidityWarning
)

func (r RiskEventType) String() string {
	switch r {
	case RiskEventDrawdown:
		return "DRAWDOWN"
	case RiskEventDailyLoss:
		return "DAILY_LOSS"
	case RiskEventConsecutiveLoss:
		return "CONSECUTIVE_LOSS"
	case RiskEventCircuitBreaker:
		return "CIRCUIT_BREAKER"
	case RiskEventPositionLimit:
		return "POSITION_LIMIT"
	case RiskEventVolatilitySpike:
		return "VOLATILITY_SPIKE"
	case RiskEventLiquidityWarning:
		return "LIQUIDITY_WARNING"
	default:
		return "UNKNOWN"
	}
}

// DrawdownInfo holds drawdown information
type DrawdownInfo struct {
	CurrentDrawdown    float64
	MaxDrawdown        float64
	DrawdownStart      time.Time
	DrawdownDuration   time.Duration
	RecoveryRequired   float64 // % gain needed to recover
}

// RiskLimits holds current risk limit status
type RiskLimits struct {
	DailyLossUsed      float64
	DailyLossLimit     float64
	DailyLossPercent   float64

	WeeklyLossUsed     float64
	WeeklyLossLimit    float64
	WeeklyLossPercent  float64

	DrawdownCurrent    float64
	DrawdownLimit      float64
	DrawdownPercent    float64

	PositionsOpen      int
	PositionsLimit     int
	PositionsPercent   float64

	IsWithinLimits     bool
	LimitBreaches      []string
}
