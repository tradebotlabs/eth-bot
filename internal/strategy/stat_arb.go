package strategy

import (
	"math"

	"github.com/eth-trading/internal/indicators"
)

// StatArbConfig holds configuration for statistical arbitrage strategy
type StatArbConfig struct {
	// Z-score settings
	ZScorePeriod       int     // Lookback for mean/std calculation
	ZScoreEntryThreshold float64 // Z-score threshold for entry
	ZScoreExitThreshold  float64 // Z-score threshold for exit

	// Mean reversion
	UseHalfLife        bool    // Use half-life for mean reversion
	MaxHalfLife        int     // Maximum acceptable half-life

	// RSI confirmation
	UseRSI             bool
	RSIPeriod          int
	RSIOverbought      float64
	RSIOversold        float64

	// Hurst exponent (mean reversion strength)
	UseHurst           bool
	HurstThreshold     float64 // Below this suggests mean reversion

	// Risk management
	StopLossZScore     float64 // Stop at this Z-score
	MaxHoldingPeriod   int     // Maximum bars to hold position

	// Stop loss / Take profit
	StopLossATRMult    float64
	TakeProfitATRMult  float64
}

// DefaultStatArbConfig returns default configuration
func DefaultStatArbConfig() *StatArbConfig {
	return &StatArbConfig{
		ZScorePeriod:        20,
		ZScoreEntryThreshold: 2.0,
		ZScoreExitThreshold:  0.5,
		UseHalfLife:         false,
		MaxHalfLife:         20,
		UseRSI:              true,
		RSIPeriod:           14,
		RSIOverbought:       70,
		RSIOversold:         30,
		UseHurst:            false,
		HurstThreshold:      0.5,
		StopLossZScore:      3.0,
		MaxHoldingPeriod:    50,
		StopLossATRMult:     2.5,
		TakeProfitATRMult:   1.5,
	}
}

// StatArbStrategy implements statistical arbitrage trading
type StatArbStrategy struct {
	BaseStrategy
	config *StatArbConfig

	// State
	entryBar    int
	currentBar  int
}

// NewStatArbStrategy creates a new stat arb strategy
func NewStatArbStrategy(config *StatArbConfig) *StatArbStrategy {
	if config == nil {
		config = DefaultStatArbConfig()
	}

	return &StatArbStrategy{
		BaseStrategy: NewBaseStrategy("stat_arb", 50, 14),
		config:       config,
	}
}

// Analyze analyzes market data for stat arb signals
func (s *StatArbStrategy) Analyze(data *MarketData) []Signal {
	if !s.enabled || len(data.Closes) < s.minData {
		return nil
	}

	s.currentBar++

	var signals []Signal

	shouldEnter, direction, strength := s.ShouldEnter(data)
	if shouldEnter {
		signal := s.CreateSignal(data, SignalTypeEntry, direction, strength, s.getEntryReason(data, direction))
		signal.StopLoss = s.CalculateStopLoss(data, direction, signal.Price)
		signal.TakeProfit = s.CalculateTakeProfit(data, direction, signal.Price)
		signals = append(signals, signal)
		s.entryBar = s.currentBar
	}

	return signals
}

// ShouldEnter checks entry conditions
func (s *StatArbStrategy) ShouldEnter(data *MarketData) (bool, Direction, float64) {
	closes := data.Closes
	if len(closes) < s.config.ZScorePeriod {
		return false, DirectionNone, 0
	}

	// Calculate Z-score
	zScore := s.calculateZScore(closes)

	var direction Direction
	var strength float64

	// Entry on extreme Z-score
	if zScore <= -s.config.ZScoreEntryThreshold {
		direction = DirectionLong // Price below mean, expect reversion up
		strength = s.calculateStrength(data, zScore, true)
	} else if zScore >= s.config.ZScoreEntryThreshold {
		direction = DirectionShort // Price above mean, expect reversion down
		strength = s.calculateStrength(data, zScore, false)
	}

	if direction == DirectionNone {
		return false, DirectionNone, 0
	}

	// RSI confirmation
	if s.config.UseRSI {
		rsi := data.Analysis.RSI.Value
		if direction == DirectionLong && rsi > s.config.RSIOversold*1.2 {
			// RSI not confirming oversold
			strength *= 0.7
		}
		if direction == DirectionShort && rsi < s.config.RSIOverbought*0.8 {
			// RSI not confirming overbought
			strength *= 0.7
		}
	}

	// Check mean reversion potential
	if s.config.UseHurst {
		hurst := s.estimateHurst(closes)
		if hurst >= s.config.HurstThreshold {
			// Not mean-reverting
			return false, DirectionNone, 0
		}
		// Lower Hurst = stronger mean reversion
		strength += (s.config.HurstThreshold - hurst) * 0.5
	}

	return true, direction, strength
}

