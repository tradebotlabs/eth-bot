package storage

import (
	"fmt"
	"sync"
)

// QueueManager manages multiple candle queues for different symbol/timeframe combinations
type QueueManager struct {
	queues           map[string]*CandleQueue
	defaultCapacity  int
	capacities       map[string]int
	mu               sync.RWMutex
}

// NewQueueManager creates a new queue manager
func NewQueueManager(defaultCapacity int, capacities map[string]int) *QueueManager {
	if defaultCapacity <= 0 {
		defaultCapacity = 200
	}
	if capacities == nil {
		capacities = make(map[string]int)
	}
	return &QueueManager{
		queues:          make(map[string]*CandleQueue),
		defaultCapacity: defaultCapacity,
		capacities:      capacities,
	}
}

// makeKey creates a unique key for symbol/timeframe combination
func makeKey(symbol, timeframe string) string {
	return fmt.Sprintf("%s_%s", symbol, timeframe)
}

// GetOrCreate returns existing queue or creates a new one
func (qm *QueueManager) GetOrCreate(symbol, timeframe string) *CandleQueue {
	key := makeKey(symbol, timeframe)

	qm.mu.Lock()
	defer qm.mu.Unlock()

	if queue, exists := qm.queues[key]; exists {
		return queue
	}

	// Determine capacity for this timeframe
	capacity := qm.defaultCapacity
	if cap, ok := qm.capacities[timeframe]; ok {
		capacity = cap
	}

	queue := NewCandleQueue(capacity)
	qm.queues[key] = queue
	return queue
}

// Get returns a queue if it exists
func (qm *QueueManager) Get(symbol, timeframe string) (*CandleQueue, bool) {
	key := makeKey(symbol, timeframe)

	qm.mu.RLock()
	defer qm.mu.RUnlock()

	queue, exists := qm.queues[key]
	return queue, exists
}

// GetAll returns all queues
func (qm *QueueManager) GetAll() map[string]*CandleQueue {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	result := make(map[string]*CandleQueue, len(qm.queues))
	for k, v := range qm.queues {
		result[k] = v
	}
	return result
}

// Remove removes a queue for the given symbol/timeframe
func (qm *QueueManager) Remove(symbol, timeframe string) bool {
	key := makeKey(symbol, timeframe)

	qm.mu.Lock()
	defer qm.mu.Unlock()

	if _, exists := qm.queues[key]; exists {
		delete(qm.queues, key)
		return true
	}
	return false
}

// Clear removes all queues
func (qm *QueueManager) Clear() {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	qm.queues = make(map[string]*CandleQueue)
}

// ClearQueue clears a specific queue without removing it
func (qm *QueueManager) ClearQueue(symbol, timeframe string) bool {
	queue, exists := qm.Get(symbol, timeframe)
	if !exists {
		return false
	}
	queue.Clear()
	return true
}

// Count returns the number of managed queues
func (qm *QueueManager) Count() int {
	qm.mu.RLock()
	defer qm.mu.RUnlock()
	return len(qm.queues)
}

// Exists checks if a queue exists for the given symbol/timeframe
func (qm *QueueManager) Exists(symbol, timeframe string) bool {
	_, exists := qm.Get(symbol, timeframe)
	return exists
}

// SetCapacity sets the capacity for a specific timeframe
func (qm *QueueManager) SetCapacity(timeframe string, capacity int) {
	qm.mu.Lock()
	defer qm.mu.Unlock()
	qm.capacities[timeframe] = capacity
}

// GetCapacity returns the capacity for a timeframe
func (qm *QueueManager) GetCapacity(timeframe string) int {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	if cap, ok := qm.capacities[timeframe]; ok {
		return cap
	}
	return qm.defaultCapacity
}

// AddCandle adds a candle to the appropriate queue
func (qm *QueueManager) AddCandle(candle Candle) {
	queue := qm.GetOrCreate(candle.Symbol, candle.Timeframe)
	queue.Push(candle)
}

