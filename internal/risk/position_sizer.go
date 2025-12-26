package risk

import (
	"math"

	"github.com/rs/zerolog/log"
)

// PositionSizer calculates appropriate position sizes
type PositionSizer struct {
	config *RiskConfig
}

// NewPositionSizer creates a new position sizer
func NewPositionSizer(config *RiskConfig) *PositionSizer {
	if config == nil {
		config = DefaultRiskConfig()
	}
	return &PositionSizer{config: config}
}

// CalculateSize calculates position size based on risk parameters
func (ps *PositionSizer) CalculateSize(params PositionSizeParams) PositionSizeResult {
	result := PositionSizeResult{}

	// Calculate stop distance
	result.StopDistance = ps.calculateStopDistance(params.EntryPrice, params.StopLoss, params.Direction)
	if result.StopDistance <= 0 {
		return result
	}

	// Calculate risk amount (% of equity)
	riskPercent := ps.config.MaxRiskPerTrade
	result.RiskAmount = params.Equity * riskPercent

	// Calculate position size based on risk
	if params.Direction == "LONG" {
		result.Size = result.RiskAmount / result.StopDistance
	} else {
		result.Size = result.RiskAmount / result.StopDistance
	}

	log.Debug().
		Float64("paramsEquity", params.Equity).
		Float64("riskPercent", riskPercent).
		Float64("stopDistance", result.StopDistance).
		Float64("riskAmount", result.RiskAmount).
		Float64("size", result.Size).
		Msg("Position sizer calculation")

	// Calculate position value
	result.Value = result.Size * params.EntryPrice

	// Apply limits
	log.Debug().
		Float64("sizeBeforeLimits", result.Size).
		Float64("riskAmountBeforeLimits", result.RiskAmount).
		Msg("Before applyLimits")

	result = ps.applyLimits(result, params)

	log.Debug().
		Float64("sizeAfterLimits", result.Size).
		Float64("riskAmountAfterLimits", result.RiskAmount).
		Msg("After applyLimits")

	// Apply volatility adjustment
	if ps.config.AdjustForVolatility && params.IsHighVolatility {
		result.Size *= ps.config.HighVolatilityReduction
		result.Value = result.Size * params.EntryPrice
		result.RiskAmount = result.Size * result.StopDistance
	}

	// Calculate risk percent
	if params.Equity > 0 {
		result.RiskPercent = result.RiskAmount / params.Equity
	}

	// Calculate leverage
	if params.Equity > 0 {
		result.Leverage = result.Value / params.Equity
	}

	return result
}

// PositionSizeParams holds parameters for position sizing
type PositionSizeParams struct {
	Equity           float64
	EntryPrice       float64
	StopLoss         float64
	TakeProfit       float64
	Direction        string // "LONG" or "SHORT"
	ATR              float64
	IsHighVolatility bool
	SignalStrength   float64 // 0-1, can scale position
}

// calculateStopDistance calculates distance to stop loss
func (ps *PositionSizer) calculateStopDistance(entry, stop float64, direction string) float64 {
	if direction == "LONG" {
		return entry - stop
	}
	return stop - entry
}

// applyLimits applies position size limits
func (ps *PositionSizer) applyLimits(result PositionSizeResult, params PositionSizeParams) PositionSizeResult {
	// Max position size as % of equity
	maxSize := params.Equity * ps.config.MaxPositionSize
	maxSizeUnits := maxSize / params.EntryPrice
	if result.Size > maxSizeUnits {
		log.Debug().
			Float64("size", result.Size).
			Float64("maxSizeUnits", maxSizeUnits).
			Msg("Limiting by MaxPositionSize")
		result.Size = maxSizeUnits
		result.Value = result.Size * params.EntryPrice
	}

	// Max position value
	if result.Value > ps.config.MaxPositionValue {
		log.Debug().
			Float64("value", result.Value).
			Float64("maxPositionValue", ps.config.MaxPositionValue).
			Msg("Limiting by MaxPositionValue")
		result.Value = ps.config.MaxPositionValue
		result.Size = result.Value / params.EntryPrice
	}

	// Recalculate risk amount after limits
	result.RiskAmount = result.Size * result.StopDistance

	// Check max risk per trade
	maxRisk := params.Equity * ps.config.MaxRiskPerTrade
	if result.RiskAmount > maxRisk {
		log.Debug().
			Float64("riskAmount", result.RiskAmount).
			Float64("maxRisk", maxRisk).
			Msg("Scaling down by MaxRiskPerTrade")
		// Scale down position
		scaleFactor := maxRisk / result.RiskAmount
		result.Size *= scaleFactor
		result.Value = result.Size * params.EntryPrice
		result.RiskAmount = maxRisk
	}

	// Max leverage
	if params.Equity > 0 {
		leverage := result.Value / params.Equity
		if leverage > ps.config.MaxLeverage {
			log.Debug().
				Float64("leverage", leverage).
				Float64("maxLeverage", ps.config.MaxLeverage).
				Msg("Limiting by MaxLeverage")
			scaleFactor := ps.config.MaxLeverage / leverage
			result.Size *= scaleFactor
			result.Value = result.Size * params.EntryPrice
			result.RiskAmount = result.Size * result.StopDistance
		}
	}

	return result
}

