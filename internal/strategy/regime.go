package strategy

import (
	"sync"

	"github.com/eth-trading/internal/indicators"
)

// MarketRegime represents different market conditions
type MarketRegime int

const (
	RegimeUnknown MarketRegime = iota
	RegimeTrending
	RegimeMeanReverting
	RegimeBreakout
	RegimeHighVolatility
	RegimeConsolidating
)

func (r MarketRegime) String() string {
	switch r {
	case RegimeTrending:
		return "TRENDING"
	case RegimeMeanReverting:
		return "MEAN_REVERTING"
	case RegimeBreakout:
		return "BREAKOUT"
	case RegimeHighVolatility:
		return "HIGH_VOLATILITY"
	case RegimeConsolidating:
		return "CONSOLIDATING"
	default:
		return "UNKNOWN"
	}
}

// RegimeConfig holds configuration for regime detection
type RegimeConfig struct {
	// ADX thresholds
	ADXTrendingThreshold    float64 // Above this = trending
	ADXWeakThreshold        float64 // Below this = ranging

	// RSI thresholds for mean reversion
	RSIOverbought           float64
	RSIOversold             float64

	// Volatility thresholds
	ATRHighVolMultiplier    float64 // Multiple of average ATR for high vol
	BBSqueezeThreshold      float64 // Bandwidth % for squeeze detection

	// Trend confirmation
	MinTrendBars            int     // Min bars for trend confirmation
	TrendSlopeThreshold     float64 // Min MA slope for trend

	// Volume thresholds
	VolumeBreakoutMultiplier float64

	// Lookback periods
	VolatilityLookback      int
	TrendLookback           int
}

// DefaultRegimeConfig returns default regime configuration
func DefaultRegimeConfig() *RegimeConfig {
	return &RegimeConfig{
		ADXTrendingThreshold:     25,
		ADXWeakThreshold:         20,
		RSIOverbought:            70,
		RSIOversold:              30,
		ATRHighVolMultiplier:     1.5,
		BBSqueezeThreshold:       0.05,
		MinTrendBars:             5,
		TrendSlopeThreshold:      0.1,
		VolumeBreakoutMultiplier: 2.0,
		VolatilityLookback:       20,
		TrendLookback:            20,
	}
}

// RegimeResult holds regime detection result
type RegimeResult struct {
	Regime        MarketRegime
	Confidence    float64 // 0-1 confidence in detection
	TrendDir      indicators.TrendDirection
	TrendStrength indicators.TrendStrength
	Volatility    VolatilityState
	Details       RegimeDetails
}

// RegimeDetails holds detailed regime analysis
type RegimeDetails struct {
	ADX           float64
	RSI           float64
	ATRPercent    float64
	BBWidth       float64
	BBSqueeze     bool
	VolumeRatio   float64
	TrendSlope    float64
	PricePosition string // Above/below key MAs
}

// VolatilityState represents volatility conditions
type VolatilityState int

const (
	VolatilityLow VolatilityState = iota
	VolatilityNormal
	VolatilityHigh
	VolatilityExpanding
	VolatilityContracting
)

func (v VolatilityState) String() string {
	switch v {
	case VolatilityLow:
		return "LOW"
	case VolatilityNormal:
		return "NORMAL"
	case VolatilityHigh:
		return "HIGH"
	case VolatilityExpanding:
		return "EXPANDING"
	case VolatilityContracting:
		return "CONTRACTING"
	default:
		return "UNKNOWN"
	}
}

// RegimeDetector detects market regimes
type RegimeDetector struct {
	config     *RegimeConfig
	indicators *indicators.Manager

	// History for regime persistence
	lastRegime  MarketRegime
	regimeCount int
	mu          sync.RWMutex
}

// NewRegimeDetector creates a new regime detector
func NewRegimeDetector(config *RegimeConfig, indicatorManager *indicators.Manager) *RegimeDetector {
	if config == nil {
		config = DefaultRegimeConfig()
	}

	return &RegimeDetector{
		config:     config,
		indicators: indicatorManager,
	}
}

