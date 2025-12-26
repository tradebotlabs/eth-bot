package strategy

import (
	"github.com/eth-trading/internal/indicators"
)

// BreakoutConfig holds configuration for breakout strategy
type BreakoutConfig struct {
	// Bollinger breakout
	BBPeriod         int
	BBStdDev         float64
	RequireSqueeze   bool    // Require BB squeeze before breakout
	SqueezeLookback  int     // Bars to look back for squeeze

	// Donchian breakout
	DonchianPeriod   int
	UseDonchian      bool

	// Volume confirmation
	RequireVolume    bool
	VolumeMultiplier float64

	// ADX filter
	MinADXForBreakout float64 // Min ADX for breakout confirmation

	// Retest
	AllowRetest      bool
	RetestBuffer     float64 // % buffer for retest

	// Stop loss / Take profit
	StopLossATRMult   float64
	TakeProfitATRMult float64
	UseRecentSwing    bool // Use recent swing low/high for stop
}

// DefaultBreakoutConfig returns default configuration
func DefaultBreakoutConfig() *BreakoutConfig {
	return &BreakoutConfig{
		BBPeriod:          20,
		BBStdDev:          2.0,
		RequireSqueeze:    true,
		SqueezeLookback:   20,
		DonchianPeriod:    20,
		UseDonchian:       false,
		RequireVolume:     true,
		VolumeMultiplier:  1.5,
		MinADXForBreakout: 20,
		AllowRetest:       false,
		RetestBuffer:      0.005,
		StopLossATRMult:   1.5,
		TakeProfitATRMult: 3.5,  // Increased from 2.5 to 3.5 for 2.33:1 R:R ratio
		UseRecentSwing:    true,
	}
}

// BreakoutStrategy implements breakout trading
type BreakoutStrategy struct {
	BaseStrategy
	config *BreakoutConfig

	// State for squeeze detection
	squeezeActive bool
	squeezeBars   int
}

// NewBreakoutStrategy creates a new breakout strategy
func NewBreakoutStrategy(config *BreakoutConfig) *BreakoutStrategy {
	if config == nil {
		config = DefaultBreakoutConfig()
	}

	return &BreakoutStrategy{
		BaseStrategy: NewBaseStrategy("breakout", 40, 14),
		config:       config,
	}
}

// Analyze analyzes market data for breakout signals
func (s *BreakoutStrategy) Analyze(data *MarketData) []Signal {
	if !s.enabled || len(data.Closes) < s.minData {
		return nil
	}

	// Update squeeze state
	s.updateSqueezeState(data)

	var signals []Signal

	shouldEnter, direction, strength := s.ShouldEnter(data)
	if shouldEnter {
		signal := s.CreateSignal(data, SignalTypeEntry, direction, strength, s.getEntryReason(data, direction))
		signal.StopLoss = s.CalculateStopLoss(data, direction, signal.Price)
		signal.TakeProfit = s.CalculateTakeProfit(data, direction, signal.Price)
		signals = append(signals, signal)
	}

	return signals
}

// updateSqueezeState updates BB squeeze tracking
func (s *BreakoutStrategy) updateSqueezeState(data *MarketData) {
	if data.Analysis.Bollinger.Squeeze {
		s.squeezeBars++
		s.squeezeActive = true
	} else {
		if s.squeezeActive {
			// Just exited squeeze
			s.squeezeActive = false
		}
		s.squeezeBars = 0
	}
}

// ShouldEnter checks entry conditions
func (s *BreakoutStrategy) ShouldEnter(data *MarketData) (bool, Direction, float64) {
	analysis := data.Analysis
	closes := data.Closes
	price := closes[len(closes)-1]

	var direction Direction
	var strength float64

	// Check for squeeze requirement
	if s.config.RequireSqueeze {
		// We need a recent squeeze that just ended
		if s.squeezeBars == 0 && !s.squeezeActive {
			// Check historical squeeze
			wasSqueezing := s.checkRecentSqueeze(data)
			if !wasSqueezing {
				return false, DirectionNone, 0
			}
		}
	}

	// Bollinger Band breakout
	if analysis.Bollinger.Breakout == indicators.BreakoutUpper {
		direction = DirectionLong
		strength = s.calculateStrength(data, true)
	} else if analysis.Bollinger.Breakout == indicators.BreakoutLower {
		direction = DirectionShort
		strength = s.calculateStrength(data, false)
	}

	// Donchian breakout (alternative)
	if s.config.UseDonchian && direction == DirectionNone {
		donchianBreakout := indicators.DonchianBreakout(data.Highs, data.Lows, closes, s.config.DonchianPeriod)
		if donchianBreakout == indicators.BreakoutUpper {
			direction = DirectionLong
			strength = s.calculateStrength(data, true)
		} else if donchianBreakout == indicators.BreakoutLower {
			direction = DirectionShort
			strength = s.calculateStrength(data, false)
		}
	}

	if direction == DirectionNone {
		return false, DirectionNone, 0
	}

	// Volume confirmation
	if s.config.RequireVolume {
		if analysis.Volume.Ratio < s.config.VolumeMultiplier {
			return false, DirectionNone, 0
		}
		strength += 0.1 // Volume boost
	}

	// ADX confirmation (trend developing)
	if analysis.ADX.ADX < s.config.MinADXForBreakout {
		strength *= 0.8 // Reduce strength without ADX confirmation
	}

	// Check for false breakout (price back inside bands)
	if direction == DirectionLong && price < analysis.Bollinger.Upper {
		return false, DirectionNone, 0
	}
	if direction == DirectionShort && price > analysis.Bollinger.Lower {
		return false, DirectionNone, 0
	}

	return true, direction, strength
}

