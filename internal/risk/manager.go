package risk

import (
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// Manager handles all risk management
type Manager struct {
	config        *RiskConfig
	positionSizer *PositionSizer
	state         *AccountState
	events        []RiskEvent
	mu            sync.RWMutex

	// Callbacks
	onRiskEvent func(RiskEvent)
}

// NewManager creates a new risk manager
func NewManager(config *RiskConfig) *Manager {
	if config == nil {
		config = DefaultRiskConfig()
	}

	return &Manager{
		config:        config,
		positionSizer: NewPositionSizer(config),
		state: &AccountState{
			PeakEquity: 0,
		},
		events: make([]RiskEvent, 0),
	}
}

// SetOnRiskEvent sets callback for risk events
func (m *Manager) SetOnRiskEvent(fn func(RiskEvent)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onRiskEvent = fn
}

// UpdateAccountState updates current account state
func (m *Manager) UpdateAccountState(equity, availableBalance, unrealizedPnL, dailyPnL, weeklyPnL float64, openPositions int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.state.Equity = equity
	m.state.AvailableBalance = availableBalance
	m.state.UnrealizedPnL = unrealizedPnL
	m.state.DailyPnL = dailyPnL
	m.state.WeeklyPnL = weeklyPnL
	m.state.OpenPositions = openPositions

	// Update peak equity
	if equity > m.state.PeakEquity {
		m.state.PeakEquity = equity
	}

	// Calculate drawdown
	if m.state.PeakEquity > 0 {
		m.state.CurrentDrawdown = (m.state.PeakEquity - equity) / m.state.PeakEquity
	}

	// Check for risk events
	m.checkRiskLimits()
}

// checkRiskLimits checks if any risk limits are breached
func (m *Manager) checkRiskLimits() {
	// Daily loss check
	dailyLossLimit := -m.state.PeakEquity * m.config.MaxDailyLoss
	if m.state.DailyPnL < dailyLossLimit {
		m.emitEvent(RiskEvent{
			Type:      RiskEventDailyLoss,
			Level:     RiskHigh,
			Message:   "Daily loss limit exceeded",
			Timestamp: time.Now(),
			Details: map[string]interface{}{
				"dailyPnL": m.state.DailyPnL,
				"limit":    dailyLossLimit,
			},
		})
	}

	// Weekly loss check
	weeklyLossLimit := -m.state.PeakEquity * m.config.MaxWeeklyLoss
	if m.state.WeeklyPnL < weeklyLossLimit {
		m.emitEvent(RiskEvent{
			Type:      RiskEventDailyLoss,
			Level:     RiskHigh,
			Message:   "Weekly loss limit exceeded",
			Timestamp: time.Now(),
			Details: map[string]interface{}{
				"weeklyPnL": m.state.WeeklyPnL,
				"limit":     weeklyLossLimit,
			},
		})
	}

	// Drawdown check
	if m.state.CurrentDrawdown > m.config.MaxTotalDrawdown {
		m.emitEvent(RiskEvent{
			Type:      RiskEventDrawdown,
			Level:     RiskCritical,
			Message:   "Maximum drawdown exceeded",
			Timestamp: time.Now(),
			Details: map[string]interface{}{
				"drawdown": m.state.CurrentDrawdown,
				"limit":    m.config.MaxTotalDrawdown,
			},
		})

		// Trigger circuit breaker
		if m.config.EnableCircuitBreaker {
			m.triggerCircuitBreaker("Maximum drawdown exceeded")
		}
	}
}

// emitEvent emits a risk event
func (m *Manager) emitEvent(event RiskEvent) {
	m.events = append(m.events, event)

	// Keep only last 100 events
	if len(m.events) > 100 {
		m.events = m.events[len(m.events)-100:]
	}

	log.Warn().
		Str("type", event.Type.String()).
		Str("level", event.Level.String()).
		Str("message", event.Message).
		Msg("Risk event")

	if m.onRiskEvent != nil {
		m.onRiskEvent(event)
	}
}

// AssessTrade assesses risk for a potential trade
func (m *Manager) AssessTrade(params TradeParams) RiskAssessment {
	m.mu.RLock()
	defer m.mu.RUnlock()

	log.Debug().
		Float64("stateEquity", m.state.Equity).
		Float64("entryPrice", params.EntryPrice).
		Float64("stopLoss", params.StopLoss).
		Float64("takeProfit", params.TakeProfit).
		Str("direction", params.Direction).
		Msg("AssessTrade called")

	assessment := RiskAssessment{
		Approved:  true,
		RiskLevel: RiskLow,
		Reasons:   make([]string, 0),
		Warnings:  make([]string, 0),
	}

	// Check if trading is halted
	if m.state.IsHalted {
		assessment.Approved = false
		assessment.RiskLevel = RiskCritical
		assessment.Reasons = append(assessment.Reasons, "Trading halted: "+m.state.HaltReason)
		return assessment
	}

	// Check position limits
	if m.state.OpenPositions >= m.config.MaxOpenPositions {
		assessment.Approved = false
		assessment.RiskLevel = RiskHigh
		assessment.Reasons = append(assessment.Reasons, "Maximum open positions reached")
		return assessment
	}

	// Check daily loss
	dailyLossLimit := m.state.PeakEquity * m.config.MaxDailyLoss
	if -m.state.DailyPnL >= dailyLossLimit*0.8 {
		assessment.Warnings = append(assessment.Warnings, "Approaching daily loss limit")
		assessment.RiskLevel = RiskMedium
	}
	if -m.state.DailyPnL >= dailyLossLimit {
		assessment.Approved = false
		assessment.RiskLevel = RiskHigh
		assessment.Reasons = append(assessment.Reasons, "Daily loss limit exceeded")
		return assessment
	}

	// Check drawdown
	if m.state.CurrentDrawdown >= m.config.MaxTotalDrawdown*0.8 {
		assessment.Warnings = append(assessment.Warnings, "Approaching maximum drawdown")
		assessment.RiskLevel = RiskMedium
	}
	if m.state.CurrentDrawdown >= m.config.MaxTotalDrawdown {
		assessment.Approved = false
		assessment.RiskLevel = RiskCritical
		assessment.Reasons = append(assessment.Reasons, "Maximum drawdown exceeded")
		return assessment
	}

	// Check consecutive losses
	if m.config.EnableCircuitBreaker {
		if m.state.ConsecutiveLosses >= m.config.ConsecutiveLossLimit-1 {
			assessment.Warnings = append(assessment.Warnings, "Near consecutive loss limit")
			assessment.RiskLevel = RiskMedium
		}
		if m.state.ConsecutiveLosses >= m.config.ConsecutiveLossLimit {
			assessment.Approved = false
			assessment.RiskLevel = RiskHigh
			assessment.Reasons = append(assessment.Reasons, "Consecutive loss limit reached")
			return assessment
		}
	}

	// Calculate position size
	sizeResult := m.positionSizer.CalculateSize(PositionSizeParams{
		Equity:           m.state.Equity,
		EntryPrice:       params.EntryPrice,
		StopLoss:         params.StopLoss,
		TakeProfit:       params.TakeProfit,
		Direction:        params.Direction,
		ATR:              params.ATR,
		IsHighVolatility: params.IsHighVolatility,
		SignalStrength:   params.SignalStrength,
	})

	assessment.AdjustedSize = sizeResult.Size
	assessment.StopLoss = params.StopLoss
	assessment.TakeProfit = params.TakeProfit
	assessment.RiskAmount = sizeResult.RiskAmount

	// Calculate reward
	var rewardDistance float64
	if params.Direction == "LONG" {
		rewardDistance = params.TakeProfit - params.EntryPrice
	} else {
		rewardDistance = params.EntryPrice - params.TakeProfit
	}
	assessment.RewardAmount = sizeResult.Size * rewardDistance

	// Risk/reward ratio
	if assessment.RiskAmount > 0 {
		assessment.RiskRewardRatio = assessment.RewardAmount / assessment.RiskAmount
	}

	// Check minimum R/R
	if assessment.RiskRewardRatio < m.config.MinRiskRewardRatio {
		assessment.Approved = false
		assessment.RiskLevel = RiskMedium
		assessment.Reasons = append(assessment.Reasons, "Risk/reward ratio below minimum")
		log.Warn().
			Float64("riskRewardRatio", assessment.RiskRewardRatio).
			Float64("minRequired", m.config.MinRiskRewardRatio).
			Float64("entryPrice", params.EntryPrice).
			Float64("stopLoss", params.StopLoss).
			Float64("takeProfit", params.TakeProfit).
			Float64("riskAmount", assessment.RiskAmount).
			Float64("rewardAmount", assessment.RewardAmount).
			Msg("Trade rejected: R/R ratio too low")
		return assessment
	}

	// Trading hours check
	if m.config.TradingHoursOnly {
		hour := time.Now().Hour()
		if hour < m.config.TradingStartHour || hour >= m.config.TradingEndHour {
			assessment.Approved = false
			assessment.Reasons = append(assessment.Reasons, "Outside trading hours")
			return assessment
		}
	}

	// Weekend check
	if m.config.AvoidWeekends {
		weekday := time.Now().Weekday()
		if weekday == time.Saturday || weekday == time.Sunday {
			assessment.Approved = false
			assessment.Reasons = append(assessment.Reasons, "Weekend trading disabled")
			return assessment
		}
	}

	return assessment
}

// TradeParams holds parameters for trade assessment
type TradeParams struct {
	Symbol           string
	Direction        string
	EntryPrice       float64
	StopLoss         float64
	TakeProfit       float64
	ATR              float64
	IsHighVolatility bool
	SignalStrength   float64
}

// RecordTrade records a completed trade for risk tracking
func (m *Manager) RecordTrade(metrics TradeMetrics) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.state.LastTradeTime = time.Now()

	if metrics.IsWin {
		m.state.ConsecutiveLosses = 0
	} else {
		m.state.ConsecutiveLosses++

		// Check circuit breaker
		if m.config.EnableCircuitBreaker && m.state.ConsecutiveLosses >= m.config.ConsecutiveLossLimit {
			m.triggerCircuitBreaker("Consecutive loss limit reached")
		}
	}
}