// Detect detects the current market regime
func (rd *RegimeDetector) Detect(opens, highs, lows, closes, volumes []float64) RegimeResult {
	rd.mu.Lock()
	defer rd.mu.Unlock()

	// Get indicator analysis
	analysis := rd.indicators.Analyze(opens, highs, lows, closes, volumes)

	// Build details
	details := RegimeDetails{
		ADX:         analysis.ADX.ADX,
		RSI:         analysis.RSI.Value,
		ATRPercent:  analysis.ATR.ATRPercent,
		BBWidth:     analysis.Bollinger.Width,
		BBSqueeze:   analysis.Bollinger.Squeeze,
		VolumeRatio: analysis.Volume.Ratio,
	}

	// Calculate trend slope
	if len(closes) >= rd.config.TrendLookback {
		details.TrendSlope, _ = indicators.LinearRegression(closes[len(closes)-rd.config.TrendLookback:])
	}

	// Determine price position relative to MAs
	details.PricePosition = rd.getPricePosition(closes)

	// Detect regime
	regime, confidence := rd.detectRegime(analysis, details)

	// Apply persistence (regime shouldn't flip-flop)
	regime, confidence = rd.applyPersistence(regime, confidence)

	return RegimeResult{
		Regime:        regime,
		Confidence:    confidence,
		TrendDir:      analysis.ADX.Direction,
		TrendStrength: analysis.ADX.Strength,
		Volatility:    rd.detectVolatilityState(analysis, details),
		Details:       details,
	}
}

// detectRegime determines the primary regime
func (rd *RegimeDetector) detectRegime(analysis indicators.AnalysisResult, details RegimeDetails) (MarketRegime, float64) {
	scores := make(map[MarketRegime]float64)

	// Trending regime
	if analysis.ADX.ADX >= rd.config.ADXTrendingThreshold && analysis.ADX.Trending {
		score := 0.5
		if analysis.ADX.Strength >= indicators.TrendStrong {
			score += 0.3
		}
		if analysis.MA.Trend == analysis.ADX.Direction {
			score += 0.2
		}
		scores[RegimeTrending] = score
	}

	// Mean reverting regime
	if analysis.ADX.ADX < rd.config.ADXWeakThreshold {
		score := 0.4
		if analysis.RSI.IsOverbought || analysis.RSI.IsOversold {
			score += 0.3
		}
		if analysis.Bollinger.PercentB < 0.2 || analysis.Bollinger.PercentB > 0.8 {
			score += 0.2
		}
		if !analysis.Volume.IsHighVolume {
			score += 0.1
		}
		scores[RegimeMeanReverting] = score
	}

	// Breakout regime
	if analysis.Bollinger.Breakout != indicators.BreakoutNone {
		score := 0.4
		if analysis.Volume.IsHighVolume {
			score += 0.3
		}
		if details.BBSqueeze {
			score += 0.2 // Squeeze preceding breakout
		}
		scores[RegimeBreakout] = score
	}

	// High volatility regime
	if analysis.ATR.HighVolatility || details.BBWidth > 0.1 {
		score := 0.5
		if analysis.ATR.ATRPercent > rd.config.ATRHighVolMultiplier {
			score += 0.2
		}
		if analysis.Volume.IsHighVolume {
			score += 0.2
		}
		scores[RegimeHighVolatility] = score
	}

	// Consolidating regime
	if details.BBSqueeze && analysis.ADX.ADX < rd.config.ADXWeakThreshold {
		score := 0.5
		if !analysis.Volume.IsHighVolume {
			score += 0.2
		}
		if analysis.RSI.Value > 40 && analysis.RSI.Value < 60 {
			score += 0.2
		}
		scores[RegimeConsolidating] = score
	}

	// Find highest scoring regime
	maxScore := 0.0
	regime := RegimeUnknown

	for r, score := range scores {
		if score > maxScore {
			maxScore = score
			regime = r
		}
	}

	return regime, maxScore
}

// detectVolatilityState determines volatility state
func (rd *RegimeDetector) detectVolatilityState(analysis indicators.AnalysisResult, details RegimeDetails) VolatilityState {
	if analysis.ATR.HighVolatility {
		return VolatilityHigh
	}

	if details.BBSqueeze {
		return VolatilityLow
	}

	// Check for expansion/contraction (would need historical data)
	if details.BBWidth > 0.08 {
		return VolatilityHigh
	}

	if details.BBWidth < 0.03 {
		return VolatilityLow
	}

	return VolatilityNormal
}

// getPricePosition determines price position relative to key levels
func (rd *RegimeDetector) getPricePosition(closes []float64) string {
	if len(closes) < 50 {
		return "UNKNOWN"
	}

	price := closes[len(closes)-1]

	// Calculate SMAs
	sma20 := indicators.SMALast(closes, 20)
	sma50 := indicators.SMALast(closes, 50)

	if price > sma20 && price > sma50 && sma20 > sma50 {
		return "BULLISH_ABOVE_MAS"
	}
	if price < sma20 && price < sma50 && sma20 < sma50 {
		return "BEARISH_BELOW_MAS"
	}
	if price > sma20 && price < sma50 {
		return "BETWEEN_MAS_BULLISH"
	}
	if price < sma20 && price > sma50 {
		return "BETWEEN_MAS_BEARISH"
	}

	return "MIXED"
}

