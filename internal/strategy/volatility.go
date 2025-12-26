package strategy

import (
	"github.com/eth-trading/internal/indicators"
)

// VolatilityConfig holds configuration for volatility strategy
type VolatilityConfig struct {
	// ATR settings
	ATRPeriod           int
	ATRHighMultiplier   float64 // Multiple of avg ATR for high volatility
	ATRLowMultiplier    float64 // Multiple of avg ATR for low volatility

	// Bollinger settings
	BBPeriod            int
	BBStdDev            float64
	BBWidthHighThreshold float64 // BB width for high vol
	BBWidthLowThreshold  float64 // BB width for low vol (squeeze)

	// Keltner Channel for squeeze detection
	KCPeriod            int
	KCMultiplier        float64

	// Strategy mode
	TradeExpansion      bool // Trade volatility expansion
	TradeContraction    bool // Trade volatility contraction

	// Stop loss / Take profit
	StopLossATRMult     float64
	TakeProfitATRMult   float64
	DynamicTargets      bool // Adjust targets based on volatility
}

// DefaultVolatilityConfig returns default configuration
func DefaultVolatilityConfig() *VolatilityConfig {
	return &VolatilityConfig{
		ATRPeriod:            14,
		ATRHighMultiplier:    1.5,
		ATRLowMultiplier:     0.7,
		BBPeriod:             20,
		BBStdDev:             2.0,
		BBWidthHighThreshold: 0.08,
		BBWidthLowThreshold:  0.03,
		KCPeriod:             20,
		KCMultiplier:         1.5,
		TradeExpansion:       true,
		TradeContraction:     true,
		StopLossATRMult:      2.0,
		TakeProfitATRMult:    3.0,
		DynamicTargets:       true,
	}
}

// VolatilityStrategy implements volatility-based trading
type VolatilityStrategy struct {
	BaseStrategy
	config *VolatilityConfig

	// State
	wasInSqueeze   bool
	squeezeEndBar  int
	currentBar     int
}

// NewVolatilityStrategy creates a new volatility strategy
func NewVolatilityStrategy(config *VolatilityConfig) *VolatilityStrategy {
	if config == nil {
		config = DefaultVolatilityConfig()
	}

	return &VolatilityStrategy{
		BaseStrategy: NewBaseStrategy("volatility", 40, 14),
		config:       config,
	}
}

