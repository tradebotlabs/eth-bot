package indicators

// VolumeAnalyzer provides volume-based analysis
type VolumeAnalyzer struct {
	period          int
	highThreshold   float64
	lowThreshold    float64
}

// NewVolumeAnalyzer creates a new volume analyzer
func NewVolumeAnalyzer(period int, highThreshold, lowThreshold float64) *VolumeAnalyzer {
	if period <= 0 {
		period = 20
	}
	if highThreshold <= 0 {
		highThreshold = 1.5
	}
	if lowThreshold <= 0 {
		lowThreshold = 0.5
	}
	return &VolumeAnalyzer{
		period:        period,
		highThreshold: highThreshold,
		lowThreshold:  lowThreshold,
	}
}

// Analyze analyzes volume
func (va *VolumeAnalyzer) Analyze(volumes []float64) VolumeResult {
	if len(volumes) < va.period {
		return VolumeResult{}
	}

	current := volumes[len(volumes)-1]
	avg := Mean(volumes[len(volumes)-va.period:])
	ratio := 1.0
	if avg > 0 {
		ratio = current / avg
	}

	return VolumeResult{
		Current:      current,
		Average:      avg,
		Ratio:        ratio,
		IsHighVolume: ratio >= va.highThreshold,
		IsLowVolume:  ratio <= va.lowThreshold,
	}
}

// VolumeRatio calculates volume ratio to average
func VolumeRatio(volumes []float64, period int) float64 {
	if len(volumes) < period {
		return 1.0
	}

	current := volumes[len(volumes)-1]
	avg := Mean(volumes[len(volumes)-period : len(volumes)-1])

	if avg == 0 {
		return 1.0
	}

	return current / avg
}

// VolumeBreakout detects volume breakout
func VolumeBreakout(volumes []float64, period int, threshold float64) bool {
	return VolumeRatio(volumes, period) >= threshold
}

// OBV calculates On-Balance Volume
func OBV(closes, volumes []float64) []float64 {
	if len(closes) != len(volumes) || len(closes) < 2 {
		return nil
	}

	obv := make([]float64, len(closes))
	obv[0] = volumes[0]

	for i := 1; i < len(closes); i++ {
		if closes[i] > closes[i-1] {
			obv[i] = obv[i-1] + volumes[i]
		} else if closes[i] < closes[i-1] {
			obv[i] = obv[i-1] - volumes[i]
		} else {
			obv[i] = obv[i-1]
		}
	}

	return obv
}

// OBVTrend analyzes OBV trend
func OBVTrend(closes, volumes []float64, lookback int) TrendDirection {
	obv := OBV(closes, volumes)
	if obv == nil || len(obv) < lookback {
		return TrendNeutral
	}

	recent := obv[len(obv)-lookback:]
	slope, _ := LinearRegression(recent)

	if slope > 0 {
		return TrendUp
	}
	if slope < 0 {
		return TrendDown
	}
	return TrendNeutral
}

// OBVDivergence detects OBV divergence with price
func OBVDivergence(closes, volumes []float64, lookback int) (bullish, bearish bool) {
	obv := OBV(closes, volumes)
	if obv == nil || len(obv) < lookback {
		return false, false
	}

	recentCloses := closes[len(closes)-lookback:]
	recentOBV := obv[len(obv)-lookback:]

	// Find local extremes
	priceSlope, _ := LinearRegression(recentCloses)
	obvSlope, _ := LinearRegression(recentOBV)

	// Bullish: price falling, OBV rising
	if priceSlope < 0 && obvSlope > 0 {
		bullish = true
	}

	// Bearish: price rising, OBV falling
	if priceSlope > 0 && obvSlope < 0 {
		bearish = true
	}

	return
}