// CalculateATRSize calculates position size using ATR-based stops
func (ps *PositionSizer) CalculateATRSize(equity, entryPrice, atr float64, atrMultiplier float64, direction string) PositionSizeResult {
	stopDistance := atr * atrMultiplier

	var stopLoss float64
	if direction == "LONG" {
		stopLoss = entryPrice - stopDistance
	} else {
		stopLoss = entryPrice + stopDistance
	}

	return ps.CalculateSize(PositionSizeParams{
		Equity:     equity,
		EntryPrice: entryPrice,
		StopLoss:   stopLoss,
		Direction:  direction,
		ATR:        atr,
	})
}

// CalculateKellySize calculates position size using Kelly Criterion
func (ps *PositionSizer) CalculateKellySize(equity, winRate, avgWin, avgLoss float64) float64 {
	if avgLoss == 0 {
		return 0
	}

	// Kelly Criterion: f* = (p * b - q) / b
	// where p = win probability, q = loss probability, b = win/loss ratio
	b := avgWin / avgLoss
	p := winRate
	q := 1 - winRate

	kelly := (p*b - q) / b

	// Apply half-Kelly for safety
	kelly = kelly * 0.5

	// Clamp to max position size
	if kelly > ps.config.MaxPositionSize {
		kelly = ps.config.MaxPositionSize
	}
	if kelly < 0 {
		kelly = 0
	}

	return equity * kelly
}

// CalculateOptimalF calculates position size using Optimal F
func (ps *PositionSizer) CalculateOptimalF(equity float64, trades []TradeMetrics) float64 {
	if len(trades) == 0 {
		return equity * ps.config.DefaultPositionSize
	}

	// Find the optimal f that maximizes geometric growth
	// Simplified: use half of max winning trade / max losing trade
	var maxWin, maxLoss float64
	for _, trade := range trades {
		if trade.PnL > maxWin {
			maxWin = trade.PnL
		}
		if trade.PnL < maxLoss {
			maxLoss = trade.PnL
		}
	}

	if maxLoss == 0 {
		return equity * ps.config.DefaultPositionSize
	}

	optimalF := 0.5 * (maxWin / (-maxLoss))
	if optimalF > ps.config.MaxPositionSize {
		optimalF = ps.config.MaxPositionSize
	}

	return equity * optimalF
}

// ScaleByStrength scales position size by signal strength
func (ps *PositionSizer) ScaleByStrength(baseSize, strength float64) float64 {
	// strength is 0-1
	// Scale between 0.5x and 1.0x based on strength
	minScale := 0.5
	scale := minScale + (1-minScale)*strength
	return baseSize * scale
}

// CalculateWithRiskReward calculates position size ensuring min risk/reward
func (ps *PositionSizer) CalculateWithRiskReward(params PositionSizeParams) (PositionSizeResult, bool) {
	result := ps.CalculateSize(params)

	// Calculate risk/reward ratio
	riskDistance := ps.calculateStopDistance(params.EntryPrice, params.StopLoss, params.Direction)

	var rewardDistance float64
	if params.Direction == "LONG" {
		rewardDistance = params.TakeProfit - params.EntryPrice
	} else {
		rewardDistance = params.EntryPrice - params.TakeProfit
	}

	if riskDistance <= 0 {
		return result, false
	}

	rrRatio := rewardDistance / riskDistance

	// Check minimum R/R
	if rrRatio < ps.config.MinRiskRewardRatio {
		return result, false
	}

	return result, true
}

// CalculateMaxLossSize calculates max position for given max loss
func (ps *PositionSizer) CalculateMaxLossSize(maxLoss, entryPrice, stopLoss float64, direction string) float64 {
	stopDistance := ps.calculateStopDistance(entryPrice, stopLoss, direction)
	if stopDistance <= 0 {
		return 0
	}
	return maxLoss / stopDistance
}

// AdjustForCorrelation adjusts position size for portfolio correlation
func (ps *PositionSizer) AdjustForCorrelation(baseSize, correlation float64) float64 {
	if correlation > ps.config.MaxCorrelation {
		// Reduce position size for high correlation
		reduction := 1 - ((correlation - ps.config.MaxCorrelation) / (1 - ps.config.MaxCorrelation))
		if reduction < 0.5 {
			reduction = 0.5
		}
		return baseSize * reduction
	}
	return baseSize
}

// RoundToLotSize rounds position size to valid lot size
func (ps *PositionSizer) RoundToLotSize(size, lotSize float64) float64 {
	if lotSize <= 0 {
		return size
	}
	return math.Floor(size/lotSize) * lotSize
}

// RoundToStepSize rounds size to step size (for crypto)
func (ps *PositionSizer) RoundToStepSize(size, stepSize float64, precision int) float64 {
	if stepSize <= 0 {
		return size
	}
	steps := math.Floor(size / stepSize)
	rounded := steps * stepSize

	// Apply precision
	multiplier := math.Pow(10, float64(precision))
	return math.Floor(rounded*multiplier) / multiplier
}
