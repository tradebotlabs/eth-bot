package indicators

import (
	"math"
	"sort"
)

// Sum calculates the sum of values
func Sum(values []float64) float64 {
	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum
}

// Mean calculates the arithmetic mean
func Mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	return Sum(values) / float64(len(values))
}

// StdDev calculates the standard deviation
func StdDev(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}

	mean := Mean(values)
	var sumSquares float64
	for _, v := range values {
		diff := v - mean
		sumSquares += diff * diff
	}
	return math.Sqrt(sumSquares / float64(len(values)))
}

// StdDevSample calculates sample standard deviation (n-1 denominator)
func StdDevSample(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}

	mean := Mean(values)
	var sumSquares float64
	for _, v := range values {
		diff := v - mean
		sumSquares += diff * diff
	}
	return math.Sqrt(sumSquares / float64(len(values)-1))
}

// Variance calculates the variance
func Variance(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}

	mean := Mean(values)
	var sumSquares float64
	for _, v := range values {
		diff := v - mean
		sumSquares += diff * diff
	}
	return sumSquares / float64(len(values))
}

// Max returns the maximum value
func Max(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	max := values[0]
	for _, v := range values[1:] {
		if v > max {
			max = v
		}
	}
	return max
}

// Min returns the minimum value
func Min(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	min := values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
	}
	return min
}