// triggerCircuitBreaker activates circuit breaker
func (m *Manager) triggerCircuitBreaker(reason string) {
	m.state.IsHalted = true
	m.state.HaltReason = reason
	m.state.HaltUntil = time.Now().Add(m.config.HaltDuration)

	m.emitEvent(RiskEvent{
		Type:      RiskEventCircuitBreaker,
		Level:     RiskCritical,
		Message:   "Circuit breaker triggered",
		Timestamp: time.Now(),
		Details: map[string]interface{}{
			"reason":    reason,
			"haltUntil": m.state.HaltUntil,
		},
	})

	log.Error().
		Str("reason", reason).
		Time("haltUntil", m.state.HaltUntil).
		Msg("Circuit breaker triggered")
}

// ResetCircuitBreaker resets the circuit breaker (manual override)
func (m *Manager) ResetCircuitBreaker() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.state.IsHalted = false
	m.state.HaltReason = ""
	m.state.HaltUntil = time.Time{}
	m.state.ConsecutiveLosses = 0

	log.Info().Msg("Circuit breaker reset")
}

// CheckCircuitBreaker checks and updates circuit breaker status
func (m *Manager) CheckCircuitBreaker() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.state.IsHalted && time.Now().After(m.state.HaltUntil) {
		m.state.IsHalted = false
		m.state.HaltReason = ""
		log.Info().Msg("Circuit breaker expired, trading resumed")
	}

	return m.state.IsHalted
}

