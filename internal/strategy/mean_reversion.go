package strategy

import (
	"github.com/eth-trading/internal/indicators"
)

// MeanReversionConfig holds configuration for mean reversion strategy
type MeanReversionConfig struct {
	// RSI settings
	RSIPeriod        int
	RSIOverbought    float64
	RSIOversold      float64
	RSIExtremeHigh   float64
	RSIExtremeLow    float64

	// Bollinger settings
	BBPeriod         int
	BBStdDev         float64
	BBEntryThreshold float64 // PercentB threshold for entry

	// ADX filter
	MaxADX           float64 // Max ADX for mean reversion (avoid trends)

	// Confirmation
	RequireRSIBB     bool // Require both RSI and BB confirmation
	RequireDivergence bool

	// Stop loss / Take profit
	StopLossATRMult   float64
	TakeProfitATRMult float64
	UseMiddleBBTarget bool // Use BB middle as target
}

// DefaultMeanReversionConfig returns default configuration
func DefaultMeanReversionConfig() *MeanReversionConfig {
	return &MeanReversionConfig{
		RSIPeriod:         14,
		RSIOverbought:     70,
		RSIOversold:       30,
		RSIExtremeHigh:    80,
		RSIExtremeLow:     20,
		BBPeriod:          20,
		BBStdDev:          2.0,
		BBEntryThreshold:  0.1, // Below 10% or above 90% of bands
		MaxADX:            25,
		RequireRSIBB:      true,
		RequireDivergence: false,
		StopLossATRMult:   1.5,
		TakeProfitATRMult: 2.0,
		UseMiddleBBTarget: true,
	}
}

// MeanReversionStrategy implements mean reversion trading
type MeanReversionStrategy struct {
	BaseStrategy
	config *MeanReversionConfig
}

// NewMeanReversionStrategy creates a new mean reversion strategy
func NewMeanReversionStrategy(config *MeanReversionConfig) *MeanReversionStrategy {
	if config == nil {
		config = DefaultMeanReversionConfig()
	}

	return &MeanReversionStrategy{
		BaseStrategy: NewBaseStrategy("mean_reversion", 30, 14),
		config:       config,
	}
}

