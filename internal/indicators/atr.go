package indicators

// ATR calculates Average True Range
type ATR struct {
	period            int
	highVolThreshold  float64
	prevATR           float64
	prevClose         float64
	count             int
	trValues          []float64
}

// NewATR creates a new ATR calculator
func NewATR(period int, highVolThreshold float64) *ATR {
	if period <= 0 {
		period = 14
	}
	if highVolThreshold <= 0 {
		highVolThreshold = 1.5
	}
	return &ATR{
		period:           period,
		highVolThreshold: highVolThreshold,
		trValues:         make([]float64, 0, period),
	}
}

// Update calculates ATR with new OHLC data
func (a *ATR) Update(high, low, close float64) ATRResult {
	if a.count == 0 {
		a.prevClose = close
		a.count++
		return ATRResult{}
	}

	tr := TrueRange(high, low, a.prevClose)
	a.prevClose = close
	a.count++

	if a.count <= a.period+1 {
		a.trValues = append(a.trValues, tr)
		if len(a.trValues) == a.period {
			a.prevATR = Mean(a.trValues)
		}
		return ATRResult{ATR: a.prevATR}
	}

	// Wilder's smoothing
	a.prevATR = (a.prevATR*float64(a.period-1) + tr) / float64(a.period)

	atrPercent := 0.0
	if close > 0 {
		atrPercent = (a.prevATR / close) * 100
	}

	return ATRResult{
		ATR:            a.prevATR,
		ATRPercent:     atrPercent,
		HighVolatility: atrPercent > a.highVolThreshold,
	}
}

// Calculate calculates ATR for OHLC series
func (a *ATR) Calculate(highs, lows, closes []float64) ATRResult {
	atr := ATRSeries(highs, lows, closes, a.period)
	if len(atr) == 0 {
		return ATRResult{}
	}

	atrValue := atr[len(atr)-1]
	close := closes[len(closes)-1]

	atrPercent := 0.0
	if close > 0 {
		atrPercent = (atrValue / close) * 100
	}

	return ATRResult{
		ATR:            atrValue,
		ATRPercent:     atrPercent,
		HighVolatility: atrPercent > a.highVolThreshold,
	}
}

// Reset resets the ATR calculator
func (a *ATR) Reset() {
	a.prevATR = 0
	a.prevClose = 0
	a.count = 0
	a.trValues = a.trValues[:0]
}

// ATRSeries calculates ATR for a series (renamed to avoid conflict with bollinger.go)
func ATRSeries(highs, lows, closes []float64, period int) []float64 {
	if len(highs) < period+1 || len(highs) != len(lows) || len(highs) != len(closes) {
		return nil
	}

	tr := TrueRanges(highs, lows, closes)
	if tr == nil {
		return nil
	}

	// Use Wilder's smoothing (SMMA)
	atr := make([]float64, len(tr)-period+1)
	atr[0] = Mean(tr[:period])

	for i := 1; i < len(atr); i++ {
		atr[i] = (atr[i-1]*float64(period-1) + tr[period-1+i]) / float64(period)
	}

	return atr
}

// ATRLast calculates only the last ATR value
func ATRLast(highs, lows, closes []float64, period int) float64 {
	atr := ATRSeries(highs, lows, closes, period)
	if len(atr) == 0 {
		return 0
	}
	return atr[len(atr)-1]
}

// ATRPercent calculates ATR as a percentage of close price
func ATRPercent(highs, lows, closes []float64, period int) []float64 {
	atr := ATRSeries(highs, lows, closes, period)
	if atr == nil {
		return nil
	}

	// Align closes with ATR
	offset := len(closes) - len(atr)
	result := make([]float64, len(atr))

	for i := 0; i < len(atr); i++ {
		if closes[offset+i] > 0 {
			result[i] = (atr[i] / closes[offset+i]) * 100
		}
	}

	return result
}

// ATRPercentLast calculates last ATR percentage
func ATRPercentLast(highs, lows, closes []float64, period int) float64 {
	atr := ATRLast(highs, lows, closes, period)
	if atr == 0 || len(closes) == 0 {
		return 0
	}
	close := closes[len(closes)-1]
	if close == 0 {
		return 0
	}
	return (atr / close) * 100
}

// ATRBands calculates ATR-based bands around price
func ATRBands(highs, lows, closes []float64, period int, multiplier float64) (upper, lower []float64) {
	atr := ATRSeries(highs, lows, closes, period)
	if atr == nil {
		return nil, nil
	}

	offset := len(closes) - len(atr)
	upper = make([]float64, len(atr))
	lower = make([]float64, len(atr))

	for i := 0; i < len(atr); i++ {
		close := closes[offset+i]
		band := multiplier * atr[i]
		upper[i] = close + band
		lower[i] = close - band
	}

	return
}

