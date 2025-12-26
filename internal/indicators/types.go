package indicators

// IndicatorResult represents the result of an indicator calculation
type IndicatorResult struct {
	Value     float64
	Signal    Signal
	Timestamp int64
}

// Signal represents trading signals
type Signal int

const (
	SignalNone Signal = iota
	SignalBuy
	SignalSell
	SignalStrongBuy
	SignalStrongSell
)

func (s Signal) String() string {
	switch s {
	case SignalBuy:
		return "BUY"
	case SignalSell:
		return "SELL"
	case SignalStrongBuy:
		return "STRONG_BUY"
	case SignalStrongSell:
		return "STRONG_SELL"
	default:
		return "NONE"
	}
}

// RSIResult holds RSI calculation result
type RSIResult struct {
	Value        float64
	IsOverbought bool
	IsOversold   bool
	Signal       Signal
}

// MACDResult holds MACD calculation result
type MACDResult struct {
	MACD      float64
	Signal    float64
	Histogram float64
	Crossover CrossoverType
}

// CrossoverType represents crossover types
type CrossoverType int

const (
	CrossoverNone CrossoverType = iota
	CrossoverBullish
	CrossoverBearish
)

func (c CrossoverType) String() string {
	switch c {
	case CrossoverBullish:
		return "BULLISH"
	case CrossoverBearish:
		return "BEARISH"
	default:
		return "NONE"
	}
}

// BollingerResult holds Bollinger Bands calculation result
type BollingerResult struct {
	Upper      float64
	Middle     float64
	Lower      float64
	Width      float64
	PercentB   float64
	Squeeze    bool
	Breakout   BreakoutType
}

// BreakoutType represents breakout types
type BreakoutType int

const (
	BreakoutNone BreakoutType = iota
	BreakoutUpper
	BreakoutLower
)

func (b BreakoutType) String() string {
	switch b {
	case BreakoutUpper:
		return "UPPER"
	case BreakoutLower:
		return "LOWER"
	default:
		return "NONE"
	}
}

// ADXResult holds ADX calculation result
type ADXResult struct {
	ADX       float64
	PlusDI    float64
	MinusDI   float64
	Trending  bool
	Strength  TrendStrength
	Direction TrendDirection
}

// TrendStrength represents trend strength levels
type TrendStrength int

const (
	TrendWeak TrendStrength = iota
	TrendModerate
	TrendStrong
	TrendVeryStrong
)

func (t TrendStrength) String() string {
	switch t {
	case TrendWeak:
		return "WEAK"
	case TrendModerate:
		return "MODERATE"
	case TrendStrong:
		return "STRONG"
	case TrendVeryStrong:
		return "VERY_STRONG"
	default:
		return "UNKNOWN"
	}
}

// TrendDirection represents trend direction
type TrendDirection int

const (
	TrendNeutral TrendDirection = iota
	TrendUp
	TrendDown
)

func (t TrendDirection) String() string {
	switch t {
	case TrendUp:
		return "UP"
	case TrendDown:
		return "DOWN"
	default:
		return "NEUTRAL"
	}
}

// ATRResult holds ATR calculation result
type ATRResult struct {
	ATR           float64
	ATRPercent    float64
	HighVolatility bool
}

// MAResult holds Moving Average result
type MAResult struct {
	Value     float64
	Trend     TrendDirection
	Crossover CrossoverType
}

// VolumeResult holds Volume analysis result
type VolumeResult struct {
	Current        float64
	Average        float64
	Ratio          float64
	IsHighVolume   bool
	IsLowVolume    bool
}

// StochResult holds Stochastic calculation result
type StochResult struct {
	K         float64
	D         float64
	Overbought bool
	Oversold   bool
	Crossover  CrossoverType
}

// SupportResistance holds support and resistance levels
type SupportResistance struct {
	Supports    []float64
	Resistances []float64
}

// PivotPoints holds pivot point calculations
type PivotPoints struct {
	Pivot float64
	R1    float64
	R2    float64
	R3    float64
	S1    float64
	S2    float64
	S3    float64
}

// IndicatorConfig holds configuration for indicators
type IndicatorConfig struct {
	// RSI
	RSIPeriod       int
	RSIOverbought   float64
	RSIOversold     float64

	// MACD
	MACDFast        int
	MACDSlow        int
	MACDSignal      int

	// Bollinger Bands
	BBPeriod        int
	BBStdDev        float64
	BBSqueezeThreshold float64

	// ADX
	ADXPeriod       int
	ADXTrendingThreshold float64

	// ATR
	ATRPeriod       int
	ATRHighVolThreshold float64

	// Moving Averages
	MAShortPeriod   int
	MAMediumPeriod  int
	MALongPeriod    int

	// Volume
	VolumePeriod    int
	VolumeHighThreshold float64
	VolumeLowThreshold  float64

	// Stochastic
	StochKPeriod    int
	StochDPeriod    int
	StochSlowing    int
	StochOverbought float64
	StochOversold   float64
}

// DefaultConfig returns default indicator configuration
func DefaultConfig() *IndicatorConfig {
	return &IndicatorConfig{
		RSIPeriod:            14,
		RSIOverbought:        70,
		RSIOversold:          30,
		MACDFast:             12,
		MACDSlow:             26,
		MACDSignal:           9,
		BBPeriod:             20,
		BBStdDev:             2.0,
		BBSqueezeThreshold:   0.05,
		ADXPeriod:            14,
		ADXTrendingThreshold: 25,
		ATRPeriod:            14,
		ATRHighVolThreshold:  1.5,
		MAShortPeriod:        10,
		MAMediumPeriod:       20,
		MALongPeriod:         50,
		VolumePeriod:         20,
		VolumeHighThreshold:  1.5,
		VolumeLowThreshold:   0.5,
		StochKPeriod:         14,
		StochDPeriod:         3,
		StochSlowing:         3,
		StochOverbought:      80,
		StochOversold:        20,
	}
}
