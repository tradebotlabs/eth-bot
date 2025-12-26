package indicators

// RSI calculates Relative Strength Index
type RSI struct {
	period     int
	overbought float64
	oversold   float64
	prevGain   float64
	prevLoss   float64
	prevClose  float64
	count      int
	values     []float64
}

// NewRSI creates a new RSI calculator
func NewRSI(period int, overbought, oversold float64) *RSI {
	if period <= 0 {
		period = 14
	}
	if overbought <= 0 {
		overbought = 70
	}
	if oversold <= 0 {
		oversold = 30
	}
	return &RSI{
		period:     period,
		overbought: overbought,
		oversold:   oversold,
		values:     make([]float64, 0, period),
	}
}

// Update calculates RSI with a new close price
func (r *RSI) Update(close float64) float64 {
	if r.count == 0 {
		r.prevClose = close
		r.count++
		return 50 // Neutral value
	}

	change := close - r.prevClose
	r.prevClose = close

	gain := MaxF(change, 0)
	loss := MaxF(-change, 0)

	if r.count < r.period {
		r.values = append(r.values, change)
		r.count++
		return 50 // Not enough data
	}

	if r.count == r.period {
		// Calculate initial averages
		var sumGain, sumLoss float64
		for _, v := range r.values {
			if v > 0 {
				sumGain += v
			} else {
				sumLoss -= v
			}
		}
		sumGain += gain
		sumLoss += loss
		r.prevGain = sumGain / float64(r.period)
		r.prevLoss = sumLoss / float64(r.period)
		r.count++
	} else {
		// Smoothed averages
		r.prevGain = (r.prevGain*float64(r.period-1) + gain) / float64(r.period)
		r.prevLoss = (r.prevLoss*float64(r.period-1) + loss) / float64(r.period)
	}

	if r.prevLoss == 0 {
		return 100
	}

	rs := r.prevGain / r.prevLoss
	return 100 - (100 / (1 + rs))
}

// Calculate calculates RSI for the latest value
func (r *RSI) Calculate(closes []float64) RSIResult {
	if len(closes) < r.period+1 {
		return RSIResult{Value: 50}
	}

	rsi := CalculateRSI(closes, r.period)
	if len(rsi) == 0 {
		return RSIResult{Value: 50}
	}

	value := rsi[len(rsi)-1]
	return RSIResult{
		Value:        value,
		IsOverbought: value >= r.overbought,
		IsOversold:   value <= r.oversold,
		Signal:       r.getSignal(value, rsi),
	}
}

// getSignal determines the trading signal
func (r *RSI) getSignal(current float64, rsi []float64) Signal {
	if len(rsi) < 2 {
		return SignalNone
	}

	prev := rsi[len(rsi)-2]

	// Crossing thresholds
	if prev > r.overbought && current <= r.overbought {
		return SignalSell
	}
	if prev < r.oversold && current >= r.oversold {
		return SignalBuy
	}

	// Extreme values
	if current >= 80 {
		return SignalStrongSell
	}
	if current <= 20 {
		return SignalStrongBuy
	}

	return SignalNone
}

// Reset resets the RSI calculator
func (r *RSI) Reset() {
	r.prevGain = 0
	r.prevLoss = 0
	r.prevClose = 0
	r.count = 0
	r.values = r.values[:0]
}

// CalculateRSI calculates RSI for a series
func CalculateRSI(closes []float64, period int) []float64 {
	if len(closes) < period+1 || period <= 0 {
		return nil
	}

	changes := Diff(closes)
	gains, losses := GainsLosses(changes)

	// Calculate initial averages
	avgGain := Mean(gains[:period])
	avgLoss := Mean(losses[:period])

	result := make([]float64, len(closes)-period)

	for i := 0; i < len(result); i++ {
		if i == 0 {
			if avgLoss == 0 {
				result[i] = 100
			} else {
				rs := avgGain / avgLoss
				result[i] = 100 - (100 / (1 + rs))
			}
		} else {
			idx := period + i - 1
			avgGain = (avgGain*float64(period-1) + gains[idx]) / float64(period)
			avgLoss = (avgLoss*float64(period-1) + losses[idx]) / float64(period)

			if avgLoss == 0 {
				result[i] = 100
			} else {
				rs := avgGain / avgLoss
				result[i] = 100 - (100 / (1 + rs))
			}
		}
	}

	return result
}

// RSILast calculates only the last RSI value (more efficient)
func RSILast(closes []float64, period int) float64 {
	rsi := CalculateRSI(closes, period)
	if len(rsi) == 0 {
		return 50
	}
	return rsi[len(rsi)-1]
}

// RSIWithDivergence detects RSI divergence with price
func RSIWithDivergence(closes []float64, period, lookback int) (rsi float64, bullishDiv, bearishDiv bool) {
	rsiValues := CalculateRSI(closes, period)
	if len(rsiValues) < lookback {
		return 50, false, false
	}

	rsi = rsiValues[len(rsiValues)-1]

	// Get recent price and RSI lows/highs
	recentCloses := closes[len(closes)-lookback:]
	recentRSI := rsiValues[len(rsiValues)-lookback:]

	// Find local extremes
	priceLow := Min(recentCloses)
	priceHigh := Max(recentCloses)
	rsiLow := Min(recentRSI)
	rsiHigh := Max(recentRSI)

	currentPrice := closes[len(closes)-1]
	currentRSI := rsi

	// Bullish divergence: price makes lower low, RSI makes higher low
	if currentPrice <= priceLow && currentRSI > rsiLow {
		bullishDiv = true
	}

	// Bearish divergence: price makes higher high, RSI makes lower high
	if currentPrice >= priceHigh && currentRSI < rsiHigh {
		bearishDiv = true
	}

	return
}

// StochRSI calculates Stochastic RSI
func StochRSI(closes []float64, rsiPeriod, stochPeriod int) []float64 {
	rsi := CalculateRSI(closes, rsiPeriod)
	if len(rsi) < stochPeriod {
		return nil
	}

	result := make([]float64, len(rsi)-stochPeriod+1)
	for i := stochPeriod - 1; i < len(rsi); i++ {
		window := rsi[i-stochPeriod+1 : i+1]
		low := Min(window)
		high := Max(window)

		if high == low {
			result[i-stochPeriod+1] = 50
		} else {
			result[i-stochPeriod+1] = 100 * (rsi[i] - low) / (high - low)
		}
	}

	return result
}
