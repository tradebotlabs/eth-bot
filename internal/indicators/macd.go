package indicators

// MACD calculates Moving Average Convergence Divergence
type MACD struct {
	fastPeriod   int
	slowPeriod   int
	signalPeriod int
	fastEMA      float64
	slowEMA      float64
	signalEMA    float64
	prevMACD     float64
	prevSignal   float64
	count        int
	values       []float64
}

// NewMACD creates a new MACD calculator
func NewMACD(fastPeriod, slowPeriod, signalPeriod int) *MACD {
	if fastPeriod <= 0 {
		fastPeriod = 12
	}
	if slowPeriod <= 0 {
		slowPeriod = 26
	}
	if signalPeriod <= 0 {
		signalPeriod = 9
	}
	return &MACD{
		fastPeriod:   fastPeriod,
		slowPeriod:   slowPeriod,
		signalPeriod: signalPeriod,
		values:       make([]float64, 0, slowPeriod),
	}
}

// Update calculates MACD with a new close price
func (m *MACD) Update(close float64) MACDResult {
	m.values = append(m.values, close)
	m.count++

	if m.count < m.slowPeriod {
		return MACDResult{}
	}

	if m.count == m.slowPeriod {
		// Initialize EMAs
		m.fastEMA = Mean(m.values[m.slowPeriod-m.fastPeriod:])
		m.slowEMA = Mean(m.values)
		m.prevMACD = m.fastEMA - m.slowEMA
		return MACDResult{MACD: m.prevMACD}
	}

	// Calculate EMAs
	fastMult := 2.0 / float64(m.fastPeriod+1)
	slowMult := 2.0 / float64(m.slowPeriod+1)
	signalMult := 2.0 / float64(m.signalPeriod+1)

	m.fastEMA = (close-m.fastEMA)*fastMult + m.fastEMA
	m.slowEMA = (close-m.slowEMA)*slowMult + m.slowEMA
	macd := m.fastEMA - m.slowEMA

	if m.count == m.slowPeriod+m.signalPeriod-1 {
		// Initialize signal line
		m.signalEMA = macd
	} else if m.count > m.slowPeriod+m.signalPeriod-1 {
		m.signalEMA = (macd-m.signalEMA)*signalMult + m.signalEMA
	}

	histogram := macd - m.signalEMA
	crossover := m.detectCrossover(macd, m.signalEMA, m.prevMACD, m.prevSignal)

	m.prevMACD = macd
	m.prevSignal = m.signalEMA

	return MACDResult{
		MACD:      macd,
		Signal:    m.signalEMA,
		Histogram: histogram,
		Crossover: crossover,
	}
}

// Calculate calculates MACD for a price series
func (m *MACD) Calculate(closes []float64) MACDResult {
	result := CalculateMACD(closes, m.fastPeriod, m.slowPeriod, m.signalPeriod)
	if result.MACD == nil || len(result.MACD) == 0 {
		return MACDResult{}
	}

	idx := len(result.MACD) - 1
	var crossover CrossoverType
	if idx > 0 {
		crossover = m.detectCrossover(
			result.MACD[idx], result.Signal[idx],
			result.MACD[idx-1], result.Signal[idx-1],
		)
	}

	return MACDResult{
		MACD:      result.MACD[idx],
		Signal:    result.Signal[idx],
		Histogram: result.Histogram[idx],
		Crossover: crossover,
	}
}

// detectCrossover detects MACD/Signal crossover
func (m *MACD) detectCrossover(macd, signal, prevMACD, prevSignal float64) CrossoverType {
	if prevMACD <= prevSignal && macd > signal {
		return CrossoverBullish
	}
	if prevMACD >= prevSignal && macd < signal {
		return CrossoverBearish
	}
	return CrossoverNone
}

// Reset resets the MACD calculator
func (m *MACD) Reset() {
	m.fastEMA = 0
	m.slowEMA = 0
	m.signalEMA = 0
	m.prevMACD = 0
	m.prevSignal = 0
	m.count = 0
	m.values = m.values[:0]
}

