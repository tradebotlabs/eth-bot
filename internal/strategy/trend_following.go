package strategy

import (
	"github.com/eth-trading/internal/indicators"
)

// TrendFollowingConfig holds configuration for trend following strategy
type TrendFollowingConfig struct {
	// ADX settings
	ADXThreshold        float64 // Min ADX for trend confirmation
	ADXStrongThreshold  float64 // ADX for strong trend

	// MA settings
	FastMAPeriod        int
	SlowMAPeriod        int
	TrendMAPeriod       int

	// MACD confirmation
	UseMACDConfirmation bool

	// Stop loss / Take profit multipliers
	StopLossATRMult     float64
	TakeProfitATRMult   float64

	// Trailing stop
	UseTrailingStop     bool
	TrailingATRMult     float64

	// Volume confirmation
	RequireVolume       bool
	VolumeThreshold     float64
}

// DefaultTrendFollowingConfig returns default configuration
func DefaultTrendFollowingConfig() *TrendFollowingConfig {
	return &TrendFollowingConfig{
		ADXThreshold:        25,
		ADXStrongThreshold:  40,
		FastMAPeriod:        10,
		SlowMAPeriod:        20,
		TrendMAPeriod:       50,
		UseMACDConfirmation: true,
		StopLossATRMult:     2.0,
		TakeProfitATRMult:   4.5,  // Increased from 3.0 to 4.5 for 2.25:1 R:R ratio
		UseTrailingStop:     true,
		TrailingATRMult:     2.5,
		RequireVolume:       false,
		VolumeThreshold:     1.2,
	}
}

// TrendFollowingStrategy implements trend following trading
type TrendFollowingStrategy struct {
	BaseStrategy
	config *TrendFollowingConfig
}

// NewTrendFollowingStrategy creates a new trend following strategy
func NewTrendFollowingStrategy(config *TrendFollowingConfig) *TrendFollowingStrategy {
	if config == nil {
		config = DefaultTrendFollowingConfig()
	}

	return &TrendFollowingStrategy{
		BaseStrategy: NewBaseStrategy("trend_following", 60, 14),
		config:       config,
	}
}