// calculateZScore calculates Z-score of current price
func (s *StatArbStrategy) calculateZScore(closes []float64) float64 {
	window := closes[len(closes)-s.config.ZScorePeriod:]
	mean := indicators.Mean(window)
	stdDev := indicators.StdDev(window)

	if stdDev == 0 {
		return 0
	}

	currentPrice := closes[len(closes)-1]
	return (currentPrice - mean) / stdDev
}

// estimateHurst estimates the Hurst exponent (simplified)
func (s *StatArbStrategy) estimateHurst(closes []float64) float64 {
	if len(closes) < 30 {
		return 0.5
	}

	// Simplified R/S analysis
	n := len(closes)
	window := closes[len(closes)-30:]

	// Calculate returns
	returns := make([]float64, len(window)-1)
	for i := 1; i < len(window); i++ {
		if window[i-1] > 0 {
			returns[i-1] = (window[i] - window[i-1]) / window[i-1]
		}
	}

	// Calculate mean and std of returns
	mean := indicators.Mean(returns)
	stdDev := indicators.StdDev(returns)

	if stdDev == 0 {
		return 0.5
	}

	// Calculate cumulative deviations
	cumDev := make([]float64, len(returns))
	cumDev[0] = returns[0] - mean
	for i := 1; i < len(returns); i++ {
		cumDev[i] = cumDev[i-1] + (returns[i] - mean)
	}

	// Range (R)
	maxDev := indicators.Max(cumDev)
	minDev := indicators.Min(cumDev)
	R := maxDev - minDev

	// R/S statistic
	RS := R / stdDev

	// Estimate Hurst (simplified: H ≈ log(R/S) / log(n))
	if RS <= 0 || n <= 1 {
		return 0.5
	}

	H := math.Log(RS) / math.Log(float64(n))

	// Clamp to valid range
	if H < 0 {
		H = 0
	}
	if H > 1 {
		H = 1
	}

	return H
}

// calculateStrength calculates signal strength
func (s *StatArbStrategy) calculateStrength(data *MarketData, zScore float64, bullish bool) float64 {
	strength := 0.5

	// Z-score contribution (more extreme = stronger)
	absZ := math.Abs(zScore)
	if absZ >= 3.0 {
		strength += 0.3
	} else if absZ >= 2.5 {
		strength += 0.2
	} else if absZ >= 2.0 {
		strength += 0.1
	}

	// RSI confirmation
	analysis := data.Analysis
	if bullish && analysis.RSI.IsOversold {
		strength += 0.15
	} else if !bullish && analysis.RSI.IsOverbought {
		strength += 0.15
	}

	// Bollinger confirmation
	if bullish && analysis.Bollinger.PercentB < 0.1 {
		strength += 0.1
	} else if !bullish && analysis.Bollinger.PercentB > 0.9 {
		strength += 0.1
	}

	// Ranging market is good for stat arb
	if !analysis.ADX.Trending {
		strength += 0.1
	}

	if strength > 1.0 {
		strength = 1.0
	}

	return strength
}