// Analyze analyzes market data for mean reversion signals
func (s *MeanReversionStrategy) Analyze(data *MarketData) []Signal {
	if !s.enabled || len(data.Closes) < s.minData {
		return nil
	}

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

// ShouldEnter checks entry conditions
func (s *MeanReversionStrategy) ShouldEnter(data *MarketData) (bool, Direction, float64) {
	analysis := data.Analysis

	// Avoid trending markets
	if analysis.ADX.ADX > s.config.MaxADX && analysis.ADX.Trending {
		return false, DirectionNone, 0
	}

	var direction Direction
	var strength float64

	rsiOversold := analysis.RSI.Value <= s.config.RSIOversold
	rsiOverbought := analysis.RSI.Value >= s.config.RSIOverbought
	rsiExtremeOversold := analysis.RSI.Value <= s.config.RSIExtremeLow
	rsiExtremeOverbought := analysis.RSI.Value >= s.config.RSIExtremeHigh

	bbLower := analysis.Bollinger.PercentB <= s.config.BBEntryThreshold
	bbUpper := analysis.Bollinger.PercentB >= (1 - s.config.BBEntryThreshold)

	// Long setup: oversold conditions
	if s.config.RequireRSIBB {
		if rsiOversold && bbLower {
			direction = DirectionLong
			strength = s.calculateStrength(analysis, true)
		} else if rsiOverbought && bbUpper {
			direction = DirectionShort
			strength = s.calculateStrength(analysis, false)
		}
	} else {
		// Either condition is sufficient
		if rsiExtremeOversold || bbLower {
			direction = DirectionLong
			strength = s.calculateStrength(analysis, true)
		} else if rsiExtremeOverbought || bbUpper {
			direction = DirectionShort
			strength = s.calculateStrength(analysis, false)
		}
	}

	if direction == DirectionNone {
		return false, DirectionNone, 0
	}

	// Check for divergence if required
	if s.config.RequireDivergence {
		_, bullDiv, bearDiv := indicators.RSIWithDivergence(data.Closes, s.config.RSIPeriod, 10)
		if direction == DirectionLong && !bullDiv {
			return false, DirectionNone, 0
		}
		if direction == DirectionShort && !bearDiv {
			return false, DirectionNone, 0
		}
		strength += 0.1 // Bonus for divergence
	}

	return true, direction, strength
}

// calculateStrength calculates signal strength
func (s *MeanReversionStrategy) calculateStrength(analysis indicators.AnalysisResult, bullish bool) float64 {
	strength := 0.5

	// RSI contribution
	if bullish {
		if analysis.RSI.Value <= s.config.RSIExtremeLow {
			strength += 0.2
		} else if analysis.RSI.Value <= s.config.RSIOversold {
			strength += 0.1
		}
	} else {
		if analysis.RSI.Value >= s.config.RSIExtremeHigh {
			strength += 0.2
		} else if analysis.RSI.Value >= s.config.RSIOverbought {
			strength += 0.1
		}
	}

	// BB contribution
	if bullish && analysis.Bollinger.PercentB < 0 {
		strength += 0.15
	} else if bullish && analysis.Bollinger.PercentB < s.config.BBEntryThreshold {
		strength += 0.1
	}

	if !bullish && analysis.Bollinger.PercentB > 1 {
		strength += 0.15
	} else if !bullish && analysis.Bollinger.PercentB > (1-s.config.BBEntryThreshold) {
		strength += 0.1
	}

	// Low ADX is good for mean reversion
	if analysis.ADX.ADX < 20 {
		strength += 0.1
	}

	// Stochastic confirmation
	if bullish && analysis.Stochastic.Oversold {
		strength += 0.1
	} else if !bullish && analysis.Stochastic.Overbought {
		strength += 0.1
	}

	if strength > 1.0 {
		strength = 1.0
	}

	return strength
}

// ShouldExit checks exit conditions
func (s *MeanReversionStrategy) ShouldExit(data *MarketData, position *Position) (bool, string) {
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

	// Mean reversion target: RSI back to neutral
	if position.Direction == DirectionLong && analysis.RSI.Value >= 50 {
		return true, "RSI reverted to neutral"
	}
	if position.Direction == DirectionShort && analysis.RSI.Value <= 50 {
		return true, "RSI reverted to neutral"
	}

	// BB middle target
	if s.config.UseMiddleBBTarget {
		if position.Direction == DirectionLong && price >= analysis.Bollinger.Middle {
			return true, "Price reverted to BB middle"
		}
		if position.Direction == DirectionShort && price <= analysis.Bollinger.Middle {
			return true, "Price reverted to BB middle"
		}
	}

	// Conditions reversed (became trending)
	if analysis.ADX.ADX > s.config.MaxADX && analysis.ADX.Trending {
		return true, "Market became trending"
	}

	return false, ""
}

// CalculateStopLoss calculates stop loss
func (s *MeanReversionStrategy) CalculateStopLoss(data *MarketData, direction Direction, entryPrice float64) float64 {
	// Use BB band as stop
	analysis := data.Analysis

	if direction == DirectionLong {
		// Stop below lower band
		stop := analysis.Bollinger.Lower - (data.Analysis.ATR.ATR * 0.5)
		atrStop := s.CalculateATRStop(data, direction, entryPrice, s.config.StopLossATRMult)
		// Use wider stop
		if stop < atrStop {
			return stop
		}
		return atrStop
	}

	// Stop above upper band
	stop := analysis.Bollinger.Upper + (data.Analysis.ATR.ATR * 0.5)
	atrStop := s.CalculateATRStop(data, direction, entryPrice, s.config.StopLossATRMult)
	if stop > atrStop {
		return stop
	}
	return atrStop
}

// CalculateTakeProfit calculates take profit
func (s *MeanReversionStrategy) CalculateTakeProfit(data *MarketData, direction Direction, entryPrice float64) float64 {
	if s.config.UseMiddleBBTarget {
		return data.Analysis.Bollinger.Middle
	}

	return s.CalculateATRTarget(data, direction, entryPrice, s.config.TakeProfitATRMult)
}

// getEntryReason returns entry reason description
func (s *MeanReversionStrategy) getEntryReason(data *MarketData, direction Direction) string {
	analysis := data.Analysis

	if direction == DirectionLong {
		return "Oversold: RSI=" + formatFloat(analysis.RSI.Value) +
			", BB%B=" + formatFloat(analysis.Bollinger.PercentB*100) +
			", ADX=" + formatFloat(analysis.ADX.ADX)
	}
	return "Overbought: RSI=" + formatFloat(analysis.RSI.Value) +
		", BB%B=" + formatFloat(analysis.Bollinger.PercentB*100) +
		", ADX=" + formatFloat(analysis.ADX.ADX)
}

// GetConfig returns strategy configuration
func (s *MeanReversionStrategy) GetConfig() interface{} {
	return s.config
}