// GetRiskLimits returns current risk limit status
func (m *Manager) GetRiskLimits() RiskLimits {
	m.mu.RLock()
	defer m.mu.RUnlock()

	limits := RiskLimits{
		DailyLossLimit:    m.state.PeakEquity * m.config.MaxDailyLoss,
		DailyLossUsed:     -m.state.DailyPnL,
		WeeklyLossLimit:   m.state.PeakEquity * m.config.MaxWeeklyLoss,
		WeeklyLossUsed:    -m.state.WeeklyPnL,
		DrawdownLimit:     m.config.MaxTotalDrawdown,
		DrawdownCurrent:   m.state.CurrentDrawdown,
		PositionsLimit:    m.config.MaxOpenPositions,
		PositionsOpen:     m.state.OpenPositions,
		IsWithinLimits:    true,
		LimitBreaches:     make([]string, 0),
	}

	// Calculate percentages
	if limits.DailyLossLimit > 0 {
		limits.DailyLossPercent = limits.DailyLossUsed / limits.DailyLossLimit
	}
	if limits.WeeklyLossLimit > 0 {
		limits.WeeklyLossPercent = limits.WeeklyLossUsed / limits.WeeklyLossLimit
	}
	if limits.DrawdownLimit > 0 {
		limits.DrawdownPercent = limits.DrawdownCurrent / limits.DrawdownLimit
	}
	if limits.PositionsLimit > 0 {
		limits.PositionsPercent = float64(limits.PositionsOpen) / float64(limits.PositionsLimit)
	}

	// Check breaches
	if limits.DailyLossPercent >= 1.0 {
		limits.IsWithinLimits = false
		limits.LimitBreaches = append(limits.LimitBreaches, "Daily loss limit exceeded")
	}
	if limits.WeeklyLossPercent >= 1.0 {
		limits.IsWithinLimits = false
		limits.LimitBreaches = append(limits.LimitBreaches, "Weekly loss limit exceeded")
	}
	if limits.DrawdownPercent >= 1.0 {
		limits.IsWithinLimits = false
		limits.LimitBreaches = append(limits.LimitBreaches, "Drawdown limit exceeded")
	}
	if limits.PositionsPercent >= 1.0 {
		limits.IsWithinLimits = false
		limits.LimitBreaches = append(limits.LimitBreaches, "Position limit reached")
	}

	return limits
}

