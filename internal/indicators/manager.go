package indicators

import (
	"sync"
)

// Manager manages all technical indicators
type Manager struct {
	config *IndicatorConfig

	// Indicator instances
	rsi      *RSI
	macd     *MACD
	bb       *BollingerBands
	adx      *ADX
	atr      *ATR
	ma       *MovingAverage
	volume   *VolumeAnalyzer
	stoch    *Stochastic

	mu sync.RWMutex
}

// NewManager creates a new indicator manager
func NewManager(config *IndicatorConfig) *Manager {
	if config == nil {
		config = DefaultConfig()
	}

	return &Manager{
		config:  config,
		rsi:     NewRSI(config.RSIPeriod, config.RSIOverbought, config.RSIOversold),
		macd:    NewMACD(config.MACDFast, config.MACDSlow, config.MACDSignal),
		bb:      NewBollingerBands(config.BBPeriod, config.BBStdDev, config.BBSqueezeThreshold),
		adx:     NewADX(config.ADXPeriod, config.ADXTrendingThreshold),
		atr:     NewATR(config.ATRPeriod, config.ATRHighVolThreshold),
		ma:      NewMovingAverage(config.MAShortPeriod, config.MAMediumPeriod, config.MALongPeriod, MATypeEMA),
		volume:  NewVolumeAnalyzer(config.VolumePeriod, config.VolumeHighThreshold, config.VolumeLowThreshold),
		stoch:   NewStochastic(config.StochKPeriod, config.StochDPeriod, config.StochSlowing, config.StochOverbought, config.StochOversold),
	}
}

// AnalysisResult holds all indicator results
type AnalysisResult struct {
	RSI        RSIResult
	MACD       MACDResult
	Bollinger  BollingerResult
	ADX        ADXResult
	ATR        ATRResult
	MA         MAResult
	Volume     VolumeResult
	Stochastic StochResult

	// Derived signals
	TrendStrength TrendStrength
	TrendDir      TrendDirection
	Momentum      Signal
	Volatility    VolatilityLevel
	OverallSignal Signal
}

// VolatilityLevel represents volatility classification
type VolatilityLevel int

const (
	VolatilityLow VolatilityLevel = iota
	VolatilityNormal
	VolatilityHigh
)

func (v VolatilityLevel) String() string {
	switch v {
	case VolatilityLow:
		return "LOW"
	case VolatilityHigh:
		return "HIGH"
	default:
		return "NORMAL"
	}
}

// Analyze performs complete technical analysis
func (m *Manager) Analyze(opens, highs, lows, closes, volumes []float64) AnalysisResult {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := AnalysisResult{}

	// Calculate individual indicators
	if len(closes) >= m.config.RSIPeriod+1 {
		result.RSI = m.rsi.Calculate(closes)
	}

	if len(closes) >= m.config.MACDSlow+m.config.MACDSignal {
		result.MACD = m.macd.Calculate(closes)
	}

	if len(closes) >= m.config.BBPeriod {
		result.Bollinger = m.bb.Calculate(closes)
	}

	if len(highs) >= m.config.ADXPeriod*2 {
		result.ADX = m.adx.Calculate(highs, lows, closes)
	}

	if len(highs) >= m.config.ATRPeriod+1 {
		result.ATR = m.atr.Calculate(highs, lows, closes)
	}

	if len(closes) >= m.config.MALongPeriod {
		result.MA = m.ma.Calculate(closes)
	}

	if len(volumes) >= m.config.VolumePeriod {
		result.Volume = m.volume.Analyze(volumes)
	}

	if len(highs) >= m.config.StochKPeriod+m.config.StochDPeriod+m.config.StochSlowing-2 {
		result.Stochastic = m.stoch.Calculate(highs, lows, closes)
	}

	// Derive composite signals
	result.TrendStrength = m.deriveTrendStrength(result)
	result.TrendDir = m.deriveTrendDirection(result)
	result.Momentum = m.deriveMomentum(result)
	result.Volatility = m.deriveVolatility(result)
	result.OverallSignal = m.deriveOverallSignal(result)

	return result
}

// deriveTrendStrength determines overall trend strength
func (m *Manager) deriveTrendStrength(result AnalysisResult) TrendStrength {
	return result.ADX.Strength
}

// deriveTrendDirection determines overall trend direction
func (m *Manager) deriveTrendDirection(result AnalysisResult) TrendDirection {
	// Use ADX direction as primary
	if result.ADX.Trending {
		return result.ADX.Direction
	}

	// Fall back to MA trend
	return result.MA.Trend
}