// ShouldExit checks exit conditions
func (s *StatArbStrategy) ShouldExit(data *MarketData, position *Position) (bool, string) {
	closes := data.Closes
	price := data.CurrentPrice

	// Stop loss
	if position.Direction == DirectionLong && price <= position.StopLoss {
		return true, "Stop loss triggered"
	}
	if position.Direction == DirectionShort && price >= position.StopLoss {
		return true, "Stop loss triggered"
	}

	// Take profit
	if position.Direction == DirectionLong && price >= position.TakeProfit {
		return true, "Take profit reached"
	}
	if position.Direction == DirectionShort && price <= position.TakeProfit {
		return true, "Take profit reached"
	}

	// Z-score exit
	zScore := s.calculateZScore(closes)

	if position.Direction == DirectionLong && zScore >= -s.config.ZScoreExitThreshold {
		return true, "Z-score reverted to mean"
	}
	if position.Direction == DirectionShort && zScore <= s.config.ZScoreExitThreshold {
		return true, "Z-score reverted to mean"
	}

	// Stop loss based on Z-score
	if position.Direction == DirectionLong && zScore <= -s.config.StopLossZScore {
		return true, "Z-score stop triggered"
	}
	if position.Direction == DirectionShort && zScore >= s.config.StopLossZScore {
		return true, "Z-score stop triggered"
	}

	// Maximum holding period
	if s.currentBar-s.entryBar >= s.config.MaxHoldingPeriod {
		return true, "Maximum holding period reached"
	}

	return false, ""
}

// CalculateStopLoss calculates stop loss
func (s *StatArbStrategy) CalculateStopLoss(data *MarketData, direction Direction, entryPrice float64) float64 {
	// Use both ATR stop and Z-score based stop
	atrStop := s.CalculateATRStop(data, direction, entryPrice, s.config.StopLossATRMult)

	// Z-score based stop
	closes := data.Closes
	window := closes[len(closes)-s.config.ZScorePeriod:]
	mean := indicators.Mean(window)
	stdDev := indicators.StdDev(window)

	var zScoreStop float64
	if direction == DirectionLong {
		zScoreStop = mean - (s.config.StopLossZScore * stdDev)
		// Use tighter stop
		if zScoreStop > atrStop {
			return atrStop
		}
		return zScoreStop
	}

	zScoreStop = mean + (s.config.StopLossZScore * stdDev)
	if zScoreStop < atrStop {
		return atrStop
	}
	return zScoreStop
}

// CalculateTakeProfit calculates take profit
func (s *StatArbStrategy) CalculateTakeProfit(data *MarketData, direction Direction, entryPrice float64) float64 {
	// Target is mean reversion
	closes := data.Closes
	window := closes[len(closes)-s.config.ZScorePeriod:]
	mean := indicators.Mean(window)

	// Target slightly beyond mean
	if direction == DirectionLong {
		return mean * 1.005 // 0.5% beyond mean
	}
	return mean * 0.995
}

// getEntryReason returns entry reason description
func (s *StatArbStrategy) getEntryReason(data *MarketData, direction Direction) string {
	closes := data.Closes
	zScore := s.calculateZScore(closes)

	if direction == DirectionLong {
		return "Z-score oversold: Z=" + formatZScore(zScore) +
			", RSI=" + formatFloat(data.Analysis.RSI.Value)
	}
	return "Z-score overbought: Z=" + formatZScore(zScore) +
		", RSI=" + formatFloat(data.Analysis.RSI.Value)
}

// GetConfig returns strategy configuration
func (s *StatArbStrategy) GetConfig() interface{} {
	return s.config
}

// GetCurrentZScore returns current Z-score
func (s *StatArbStrategy) GetCurrentZScore(closes []float64) float64 {
	if len(closes) < s.config.ZScorePeriod {
		return 0
	}
	return s.calculateZScore(closes)
}

// formatZScore formats Z-score for display
func formatZScore(z float64) string {
	sign := ""
	if z >= 0 {
		sign = "+"
	}
	absZ := z
	if z < 0 {
		absZ = -z
	}
	whole := int(absZ)
	frac := int((absZ - float64(whole)) * 100)
	return sign + string([]byte{
		byte(whole + '0'),
		'.',
		byte(frac/10 + '0'),
		byte(frac%10 + '0'),
	})
}
