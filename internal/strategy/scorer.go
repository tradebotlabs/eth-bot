package strategy

import (
	"math"
	"sort"
	"sync"

	"github.com/eth-trading/internal/indicators"
)

// clampConfidence ensures confidence is in the valid range [0, 1]
func clampConfidence(confidence float64) float64 {
	return math.Min(1.0, math.Max(0.0, confidence))
}

// ScorerConfig holds configuration for strategy scorer
type ScorerConfig struct {
	// Weights for each strategy
	Weights map[string]float64

	// Minimum scores
	MinScoreForEntry float64
	MinConfidence    float64

	// Conflict resolution
	ConflictMode     ConflictMode

	// Regime adjustments
	UseRegimeWeights bool
	RegimeWeights    map[MarketRegime]map[string]float64
}

// ConflictMode determines how conflicting signals are handled
type ConflictMode int

const (
	ConflictModeHighestScore ConflictMode = iota // Highest score wins
	ConflictModeConsensus                        // Require consensus
	ConflictModeNoTrade                          // No trade on conflict
	ConflictModeAverage                          // Average the signals
)

// DefaultScorerConfig returns default scorer configuration
func DefaultScorerConfig() *ScorerConfig {
	return &ScorerConfig{
		Weights: map[string]float64{
			"trend_following": 1.0,
			"mean_reversion":  1.0,
			"breakout":        1.0,
			"volatility":      0.8,
			"stat_arb":        0.8,
		},
		MinScoreForEntry: 0.5,
		MinConfidence:    0.4,
		ConflictMode:     ConflictModeHighestScore,
		UseRegimeWeights: true,
		RegimeWeights: map[MarketRegime]map[string]float64{
			RegimeTrending: {
				"trend_following": 1.5,
				"mean_reversion":  0.3,
				"breakout":        1.0,
				"volatility":      0.8,
				"stat_arb":        0.3,
			},
			RegimeMeanReverting: {
				"trend_following": 0.3,
				"mean_reversion":  1.5,
				"breakout":        0.5,
				"volatility":      0.8,
				"stat_arb":        1.2,
			},
			RegimeBreakout: {
				"trend_following": 1.0,
				"mean_reversion":  0.3,
				"breakout":        1.5,
				"volatility":      1.2,
				"stat_arb":        0.5,
			},
			RegimeHighVolatility: {
				"trend_following": 0.8,
				"mean_reversion":  0.8,
				"breakout":        1.0,
				"volatility":      1.5,
				"stat_arb":        0.5,
			},
			RegimeConsolidating: {
				"trend_following": 0.3,
				"mean_reversion":  1.2,
				"breakout":        0.5,
				"volatility":      0.5,
				"stat_arb":        1.0,
			},
		},
	}
}

// Scorer scores and combines signals from multiple strategies
type Scorer struct {
	config     *ScorerConfig
	strategies map[string]Strategy
	mu         sync.RWMutex
}

// NewScorer creates a new strategy scorer
func NewScorer(config *ScorerConfig) *Scorer {
	if config == nil {
		config = DefaultScorerConfig()
	}

	return &Scorer{
		config:     config,
		strategies: make(map[string]Strategy),
	}
}

// AddStrategy adds a strategy to the scorer
func (s *Scorer) AddStrategy(strategy Strategy) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.strategies[strategy.Name()] = strategy
}

// RemoveStrategy removes a strategy
func (s *Scorer) RemoveStrategy(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.strategies, name)
}

// Score evaluates all strategies and returns combined result
func (s *Scorer) Score(data *MarketData, regime RegimeResult) CombinedScore {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var allSignals []Signal
	strategyScores := make(map[string]ScoreResult)

	// Get signals from each strategy
	for name, strategy := range s.strategies {
		if !strategy.IsEnabled() {
			continue
		}

		signals := strategy.Analyze(data)
		if len(signals) == 0 {
			continue
		}

		// Calculate strategy weight
		weight := s.getWeight(name, regime.Regime)

		// Score each signal
		for i := range signals {
			signals[i].Strength *= weight
			signals[i].Confidence = clampConfidence(signals[i].Confidence * weight)
		}

		allSignals = append(allSignals, signals...)

		// Record strategy score
		if len(signals) > 0 {
			strategyScores[name] = ScoreResult{
				Strategy:   name,
				Score:      signals[0].Strength,
				Confidence: signals[0].Confidence,
				Direction:  signals[0].Direction,
				Signals:    signals,
			}
		}
	}

	// Combine signals
	return s.combineSignals(allSignals, strategyScores, regime)
}

// getWeight returns strategy weight adjusted for regime
func (s *Scorer) getWeight(strategyName string, regime MarketRegime) float64 {
	baseWeight := 1.0
	if w, ok := s.config.Weights[strategyName]; ok {
		baseWeight = w
	}

	if !s.config.UseRegimeWeights {
		return baseWeight
	}

	if regimeWeights, ok := s.config.RegimeWeights[regime]; ok {
		if w, ok := regimeWeights[strategyName]; ok {
			return baseWeight * w
		}
	}

	return baseWeight
}

// CombinedScore holds the combined scoring result
type CombinedScore struct {
	// Final decision
	ShouldTrade   bool
	Direction     Direction
	Score         float64
	Confidence    float64

	// Best signal
	BestSignal    *Signal

	// Strategy breakdown
	Scores        map[string]ScoreResult

	// Signal counts
	LongSignals   int
	ShortSignals  int
	TotalSignals  int

	// Conflict info
	HasConflict   bool
	ConflictLevel float64

	// Regime
	Regime        MarketRegime
}