// deriveMomentum determines momentum signal
func (m *Manager) deriveMomentum(result AnalysisResult) Signal {
	buySignals := 0
	sellSignals := 0

	// RSI
	if result.RSI.IsOversold {
		buySignals++
	} else if result.RSI.IsOverbought {
		sellSignals++
	}

	// MACD
	if result.MACD.Crossover == CrossoverBullish {
		buySignals++
	} else if result.MACD.Crossover == CrossoverBearish {
		sellSignals++
	}
	if result.MACD.Histogram > 0 {
		buySignals++
	} else if result.MACD.Histogram < 0 {
		sellSignals++
	}

	// Stochastic
	if result.Stochastic.Oversold && result.Stochastic.Crossover == CrossoverBullish {
		buySignals += 2
	} else if result.Stochastic.Overbought && result.Stochastic.Crossover == CrossoverBearish {
		sellSignals += 2
	}

	// Determine signal
	if buySignals >= 3 {
		return SignalStrongBuy
	}
	if buySignals >= 2 {
		return SignalBuy
	}
	if sellSignals >= 3 {
		return SignalStrongSell
	}
	if sellSignals >= 2 {
		return SignalSell
	}

	return SignalNone
}

// deriveVolatility determines volatility level
func (m *Manager) deriveVolatility(result AnalysisResult) VolatilityLevel {
	if result.ATR.HighVolatility {
		return VolatilityHigh
	}

	// Check Bollinger squeeze
	if result.Bollinger.Squeeze {
		return VolatilityLow
	}

	// Check bandwidth
	if result.Bollinger.Width < 0.03 {
		return VolatilityLow
	}
	if result.Bollinger.Width > 0.1 {
		return VolatilityHigh
	}

	return VolatilityNormal
}

// deriveOverallSignal derives overall trading signal
func (m *Manager) deriveOverallSignal(result AnalysisResult) Signal {
	// Weight different factors
	score := 0.0

	// Trend contribution (weight: 40%)
	if result.TrendDir == TrendUp && result.TrendStrength >= TrendModerate {
		score += 40
	} else if result.TrendDir == TrendDown && result.TrendStrength >= TrendModerate {
		score -= 40
	}

	// Momentum contribution (weight: 30%)
	switch result.Momentum {
	case SignalStrongBuy:
		score += 30
	case SignalBuy:
		score += 15
	case SignalSell:
		score -= 15
	case SignalStrongSell:
		score -= 30
	}

	// Mean reversion signals (weight: 20%)
	if result.Bollinger.Breakout == BreakoutLower && result.RSI.IsOversold {
		score += 20
	} else if result.Bollinger.Breakout == BreakoutUpper && result.RSI.IsOverbought {
		score -= 20
	}

	// Volume confirmation (weight: 10%)
	if result.Volume.IsHighVolume {
		if score > 0 {
			score += 10
		} else if score < 0 {
			score -= 10
		}
	}

	// Determine signal based on score
	if score >= 60 {
		return SignalStrongBuy
	}
	if score >= 30 {
		return SignalBuy
	}
	if score <= -60 {
		return SignalStrongSell
	}
	if score <= -30 {
		return SignalSell
	}

	return SignalNone
}

// GetConfig returns the current configuration
func (m *Manager) GetConfig() *IndicatorConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// UpdateConfig updates the configuration and recreates indicators
func (m *Manager) UpdateConfig(config *IndicatorConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.config = config
	m.rsi = NewRSI(config.RSIPeriod, config.RSIOverbought, config.RSIOversold)
	m.macd = NewMACD(config.MACDFast, config.MACDSlow, config.MACDSignal)
	m.bb = NewBollingerBands(config.BBPeriod, config.BBStdDev, config.BBSqueezeThreshold)
	m.adx = NewADX(config.ADXPeriod, config.ADXTrendingThreshold)
	m.atr = NewATR(config.ATRPeriod, config.ATRHighVolThreshold)
	m.ma = NewMovingAverage(config.MAShortPeriod, config.MAMediumPeriod, config.MALongPeriod, MATypeEMA)
	m.volume = NewVolumeAnalyzer(config.VolumePeriod, config.VolumeHighThreshold, config.VolumeLowThreshold)
	m.stoch = NewStochastic(config.StochKPeriod, config.StochDPeriod, config.StochSlowing, config.StochOverbought, config.StochOversold)
}

// QuickAnalysis performs a lightweight analysis for high-frequency updates
type QuickAnalysisResult struct {
	RSI         float64
	MACD        float64
	MACDSignal  float64
	BBPercentB  float64
	ADX         float64
	ATR         float64
	VolumeRatio float64
}