// VWAP calculates Volume Weighted Average Price
func VWAP(highs, lows, closes, volumes []float64) []float64 {
	n := len(closes)
	if n == 0 || len(highs) != n || len(lows) != n || len(volumes) != n {
		return nil
	}

	vwap := make([]float64, n)
	var cumPV, cumVol float64

	for i := 0; i < n; i++ {
		typicalPrice := (highs[i] + lows[i] + closes[i]) / 3
		cumPV += typicalPrice * volumes[i]
		cumVol += volumes[i]

		if cumVol > 0 {
			vwap[i] = cumPV / cumVol
		}
	}

	return vwap
}

// VWAPWithBands calculates VWAP with standard deviation bands
type VWAPData struct {
	VWAP      []float64
	Upper     []float64
	Lower     []float64
	UpperDev2 []float64
	LowerDev2 []float64
}

// VWAPBands calculates VWAP with bands
func VWAPBands(highs, lows, closes, volumes []float64, stdDevMult1, stdDevMult2 float64) VWAPData {
	n := len(closes)
	if n == 0 || len(highs) != n || len(lows) != n || len(volumes) != n {
		return VWAPData{}
	}

	data := VWAPData{
		VWAP:      make([]float64, n),
		Upper:     make([]float64, n),
		Lower:     make([]float64, n),
		UpperDev2: make([]float64, n),
		LowerDev2: make([]float64, n),
	}

	var cumPV, cumVol, cumPV2 float64

	for i := 0; i < n; i++ {
		typicalPrice := (highs[i] + lows[i] + closes[i]) / 3
		cumPV += typicalPrice * volumes[i]
		cumPV2 += typicalPrice * typicalPrice * volumes[i]
		cumVol += volumes[i]

		if cumVol > 0 {
			data.VWAP[i] = cumPV / cumVol
			variance := (cumPV2 / cumVol) - (data.VWAP[i] * data.VWAP[i])
			if variance < 0 {
				variance = 0
			}
			stdDev := sqrt(variance)

			data.Upper[i] = data.VWAP[i] + stdDevMult1*stdDev
			data.Lower[i] = data.VWAP[i] - stdDevMult1*stdDev
			data.UpperDev2[i] = data.VWAP[i] + stdDevMult2*stdDev
			data.LowerDev2[i] = data.VWAP[i] - stdDevMult2*stdDev
		}
	}

	return data
}

// ADLine calculates Accumulation/Distribution Line
func ADLine(highs, lows, closes, volumes []float64) []float64 {
	n := len(closes)
	if n == 0 || len(highs) != n || len(lows) != n || len(volumes) != n {
		return nil
	}

	ad := make([]float64, n)

	for i := 0; i < n; i++ {
		hlRange := highs[i] - lows[i]
		if hlRange > 0 {
			mfMultiplier := ((closes[i] - lows[i]) - (highs[i] - closes[i])) / hlRange
			mfVolume := mfMultiplier * volumes[i]
			if i == 0 {
				ad[i] = mfVolume
			} else {
				ad[i] = ad[i-1] + mfVolume
			}
		} else {
			if i > 0 {
				ad[i] = ad[i-1]
			}
		}
	}

	return ad
}

// ChaikinMF calculates Chaikin Money Flow
func ChaikinMF(highs, lows, closes, volumes []float64, period int) []float64 {
	n := len(closes)
	if n < period || len(highs) != n || len(lows) != n || len(volumes) != n {
		return nil
	}

	// Calculate money flow multiplier and money flow volume
	mfv := make([]float64, n)
	for i := 0; i < n; i++ {
		hlRange := highs[i] - lows[i]
		if hlRange > 0 {
			mfMultiplier := ((closes[i] - lows[i]) - (highs[i] - closes[i])) / hlRange
			mfv[i] = mfMultiplier * volumes[i]
		}
	}

	// Calculate CMF
	result := make([]float64, n-period+1)
	for i := period - 1; i < n; i++ {
		var sumMFV, sumVol float64
		for j := i - period + 1; j <= i; j++ {
			sumMFV += mfv[j]
			sumVol += volumes[j]
		}
		if sumVol > 0 {
			result[i-period+1] = sumMFV / sumVol
		}
	}

	return result
}

