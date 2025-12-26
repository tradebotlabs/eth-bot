package strategy

import (
	"sync"
	"time"

	"github.com/eth-trading/internal/indicators"
	"github.com/rs/zerolog/log"
)

// ManagerConfig holds configuration for strategy manager
type ManagerConfig struct {
	// Strategies to enable
	EnableTrendFollowing bool
	EnableMeanReversion  bool
	EnableBreakout       bool
	EnableVolatility     bool
	EnableStatArb        bool

	// Strategy configs
	TrendFollowingConfig *TrendFollowingConfig
	MeanReversionConfig  *MeanReversionConfig
	BreakoutConfig       *BreakoutConfig
	VolatilityConfig     *VolatilityConfig
	StatArbConfig        *StatArbConfig

	// Scorer config
	ScorerConfig *ScorerConfig

	// Regime config
	RegimeConfig *RegimeConfig

	// General settings
	MinDataPoints int
}

// DefaultManagerConfig returns default manager configuration
func DefaultManagerConfig() *ManagerConfig {
	return &ManagerConfig{
		EnableTrendFollowing: true,
		EnableMeanReversion:  true,
		EnableBreakout:       true,
		EnableVolatility:     true,
		EnableStatArb:        true,
		TrendFollowingConfig: DefaultTrendFollowingConfig(),
		MeanReversionConfig:  DefaultMeanReversionConfig(),
		BreakoutConfig:       DefaultBreakoutConfig(),
		VolatilityConfig:     DefaultVolatilityConfig(),
		StatArbConfig:        DefaultStatArbConfig(),
		ScorerConfig:         DefaultScorerConfig(),
		RegimeConfig:         DefaultRegimeConfig(),
		MinDataPoints:        60,
	}
}

// Manager manages all trading strategies
type Manager struct {
	config         *ManagerConfig
	indicators     *indicators.Manager
	regimeDetector *RegimeDetector
	scorer         *Scorer
	strategies     map[string]Strategy

	// State
	lastResult     *AnalysisOutput
	lastRegime     RegimeResult
	regimeHistory  *RegimeHistory

	mu sync.RWMutex
}

// NewManager creates a new strategy manager
func NewManager(config *ManagerConfig, indicatorConfig *indicators.IndicatorConfig) *Manager {
	if config == nil {
		config = DefaultManagerConfig()
	}

	indicatorManager := indicators.NewManager(indicatorConfig)

	m := &Manager{
		config:        config,
		indicators:    indicatorManager,
		scorer:        NewScorer(config.ScorerConfig),
		strategies:    make(map[string]Strategy),
		regimeHistory: NewRegimeHistory(100),
	}

	// Create regime detector
	m.regimeDetector = NewRegimeDetector(config.RegimeConfig, indicatorManager)

	// Initialize strategies
	m.initStrategies()

	return m
}

// initStrategies initializes all enabled strategies
func (m *Manager) initStrategies() {
	if m.config.EnableTrendFollowing {
		s := NewTrendFollowingStrategy(m.config.TrendFollowingConfig)
		m.strategies[s.Name()] = s
		m.scorer.AddStrategy(s)
	}

	if m.config.EnableMeanReversion {
		s := NewMeanReversionStrategy(m.config.MeanReversionConfig)
		m.strategies[s.Name()] = s
		m.scorer.AddStrategy(s)
	}

	if m.config.EnableBreakout {
		s := NewBreakoutStrategy(m.config.BreakoutConfig)
		m.strategies[s.Name()] = s
		m.scorer.AddStrategy(s)
	}

	if m.config.EnableVolatility {
		s := NewVolatilityStrategy(m.config.VolatilityConfig)
		m.strategies[s.Name()] = s
		m.scorer.AddStrategy(s)
	}

	if m.config.EnableStatArb {
		s := NewStatArbStrategy(m.config.StatArbConfig)
		m.strategies[s.Name()] = s
		m.scorer.AddStrategy(s)
	}

	log.Info().Int("count", len(m.strategies)).Msg("Strategies initialized")
}