// MACDData holds complete MACD calculation
type MACDData struct {
	MACD      []float64
	Signal    []float64
	Histogram []float64
}

// CalculateMACD calculates MACD for a series
func CalculateMACD(closes []float64, fastPeriod, slowPeriod, signalPeriod int) MACDData {
	if len(closes) < slowPeriod+signalPeriod {
		return MACDData{}
	}

	// Calculate EMAs
	fastEMA := EMA(closes, fastPeriod)
	slowEMA := EMA(closes, slowPeriod)

	if fastEMA == nil || slowEMA == nil {
		return MACDData{}
	}

	// Align lengths
	offset := len(fastEMA) - len(slowEMA)
	macdLine := make([]float64, len(slowEMA))
	for i := 0; i < len(slowEMA); i++ {
		macdLine[i] = fastEMA[i+offset] - slowEMA[i]
	}

	// Calculate signal line
	signalLine := EMA(macdLine, signalPeriod)
	if signalLine == nil {
		return MACDData{MACD: macdLine}
	}

	// Calculate histogram
	offset = len(macdLine) - len(signalLine)
	histogram := make([]float64, len(signalLine))
	for i := 0; i < len(signalLine); i++ {
		histogram[i] = macdLine[i+offset] - signalLine[i]
	}

	return MACDData{
		MACD:      macdLine[offset:],
		Signal:    signalLine,
		Histogram: histogram,
	}
}

// MACDLast calculates only the last MACD values
func MACDLast(closes []float64, fastPeriod, slowPeriod, signalPeriod int) MACDResult {
	data := CalculateMACD(closes, fastPeriod, slowPeriod, signalPeriod)
	if len(data.MACD) == 0 {
		return MACDResult{}
	}

	idx := len(data.MACD) - 1
	var crossover CrossoverType
	if idx > 0 {
		if data.MACD[idx-1] <= data.Signal[idx-1] && data.MACD[idx] > data.Signal[idx] {
			crossover = CrossoverBullish
		} else if data.MACD[idx-1] >= data.Signal[idx-1] && data.MACD[idx] < data.Signal[idx] {
			crossover = CrossoverBearish
		}
	}

	return MACDResult{
		MACD:      data.MACD[idx],
		Signal:    data.Signal[idx],
		Histogram: data.Histogram[idx],
		Crossover: crossover,
	}
}

// MACDWithDivergence detects MACD divergence with price
func MACDWithDivergence(closes []float64, fastPeriod, slowPeriod, signalPeriod, lookback int) (result MACDResult, bullishDiv, bearishDiv bool) {
	data := CalculateMACD(closes, fastPeriod, slowPeriod, signalPeriod)
	if len(data.MACD) < lookback {
		return MACDResult{}, false, false
	}

	idx := len(data.MACD) - 1
	result = MACDResult{
		MACD:      data.MACD[idx],
		Signal:    data.Signal[idx],
		Histogram: data.Histogram[idx],
	}

	// Get recent histogram for divergence detection
	recentHist := data.Histogram[len(data.Histogram)-lookback:]
	recentCloses := closes[len(closes)-lookback:]

	histLow := Min(recentHist)
	histHigh := Max(recentHist)
	priceLow := Min(recentCloses)
	priceHigh := Max(recentCloses)

	currentHist := data.Histogram[idx]
	currentPrice := closes[len(closes)-1]

	// Bullish divergence
	if currentPrice <= priceLow && currentHist > histLow {
		bullishDiv = true
	}

	// Bearish divergence
	if currentPrice >= priceHigh && currentHist < histHigh {
		bearishDiv = true
	}

	return
}

// MACDHistogramTrend analyzes histogram trend
func MACDHistogramTrend(histogram []float64, lookback int) TrendDirection {
	if len(histogram) < lookback {
		return TrendNeutral
	}

	recent := histogram[len(histogram)-lookback:]
	slope, _ := LinearRegression(recent)

	if slope > 0 {
		return TrendUp
	} else if slope < 0 {
		return TrendDown
	}
	return TrendNeutral
}