// MFI calculates Money Flow Index
func MFI(highs, lows, closes, volumes []float64, period int) []float64 {
	n := len(closes)
	if n < period+1 || len(highs) != n || len(lows) != n || len(volumes) != n {
		return nil
	}

	// Calculate typical prices and money flow
	typicalPrices := make([]float64, n)
	moneyFlow := make([]float64, n)

	for i := 0; i < n; i++ {
		typicalPrices[i] = (highs[i] + lows[i] + closes[i]) / 3
		moneyFlow[i] = typicalPrices[i] * volumes[i]
	}

	// Calculate MFI
	result := make([]float64, n-period)

	for i := period; i < n; i++ {
		var posFlow, negFlow float64

		for j := i - period + 1; j <= i; j++ {
			if typicalPrices[j] > typicalPrices[j-1] {
				posFlow += moneyFlow[j]
			} else if typicalPrices[j] < typicalPrices[j-1] {
				negFlow += moneyFlow[j]
			}
		}

		if negFlow == 0 {
			result[i-period] = 100
		} else {
			moneyRatio := posFlow / negFlow
			result[i-period] = 100 - (100 / (1 + moneyRatio))
		}
	}

	return result
}

// ForceIndex calculates Force Index
func ForceIndex(closes, volumes []float64, period int) []float64 {
	if len(closes) < 2 || len(closes) != len(volumes) {
		return nil
	}

	// Raw force index
	rawFI := make([]float64, len(closes)-1)
	for i := 1; i < len(closes); i++ {
		rawFI[i-1] = (closes[i] - closes[i-1]) * volumes[i]
	}

	// Apply EMA
	return EMA(rawFI, period)
}

// EaseOfMovement calculates Ease of Movement
func EaseOfMovement(highs, lows, volumes []float64, period int) []float64 {
	n := len(highs)
	if n < 2 || len(lows) != n || len(volumes) != n {
		return nil
	}

	emv := make([]float64, n-1)
	for i := 1; i < n; i++ {
		dm := ((highs[i] + lows[i]) / 2) - ((highs[i-1] + lows[i-1]) / 2)
		boxRatio := (volumes[i] / 1e8) / (highs[i] - lows[i])
		if boxRatio != 0 {
			emv[i-1] = dm / boxRatio
		}
	}

	return SMA(emv, period)
}

// VolumeProfile calculates volume profile (simplified)
type VolumeProfileLevel struct {
	Price  float64
	Volume float64
}

// VolumeProfile calculates volume at price levels
func VolumeProfile(highs, lows, closes, volumes []float64, numLevels int) []VolumeProfileLevel {
	if len(closes) == 0 || numLevels <= 0 {
		return nil
	}

	// Find price range
	minPrice := Min(lows)
	maxPrice := Max(highs)
	priceRange := maxPrice - minPrice

	if priceRange == 0 {
		return nil
	}

	levelSize := priceRange / float64(numLevels)

	// Initialize levels
	levels := make([]VolumeProfileLevel, numLevels)
	for i := 0; i < numLevels; i++ {
		levels[i].Price = minPrice + levelSize*float64(i) + levelSize/2
	}

	// Distribute volume to levels
	for i := 0; i < len(closes); i++ {
		typicalPrice := (highs[i] + lows[i] + closes[i]) / 3
		levelIdx := int((typicalPrice - minPrice) / levelSize)
		if levelIdx >= numLevels {
			levelIdx = numLevels - 1
		}
		if levelIdx < 0 {
			levelIdx = 0
		}
		levels[levelIdx].Volume += volumes[i]
	}

	return levels
}

// PointOfControl finds the price level with highest volume
func PointOfControl(highs, lows, closes, volumes []float64, numLevels int) float64 {
	profile := VolumeProfile(highs, lows, closes, volumes, numLevels)
	if profile == nil {
		return 0
	}

	var maxVol float64
	var poc float64

	for _, level := range profile {
		if level.Volume > maxVol {
			maxVol = level.Volume
			poc = level.Price
		}
	}

	return poc
}