// AnalysisOutput holds complete analysis output
type AnalysisOutput struct {
	Timestamp     time.Time
	Symbol        string
	Timeframe     string

	// Market regime
	Regime        RegimeResult

	// Indicators
	Indicators    indicators.AnalysisResult

	// Strategy scores
	Score         CombinedScore

	// Final recommendation
	Recommendation Recommendation
}

// Recommendation represents the final trading recommendation
type Recommendation struct {
	Action      Action
	Direction   Direction
	Confidence  float64
	Price       float64
	StopLoss    float64
	TakeProfit  float64
	Reason      string
	Strategy    string
}

// Action represents recommended action
type Action int

const (
	ActionNone Action = iota
	ActionBuy
	ActionSell
	ActionHold
	ActionClose
)

func (a Action) String() string {
	switch a {
	case ActionBuy:
		return "BUY"
	case ActionSell:
		return "SELL"
	case ActionHold:
		return "HOLD"
	case ActionClose:
		return "CLOSE"
	default:
		return "NONE"
	}
}

// Analyze performs complete market analysis
func (m *Manager) Analyze(symbol, timeframe string, opens, highs, lows, closes, volumes []float64, currentPrice float64) *AnalysisOutput {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(closes) < m.config.MinDataPoints {
		log.Debug().Int("available", len(closes)).Int("required", m.config.MinDataPoints).Msg("Insufficient data for analysis")
		return nil
	}

	// Create market data
	data := &MarketData{
		Symbol:       symbol,
		Timeframe:    timeframe,
		Timestamp:    time.Now(),
		Opens:        opens,
		Highs:        highs,
		Lows:         lows,
		Closes:       closes,
		Volumes:      volumes,
		CurrentPrice: currentPrice,
	}

	if data.CurrentPrice == 0 {
		data.CurrentPrice = closes[len(closes)-1]
	}

	// Get indicator analysis
	data.Analysis = m.indicators.Analyze(opens, highs, lows, closes, volumes)

	// Detect regime
	regime := m.regimeDetector.Detect(opens, highs, lows, closes, volumes)
	data.Regime = regime
	m.lastRegime = regime

	// Record regime history
	m.regimeHistory.Add(time.Now().Unix(), regime.Regime, regime.Confidence)

	// Score strategies
	score := m.scorer.Score(data, regime)

	// Generate recommendation
	recommendation := m.generateRecommendation(data, score, regime)

	output := &AnalysisOutput{
		Timestamp:      time.Now(),
		Symbol:         symbol,
		Timeframe:      timeframe,
		Regime:         regime,
		Indicators:     data.Analysis,
		Score:          score,
		Recommendation: recommendation,
	}

	m.lastResult = output
	return output
}

// generateRecommendation generates final recommendation
func (m *Manager) generateRecommendation(data *MarketData, score CombinedScore, regime RegimeResult) Recommendation {
	rec := Recommendation{
		Action:     ActionNone,
		Direction:  DirectionNone,
		Confidence: score.Confidence,
		Price:      data.CurrentPrice,
	}

	// Check if we should avoid trading
	if m.regimeDetector.ShouldAvoidTrading(regime) {
		rec.Reason = "Unfavorable market regime"
		return rec
	}

	if !score.ShouldTrade {
		rec.Reason = "No trade signal"
		return rec
	}

	// Generate action based on direction
	if score.Direction == DirectionLong {
		rec.Action = ActionBuy
		rec.Direction = DirectionLong
	} else if score.Direction == DirectionShort {
		rec.Action = ActionSell
		rec.Direction = DirectionShort
	}

	// Get stop loss and take profit from best signal
	if score.BestSignal != nil {
		rec.StopLoss = score.BestSignal.StopLoss
		rec.TakeProfit = score.BestSignal.TakeProfit
		rec.Reason = score.BestSignal.Reason
		rec.Strategy = score.BestSignal.Strategy
	}

	return rec
}