// ATRTrailingStop calculates ATR-based trailing stop
func ATRTrailingStop(highs, lows, closes []float64, period int, multiplier float64, isLong bool) []float64 {
	atr := ATRSeries(highs, lows, closes, period)
	if atr == nil {
		return nil
	}

	offset := len(closes) - len(atr)
	stops := make([]float64, len(atr))

	for i := 0; i < len(atr); i++ {
		close := closes[offset+i]
		atrDist := multiplier * atr[i]

		if isLong {
			stops[i] = close - atrDist
			// Ratchet up
			if i > 0 && stops[i] < stops[i-1] {
				stops[i] = stops[i-1]
			}
		} else {
			stops[i] = close + atrDist
			// Ratchet down
			if i > 0 && stops[i] > stops[i-1] {
				stops[i] = stops[i-1]
			}
		}
	}

	return stops
}

// VolatilityRatio calculates current volatility relative to average
func VolatilityRatio(highs, lows, closes []float64, shortPeriod, longPeriod int) float64 {
	shortATR := ATRLast(highs, lows, closes, shortPeriod)
	longATR := ATRLast(highs, lows, closes, longPeriod)

	if longATR == 0 {
		return 1.0
	}

	return shortATR / longATR
}

// NormalizedATR normalizes ATR for comparison across assets
func NormalizedATR(highs, lows, closes []float64, period int) []float64 {
	atr := ATRSeries(highs, lows, closes, period)
	if atr == nil {
		return nil
	}

	offset := len(closes) - len(atr)
	result := make([]float64, len(atr))

	for i := 0; i < len(atr); i++ {
		if closes[offset+i] > 0 {
			result[i] = atr[i] / closes[offset+i]
		}
	}

	return result
}

// ATRExpansion detects ATR expansion (volatility increase)
func ATRExpansion(highs, lows, closes []float64, period, lookback int, threshold float64) bool {
	atr := ATRSeries(highs, lows, closes, period)
	if len(atr) < lookback {
		return false
	}

	recent := atr[len(atr)-lookback:]
	current := atr[len(atr)-1]
	avg := Mean(recent[:len(recent)-1])

	return current > avg*threshold
}

// ATRContraction detects ATR contraction (volatility decrease)
func ATRContraction(highs, lows, closes []float64, period, lookback int, threshold float64) bool {
	atr := ATRSeries(highs, lows, closes, period)
	if len(atr) < lookback {
		return false
	}

	recent := atr[len(atr)-lookback:]
	current := atr[len(atr)-1]
	avg := Mean(recent[:len(recent)-1])

	return current < avg*threshold
}

// HistoricalVolatility calculates historical volatility (annualized)
func HistoricalVolatility(closes []float64, period int) float64 {
	if len(closes) < period+1 {
		return 0
	}

	// Calculate log returns
	returns := make([]float64, len(closes)-1)
	for i := 1; i < len(closes); i++ {
		if closes[i-1] > 0 && closes[i] > 0 {
			returns[i-1] = (closes[i] - closes[i-1]) / closes[i-1] * 100
		}
	}

	// Get recent returns
	recentReturns := returns[len(returns)-period:]

	// Calculate standard deviation of returns
	stdDev := StdDev(recentReturns)

	// Annualize (assuming daily data, 365 trading days for crypto)
	return stdDev * sqrt(365)
}

// RealizedVolatility calculates realized volatility from intraday data
func RealizedVolatility(closes []float64, period int) float64 {
	if len(closes) < period+1 {
		return 0
	}

	// Calculate squared log returns
	var sumSquaredReturns float64
	for i := len(closes) - period; i < len(closes); i++ {
		if closes[i-1] > 0 && closes[i] > 0 {
			ret := (closes[i] - closes[i-1]) / closes[i-1]
			sumSquaredReturns += ret * ret
		}
	}

	return sqrt(sumSquaredReturns) * 100
}

// ChaikinVolatility calculates Chaikin Volatility
func ChaikinVolatility(highs, lows []float64, period, smoothPeriod int) []float64 {
	if len(highs) < period+smoothPeriod || len(highs) != len(lows) {
		return nil
	}

	// Calculate high-low range
	hlRange := make([]float64, len(highs))
	for i := 0; i < len(highs); i++ {
		hlRange[i] = highs[i] - lows[i]
	}

	// EMA of high-low range
	emaHL := EMA(hlRange, smoothPeriod)
	if emaHL == nil {
		return nil
	}

	// Rate of change
	result := make([]float64, len(emaHL)-period)
	for i := period; i < len(emaHL); i++ {
		if emaHL[i-period] != 0 {
			result[i-period] = ((emaHL[i] - emaHL[i-period]) / emaHL[i-period]) * 100
		}
	}

	return result
}
