package indicators

// MovingAverage provides moving average calculations
type MovingAverage struct {
	shortPeriod  int
	mediumPeriod int
	longPeriod   int
	maType       MAType
}

// MAType represents moving average type
type MAType int

const (
	MATypeSMA MAType = iota
	MATypeEMA
	MATypeWMA
	MATypeDEMA
	MATypeTEMA
)

// NewMovingAverage creates a new moving average calculator
func NewMovingAverage(shortPeriod, mediumPeriod, longPeriod int, maType MAType) *MovingAverage {
	if shortPeriod <= 0 {
		shortPeriod = 10
	}
	if mediumPeriod <= 0 {
		mediumPeriod = 20
	}
	if longPeriod <= 0 {
		longPeriod = 50
	}
	return &MovingAverage{
		shortPeriod:  shortPeriod,
		mediumPeriod: mediumPeriod,
		longPeriod:   longPeriod,
		maType:       maType,
	}
}

// Calculate calculates MAs and returns the result
func (ma *MovingAverage) Calculate(closes []float64) MAResult {
	if len(closes) < ma.longPeriod {
		return MAResult{}
	}

	var shortMA, longMA float64

	switch ma.maType {
	case MATypeEMA:
		shortEMA := EMA(closes, ma.shortPeriod)
		longEMA := EMA(closes, ma.longPeriod)
		if shortEMA != nil && longEMA != nil {
			shortMA = shortEMA[len(shortEMA)-1]
			longMA = longEMA[len(longEMA)-1]
		}
	case MATypeWMA:
		shortWMA := WMA(closes, ma.shortPeriod)
		longWMA := WMA(closes, ma.longPeriod)
		if shortWMA != nil && longWMA != nil {
			shortMA = shortWMA[len(shortWMA)-1]
			longMA = longWMA[len(longWMA)-1]
		}
	case MATypeDEMA:
		shortDEMA := DEMA(closes, ma.shortPeriod)
		longDEMA := DEMA(closes, ma.longPeriod)
		if shortDEMA != nil && longDEMA != nil {
			shortMA = shortDEMA[len(shortDEMA)-1]
			longMA = longDEMA[len(longDEMA)-1]
		}
	case MATypeTEMA:
		shortTEMA := TEMA(closes, ma.shortPeriod)
		longTEMA := TEMA(closes, ma.longPeriod)
		if shortTEMA != nil && longTEMA != nil {
			shortMA = shortTEMA[len(shortTEMA)-1]
			longMA = longTEMA[len(longTEMA)-1]
		}
	default: // SMA
		shortSMA := SMA(closes, ma.shortPeriod)
		longSMA := SMA(closes, ma.longPeriod)
		if shortSMA != nil && longSMA != nil {
			shortMA = shortSMA[len(shortSMA)-1]
			longMA = longSMA[len(longSMA)-1]
		}
	}

	return MAResult{
		Value: shortMA,
		Trend: ma.getTrend(closes[len(closes)-1], shortMA, longMA),
		Crossover: ma.detectCrossover(closes),
	}
}

// getTrend determines trend based on MA relationship
func (ma *MovingAverage) getTrend(price, shortMA, longMA float64) TrendDirection {
	if price > shortMA && shortMA > longMA {
		return TrendUp
	}
	if price < shortMA && shortMA < longMA {
		return TrendDown
	}
	return TrendNeutral
}

// detectCrossover detects MA crossover
func (ma *MovingAverage) detectCrossover(closes []float64) CrossoverType {
	if len(closes) < ma.longPeriod+1 {
		return CrossoverNone
	}

	var shortMA, longMA, prevShortMA, prevLongMA float64

	switch ma.maType {
	case MATypeEMA:
		short := EMA(closes, ma.shortPeriod)
		long := EMA(closes, ma.longPeriod)
		if short == nil || long == nil || len(short) < 2 || len(long) < 2 {
			return CrossoverNone
		}
		offset := len(short) - len(long)
		shortMA = short[len(short)-1]
		longMA = long[len(long)-1]
		prevShortMA = short[len(short)-2]
		prevLongMA = long[len(long)-2]
		if offset > 0 && len(short)-2 >= offset {
			prevShortMA = short[len(short)-2]
		}
	default:
		short := SMA(closes, ma.shortPeriod)
		long := SMA(closes, ma.longPeriod)
		if short == nil || long == nil || len(short) < 2 || len(long) < 2 {
			return CrossoverNone
		}
		offset := len(short) - len(long)
		shortMA = short[len(short)-1]
		longMA = long[len(long)-1]
		if offset > 0 && len(short) >= offset+2 {
			prevShortMA = short[len(short)-2]
		}
		prevLongMA = long[len(long)-2]
	}

	// Golden cross (short crosses above long)
	if prevShortMA <= prevLongMA && shortMA > longMA {
		return CrossoverBullish
	}
	// Death cross (short crosses below long)
	if prevShortMA >= prevLongMA && shortMA < longMA {
		return CrossoverBearish
	}

	return CrossoverNone
}