// applyPersistence applies regime persistence to prevent flip-flopping
func (rd *RegimeDetector) applyPersistence(regime MarketRegime, confidence float64) (MarketRegime, float64) {
	if regime == rd.lastRegime {
		rd.regimeCount++
		// Increase confidence for consistent regime
		confidence = clampConfidence(confidence + float64(rd.regimeCount)*0.05)
	} else {
		// Require higher confidence to change regime
		if confidence < 0.6 && rd.regimeCount >= 3 {
			// Stick with old regime if new confidence is low
			regime = rd.lastRegime
			rd.regimeCount++
		} else {
			rd.lastRegime = regime
			rd.regimeCount = 1
		}
	}

	return regime, confidence
}

// GetRecommendedStrategies returns strategies suitable for current regime
func (rd *RegimeDetector) GetRecommendedStrategies(regime MarketRegime) []string {
	switch regime {
	case RegimeTrending:
		return []string{"trend_following", "momentum"}
	case RegimeMeanReverting:
		return []string{"mean_reversion", "rsi_divergence"}
	case RegimeBreakout:
		return []string{"breakout", "momentum"}
	case RegimeHighVolatility:
		return []string{"volatility", "breakout"}
	case RegimeConsolidating:
		return []string{"range_trading", "mean_reversion"}
	default:
		return []string{}
	}
}

// ShouldAvoidTrading returns true if trading should be avoided
func (rd *RegimeDetector) ShouldAvoidTrading(result RegimeResult) bool {
	// Avoid trading in unknown regime
	if result.Regime == RegimeUnknown {
		return true
	}

	// Low confidence
	if result.Confidence < 0.4 {
		return true
	}

	return false
}

// RegimeHistory tracks regime changes over time
type RegimeHistory struct {
	entries []RegimeEntry
	maxSize int
	mu      sync.RWMutex
}

// RegimeEntry represents a historical regime entry
type RegimeEntry struct {
	Timestamp  int64
	Regime     MarketRegime
	Confidence float64
	Duration   int // Number of periods
}

// NewRegimeHistory creates a new regime history tracker
func NewRegimeHistory(maxSize int) *RegimeHistory {
	if maxSize <= 0 {
		maxSize = 100
	}
	return &RegimeHistory{
		entries: make([]RegimeEntry, 0, maxSize),
		maxSize: maxSize,
	}
}

// Add adds a regime entry
func (rh *RegimeHistory) Add(timestamp int64, regime MarketRegime, confidence float64) {
	rh.mu.Lock()
	defer rh.mu.Unlock()

	// Check if same regime as last entry
	if len(rh.entries) > 0 {
		last := &rh.entries[len(rh.entries)-1]
		if last.Regime == regime {
			last.Duration++
			last.Confidence = (last.Confidence + confidence) / 2
			return
		}
	}

	// New regime
	entry := RegimeEntry{
		Timestamp:  timestamp,
		Regime:     regime,
		Confidence: confidence,
		Duration:   1,
	}

	rh.entries = append(rh.entries, entry)

	// Trim if needed
	if len(rh.entries) > rh.maxSize {
		rh.entries = rh.entries[1:]
	}
}

// GetRecent returns recent regime entries
func (rh *RegimeHistory) GetRecent(n int) []RegimeEntry {
	rh.mu.RLock()
	defer rh.mu.RUnlock()

	if n > len(rh.entries) {
		n = len(rh.entries)
	}

	result := make([]RegimeEntry, n)
	copy(result, rh.entries[len(rh.entries)-n:])
	return result
}

// GetDominantRegime returns the most common regime over recent periods
func (rh *RegimeHistory) GetDominantRegime(periods int) MarketRegime {
	entries := rh.GetRecent(periods)
	if len(entries) == 0 {
		return RegimeUnknown
	}

	counts := make(map[MarketRegime]int)
	for _, e := range entries {
		counts[e.Regime] += e.Duration
	}

	maxCount := 0
	dominant := RegimeUnknown
	for regime, count := range counts {
		if count > maxCount {
			maxCount = count
			dominant = regime
		}
	}

	return dominant
}