// checkRecentSqueeze checks for recent squeeze
func (s *BreakoutStrategy) checkRecentSqueeze(data *MarketData) bool {
	if len(data.Closes) < s.config.SqueezeLookback+s.config.BBPeriod {
		return false
	}

	// Calculate BB widths for lookback period
	for i := 1; i <= s.config.SqueezeLookback; i++ {
		endIdx := len(data.Closes) - i
		if endIdx < s.config.BBPeriod {
			break
		}

		window := data.Closes[endIdx-s.config.BBPeriod : endIdx]
		bb := indicators.BollingerLast(window, s.config.BBPeriod, s.config.BBStdDev)
		if bb.Squeeze {
			return true
		}
	}

	return false
}

// calculateStrength calculates signal strength
func (s *BreakoutStrategy) calculateStrength(data *MarketData, bullish bool) float64 {
	analysis := data.Analysis
	strength := 0.5

	// Volume contribution
	if analysis.Volume.IsHighVolume {
		strength += 0.2
	} else if analysis.Volume.Ratio >= s.config.VolumeMultiplier {
		strength += 0.1
	}

	// Squeeze quality (longer squeeze = potentially stronger breakout)
	if s.squeezeBars >= 10 {
		strength += 0.15
	} else if s.squeezeBars >= 5 {
		strength += 0.1
	}

	// ADX contribution
	if analysis.ADX.ADX >= 25 {
		strength += 0.1
	}

	// Direction alignment
	if bullish && analysis.ADX.Direction == indicators.TrendUp {
		strength += 0.05
	} else if !bullish && analysis.ADX.Direction == indicators.TrendDown {
		strength += 0.05
	}

	// MACD alignment
	if bullish && analysis.MACD.Histogram > 0 {
		strength += 0.05
	} else if !bullish && analysis.MACD.Histogram < 0 {
		strength += 0.05
	}

	if strength > 1.0 {
		strength = 1.0
	}

	return strength
}

// ShouldExit checks exit conditions
func (s *BreakoutStrategy) ShouldExit(data *MarketData, position *Position) (bool, string) {
	analysis := data.Analysis
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

	// Breakout failure (price back in bands)
	if position.Direction == DirectionLong && price < analysis.Bollinger.Middle {
		return true, "Breakout failed - price below BB middle"
	}
	if position.Direction == DirectionShort && price > analysis.Bollinger.Middle {
		return true, "Breakout failed - price above BB middle"
	}

	// Momentum loss
	if position.Direction == DirectionLong && analysis.MACD.Crossover == indicators.CrossoverBearish {
		return true, "MACD bearish crossover"
	}
	if position.Direction == DirectionShort && analysis.MACD.Crossover == indicators.CrossoverBullish {
		return true, "MACD bullish crossover"
	}

	return false, ""
}

// CalculateStopLoss calculates stop loss
func (s *BreakoutStrategy) CalculateStopLoss(data *MarketData, direction Direction, entryPrice float64) float64 {
	analysis := data.Analysis

	if s.config.UseRecentSwing {
		// Use recent swing low/high
		lookback := 10
		if len(data.Lows) >= lookback {
			if direction == DirectionLong {
				recentLow := indicators.Min(data.Lows[len(data.Lows)-lookback:])
				buffer := analysis.ATR.ATR * 0.5
				return recentLow - buffer
			}
			recentHigh := indicators.Max(data.Highs[len(data.Highs)-lookback:])
			buffer := analysis.ATR.ATR * 0.5
			return recentHigh + buffer
		}
	}

	// Use BB band as stop
	if direction == DirectionLong {
		return analysis.Bollinger.Middle - (analysis.ATR.ATR * 0.5)
	}
	return analysis.Bollinger.Middle + (analysis.ATR.ATR * 0.5)
}

// CalculateTakeProfit calculates take profit
func (s *BreakoutStrategy) CalculateTakeProfit(data *MarketData, direction Direction, entryPrice float64) float64 {
	return s.CalculateATRTarget(data, direction, entryPrice, s.config.TakeProfitATRMult)
}

// getEntryReason returns entry reason description
func (s *BreakoutStrategy) getEntryReason(data *MarketData, direction Direction) string {
	analysis := data.Analysis

	breakoutType := "BB"
	if s.config.UseDonchian {
		breakoutType = "Donchian"
	}

	if direction == DirectionLong {
		return breakoutType + " upper breakout, Vol=" + formatFloat(analysis.Volume.Ratio*100) +
			"%, ADX=" + formatFloat(analysis.ADX.ADX)
	}
	return breakoutType + " lower breakout, Vol=" + formatFloat(analysis.Volume.Ratio*100) +
		"%, ADX=" + formatFloat(analysis.ADX.ADX)
}

// GetConfig returns strategy configuration
func (s *BreakoutStrategy) GetConfig() interface{} {
	return s.config
}
