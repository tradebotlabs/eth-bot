package indicators

// ADX calculates Average Directional Index
type ADX struct {
	period           int
	trendingThreshold float64
	prevHigh         float64
	prevLow          float64
	prevClose        float64
	prevPlusDM       float64
	prevMinusDM      float64
	prevTR           float64
	prevDX           []float64
	prevADX          float64
	count            int
}

// NewADX creates a new ADX calculator
func NewADX(period int, trendingThreshold float64) *ADX {
	if period <= 0 {
		period = 14
	}
	if trendingThreshold <= 0 {
		trendingThreshold = 25
	}
	return &ADX{
		period:           period,
		trendingThreshold: trendingThreshold,
		prevDX:           make([]float64, 0, period),
	}
}

// Update calculates ADX with new OHLC data
func (a *ADX) Update(high, low, close float64) ADXResult {
	if a.count == 0 {
		a.prevHigh = high
		a.prevLow = low
		a.prevClose = close
		a.count++
		return ADXResult{}
	}

	// Calculate directional movement
	upMove := high - a.prevHigh
	downMove := a.prevLow - low

	plusDM := 0.0
	minusDM := 0.0

	if upMove > downMove && upMove > 0 {
		plusDM = upMove
	}
	if downMove > upMove && downMove > 0 {
		minusDM = downMove
	}

	// Calculate true range
	tr := TrueRange(high, low, a.prevClose)

	a.prevHigh = high
	a.prevLow = low
	a.prevClose = close
	a.count++

	if a.count <= a.period {
		a.prevPlusDM += plusDM
		a.prevMinusDM += minusDM
		a.prevTR += tr
		return ADXResult{}
	}

	// Smoothed values
	a.prevPlusDM = a.prevPlusDM - (a.prevPlusDM / float64(a.period)) + plusDM
	a.prevMinusDM = a.prevMinusDM - (a.prevMinusDM / float64(a.period)) + minusDM
	a.prevTR = a.prevTR - (a.prevTR / float64(a.period)) + tr

	// Calculate +DI and -DI
	plusDI := 0.0
	minusDI := 0.0
	if a.prevTR > 0 {
		plusDI = 100 * a.prevPlusDM / a.prevTR
		minusDI = 100 * a.prevMinusDM / a.prevTR
	}

	// Calculate DX
	dx := 0.0
	diSum := plusDI + minusDI
	if diSum > 0 {
		dx = 100 * Abs(plusDI-minusDI) / diSum
	}

	// Store DX for ADX calculation
	a.prevDX = append(a.prevDX, dx)
	if len(a.prevDX) > a.period {
		a.prevDX = a.prevDX[1:]
	}

	// Calculate ADX
	adx := 0.0
	if len(a.prevDX) == a.period {
		if a.prevADX == 0 {
			adx = Mean(a.prevDX)
		} else {
			adx = (a.prevADX*float64(a.period-1) + dx) / float64(a.period)
		}
		a.prevADX = adx
	}

	return ADXResult{
		ADX:       adx,
		PlusDI:    plusDI,
		MinusDI:   minusDI,
		Trending:  adx >= a.trendingThreshold,
		Strength:  a.getStrength(adx),
		Direction: a.getDirection(plusDI, minusDI),
	}
}

// Calculate calculates ADX for OHLC series
func (a *ADX) Calculate(highs, lows, closes []float64) ADXResult {
	result := CalculateADX(highs, lows, closes, a.period)
	if len(result.ADX) == 0 {
		return ADXResult{}
	}

	idx := len(result.ADX) - 1
	adx := result.ADX[idx]
	plusDI := result.PlusDI[idx]
	minusDI := result.MinusDI[idx]

	return ADXResult{
		ADX:       adx,
		PlusDI:    plusDI,
		MinusDI:   minusDI,
		Trending:  adx >= a.trendingThreshold,
		Strength:  a.getStrength(adx),
		Direction: a.getDirection(plusDI, minusDI),
	}
}

// getStrength determines trend strength from ADX value
func (a *ADX) getStrength(adx float64) TrendStrength {
	if adx >= 50 {
		return TrendVeryStrong
	}
	if adx >= 40 {
		return TrendStrong
	}
	if adx >= 25 {
		return TrendModerate
	}
	return TrendWeak
}

// getDirection determines trend direction from DI values
func (a *ADX) getDirection(plusDI, minusDI float64) TrendDirection {
	if plusDI > minusDI {
		return TrendUp
	}
	if minusDI > plusDI {
		return TrendDown
	}
	return TrendNeutral
}

// Reset resets the ADX calculator
func (a *ADX) Reset() {
	a.prevHigh = 0
	a.prevLow = 0
	a.prevClose = 0
	a.prevPlusDM = 0
	a.prevMinusDM = 0
	a.prevTR = 0
	a.prevDX = a.prevDX[:0]
	a.prevADX = 0
	a.count = 0
}

// ADXData holds complete ADX calculation
type ADXData struct {
	ADX     []float64
	PlusDI  []float64
	MinusDI []float64
	DX      []float64
}