// Analyze analyzes market data for trend following signals
func (s *TrendFollowingStrategy) Analyze(data *MarketData) []Signal {
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
func (s *TrendFollowingStrategy) ShouldEnter(data *MarketData) (bool, Direction, float64) {
	analysis := data.Analysis

	// Check if in trending regime
	if !analysis.ADX.Trending || analysis.ADX.ADX < s.config.ADXThreshold {
		return false, DirectionNone, 0
	}

	// Calculate MAs
	closes := data.Closes
	fastMA := indicators.SMALast(closes, s.config.FastMAPeriod)
	slowMA := indicators.SMALast(closes, s.config.SlowMAPeriod)
	trendMA := indicators.SMALast(closes, s.config.TrendMAPeriod)
	price := closes[len(closes)-1]

	// Check for MA crossover and alignment
	var direction Direction
	var strength float64

	// Bullish conditions
	if fastMA > slowMA && price > trendMA && analysis.ADX.Direction == indicators.TrendUp {
		direction = DirectionLong
		strength = s.calculateStrength(analysis, true)
	}

	// Bearish conditions
	if fastMA < slowMA && price < trendMA && analysis.ADX.Direction == indicators.TrendDown {
		direction = DirectionShort
		strength = s.calculateStrength(analysis, false)
	}

	if direction == DirectionNone {
		return false, DirectionNone, 0
	}

	// MACD confirmation
	if s.config.UseMACDConfirmation {
		if direction == DirectionLong && analysis.MACD.MACD < analysis.MACD.Signal {
			return false, DirectionNone, 0
		}
		if direction == DirectionShort && analysis.MACD.MACD > analysis.MACD.Signal {
			return false, DirectionNone, 0
		}
	}

	// Volume confirmation
	if s.config.RequireVolume && analysis.Volume.Ratio < s.config.VolumeThreshold {
		strength *= 0.7 // Reduce strength without volume
	}

	return true, direction, strength
}

// calculateStrength calculates signal strength
func (s *TrendFollowingStrategy) calculateStrength(analysis indicators.AnalysisResult, bullish bool) float64 {
	strength := 0.5

	// ADX contribution
	if analysis.ADX.ADX >= s.config.ADXStrongThreshold {
		strength += 0.2
	} else if analysis.ADX.ADX >= s.config.ADXThreshold {
		strength += 0.1
	}

	// Trend strength contribution
	switch analysis.ADX.Strength {
	case indicators.TrendVeryStrong:
		strength += 0.15
	case indicators.TrendStrong:
		strength += 0.1
	case indicators.TrendModerate:
		strength += 0.05
	}

	// MACD confirmation
	if bullish && analysis.MACD.Histogram > 0 {
		strength += 0.1
	} else if !bullish && analysis.MACD.Histogram < 0 {
		strength += 0.1
	}

	// Volume boost
	if analysis.Volume.IsHighVolume {
		strength += 0.05
	}

	if strength > 1.0 {
		strength = 1.0
	}

	return strength
}

// ShouldExit checks exit conditions
func (s *TrendFollowingStrategy) ShouldExit(data *MarketData, position *Position) (bool, string) {
	analysis := data.Analysis
	price := data.CurrentPrice

	// Stop loss hit
	if position.Direction == DirectionLong && price <= position.StopLoss {
		return true, "Stop loss triggered"
	}
	if position.Direction == DirectionShort && price >= position.StopLoss {
		return true, "Stop loss triggered"
	}

	// Take profit hit
	if position.Direction == DirectionLong && price >= position.TakeProfit {
		return true, "Take profit reached"
	}
	if position.Direction == DirectionShort && price <= position.TakeProfit {
		return true, "Take profit reached"
	}

	// Trend reversal
	if !analysis.ADX.Trending {
		return true, "Trend ended (ADX < threshold)"
	}

	if position.Direction == DirectionLong && analysis.ADX.Direction == indicators.TrendDown {
		return true, "Trend reversed to bearish"
	}
	if position.Direction == DirectionShort && analysis.ADX.Direction == indicators.TrendUp {
		return true, "Trend reversed to bullish"
	}

	// MA crossover against position
	closes := data.Closes
	fastMA := indicators.SMALast(closes, s.config.FastMAPeriod)
	slowMA := indicators.SMALast(closes, s.config.SlowMAPeriod)

	if position.Direction == DirectionLong && fastMA < slowMA {
		return true, "Bearish MA crossover"
	}
	if position.Direction == DirectionShort && fastMA > slowMA {
		return true, "Bullish MA crossover"
	}

	return false, ""
}

// CalculateStopLoss calculates stop loss
func (s *TrendFollowingStrategy) CalculateStopLoss(data *MarketData, direction Direction, entryPrice float64) float64 {
	return s.CalculateATRStop(data, direction, entryPrice, s.config.StopLossATRMult)
}

// CalculateTakeProfit calculates take profit
func (s *TrendFollowingStrategy) CalculateTakeProfit(data *MarketData, direction Direction, entryPrice float64) float64 {
	return s.CalculateATRTarget(data, direction, entryPrice, s.config.TakeProfitATRMult)
}

// getEntryReason returns entry reason description
func (s *TrendFollowingStrategy) getEntryReason(data *MarketData, direction Direction) string {
	analysis := data.Analysis

	if direction == DirectionLong {
		return "Bullish trend: ADX=" + formatFloat(analysis.ADX.ADX) +
			", +DI > -DI, Price above trend MA, MACD bullish"
	}
	return "Bearish trend: ADX=" + formatFloat(analysis.ADX.ADX) +
		", -DI > +DI, Price below trend MA, MACD bearish"
}

// GetConfig returns strategy configuration
func (s *TrendFollowingStrategy) GetConfig() interface{} {
	return s.config
}

// UpdateTrailingStop calculates new trailing stop
func (s *TrendFollowingStrategy) UpdateTrailingStop(data *MarketData, position *Position) float64 {
	if !s.config.UseTrailingStop {
		return position.StopLoss
	}

	atr := data.Analysis.ATR.ATR
	price := data.CurrentPrice

	var newStop float64
	if position.Direction == DirectionLong {
		newStop = price - (atr * s.config.TrailingATRMult)
		if newStop > position.StopLoss {
			return newStop
		}
	} else {
		newStop = price + (atr * s.config.TrailingATRMult)
		if newStop < position.StopLoss {
			return newStop
		}
	}

	return position.StopLoss
}

// Helper function to format float
func formatFloat(f float64) string {
	return string([]byte{
		byte(int(f)/10 + '0'),
		byte(int(f)%10 + '0'),
	})
}
