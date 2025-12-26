package indicators

import "math"

// BollingerBands calculates Bollinger Bands
type BollingerBands struct {
	period           int
	stdDevMultiplier float64
	squeezeThreshold float64
	values           []float64
	prevWidth        float64
}

// NewBollingerBands creates a new Bollinger Bands calculator
func NewBollingerBands(period int, stdDevMultiplier, squeezeThreshold float64) *BollingerBands {
	if period <= 0 {
		period = 20
	}
	if stdDevMultiplier <= 0 {
		stdDevMultiplier = 2.0
	}
	if squeezeThreshold <= 0 {
		squeezeThreshold = 0.05
	}
	return &BollingerBands{
		period:           period,
		stdDevMultiplier: stdDevMultiplier,
		squeezeThreshold: squeezeThreshold,
		values:           make([]float64, 0, period),
	}
}

// Update calculates Bollinger Bands with a new close price
func (bb *BollingerBands) Update(close float64) BollingerResult {
	bb.values = append(bb.values, close)
	if len(bb.values) > bb.period {
		bb.values = bb.values[1:]
	}

	if len(bb.values) < bb.period {
		return BollingerResult{
			Upper:  close,
			Middle: close,
			Lower:  close,
		}
	}

	middle := Mean(bb.values)
	stdDev := StdDev(bb.values)
	upper := middle + bb.stdDevMultiplier*stdDev
	lower := middle - bb.stdDevMultiplier*stdDev
	width := (upper - lower) / middle

	result := BollingerResult{
		Upper:    upper,
		Middle:   middle,
		Lower:    lower,
		Width:    width,
		PercentB: bb.calculatePercentB(close, upper, lower),
		Squeeze:  width < bb.squeezeThreshold,
		Breakout: bb.detectBreakout(close, upper, lower),
	}

	bb.prevWidth = width
	return result
}

// Calculate calculates Bollinger Bands for a price series
func (bb *BollingerBands) Calculate(closes []float64) BollingerResult {
	if len(closes) < bb.period {
		return BollingerResult{}
	}

	data := CalculateBollingerBands(closes, bb.period, bb.stdDevMultiplier)
	if len(data.Upper) == 0 {
		return BollingerResult{}
	}

	idx := len(data.Upper) - 1
	close := closes[len(closes)-1]
	width := data.Width[idx]

	return BollingerResult{
		Upper:    data.Upper[idx],
		Middle:   data.Middle[idx],
		Lower:    data.Lower[idx],
		Width:    width,
		PercentB: data.PercentB[idx],
		Squeeze:  width < bb.squeezeThreshold,
		Breakout: bb.detectBreakout(close, data.Upper[idx], data.Lower[idx]),
	}
}

// calculatePercentB calculates %B indicator
func (bb *BollingerBands) calculatePercentB(close, upper, lower float64) float64 {
	if upper == lower {
		return 0.5
	}
	return (close - lower) / (upper - lower)
}

// detectBreakout detects breakout from bands
func (bb *BollingerBands) detectBreakout(close, upper, lower float64) BreakoutType {
	if close > upper {
		return BreakoutUpper
	}
	if close < lower {
		return BreakoutLower
	}
	return BreakoutNone
}

// Reset resets the calculator
func (bb *BollingerBands) Reset() {
	bb.values = bb.values[:0]
	bb.prevWidth = 0
}

// BollingerData holds complete Bollinger Bands data
type BollingerData struct {
	Upper    []float64
	Middle   []float64
	Lower    []float64
	Width    []float64
	PercentB []float64
}

// CalculateBollingerBands calculates Bollinger Bands for a series
func CalculateBollingerBands(closes []float64, period int, stdDevMultiplier float64) BollingerData {
	if len(closes) < period || period <= 0 {
		return BollingerData{}
	}

	length := len(closes) - period + 1
	result := BollingerData{
		Upper:    make([]float64, length),
		Middle:   make([]float64, length),
		Lower:    make([]float64, length),
		Width:    make([]float64, length),
		PercentB: make([]float64, length),
	}

	for i := 0; i < length; i++ {
		window := closes[i : i+period]
		middle := Mean(window)
		stdDev := StdDev(window)
		upper := middle + stdDevMultiplier*stdDev
		lower := middle - stdDevMultiplier*stdDev

		result.Upper[i] = upper
		result.Middle[i] = middle
		result.Lower[i] = lower

		if middle != 0 {
			result.Width[i] = (upper - lower) / middle
		}

		if upper != lower {
			result.PercentB[i] = (closes[i+period-1] - lower) / (upper - lower)
		} else {
			result.PercentB[i] = 0.5
		}
	}

	return result
}