// CalculateADX calculates ADX for a series
func CalculateADX(highs, lows, closes []float64, period int) ADXData {
	if len(highs) < period*2 || len(highs) != len(lows) || len(highs) != len(closes) {
		return ADXData{}
	}

	n := len(highs)

	// Calculate +DM and -DM
	plusDM := make([]float64, n-1)
	minusDM := make([]float64, n-1)
	tr := make([]float64, n-1)

	for i := 1; i < n; i++ {
		upMove := highs[i] - highs[i-1]
		downMove := lows[i-1] - lows[i]

		if upMove > downMove && upMove > 0 {
			plusDM[i-1] = upMove
		}
		if downMove > upMove && downMove > 0 {
			minusDM[i-1] = downMove
		}

		tr[i-1] = TrueRange(highs[i], lows[i], closes[i-1])
	}

	// Smooth +DM, -DM, and TR
	smoothedPlusDM := wilder(plusDM, period)
	smoothedMinusDM := wilder(minusDM, period)
	smoothedTR := wilder(tr, period)

	if smoothedPlusDM == nil || smoothedMinusDM == nil || smoothedTR == nil {
		return ADXData{}
	}

	// Calculate +DI and -DI
	length := len(smoothedTR)
	plusDI := make([]float64, length)
	minusDI := make([]float64, length)
	dx := make([]float64, length)

	for i := 0; i < length; i++ {
		if smoothedTR[i] > 0 {
			plusDI[i] = 100 * smoothedPlusDM[i] / smoothedTR[i]
			minusDI[i] = 100 * smoothedMinusDM[i] / smoothedTR[i]
		}

		diSum := plusDI[i] + minusDI[i]
		if diSum > 0 {
			dx[i] = 100 * Abs(plusDI[i]-minusDI[i]) / diSum
		}
	}

	// Calculate ADX (smoothed DX)
	adx := wilder(dx, period)
	if adx == nil {
		return ADXData{}
	}

	// Align lengths
	offset := length - len(adx)

	return ADXData{
		ADX:     adx,
		PlusDI:  plusDI[offset:],
		MinusDI: minusDI[offset:],
		DX:      dx[offset:],
	}
}

// wilder applies Wilder's smoothing (similar to SMMA)
func wilder(values []float64, period int) []float64 {
	if len(values) < period || period <= 0 {
		return nil
	}

	result := make([]float64, len(values)-period+1)
	result[0] = Sum(values[:period])

	for i := 1; i < len(result); i++ {
		result[i] = result[i-1] - (result[i-1] / float64(period)) + values[period-1+i]
	}

	return result
}

// ADXLast calculates only the last ADX values
func ADXLast(highs, lows, closes []float64, period int) ADXResult {
	data := CalculateADX(highs, lows, closes, period)
	if len(data.ADX) == 0 {
		return ADXResult{}
	}

	idx := len(data.ADX) - 1
	adx := data.ADX[idx]
	plusDI := data.PlusDI[idx]
	minusDI := data.MinusDI[idx]

	return ADXResult{
		ADX:     adx,
		PlusDI:  plusDI,
		MinusDI: minusDI,
		Trending: adx >= 25,
		Strength: func() TrendStrength {
			if adx >= 50 {
				return TrendVeryStrong
			}
			if adx >= 40 {
				return TrendStrong
			}
			if adx >= 25 {
				return TrendModerate
			}
			return TrendWeak
		}(),
		Direction: func() TrendDirection {
			if plusDI > minusDI {
				return TrendUp
			}
			if minusDI > plusDI {
				return TrendDown
			}
			return TrendNeutral
		}(),
	}
}

// DMICrossover detects +DI/-DI crossover
func DMICrossover(highs, lows, closes []float64, period int) CrossoverType {
	data := CalculateADX(highs, lows, closes, period)
	if len(data.PlusDI) < 2 {
		return CrossoverNone
	}

	idx := len(data.PlusDI) - 1
	prevPlusDI := data.PlusDI[idx-1]
	prevMinusDI := data.MinusDI[idx-1]
	currPlusDI := data.PlusDI[idx]
	currMinusDI := data.MinusDI[idx]

	// Bullish crossover: +DI crosses above -DI
	if prevPlusDI <= prevMinusDI && currPlusDI > currMinusDI {
		return CrossoverBullish
	}

	// Bearish crossover: +DI crosses below -DI
	if prevPlusDI >= prevMinusDI && currPlusDI < currMinusDI {
		return CrossoverBearish
	}

	return CrossoverNone
}

// ADXRising checks if ADX is rising
func ADXRising(highs, lows, closes []float64, period, lookback int) bool {
	data := CalculateADX(highs, lows, closes, period)
	if len(data.ADX) < lookback {
		return false
	}

	recent := data.ADX[len(data.ADX)-lookback:]
	slope, _ := LinearRegression(recent)
	return slope > 0
}

// ADXWithStrength categorizes ADX strength with more granularity
func ADXWithStrength(adx float64) (TrendStrength, string) {
	switch {
	case adx >= 75:
		return TrendVeryStrong, "Extremely Strong Trend"
	case adx >= 50:
		return TrendVeryStrong, "Very Strong Trend"
	case adx >= 40:
		return TrendStrong, "Strong Trend"
	case adx >= 25:
		return TrendModerate, "Moderate Trend"
	case adx >= 20:
		return TrendWeak, "Weak Trend"
	default:
		return TrendWeak, "No Clear Trend"
	}
}
