package indicators

// Stochastic calculates Stochastic Oscillator
type Stochastic struct {
	kPeriod    int
	dPeriod    int
	slowing    int
	overbought float64
	oversold   float64
}

// NewStochastic creates a new Stochastic calculator
func NewStochastic(kPeriod, dPeriod, slowing int, overbought, oversold float64) *Stochastic {
	if kPeriod <= 0 {
		kPeriod = 14
	}
	if dPeriod <= 0 {
		dPeriod = 3
	}
	if slowing <= 0 {
		slowing = 3
	}
	if overbought <= 0 {
		overbought = 80
	}
	if oversold <= 0 {
		oversold = 20
	}
	return &Stochastic{
		kPeriod:    kPeriod,
		dPeriod:    dPeriod,
		slowing:    slowing,
		overbought: overbought,
		oversold:   oversold,
	}
}

// Calculate calculates Stochastic for a series
func (s *Stochastic) Calculate(highs, lows, closes []float64) StochResult {
	data := CalculateStochastic(highs, lows, closes, s.kPeriod, s.dPeriod, s.slowing)
	if len(data.K) == 0 {
		return StochResult{}
	}

	idx := len(data.K) - 1
	k := data.K[idx]
	d := data.D[idx]

	var crossover CrossoverType
	if idx > 0 {
		crossover = s.detectCrossover(data.K, data.D)
	}

	return StochResult{
		K:          k,
		D:          d,
		Overbought: k >= s.overbought,
		Oversold:   k <= s.oversold,
		Crossover:  crossover,
	}
}

// detectCrossover detects %K/%D crossover
func (s *Stochastic) detectCrossover(k, d []float64) CrossoverType {
	if len(k) < 2 || len(d) < 2 {
		return CrossoverNone
	}

	idx := len(k) - 1

	// Bullish: K crosses above D
	if k[idx-1] <= d[idx-1] && k[idx] > d[idx] {
		return CrossoverBullish
	}

	// Bearish: K crosses below D
	if k[idx-1] >= d[idx-1] && k[idx] < d[idx] {
		return CrossoverBearish
	}

	return CrossoverNone
}

// StochData holds complete Stochastic data
type StochData struct {
	K []float64
	D []float64
}

// CalculateStochastic calculates Stochastic Oscillator
func CalculateStochastic(highs, lows, closes []float64, kPeriod, dPeriod, slowing int) StochData {
	n := len(closes)
	if n < kPeriod+slowing+dPeriod-2 || len(highs) != n || len(lows) != n {
		return StochData{}
	}

	// Calculate raw %K
	rawK := make([]float64, n-kPeriod+1)
	for i := kPeriod - 1; i < n; i++ {
		high := Max(highs[i-kPeriod+1 : i+1])
		low := Min(lows[i-kPeriod+1 : i+1])

		if high == low {
			rawK[i-kPeriod+1] = 50
		} else {
			rawK[i-kPeriod+1] = 100 * (closes[i] - low) / (high - low)
		}
	}

	// Apply slowing (SMA of raw K)
	slowK := SMA(rawK, slowing)
	if slowK == nil {
		return StochData{}
	}

	// Calculate %D (SMA of slow K)
	slowD := SMA(slowK, dPeriod)
	if slowD == nil {
		return StochData{K: slowK}
	}

	// Align lengths
	offset := len(slowK) - len(slowD)

	return StochData{
		K: slowK[offset:],
		D: slowD,
	}
}

// StochLast calculates last Stochastic values
func StochLast(highs, lows, closes []float64, kPeriod, dPeriod, slowing int) StochResult {
	data := CalculateStochastic(highs, lows, closes, kPeriod, dPeriod, slowing)
	if len(data.K) == 0 {
		return StochResult{}
	}

	idx := len(data.K) - 1
	return StochResult{
		K:          data.K[idx],
		D:          data.D[idx],
		Overbought: data.K[idx] >= 80,
		Oversold:   data.K[idx] <= 20,
	}
}

// FastStochastic calculates Fast Stochastic (no slowing)
func FastStochastic(highs, lows, closes []float64, kPeriod, dPeriod int) StochData {
	return CalculateStochastic(highs, lows, closes, kPeriod, dPeriod, 1)
}

// SlowStochastic calculates Slow Stochastic (with slowing)
func SlowStochastic(highs, lows, closes []float64, kPeriod, dPeriod, slowing int) StochData {
	return CalculateStochastic(highs, lows, closes, kPeriod, dPeriod, slowing)
}

// FullStochastic calculates Full Stochastic with all parameters configurable
func FullStochastic(highs, lows, closes []float64, kPeriod, kSlowing, dPeriod int) StochData {
	return CalculateStochastic(highs, lows, closes, kPeriod, dPeriod, kSlowing)
}

// StochWithDivergence detects Stochastic divergence with price
func StochWithDivergence(highs, lows, closes []float64, kPeriod, dPeriod, slowing, lookback int) (result StochResult, bullishDiv, bearishDiv bool) {
	data := CalculateStochastic(highs, lows, closes, kPeriod, dPeriod, slowing)
	if len(data.K) < lookback {
		return StochResult{}, false, false
	}

	idx := len(data.K) - 1
	result = StochResult{
		K: data.K[idx],
		D: data.D[idx],
	}

	// Get recent values
	recentK := data.K[len(data.K)-lookback:]
	recentCloses := closes[len(closes)-lookback:]

	// Find local extremes
	kLow := Min(recentK)
	kHigh := Max(recentK)
	priceLow := Min(recentCloses)
	priceHigh := Max(recentCloses)

	currentK := data.K[idx]
	currentPrice := closes[len(closes)-1]

	// Bullish divergence
	if currentPrice <= priceLow && currentK > kLow {
		bullishDiv = true
	}

	// Bearish divergence
	if currentPrice >= priceHigh && currentK < kHigh {
		bearishDiv = true
	}

	return
}