// BollingerLast calculates only the last Bollinger Bands values
func BollingerLast(closes []float64, period int, stdDevMultiplier float64) BollingerResult {
	if len(closes) < period {
		return BollingerResult{}
	}

	window := closes[len(closes)-period:]
	middle := Mean(window)
	stdDev := StdDev(window)
	upper := middle + stdDevMultiplier*stdDev
	lower := middle - stdDevMultiplier*stdDev

	close := closes[len(closes)-1]
	width := 0.0
	if middle != 0 {
		width = (upper - lower) / middle
	}

	percentB := 0.5
	if upper != lower {
		percentB = (close - lower) / (upper - lower)
	}

	return BollingerResult{
		Upper:    upper,
		Middle:   middle,
		Lower:    lower,
		Width:    width,
		PercentB: percentB,
		Breakout: func() BreakoutType {
			if close > upper {
				return BreakoutUpper
			}
			if close < lower {
				return BreakoutLower
			}
			return BreakoutNone
		}(),
	}
}

// BollingerBandwidth calculates bandwidth indicator
func BollingerBandwidth(closes []float64, period int, stdDevMultiplier float64) []float64 {
	data := CalculateBollingerBands(closes, period, stdDevMultiplier)
	return data.Width
}

// BollingerSqueeze detects squeeze conditions
func BollingerSqueeze(closes []float64, period int, stdDevMultiplier float64, lookback int) bool {
	widths := BollingerBandwidth(closes, period, stdDevMultiplier)
	if len(widths) < lookback {
		return false
	}

	recent := widths[len(widths)-lookback:]
	current := widths[len(widths)-1]
	minWidth := Min(recent)

	// Squeeze when current width is near minimum
	return current <= minWidth*1.05
}

// KeltnerChannel calculates Keltner Channels (often used with BB for squeeze)
type KeltnerChannel struct {
	period     int
	multiplier float64
}

// NewKeltnerChannel creates a new Keltner Channel calculator
func NewKeltnerChannel(period int, multiplier float64) *KeltnerChannel {
	if period <= 0 {
		period = 20
	}
	if multiplier <= 0 {
		multiplier = 1.5
	}
	return &KeltnerChannel{
		period:     period,
		multiplier: multiplier,
	}
}

// KeltnerData holds Keltner Channel data
type KeltnerData struct {
	Upper  []float64
	Middle []float64
	Lower  []float64
}

// Calculate calculates Keltner Channels
func (kc *KeltnerChannel) Calculate(highs, lows, closes []float64) KeltnerData {
	if len(closes) < kc.period {
		return KeltnerData{}
	}

	// Calculate EMA of close
	middle := EMA(closes, kc.period)
	if middle == nil {
		return KeltnerData{}
	}

	// Calculate ATR
	atr := CalculateATR(highs, lows, closes, kc.period)
	if len(atr) == 0 {
		return KeltnerData{}
	}

	// Align lengths
	offset := len(middle) - len(atr)
	if offset < 0 {
		offset = 0
	}

	minLen := len(atr)
	if len(middle)-offset < minLen {
		minLen = len(middle) - offset
	}

	result := KeltnerData{
		Upper:  make([]float64, minLen),
		Middle: make([]float64, minLen),
		Lower:  make([]float64, minLen),
	}

	for i := 0; i < minLen; i++ {
		result.Middle[i] = middle[i+offset]
		result.Upper[i] = middle[i+offset] + kc.multiplier*atr[i]
		result.Lower[i] = middle[i+offset] - kc.multiplier*atr[i]
	}

	return result
}