// GoldenCross detects golden cross (short MA crosses above long MA)
func GoldenCross(closes []float64, shortPeriod, longPeriod int, maType MAType) bool {
	ma := NewMovingAverage(shortPeriod, 0, longPeriod, maType)
	return ma.detectCrossover(closes) == CrossoverBullish
}

// DeathCross detects death cross (short MA crosses below long MA)
func DeathCross(closes []float64, shortPeriod, longPeriod int, maType MAType) bool {
	ma := NewMovingAverage(shortPeriod, 0, longPeriod, maType)
	return ma.detectCrossover(closes) == CrossoverBearish
}

// PriceAboveMA checks if price is above moving average
func PriceAboveMA(closes []float64, period int, maType MAType) bool {
	if len(closes) < period {
		return false
	}

	price := closes[len(closes)-1]
	var ma float64

	switch maType {
	case MATypeEMA:
		ema := EMA(closes, period)
		if ema != nil {
			ma = ema[len(ema)-1]
		}
	default:
		sma := SMA(closes, period)
		if sma != nil {
			ma = sma[len(sma)-1]
		}
	}

	return price > ma
}

// PriceBelowMA checks if price is below moving average
func PriceBelowMA(closes []float64, period int, maType MAType) bool {
	if len(closes) < period {
		return false
	}

	price := closes[len(closes)-1]
	var ma float64

	switch maType {
	case MATypeEMA:
		ema := EMA(closes, period)
		if ema != nil {
			ma = ema[len(ema)-1]
		}
	default:
		sma := SMA(closes, period)
		if sma != nil {
			ma = sma[len(sma)-1]
		}
	}

	return price < ma
}

// MASlope calculates the slope of a moving average
func MASlope(closes []float64, period, lookback int, maType MAType) float64 {
	if len(closes) < period+lookback {
		return 0
	}

	var maValues []float64

	switch maType {
	case MATypeEMA:
		maValues = EMA(closes, period)
	default:
		maValues = SMA(closes, period)
	}

	if maValues == nil || len(maValues) < lookback {
		return 0
	}

	recent := maValues[len(maValues)-lookback:]
	slope, _ := LinearRegression(recent)
	return slope
}

// MATrend determines trend based on MA slope
func MATrend(closes []float64, period, lookback int, maType MAType) TrendDirection {
	slope := MASlope(closes, period, lookback, maType)

	if slope > 0 {
		return TrendUp
	}
	if slope < 0 {
		return TrendDown
	}
	return TrendNeutral
}

// MADistance calculates distance from price to MA as percentage
func MADistance(closes []float64, period int, maType MAType) float64 {
	if len(closes) < period {
		return 0
	}

	price := closes[len(closes)-1]
	var ma float64

	switch maType {
	case MATypeEMA:
		ema := EMA(closes, period)
		if ema != nil {
			ma = ema[len(ema)-1]
		}
	default:
		sma := SMA(closes, period)
		if sma != nil {
			ma = sma[len(sma)-1]
		}
	}

	if ma == 0 {
		return 0
	}

	return ((price - ma) / ma) * 100
}

// MultiMAAlignment checks alignment of multiple MAs
func MultiMAAlignment(closes []float64, periods []int, maType MAType) TrendDirection {
	if len(periods) < 2 {
		return TrendNeutral
	}

	// Sort periods ascending
	sortedPeriods := make([]int, len(periods))
	copy(sortedPeriods, periods)
	for i := 0; i < len(sortedPeriods)-1; i++ {
		for j := i + 1; j < len(sortedPeriods); j++ {
			if sortedPeriods[i] > sortedPeriods[j] {
				sortedPeriods[i], sortedPeriods[j] = sortedPeriods[j], sortedPeriods[i]
			}
		}
	}

	// Check if needed data available
	maxPeriod := sortedPeriods[len(sortedPeriods)-1]
	if len(closes) < maxPeriod {
		return TrendNeutral
	}

	// Calculate MAs
	mas := make([]float64, len(sortedPeriods))
	for i, period := range sortedPeriods {
		switch maType {
		case MATypeEMA:
			ema := EMA(closes, period)
			if ema != nil {
				mas[i] = ema[len(ema)-1]
			}
		default:
			sma := SMA(closes, period)
			if sma != nil {
				mas[i] = sma[len(sma)-1]
			}
		}
	}

	// Check bullish alignment (shorter MA > longer MA)
	bullish := true
	for i := 0; i < len(mas)-1; i++ {
		if mas[i] <= mas[i+1] {
			bullish = false
			break
		}
	}
	if bullish {
		return TrendUp
	}

	// Check bearish alignment (shorter MA < longer MA)
	bearish := true
	for i := 0; i < len(mas)-1; i++ {
		if mas[i] >= mas[i+1] {
			bearish = false
			break
		}
	}
	if bearish {
		return TrendDown
	}

	return TrendNeutral
}

// RibbonStatus analyzes MA ribbon status
type RibbonStatus struct {
	Aligned   bool
	Direction TrendDirection
	Expanding bool
	Spread    float64
}