// Analyze analyzes market data for volatility signals
func (s *VolatilityStrategy) Analyze(data *MarketData) []Signal {
	if !s.enabled || len(data.Closes) < s.minData {
		return nil
	}

	s.currentBar++
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

// updateSqueezeState updates squeeze tracking
func (s *VolatilityStrategy) updateSqueezeState(data *MarketData) {
	// Check TTM squeeze
	squeeze, _ := indicators.TTMSqueeze(
		data.Highs, data.Lows, data.Closes,
		s.config.BBPeriod, s.config.BBStdDev,
		s.config.KCPeriod, s.config.KCMultiplier,
	)

	if squeeze {
		s.wasInSqueeze = true
	} else if s.wasInSqueeze {
		// Just exited squeeze
		s.squeezeEndBar = s.currentBar
		s.wasInSqueeze = false
	}
}

// ShouldEnter checks entry conditions
func (s *VolatilityStrategy) ShouldEnter(data *MarketData) (bool, Direction, float64) {
	analysis := data.Analysis

	var direction Direction
	var strength float64

	// Volatility expansion trade
	if s.config.TradeExpansion {
		dir, str := s.checkExpansionEntry(data)
		if dir != DirectionNone {
			direction = dir
			strength = str
		}
	}

	// Volatility contraction trade (squeeze breakout)
	if s.config.TradeContraction && direction == DirectionNone {
		dir, str := s.checkSqueezeBreakout(data)
		if dir != DirectionNone {
			direction = dir
			strength = str
		}
	}

	if direction == DirectionNone {
		return false, DirectionNone, 0
	}

	// Confirm with trend
	if direction == DirectionLong && analysis.ADX.Direction == indicators.TrendDown && analysis.ADX.Trending {
		strength *= 0.7 // Reduce strength against trend
	}
	if direction == DirectionShort && analysis.ADX.Direction == indicators.TrendUp && analysis.ADX.Trending {
		strength *= 0.7
	}

	return true, direction, strength
}

// checkExpansionEntry checks for volatility expansion entry
func (s *VolatilityStrategy) checkExpansionEntry(data *MarketData) (Direction, float64) {
	analysis := data.Analysis

	// High ATR
	isHighATR := analysis.ATR.ATRPercent > (s.config.ATRHighMultiplier * 100)

	// Wide BB
	isWideBB := analysis.Bollinger.Width > s.config.BBWidthHighThreshold

	if !isHighATR && !isWideBB {
		return DirectionNone, 0
	}

	// Determine direction from price action and momentum
	closes := data.Closes
	price := closes[len(closes)-1]
	prevPrice := closes[len(closes)-2]

	var direction Direction
	strength := 0.5

	// Strong directional move with high volatility
	if price > prevPrice && analysis.MACD.Histogram > 0 {
		direction = DirectionLong
	} else if price < prevPrice && analysis.MACD.Histogram < 0 {
		direction = DirectionShort
	} else {
		return DirectionNone, 0
	}

	// Strength based on volatility level
	if isHighATR && isWideBB {
		strength += 0.2
	} else if isHighATR || isWideBB {
		strength += 0.1
	}

	// Volume confirmation
	if analysis.Volume.IsHighVolume {
		strength += 0.15
	}

	// Momentum confirmation
	if direction == DirectionLong && analysis.RSI.Value > 50 && analysis.RSI.Value < 70 {
		strength += 0.1
	} else if direction == DirectionShort && analysis.RSI.Value < 50 && analysis.RSI.Value > 30 {
		strength += 0.1
	}

	return direction, strength
}

// checkSqueezeBreakout checks for squeeze breakout entry
func (s *VolatilityStrategy) checkSqueezeBreakout(data *MarketData) (Direction, float64) {
	analysis := data.Analysis

	// Check if we just exited a squeeze (within last 3 bars)
	if s.currentBar-s.squeezeEndBar > 3 {
		return DirectionNone, 0
	}

	// Need momentum to confirm breakout direction
	squeeze, momentum := indicators.TTMSqueeze(
		data.Highs, data.Lows, data.Closes,
		s.config.BBPeriod, s.config.BBStdDev,
		s.config.KCPeriod, s.config.KCMultiplier,
	)

	if squeeze {
		// Still in squeeze
		return DirectionNone, 0
	}

	var direction Direction
	strength := 0.6 // Base strength for squeeze breakout

	// Direction from momentum
	if momentum > 0 {
		direction = DirectionLong
	} else if momentum < 0 {
		direction = DirectionShort
	} else {
		return DirectionNone, 0
	}

	// Confirm with BB breakout
	if direction == DirectionLong && analysis.Bollinger.Breakout != indicators.BreakoutUpper {
		// No upper breakout yet
		if analysis.Bollinger.PercentB < 0.8 {
			return DirectionNone, 0
		}
	}
	if direction == DirectionShort && analysis.Bollinger.Breakout != indicators.BreakoutLower {
		if analysis.Bollinger.PercentB > 0.2 {
			return DirectionNone, 0
		}
	}

	// Volume confirmation
	if analysis.Volume.IsHighVolume {
		strength += 0.2
	}

	// ADX confirmation
	if analysis.ADX.ADX > 20 {
		strength += 0.1
	}

	return direction, strength
}

// ShouldExit checks exit conditions
func (s *VolatilityStrategy) ShouldExit(data *MarketData, position *Position) (bool, string) {
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

	// Volatility contraction (for expansion trades)
	if analysis.ATR.ATRPercent < (s.config.ATRLowMultiplier * 100) {
		return true, "Volatility contracted"
	}

	// Momentum reversal
	if position.Direction == DirectionLong && analysis.MACD.Histogram < 0 {
		return true, "Momentum turned negative"
	}
	if position.Direction == DirectionShort && analysis.MACD.Histogram > 0 {
		return true, "Momentum turned positive"
	}

	return false, ""
}

// CalculateStopLoss calculates stop loss
func (s *VolatilityStrategy) CalculateStopLoss(data *MarketData, direction Direction, entryPrice float64) float64 {
	multiplier := s.config.StopLossATRMult

	// Wider stops in high volatility
	if data.Analysis.ATR.HighVolatility {
		multiplier *= 1.2
	}

	return s.CalculateATRStop(data, direction, entryPrice, multiplier)
}

// CalculateTakeProfit calculates take profit
func (s *VolatilityStrategy) CalculateTakeProfit(data *MarketData, direction Direction, entryPrice float64) float64 {
	multiplier := s.config.TakeProfitATRMult

	if s.config.DynamicTargets {
		// Larger targets in high volatility
		if data.Analysis.ATR.HighVolatility {
			multiplier *= 1.3
		}
		// Smaller targets in low volatility
		if data.Analysis.Bollinger.Width < s.config.BBWidthLowThreshold {
			multiplier *= 0.8
		}
	}

	return s.CalculateATRTarget(data, direction, entryPrice, multiplier)
}

// getEntryReason returns entry reason description
func (s *VolatilityStrategy) getEntryReason(data *MarketData, direction Direction) string {
	analysis := data.Analysis

	// Determine if squeeze or expansion trade
	var tradeType string
	if s.currentBar-s.squeezeEndBar <= 3 {
		tradeType = "Squeeze breakout"
	} else {
		tradeType = "Volatility expansion"
	}

	if direction == DirectionLong {
		return tradeType + " LONG: ATR%=" + formatFloat(analysis.ATR.ATRPercent) +
			", BBWidth=" + formatFloat(analysis.Bollinger.Width*100)
	}
	return tradeType + " SHORT: ATR%=" + formatFloat(analysis.ATR.ATRPercent) +
		", BBWidth=" + formatFloat(analysis.Bollinger.Width*100)
}

// GetConfig returns strategy configuration
func (s *VolatilityStrategy) GetConfig() interface{} {
	return s.config
}

// IsInSqueeze returns true if currently in squeeze
func (s *VolatilityStrategy) IsInSqueeze() bool {
	return s.wasInSqueeze
}