// QuickAnalyze performs a quick analysis with just the key values
func (m *Manager) QuickAnalyze(opens, highs, lows, closes, volumes []float64) QuickAnalysisResult {
	result := QuickAnalysisResult{}

	if len(closes) >= m.config.RSIPeriod+1 {
		result.RSI = RSILast(closes, m.config.RSIPeriod)
	}

	if len(closes) >= m.config.MACDSlow+m.config.MACDSignal {
		macd := MACDLast(closes, m.config.MACDFast, m.config.MACDSlow, m.config.MACDSignal)
		result.MACD = macd.MACD
		result.MACDSignal = macd.Signal
	}

	if len(closes) >= m.config.BBPeriod {
		bb := BollingerLast(closes, m.config.BBPeriod, m.config.BBStdDev)
		result.BBPercentB = bb.PercentB
	}

	if len(highs) >= m.config.ADXPeriod*2 {
		adx := ADXLast(highs, lows, closes, m.config.ADXPeriod)
		result.ADX = adx.ADX
	}

	if len(highs) >= m.config.ATRPeriod+1 {
		result.ATR = ATRLast(highs, lows, closes, m.config.ATRPeriod)
	}

	if len(volumes) >= m.config.VolumePeriod {
		result.VolumeRatio = VolumeRatio(volumes, m.config.VolumePeriod)
	}

	return result
}

// IndicatorSnapshot represents current indicator values for display
type IndicatorSnapshot struct {
	Symbol    string
	Timeframe string

	// Values
	RSI        float64
	MACD       float64
	MACDSignal float64
	MACDHist   float64
	BBUpper    float64
	BBMiddle   float64
	BBLower    float64
	BBWidth    float64
	ADX        float64
	PlusDI     float64
	MinusDI    float64
	ATR        float64
	ATRPercent float64
	StochK     float64
	StochD     float64

	// Signals
	RSISignal   string
	MACDSignal_ string
	ADXTrend    string
	BBStatus    string
	StochSignal string
}

// GetSnapshot returns current indicator values as a snapshot
func (m *Manager) GetSnapshot(symbol, timeframe string, opens, highs, lows, closes, volumes []float64) IndicatorSnapshot {
	analysis := m.Analyze(opens, highs, lows, closes, volumes)

	return IndicatorSnapshot{
		Symbol:      symbol,
		Timeframe:   timeframe,
		RSI:         analysis.RSI.Value,
		MACD:        analysis.MACD.MACD,
		MACDSignal:  analysis.MACD.Signal,
		MACDHist:    analysis.MACD.Histogram,
		BBUpper:     analysis.Bollinger.Upper,
		BBMiddle:    analysis.Bollinger.Middle,
		BBLower:     analysis.Bollinger.Lower,
		BBWidth:     analysis.Bollinger.Width,
		ADX:         analysis.ADX.ADX,
		PlusDI:      analysis.ADX.PlusDI,
		MinusDI:     analysis.ADX.MinusDI,
		ATR:         analysis.ATR.ATR,
		ATRPercent:  analysis.ATR.ATRPercent,
		StochK:      analysis.Stochastic.K,
		StochD:      analysis.Stochastic.D,
		RSISignal:   m.getRSISignalText(analysis.RSI),
		MACDSignal_: analysis.MACD.Crossover.String(),
		ADXTrend:    m.getADXTrendText(analysis.ADX),
		BBStatus:    m.getBBStatusText(analysis.Bollinger),
		StochSignal: m.getStochSignalText(analysis.Stochastic),
	}
}

func (m *Manager) getRSISignalText(rsi RSIResult) string {
	if rsi.IsOverbought {
		return "OVERBOUGHT"
	}
	if rsi.IsOversold {
		return "OVERSOLD"
	}
	return "NEUTRAL"
}

func (m *Manager) getADXTrendText(adx ADXResult) string {
	if !adx.Trending {
		return "NO TREND"
	}
	return adx.Direction.String() + " " + adx.Strength.String()
}

func (m *Manager) getBBStatusText(bb BollingerResult) string {
	if bb.Squeeze {
		return "SQUEEZE"
	}
	if bb.Breakout == BreakoutUpper {
		return "UPPER BREAKOUT"
	}
	if bb.Breakout == BreakoutLower {
		return "LOWER BREAKOUT"
	}
	return "NORMAL"
}

func (m *Manager) getStochSignalText(stoch StochResult) string {
	if stoch.Overbought {
		return "OVERBOUGHT"
	}
	if stoch.Oversold {
		return "OVERSOLD"
	}
	return "NEUTRAL"
}
