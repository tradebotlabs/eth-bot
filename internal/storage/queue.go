package storage

import (
	"sync"
	"time"
)

// CandleQueue is a thread-safe circular queue for storing candles
type CandleQueue struct {
	buffer   []Candle
	capacity int
	head     int // Points to oldest element
	tail     int // Points to next write position
	size     int // Current number of elements
	mu       sync.RWMutex
}

// NewCandleQueue creates a new circular queue with the given capacity
func NewCandleQueue(capacity int) *CandleQueue {
	if capacity <= 0 {
		capacity = 200 // default capacity
	}
	return &CandleQueue{
		buffer:   make([]Candle, capacity),
		capacity: capacity,
		head:     0,
		tail:     0,
		size:     0,
	}
}

// Push adds a candle to the queue (overwrites oldest if full)
func (q *CandleQueue) Push(candle Candle) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.buffer[q.tail] = candle
	q.tail = (q.tail + 1) % q.capacity

	if q.size < q.capacity {
		q.size++
	} else {
		// Queue is full, move head forward (discard oldest)
		q.head = (q.head + 1) % q.capacity
	}
}

// GetAll returns all candles in order (oldest to newest)
func (q *CandleQueue) GetAll() []Candle {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if q.size == 0 {
		return nil
	}

	result := make([]Candle, q.size)
	for i := 0; i < q.size; i++ {
		idx := (q.head + i) % q.capacity
		result[i] = q.buffer[idx]
	}
	return result
}

// GetLast returns the last N candles (oldest to newest)
func (q *CandleQueue) GetLast(n int) []Candle {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if q.size == 0 {
		return nil
	}

	if n > q.size {
		n = q.size
	}

	result := make([]Candle, n)
	startIdx := q.size - n
	for i := 0; i < n; i++ {
		idx := (q.head + startIdx + i) % q.capacity
		result[i] = q.buffer[idx]
	}
	return result
}

// GetLatest returns the most recent candle
func (q *CandleQueue) GetLatest() (Candle, bool) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if q.size == 0 {
		return Candle{}, false
	}

	idx := (q.tail - 1 + q.capacity) % q.capacity
	return q.buffer[idx], true
}

// GetOldest returns the oldest candle in the queue
func (q *CandleQueue) GetOldest() (Candle, bool) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if q.size == 0 {
		return Candle{}, false
	}

	return q.buffer[q.head], true
}

// GetAt returns the candle at the given index (0 = oldest)
func (q *CandleQueue) GetAt(index int) (Candle, bool) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if index < 0 || index >= q.size {
		return Candle{}, false
	}

	idx := (q.head + index) % q.capacity
	return q.buffer[idx], true
}

// GetFromEnd returns the candle at the given index from the end (0 = newest)
func (q *CandleQueue) GetFromEnd(index int) (Candle, bool) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if index < 0 || index >= q.size {
		return Candle{}, false
	}

	idx := (q.tail - 1 - index + q.capacity) % q.capacity
	return q.buffer[idx], true
}

// UpdateLatest updates the most recent candle (for live candle updates)
func (q *CandleQueue) UpdateLatest(candle Candle) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.size == 0 {
		return false
	}

	idx := (q.tail - 1 + q.capacity) % q.capacity
	q.buffer[idx] = candle
	return true
}

// Size returns the current number of elements
func (q *CandleQueue) Size() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.size
}

// Capacity returns the maximum capacity
func (q *CandleQueue) Capacity() int {
	return q.capacity
}

// IsFull returns true if the queue is at capacity
func (q *CandleQueue) IsFull() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.size == q.capacity
}

// IsEmpty returns true if the queue has no elements
func (q *CandleQueue) IsEmpty() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.size == 0
}

// Clear removes all elements from the queue
func (q *CandleQueue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.head = 0
	q.tail = 0
	q.size = 0
}

// GetWindow returns candles within a time window
func (q *CandleQueue) GetWindow(from, to time.Time) []Candle {
	q.mu.RLock()
	defer q.mu.RUnlock()

	var result []Candle
	for i := 0; i < q.size; i++ {
		idx := (q.head + i) % q.capacity
		candle := q.buffer[idx]
		if (candle.OpenTime.Equal(from) || candle.OpenTime.After(from)) &&
			(candle.OpenTime.Equal(to) || candle.OpenTime.Before(to)) {
			result = append(result, candle)
		}
	}
	return result
}

// GetSince returns all candles since the given time
func (q *CandleQueue) GetSince(since time.Time) []Candle {
	q.mu.RLock()
	defer q.mu.RUnlock()

	var result []Candle
	for i := 0; i < q.size; i++ {
		idx := (q.head + i) % q.capacity
		candle := q.buffer[idx]
		if candle.OpenTime.Equal(since) || candle.OpenTime.After(since) {
			result = append(result, candle)
		}
	}
	return result
}

// GetCloses returns just the close prices (oldest to newest)
func (q *CandleQueue) GetCloses() []float64 {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if q.size == 0 {
		return nil
	}

	closes := make([]float64, q.size)
	for i := 0; i < q.size; i++ {
		idx := (q.head + i) % q.capacity
		closes[i] = q.buffer[idx].Close
	}
	return closes
}

// GetOpens returns just the open prices (oldest to newest)
func (q *CandleQueue) GetOpens() []float64 {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if q.size == 0 {
		return nil
	}

	opens := make([]float64, q.size)
	for i := 0; i < q.size; i++ {
		idx := (q.head + i) % q.capacity
		opens[i] = q.buffer[idx].Open
	}
	return opens
}