// AnalyzePosition analyzes an open position for exit signals
func (m *Manager) AnalyzePosition(position *Position, data *MarketData) (bool, string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check each strategy for exit signal
	for name, strategy := range m.strategies {
		if !strategy.IsEnabled() {
			continue
		}

		// Only check the strategy that opened the position
		if position.Strategy != "" && position.Strategy != name {
			continue
		}

		shouldExit, reason := strategy.ShouldExit(data, position)
		if shouldExit {
			return true, reason
		}
	}

	return false, ""
}

// GetLastResult returns the last analysis result
func (m *Manager) GetLastResult() *AnalysisOutput {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastResult
}

// GetLastRegime returns the last detected regime
func (m *Manager) GetLastRegime() RegimeResult {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastRegime
}

// GetStrategies returns all strategies
func (m *Manager) GetStrategies() map[string]Strategy {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]Strategy)
	for k, v := range m.strategies {
		result[k] = v
	}
	return result
}

// EnableStrategy enables a strategy
func (m *Manager) EnableStrategy(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if s, ok := m.strategies[name]; ok {
		s.SetEnabled(true)
		log.Info().Str("strategy", name).Msg("Strategy enabled")
	}
}

// DisableStrategy disables a strategy
func (m *Manager) DisableStrategy(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if s, ok := m.strategies[name]; ok {
		s.SetEnabled(false)
		log.Info().Str("strategy", name).Msg("Strategy disabled")
	}
}

// GetIndicators returns indicator manager
func (m *Manager) GetIndicators() *indicators.Manager {
	return m.indicators
}

// GetScorer returns strategy scorer
func (m *Manager) GetScorer() *Scorer {
	return m.scorer
}

// GetRegimeDetector returns regime detector
func (m *Manager) GetRegimeDetector() *RegimeDetector {
	return m.regimeDetector
}

// GetRegimeHistory returns regime history
func (m *Manager) GetRegimeHistory() *RegimeHistory {
	return m.regimeHistory
}

// StrategyStatus holds strategy status information
type StrategyStatus struct {
	Name       string
	Enabled    bool
	LastSignal *Signal
	Weight     float64
}

// GetStrategyStatuses returns status of all strategies
func (m *Manager) GetStrategyStatuses() []StrategyStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var statuses []StrategyStatus
	for name, strategy := range m.strategies {
		status := StrategyStatus{
			Name:    name,
			Enabled: strategy.IsEnabled(),
			Weight:  m.scorer.config.Weights[name],
		}
		statuses = append(statuses, status)
	}

	return statuses
}

// Summary returns a summary of current market state
type Summary struct {
	Symbol         string
	Timeframe      string
	Price          float64
	Regime         string
	RegimeConf     float64
	TrendDir       string
	TrendStrength  string
	Volatility     string
	RSI            float64
	MACD           float64
	ADX            float64
	Action         string
	Confidence     float64
	ActiveSignals  int
}

// GetSummary returns current market summary
func (m *Manager) GetSummary() *Summary {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.lastResult == nil {
		return nil
	}

	return &Summary{
		Symbol:        m.lastResult.Symbol,
		Timeframe:     m.lastResult.Timeframe,
		Price:         m.lastResult.Recommendation.Price,
		Regime:        m.lastResult.Regime.Regime.String(),
		RegimeConf:    m.lastResult.Regime.Confidence,
		TrendDir:      m.lastResult.Regime.TrendDir.String(),
		TrendStrength: m.lastResult.Regime.TrendStrength.String(),
		Volatility:    m.lastResult.Regime.Volatility.String(),
		RSI:           m.lastResult.Indicators.RSI.Value,
		MACD:          m.lastResult.Indicators.MACD.MACD,
		ADX:           m.lastResult.Indicators.ADX.ADX,
		Action:        m.lastResult.Recommendation.Action.String(),
		Confidence:    m.lastResult.Recommendation.Confidence,
		ActiveSignals: m.lastResult.Score.TotalSignals,
	}
}