// combineSignals combines signals based on configuration
func (s *Scorer) combineSignals(signals []Signal, scores map[string]ScoreResult, regime RegimeResult) CombinedScore {
	result := CombinedScore{
		Scores: scores,
		Regime: regime.Regime,
	}

	if len(signals) == 0 {
		return result
	}

	// Count signals by direction
	var longScore, shortScore float64
	for _, sig := range signals {
		result.TotalSignals++
		if sig.Direction == DirectionLong {
			result.LongSignals++
			longScore += sig.Strength
		} else if sig.Direction == DirectionShort {
			result.ShortSignals++
			shortScore += sig.Strength
		}
	}

	// Check for conflict
	if result.LongSignals > 0 && result.ShortSignals > 0 {
		result.HasConflict = true
		total := longScore + shortScore
		if total > 0 {
			result.ConflictLevel = 1 - (indicators.Abs(longScore-shortScore) / total)
		}
	}

	// Apply conflict mode
	switch s.config.ConflictMode {
	case ConflictModeHighestScore:
		result = s.resolveByHighestScore(signals, result)
	case ConflictModeConsensus:
		result = s.resolveByConsensus(signals, result)
	case ConflictModeNoTrade:
		if result.HasConflict {
			return result
		}
		result = s.resolveByHighestScore(signals, result)
	case ConflictModeAverage:
		result = s.resolveByAverage(signals, result)
	}

	// Check minimum thresholds
	if result.Score < s.config.MinScoreForEntry {
		result.ShouldTrade = false
	}
	if result.Confidence < s.config.MinConfidence {
		result.ShouldTrade = false
	}

	return result
}

// resolveByHighestScore selects the highest scoring signal
func (s *Scorer) resolveByHighestScore(signals []Signal, result CombinedScore) CombinedScore {
	// Sort by strength
	sort.Slice(signals, func(i, j int) bool {
		return signals[i].Strength > signals[j].Strength
	})

	best := signals[0]
	result.BestSignal = &best
	result.Direction = best.Direction
	result.Score = best.Strength
	result.Confidence = clampConfidence(best.Confidence)
	result.ShouldTrade = best.Strength >= s.config.MinScoreForEntry

	return result
}

// resolveByConsensus requires majority agreement
func (s *Scorer) resolveByConsensus(signals []Signal, result CombinedScore) CombinedScore {
	// Need >50% agreement on direction
	if result.LongSignals > result.ShortSignals*2 {
		result.Direction = DirectionLong
		result.ShouldTrade = true
	} else if result.ShortSignals > result.LongSignals*2 {
		result.Direction = DirectionShort
		result.ShouldTrade = true
	} else {
		// No consensus
		return result
	}

	// Calculate average score for winning direction
	var totalScore, count float64
	for _, sig := range signals {
		if sig.Direction == result.Direction {
			totalScore += sig.Strength
			count++
		}
	}

	if count > 0 {
		result.Score = totalScore / count
		result.Confidence = clampConfidence(count / float64(len(signals)))
	}

	// Find best signal in winning direction
	for i := range signals {
		if signals[i].Direction == result.Direction {
			if result.BestSignal == nil || signals[i].Strength > result.BestSignal.Strength {
				result.BestSignal = &signals[i]
			}
		}
	}

	return result
}

// resolveByAverage averages all signals
func (s *Scorer) resolveByAverage(signals []Signal, result CombinedScore) CombinedScore {
	var netScore float64 // Positive = long, negative = short
	var totalConfidence float64

	for _, sig := range signals {
		if sig.Direction == DirectionLong {
			netScore += sig.Strength
		} else if sig.Direction == DirectionShort {
			netScore -= sig.Strength
		}
		totalConfidence += sig.Confidence
	}

	result.Score = indicators.Abs(netScore) / float64(len(signals))
	result.Confidence = clampConfidence(totalConfidence / float64(len(signals)))

	if netScore > 0 {
		result.Direction = DirectionLong
	} else if netScore < 0 {
		result.Direction = DirectionShort
	}

	result.ShouldTrade = result.Score >= s.config.MinScoreForEntry

	// Find best signal in winning direction
	for i := range signals {
		if signals[i].Direction == result.Direction {
			if result.BestSignal == nil || signals[i].Strength > result.BestSignal.Strength {
				result.BestSignal = &signals[i]
			}
		}
	}

	return result
}

// GetStrategies returns all registered strategies
func (s *Scorer) GetStrategies() map[string]Strategy {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]Strategy)
	for k, v := range s.strategies {
		result[k] = v
	}
	return result
}

// SetStrategyEnabled enables/disables a strategy
func (s *Scorer) SetStrategyEnabled(name string, enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if strategy, ok := s.strategies[name]; ok {
		strategy.SetEnabled(enabled)
	}
}

// UpdateWeights updates strategy weights
func (s *Scorer) UpdateWeights(weights map[string]float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config.Weights = weights
}

// GetConfig returns scorer configuration
func (s *Scorer) GetConfig() *ScorerConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

// SetConfig updates scorer configuration
func (s *Scorer) SetConfig(config *ScorerConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = config
}