// GetHighs returns just the high prices (oldest to newest)
func (q *CandleQueue) GetHighs() []float64 {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if q.size == 0 {
		return nil
	}

	highs := make([]float64, q.size)
	for i := 0; i < q.size; i++ {
		idx := (q.head + i) % q.capacity
		highs[i] = q.buffer[idx].High
	}
	return highs
}

// GetLows returns just the low prices (oldest to newest)
func (q *CandleQueue) GetLows() []float64 {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if q.size == 0 {
		return nil
	}

	lows := make([]float64, q.size)
	for i := 0; i < q.size; i++ {
		idx := (q.head + i) % q.capacity
		lows[i] = q.buffer[idx].Low
	}
	return lows
}

// GetVolumes returns just the volumes (oldest to newest)
func (q *CandleQueue) GetVolumes() []float64 {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if q.size == 0 {
		return nil
	}

	volumes := make([]float64, q.size)
	for i := 0; i < q.size; i++ {
		idx := (q.head + i) % q.capacity
		volumes[i] = q.buffer[idx].Volume
	}
	return volumes
}

// GetOHLCV returns all price and volume arrays at once
func (q *CandleQueue) GetOHLCV() (opens, highs, lows, closes, volumes []float64) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if q.size == 0 {
		return nil, nil, nil, nil, nil
	}

	opens = make([]float64, q.size)
	highs = make([]float64, q.size)
	lows = make([]float64, q.size)
	closes = make([]float64, q.size)
	volumes = make([]float64, q.size)

	for i := 0; i < q.size; i++ {
		idx := (q.head + i) % q.capacity
		c := q.buffer[idx]
		opens[i] = c.Open
		highs[i] = c.High
		lows[i] = c.Low
		closes[i] = c.Close
		volumes[i] = c.Volume
	}
	return
}

// GetLastNCloses returns the last N close prices (oldest to newest)
func (q *CandleQueue) GetLastNCloses(n int) []float64 {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if q.size == 0 {
		return nil
	}

	if n > q.size {
		n = q.size
	}

	closes := make([]float64, n)
	startIdx := q.size - n
	for i := 0; i < n; i++ {
		idx := (q.head + startIdx + i) % q.capacity
		closes[i] = q.buffer[idx].Close
	}
	return closes
}

// GetTypicalPrices returns typical prices (HLC/3) for all candles
func (q *CandleQueue) GetTypicalPrices() []float64 {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if q.size == 0 {
		return nil
	}

	prices := make([]float64, q.size)
	for i := 0; i < q.size; i++ {
		idx := (q.head + i) % q.capacity
		c := q.buffer[idx]
		prices[i] = (c.High + c.Low + c.Close) / 3
	}
	return prices
}

// GetTrueRanges calculates true range for all candles
func (q *CandleQueue) GetTrueRanges() []float64 {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if q.size < 2 {
		return nil
	}

	ranges := make([]float64, q.size-1)
	for i := 1; i < q.size; i++ {
		prevIdx := (q.head + i - 1) % q.capacity
		currIdx := (q.head + i) % q.capacity
		prevClose := q.buffer[prevIdx].Close
		curr := q.buffer[currIdx]
		ranges[i-1] = curr.TrueRange(prevClose)
	}
	return ranges
}

// HasEnoughData returns true if queue has at least n candles
func (q *CandleQueue) HasEnoughData(n int) bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.size >= n
}

// FindByTime finds a candle by its open time
func (q *CandleQueue) FindByTime(openTime time.Time) (Candle, bool) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	for i := 0; i < q.size; i++ {
		idx := (q.head + i) % q.capacity
		if q.buffer[idx].OpenTime.Equal(openTime) {
			return q.buffer[idx], true
		}
	}
	return Candle{}, false
}

// ForEach iterates over all candles (oldest to newest) with a callback
func (q *CandleQueue) ForEach(fn func(index int, candle Candle) bool) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	for i := 0; i < q.size; i++ {
		idx := (q.head + i) % q.capacity
		if !fn(i, q.buffer[idx]) {
			break
		}
	}
}

// Stats returns basic statistics about the candles in the queue
type QueueStats struct {
	Count      int
	AvgClose   float64
	HighestHigh float64
	LowestLow   float64
	TotalVolume float64
	TimeSpan   time.Duration
}

// GetStats calculates basic statistics for the queue
func (q *CandleQueue) GetStats() QueueStats {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if q.size == 0 {
		return QueueStats{}
	}

	var stats QueueStats
	stats.Count = q.size
	stats.HighestHigh = q.buffer[q.head].High
	stats.LowestLow = q.buffer[q.head].Low

	var sumClose float64
	for i := 0; i < q.size; i++ {
		idx := (q.head + i) % q.capacity
		c := q.buffer[idx]
		sumClose += c.Close
		stats.TotalVolume += c.Volume
		if c.High > stats.HighestHigh {
			stats.HighestHigh = c.High
		}
		if c.Low < stats.LowestLow {
			stats.LowestLow = c.Low
		}
	}
	stats.AvgClose = sumClose / float64(q.size)

	// Calculate time span
	oldest := q.buffer[q.head]
	newest := q.buffer[(q.tail-1+q.capacity)%q.capacity]
	stats.TimeSpan = newest.CloseTime.Sub(oldest.OpenTime)

	return stats
}