// Williams %R (similar to Stochastic but inverted)
func WilliamsR(highs, lows, closes []float64, period int) []float64 {
	n := len(closes)
	if n < period || len(highs) != n || len(lows) != n {
		return nil
	}

	result := make([]float64, n-period+1)

	for i := period - 1; i < n; i++ {
		high := Max(highs[i-period+1 : i+1])
		low := Min(lows[i-period+1 : i+1])

		if high == low {
			result[i-period+1] = -50
		} else {
			result[i-period+1] = -100 * (high - closes[i]) / (high - low)
		}
	}

	return result
}

// WilliamsRLast calculates last Williams %R value
func WilliamsRLast(highs, lows, closes []float64, period int) float64 {
	wr := WilliamsR(highs, lows, closes, period)
	if len(wr) == 0 {
		return -50
	}
	return wr[len(wr)-1]
}

// CCI calculates Commodity Channel Index
func CCI(highs, lows, closes []float64, period int) []float64 {
	n := len(closes)
	if n < period || len(highs) != n || len(lows) != n {
		return nil
	}

	// Calculate typical prices
	tp := make([]float64, n)
	for i := 0; i < n; i++ {
		tp[i] = (highs[i] + lows[i] + closes[i]) / 3
	}

	result := make([]float64, n-period+1)

	for i := period - 1; i < n; i++ {
		window := tp[i-period+1 : i+1]
		sma := Mean(window)

		// Calculate mean deviation
		var sumDev float64
		for _, v := range window {
			sumDev += Abs(v - sma)
		}
		meanDev := sumDev / float64(period)

		if meanDev == 0 {
			result[i-period+1] = 0
		} else {
			result[i-period+1] = (tp[i] - sma) / (0.015 * meanDev)
		}
	}

	return result
}

// CCILast calculates last CCI value
func CCILast(highs, lows, closes []float64, period int) float64 {
	cci := CCI(highs, lows, closes, period)
	if len(cci) == 0 {
		return 0
	}
	return cci[len(cci)-1]
}

// UltimateOscillator calculates Ultimate Oscillator
func UltimateOscillator(highs, lows, closes []float64, period1, period2, period3 int) []float64 {
	n := len(closes)
	maxPeriod := period3
	if n < maxPeriod+1 || len(highs) != n || len(lows) != n {
		return nil
	}

	// Calculate BP and TR
	bp := make([]float64, n-1)
	tr := make([]float64, n-1)

	for i := 1; i < n; i++ {
		trueLow := MinF(lows[i], closes[i-1])
		trueHigh := MaxF(highs[i], closes[i-1])
		bp[i-1] = closes[i] - trueLow
		tr[i-1] = trueHigh - trueLow
	}

	// Calculate rolling sums
	result := make([]float64, n-maxPeriod)

	for i := maxPeriod - 1; i < len(bp); i++ {
		var bp1, tr1, bp2, tr2, bp3, tr3 float64

		for j := 0; j < period1; j++ {
			bp1 += bp[i-j]
			tr1 += tr[i-j]
		}
		for j := 0; j < period2; j++ {
			bp2 += bp[i-j]
			tr2 += tr[i-j]
		}
		for j := 0; j < period3; j++ {
			bp3 += bp[i-j]
			tr3 += tr[i-j]
		}

		var avg1, avg2, avg3 float64
		if tr1 > 0 {
			avg1 = bp1 / tr1
		}
		if tr2 > 0 {
			avg2 = bp2 / tr2
		}
		if tr3 > 0 {
			avg3 = bp3 / tr3
		}

		result[i-maxPeriod+1] = 100 * (4*avg1 + 2*avg2 + avg3) / 7
	}

	return result
}

// AwesomeOscillator calculates Awesome Oscillator
func AwesomeOscillator(highs, lows []float64, fastPeriod, slowPeriod int) []float64 {
	if len(highs) != len(lows) || len(highs) < slowPeriod {
		return nil
	}

	// Calculate median price
	mp := make([]float64, len(highs))
	for i := 0; i < len(highs); i++ {
		mp[i] = (highs[i] + lows[i]) / 2
	}

	// Calculate SMAs
	fastSMA := SMA(mp, fastPeriod)
	slowSMA := SMA(mp, slowPeriod)

	if fastSMA == nil || slowSMA == nil {
		return nil
	}

	// Align and calculate AO
	offset := len(fastSMA) - len(slowSMA)
	result := make([]float64, len(slowSMA))

	for i := 0; i < len(slowSMA); i++ {
		result[i] = fastSMA[i+offset] - slowSMA[i]
	}

	return result
}

// AcceleratorOscillator calculates Accelerator Oscillator
func AcceleratorOscillator(highs, lows []float64, fastPeriod, slowPeriod, acPeriod int) []float64 {
	ao := AwesomeOscillator(highs, lows, fastPeriod, slowPeriod)
	if ao == nil || len(ao) < acPeriod {
		return nil
	}

	// AC = AO - SMA(AO)
	aoSMA := SMA(ao, acPeriod)
	if aoSMA == nil {
		return nil
	}

	offset := len(ao) - len(aoSMA)
	result := make([]float64, len(aoSMA))

	for i := 0; i < len(aoSMA); i++ {
		result[i] = ao[i+offset] - aoSMA[i]
	}

	return result
}