// MaxF returns the maximum of two floats
func MaxF(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// MinF returns the minimum of two floats
func MinF(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// Abs returns the absolute value
func Abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

// Median calculates the median value
func Median(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	n := len(sorted)
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return sorted[n/2]
}

// SMA calculates Simple Moving Average
func SMA(values []float64, period int) []float64 {
	if len(values) < period || period <= 0 {
		return nil
	}

	result := make([]float64, len(values)-period+1)
	sum := Sum(values[:period])
	result[0] = sum / float64(period)

	for i := period; i < len(values); i++ {
		sum = sum - values[i-period] + values[i]
		result[i-period+1] = sum / float64(period)
	}

	return result
}

// SMALast calculates the last SMA value
func SMALast(values []float64, period int) float64 {
	if len(values) < period || period <= 0 {
		return 0
	}
	return Mean(values[len(values)-period:])
}

// EMA calculates Exponential Moving Average
func EMA(values []float64, period int) []float64 {
	if len(values) < period || period <= 0 {
		return nil
	}

	result := make([]float64, len(values))
	multiplier := 2.0 / float64(period+1)

	// Start with SMA for first value
	result[period-1] = Mean(values[:period])

	for i := period; i < len(values); i++ {
		result[i] = (values[i]-result[i-1])*multiplier + result[i-1]
	}

	return result[period-1:]
}

// EMALast calculates the last EMA value given previous EMA
func EMALast(value, prevEMA float64, period int) float64 {
	multiplier := 2.0 / float64(period+1)
	return (value-prevEMA)*multiplier + prevEMA
}

// EMAWithSeed calculates EMA starting from a seed value
func EMAWithSeed(values []float64, period int, seed float64) []float64 {
	if len(values) == 0 || period <= 0 {
		return nil
	}

	result := make([]float64, len(values))
	multiplier := 2.0 / float64(period+1)
	result[0] = (values[0]-seed)*multiplier + seed

	for i := 1; i < len(values); i++ {
		result[i] = (values[i]-result[i-1])*multiplier + result[i-1]
	}

	return result
}

// WMA calculates Weighted Moving Average
func WMA(values []float64, period int) []float64 {
	if len(values) < period || period <= 0 {
		return nil
	}

	result := make([]float64, len(values)-period+1)
	weights := float64(period * (period + 1) / 2)

	for i := period - 1; i < len(values); i++ {
		var sum float64
		for j := 0; j < period; j++ {
			sum += values[i-period+1+j] * float64(j+1)
		}
		result[i-period+1] = sum / weights
	}

	return result
}

// DEMA calculates Double Exponential Moving Average
func DEMA(values []float64, period int) []float64 {
	ema1 := EMA(values, period)
	if ema1 == nil {
		return nil
	}

	ema2 := EMA(ema1, period)
	if ema2 == nil {
		return nil
	}

	// Adjust lengths
	offset := len(ema1) - len(ema2)
	result := make([]float64, len(ema2))
	for i := 0; i < len(ema2); i++ {
		result[i] = 2*ema1[i+offset] - ema2[i]
	}

	return result
}

// TEMA calculates Triple Exponential Moving Average
func TEMA(values []float64, period int) []float64 {
	ema1 := EMA(values, period)
	if ema1 == nil {
		return nil
	}

	ema2 := EMA(ema1, period)
	if ema2 == nil {
		return nil
	}

	ema3 := EMA(ema2, period)
	if ema3 == nil {
		return nil
	}

	// Adjust lengths
	offset1 := len(ema1) - len(ema3)
	offset2 := len(ema2) - len(ema3)
	result := make([]float64, len(ema3))
	for i := 0; i < len(ema3); i++ {
		result[i] = 3*ema1[i+offset1] - 3*ema2[i+offset2] + ema3[i]
	}

	return result
}

// SMMA calculates Smoothed Moving Average (used in RSI, ATR)
func SMMA(values []float64, period int) []float64 {
	if len(values) < period || period <= 0 {
		return nil
	}

	result := make([]float64, len(values)-period+1)
	result[0] = Mean(values[:period])

	for i := 1; i < len(result); i++ {
		result[i] = (result[i-1]*float64(period-1) + values[period-1+i]) / float64(period)
	}

	return result
}

// TrueRange calculates true range
func TrueRange(high, low, prevClose float64) float64 {
	tr1 := high - low
	tr2 := Abs(high - prevClose)
	tr3 := Abs(low - prevClose)
	return MaxF(tr1, MaxF(tr2, tr3))
}

// TrueRanges calculates true ranges for a series
func TrueRanges(highs, lows, closes []float64) []float64 {
	if len(highs) < 2 || len(highs) != len(lows) || len(highs) != len(closes) {
		return nil
	}

	result := make([]float64, len(highs)-1)
	for i := 1; i < len(highs); i++ {
		result[i-1] = TrueRange(highs[i], lows[i], closes[i-1])
	}
	return result
}

// RollingMax calculates rolling maximum
func RollingMax(values []float64, period int) []float64 {
	if len(values) < period || period <= 0 {
		return nil
	}

	result := make([]float64, len(values)-period+1)
	for i := period - 1; i < len(values); i++ {
		result[i-period+1] = Max(values[i-period+1 : i+1])
	}
	return result
}

// RollingMin calculates rolling minimum
func RollingMin(values []float64, period int) []float64 {
	if len(values) < period || period <= 0 {
		return nil
	}

	result := make([]float64, len(values)-period+1)
	for i := period - 1; i < len(values); i++ {
		result[i-period+1] = Min(values[i-period+1 : i+1])
	}
	return result
}

// Diff calculates the difference between consecutive values
func Diff(values []float64) []float64 {
	if len(values) < 2 {
		return nil
	}

	result := make([]float64, len(values)-1)
	for i := 1; i < len(values); i++ {
		result[i-1] = values[i] - values[i-1]
	}
	return result
}

// GainsLosses separates gains and losses from price changes
func GainsLosses(changes []float64) (gains, losses []float64) {
	gains = make([]float64, len(changes))
	losses = make([]float64, len(changes))

	for i, change := range changes {
		if change > 0 {
			gains[i] = change
		} else {
			losses[i] = -change
		}
	}
	return
}

// Percentile calculates the nth percentile
func Percentile(values []float64, percentile float64) float64 {
	if len(values) == 0 || percentile < 0 || percentile > 100 {
		return 0
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	index := (percentile / 100) * float64(len(sorted)-1)
	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))

	if lower == upper {
		return sorted[lower]
	}

	weight := index - float64(lower)
	return sorted[lower]*(1-weight) + sorted[upper]*weight
}

// ZScore calculates z-score
func ZScore(value, mean, stdDev float64) float64 {
	if stdDev == 0 {
		return 0
	}
	return (value - mean) / stdDev
}

// Normalize normalizes values to 0-1 range
func Normalize(values []float64) []float64 {
	if len(values) == 0 {
		return nil
	}

	min := Min(values)
	max := Max(values)
	rangeVal := max - min

	if rangeVal == 0 {
		result := make([]float64, len(values))
		for i := range result {
			result[i] = 0.5
		}
		return result
	}

	result := make([]float64, len(values))
	for i, v := range values {
		result[i] = (v - min) / rangeVal
	}
	return result
}

// Correlation calculates Pearson correlation coefficient
func Correlation(x, y []float64) float64 {
	if len(x) != len(y) || len(x) < 2 {
		return 0
	}

	meanX := Mean(x)
	meanY := Mean(y)

	var sumXY, sumX2, sumY2 float64
	for i := 0; i < len(x); i++ {
		dx := x[i] - meanX
		dy := y[i] - meanY
		sumXY += dx * dy
		sumX2 += dx * dx
		sumY2 += dy * dy
	}

	if sumX2 == 0 || sumY2 == 0 {
		return 0
	}

	return sumXY / math.Sqrt(sumX2*sumY2)
}

// LinearRegression calculates linear regression slope and intercept
func LinearRegression(values []float64) (slope, intercept float64) {
	n := len(values)
	if n < 2 {
		return 0, 0
	}

	var sumX, sumY, sumXY, sumX2 float64
	for i, v := range values {
		x := float64(i)
		sumX += x
		sumY += v
		sumXY += x * v
		sumX2 += x * x
	}

	nf := float64(n)
	denom := nf*sumX2 - sumX*sumX
	if denom == 0 {
		return 0, Mean(values)
	}

	slope = (nf*sumXY - sumX*sumY) / denom
	intercept = (sumY - slope*sumX) / nf
	return
}

// LinearRegressionValue calculates the regression value at a specific index
func LinearRegressionValue(values []float64, index int) float64 {
	slope, intercept := LinearRegression(values)
	return slope*float64(index) + intercept
}