// TTMSqueeze detects TTM Squeeze (BB inside KC)
func TTMSqueeze(highs, lows, closes []float64, bbPeriod int, bbStdDev float64, kcPeriod int, kcMultiplier float64) (squeeze bool, momentum float64) {
	if len(closes) < bbPeriod || len(closes) < kcPeriod {
		return false, 0
	}

	// Calculate Bollinger Bands
	bb := BollingerLast(closes, bbPeriod, bbStdDev)

	// Calculate Keltner Channel
	kc := NewKeltnerChannel(kcPeriod, kcMultiplier)
	kcData := kc.Calculate(highs, lows, closes)
	if len(kcData.Upper) == 0 {
		return false, 0
	}

	idx := len(kcData.Upper) - 1

	// Squeeze when BB is inside KC
	squeeze = bb.Lower > kcData.Lower[idx] && bb.Upper < kcData.Upper[idx]

	// Calculate momentum (linear regression of price)
	lookback := 20
	if len(closes) < lookback {
		lookback = len(closes)
	}
	recentCloses := closes[len(closes)-lookback:]

	// Momentum is the deviation from regression line
	slope, intercept := LinearRegression(recentCloses)
	expectedPrice := slope*float64(lookback-1) + intercept
	momentum = closes[len(closes)-1] - expectedPrice

	return
}

// DonchianChannel calculates Donchian Channels
type DonchianChannel struct {
	period int
}

// NewDonchianChannel creates a new Donchian Channel calculator
func NewDonchianChannel(period int) *DonchianChannel {
	if period <= 0 {
		period = 20
	}
	return &DonchianChannel{period: period}
}

// DonchianData holds Donchian Channel data
type DonchianData struct {
	Upper  []float64
	Middle []float64
	Lower  []float64
}

// Calculate calculates Donchian Channels
func (dc *DonchianChannel) Calculate(highs, lows []float64) DonchianData {
	if len(highs) < dc.period || len(highs) != len(lows) {
		return DonchianData{}
	}

	upper := RollingMax(highs, dc.period)
	lower := RollingMin(lows, dc.period)

	middle := make([]float64, len(upper))
	for i := range upper {
		middle[i] = (upper[i] + lower[i]) / 2
	}

	return DonchianData{
		Upper:  upper,
		Middle: middle,
		Lower:  lower,
	}
}

// DonchianBreakout detects Donchian channel breakouts
func DonchianBreakout(highs, lows, closes []float64, period int) BreakoutType {
	dc := NewDonchianChannel(period)
	data := dc.Calculate(highs, lows)
	if len(data.Upper) == 0 {
		return BreakoutNone
	}

	// Use second-to-last channel values to detect breakout
	if len(data.Upper) < 2 {
		return BreakoutNone
	}

	currentClose := closes[len(closes)-1]
	prevUpper := data.Upper[len(data.Upper)-2]
	prevLower := data.Lower[len(data.Lower)-2]

	if currentClose > prevUpper {
		return BreakoutUpper
	}
	if currentClose < prevLower {
		return BreakoutLower
	}
	return BreakoutNone
}

// BandDistance calculates distance from price to band (as percentage)
func BandDistance(price, upper, lower float64) (distUpper, distLower float64) {
	bandRange := upper - lower
	if bandRange == 0 {
		return 0, 0
	}
	distUpper = (upper - price) / bandRange
	distLower = (price - lower) / bandRange
	return
}

// BollingerMomentum calculates momentum based on Bollinger Band position
func BollingerMomentum(closes []float64, period int, stdDevMultiplier float64, lookback int) float64 {
	data := CalculateBollingerBands(closes, period, stdDevMultiplier)
	if len(data.PercentB) < lookback {
		return 0
	}

	recent := data.PercentB[len(data.PercentB)-lookback:]
	slope, _ := LinearRegression(recent)
	return slope
}

// Helper for ATR calculation (forward declaration, defined in atr.go)
func CalculateATR(highs, lows, closes []float64, period int) []float64 {
	if len(highs) < period+1 || len(highs) != len(lows) || len(highs) != len(closes) {
		return nil
	}

	tr := TrueRanges(highs, lows, closes)
	if tr == nil {
		return nil
	}

	// Use SMMA for ATR
	atr := make([]float64, len(tr)-period+1)
	atr[0] = Mean(tr[:period])

	for i := 1; i < len(atr); i++ {
		atr[i] = (atr[i-1]*float64(period-1) + tr[period-1+i]) / float64(period)
	}

	return atr
}

// Helper to get last N values
func lastN(values []float64, n int) []float64 {
	if len(values) <= n {
		return values
	}
	return values[len(values)-n:]
}

// Sqrt helper
func sqrt(x float64) float64 {
	return math.Sqrt(x)
}