// MARibbon analyzes multiple MAs as a ribbon
func MARibbon(closes []float64, periods []int, maType MAType, lookback int) RibbonStatus {
	status := RibbonStatus{
		Direction: TrendNeutral,
	}

	if len(periods) < 2 {
		return status
	}

	maxPeriod := 0
	for _, p := range periods {
		if p > maxPeriod {
			maxPeriod = p
		}
	}

	if len(closes) < maxPeriod+lookback {
		return status
	}

	// Calculate current MAs
	currentMAs := make([]float64, len(periods))
	for i, period := range periods {
		switch maType {
		case MATypeEMA:
			ema := EMA(closes, period)
			if ema != nil {
				currentMAs[i] = ema[len(ema)-1]
			}
		default:
			sma := SMA(closes, period)
			if sma != nil {
				currentMAs[i] = sma[len(sma)-1]
			}
		}
	}

	// Calculate spread
	maxMA := Max(currentMAs)
	minMA := Min(currentMAs)
	if minMA > 0 {
		status.Spread = (maxMA - minMA) / minMA * 100
	}

	// Check alignment
	status.Direction = MultiMAAlignment(closes, periods, maType)
	status.Aligned = status.Direction != TrendNeutral

	// Check if expanding (compare with lookback ago)
	if lookback > 0 && len(closes) > maxPeriod+lookback {
		prevCloses := closes[:len(closes)-lookback]
		prevMAs := make([]float64, len(periods))
		for i, period := range periods {
			switch maType {
			case MATypeEMA:
				ema := EMA(prevCloses, period)
				if ema != nil {
					prevMAs[i] = ema[len(ema)-1]
				}
			default:
				sma := SMA(prevCloses, period)
				if sma != nil {
					prevMAs[i] = sma[len(sma)-1]
				}
			}
		}

		prevMaxMA := Max(prevMAs)
		prevMinMA := Min(prevMAs)
		var prevSpread float64
		if prevMinMA > 0 {
			prevSpread = (prevMaxMA - prevMinMA) / prevMinMA * 100
		}

		status.Expanding = status.Spread > prevSpread
	}

	return status
}

// VWMA calculates Volume Weighted Moving Average
func VWMA(closes, volumes []float64, period int) []float64 {
	if len(closes) < period || len(closes) != len(volumes) {
		return nil
	}

	result := make([]float64, len(closes)-period+1)

	for i := period - 1; i < len(closes); i++ {
		var sumPV, sumV float64
		for j := i - period + 1; j <= i; j++ {
			sumPV += closes[j] * volumes[j]
			sumV += volumes[j]
		}
		if sumV > 0 {
			result[i-period+1] = sumPV / sumV
		}
	}

	return result
}

// HullMA calculates Hull Moving Average
func HullMA(closes []float64, period int) []float64 {
	if len(closes) < period {
		return nil
	}

	halfPeriod := period / 2
	sqrtPeriod := int(sqrt(float64(period)))

	// WMA of half period
	wmaHalf := WMA(closes, halfPeriod)
	// WMA of full period
	wmaFull := WMA(closes, period)

	if wmaHalf == nil || wmaFull == nil {
		return nil
	}

	// Align lengths
	offset := len(wmaHalf) - len(wmaFull)
	if offset < 0 {
		return nil
	}

	// 2 * WMA(half) - WMA(full)
	raw := make([]float64, len(wmaFull))
	for i := 0; i < len(wmaFull); i++ {
		raw[i] = 2*wmaHalf[i+offset] - wmaFull[i]
	}

	// WMA of raw with sqrt period
	return WMA(raw, sqrtPeriod)
}

// ALMA calculates Arnaud Legoux Moving Average
func ALMA(closes []float64, period int, offset, sigma float64) []float64 {
	if len(closes) < period {
		return nil
	}

	// Calculate weights
	m := offset * float64(period-1)
	s := float64(period) / sigma
	weights := make([]float64, period)
	var sumW float64

	for i := 0; i < period; i++ {
		w := Exp(-((float64(i) - m) * (float64(i) - m)) / (2 * s * s))
		weights[i] = w
		sumW += w
	}

	// Normalize weights
	for i := range weights {
		weights[i] /= sumW
	}

	// Calculate ALMA
	result := make([]float64, len(closes)-period+1)
	for i := period - 1; i < len(closes); i++ {
		var sum float64
		for j := 0; j < period; j++ {
			sum += closes[i-period+1+j] * weights[j]
		}
		result[i-period+1] = sum
	}

	return result
}

// Exp calculates e^x
func Exp(x float64) float64 {
	// Simple approximation for small x
	if x < -20 {
		return 0
	}
	if x > 20 {
		return 1e9
	}

	// Taylor series approximation
	result := 1.0
	term := 1.0
	for i := 1; i < 20; i++ {
		term *= x / float64(i)
		result += term
		if term < 1e-10 && term > -1e-10 {
			break
		}
	}
	return result
}