// UpdateLatestCandle updates the latest candle in the appropriate queue
func (qm *QueueManager) UpdateLatestCandle(candle Candle) bool {
	queue, exists := qm.Get(candle.Symbol, candle.Timeframe)
	if !exists {
		return false
	}

	// Check if we should update or push
	latest, hasLatest := queue.GetLatest()
	if hasLatest && latest.OpenTime.Equal(candle.OpenTime) {
		return queue.UpdateLatest(candle)
	}

	// New candle, push it
	queue.Push(candle)
	return true
}

// GetCandles returns candles for a symbol/timeframe
func (qm *QueueManager) GetCandles(symbol, timeframe string) []Candle {
	queue, exists := qm.Get(symbol, timeframe)
	if !exists {
		return nil
	}
	return queue.GetAll()
}

// GetLastCandles returns the last N candles for a symbol/timeframe
func (qm *QueueManager) GetLastCandles(symbol, timeframe string, n int) []Candle {
	queue, exists := qm.Get(symbol, timeframe)
	if !exists {
		return nil
	}
	return queue.GetLast(n)
}

// GetLatestCandle returns the latest candle for a symbol/timeframe
func (qm *QueueManager) GetLatestCandle(symbol, timeframe string) (Candle, bool) {
	queue, exists := qm.Get(symbol, timeframe)
	if !exists {
		return Candle{}, false
	}
	return queue.GetLatest()
}

// GetOHLCV returns OHLCV data for a symbol/timeframe
func (qm *QueueManager) GetOHLCV(symbol, timeframe string) (opens, highs, lows, closes, volumes []float64) {
	queue, exists := qm.Get(symbol, timeframe)
	if !exists {
		return nil, nil, nil, nil, nil
	}
	return queue.GetOHLCV()
}

// HasEnoughData checks if a queue has enough data for indicators
func (qm *QueueManager) HasEnoughData(symbol, timeframe string, n int) bool {
	queue, exists := qm.Get(symbol, timeframe)
	if !exists {
		return false
	}
	return queue.HasEnoughData(n)
}

// GetStats returns statistics for all queues
func (qm *QueueManager) GetStats() map[string]QueueStats {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	stats := make(map[string]QueueStats, len(qm.queues))
	for key, queue := range qm.queues {
		stats[key] = queue.GetStats()
	}
	return stats
}

// QueueInfo holds information about a managed queue
type QueueInfo struct {
	Symbol    string
	Timeframe string
	Size      int
	Capacity  int
	IsFull    bool
}

// GetInfo returns information about all managed queues
func (qm *QueueManager) GetInfo() []QueueInfo {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	info := make([]QueueInfo, 0, len(qm.queues))
	for key, queue := range qm.queues {
		// Parse key back to symbol and timeframe
		var symbol, timeframe string
		fmt.Sscanf(key, "%s_%s", &symbol, &timeframe)

		info = append(info, QueueInfo{
			Symbol:    symbol,
			Timeframe: timeframe,
			Size:      queue.Size(),
			Capacity:  queue.Capacity(),
			IsFull:    queue.IsFull(),
		})
	}
	return info
}

// DefaultCapacities returns recommended capacities by timeframe
var DefaultCapacities = map[string]int{
	"1m":  500, // ~8 hours of 1m candles
	"5m":  300, // ~25 hours of 5m candles
	"15m": 200, // ~50 hours of 15m candles
	"30m": 200, // ~4 days of 30m candles
	"1h":  200, // ~8 days of 1h candles
	"4h":  150, // ~25 days of 4h candles
	"1d":  100, // ~100 days of daily candles
	"1w":  52,  // ~1 year of weekly candles
}

// NewQueueManagerWithDefaults creates a queue manager with default capacities
func NewQueueManagerWithDefaults() *QueueManager {
	return NewQueueManager(200, DefaultCapacities)
}