// GetDrawdownInfo returns current drawdown information
func (m *Manager) GetDrawdownInfo() DrawdownInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	info := DrawdownInfo{
		CurrentDrawdown: m.state.CurrentDrawdown,
		MaxDrawdown:     m.config.MaxTotalDrawdown,
	}

	// Calculate recovery required
	if m.state.CurrentDrawdown > 0 {
		info.RecoveryRequired = m.state.CurrentDrawdown / (1 - m.state.CurrentDrawdown)
	}

	return info
}

// GetAccountState returns current account state
func (m *Manager) GetAccountState() AccountState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return *m.state
}

// GetRecentEvents returns recent risk events
func (m *Manager) GetRecentEvents(n int) []RiskEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if n > len(m.events) {
		n = len(m.events)
	}

	events := make([]RiskEvent, n)
	copy(events, m.events[len(m.events)-n:])
	return events
}

// GetPositionSizer returns the position sizer
func (m *Manager) GetPositionSizer() *PositionSizer {
	return m.positionSizer
}

// UpdateConfig updates risk configuration
func (m *Manager) UpdateConfig(config *RiskConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.config = config
	m.positionSizer = NewPositionSizer(config)

	log.Info().Msg("Risk configuration updated")
}

// GetConfig returns current configuration
func (m *Manager) GetConfig() *RiskConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// IsHalted returns whether trading is halted
func (m *Manager) IsHalted() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.state.IsHalted
}

// ResetDailyStats resets daily statistics (call at start of trading day)
func (m *Manager) ResetDailyStats() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.state.DailyPnL = 0
	log.Info().Msg("Daily risk stats reset")
}

// ResetWeeklyStats resets weekly statistics
func (m *Manager) ResetWeeklyStats() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.state.WeeklyPnL = 0
	m.state.ConsecutiveLosses = 0
	log.Info().Msg("Weekly risk stats reset")
}
