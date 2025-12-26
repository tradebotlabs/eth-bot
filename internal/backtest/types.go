package backtest

import (
	"time"

	"github.com/eth-trading/internal/strategy"
)

// Candle represents a single OHLCV candle
type Candle struct {
	Timestamp time.Time
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
}

// HistoricalData holds historical candle data
type HistoricalData struct {
	Symbol    string
	Timeframe string
	Candles   []Candle
}

// Position represents an open position in backtest
type Position struct {
	ID         int64
	Symbol     string
	Strategy   string
	Direction  strategy.Direction
	EntryPrice float64
	EntryTime  time.Time
	Quantity   float64
	StopLoss   float64
	TakeProfit float64
	Commission float64
}

// Trade represents a completed trade
type Trade struct {
	ID            int64
	Symbol        string
	Strategy      string
	Direction     string
	EntryTime     time.Time
	ExitTime      time.Time
	EntryPrice    float64
	ExitPrice     float64
	Quantity      float64
	NetProfit     float64
	ReturnPercent float64
	ExitReason    string
	Commission    float64
}

// EquityPoint represents a point on the equity curve
type EquityPoint struct {
	Timestamp time.Time
	Equity    float64
	Cash      float64
	Drawdown  float64
}

// Metrics holds backtest performance metrics
type Metrics struct {
	TotalReturn      float64
	AnnualizedReturn float64
	SharpeRatio      float64
	SortinoRatio     float64
	CalmarRatio      float64
	MaxDrawdown      float64
	TotalTrades      int
	WinningTrades    int
	LosingTrades     int
	WinRate          float64
	ProfitFactor     float64
	AvgWin           float64
	AvgLoss          float64
	LargestWin       float64
	LargestLoss      float64
	AvgHoldingTime   string
	Expectancy       float64
	RecoveryFactor   float64
	StartingCapital  float64
	EndingCapital    float64
	NetProfit        float64
}

// StrategyStats holds per-strategy statistics
type StrategyStats struct {
	Name         string
	TotalTrades  int
	WinRate      float64
	ProfitFactor float64
	NetProfit    float64
	Contribution float64
}

// Result holds complete backtest results
type Result struct {
	Config         *Config
	Metrics        *Metrics
	EquityCurve    []EquityPoint
	Trades         []Trade
	MonthlyReturns map[string]float64
	StrategyStats  map[string]StrategyStats
	StartTime      time.Time
	EndTime        time.Time
	ExecutionTime  time.Duration
}

// Portfolio represents the trading portfolio during backtest
type Portfolio struct {
	Cash          float64
	Positions     []*Position
	PeakEquity    float64
	InitialEquity float64
}

// NewPortfolio creates a new portfolio
func NewPortfolio(initialCash float64) *Portfolio {
	return &Portfolio{
		Cash:          initialCash,
		Positions:     []*Position{},
		PeakEquity:    initialCash,
		InitialEquity: initialCash,
	}
}

// GetEquity returns total portfolio equity
func (p *Portfolio) GetEquity() float64 {
	equity := p.Cash
	for range p.Positions {
		// For simplicity, not calculating unrealized P&L here
		// In a full implementation, would mark positions to market
	}
	return equity
}

// GetDrawdown returns current drawdown
func (p *Portfolio) GetDrawdown() float64 {
	equity := p.GetEquity()
	if equity > p.PeakEquity {
		p.PeakEquity = equity
	}
	
	if p.PeakEquity == 0 {
		return 0
	}
	
	return (p.PeakEquity - equity) / p.PeakEquity
}

// UpdatePrice updates portfolio with current market price
func (p *Portfolio) UpdatePrice(price float64) {
	// Update unrealized P&L for open positions
	// In a full implementation would update position values
}

// OpenPosition opens a new position
func (p *Portfolio) OpenPosition(pos *Position, cost float64) {
	p.Cash -= cost
	p.Positions = append(p.Positions, pos)
}

// ClosePosition closes a position by ID
func (p *Portfolio) ClosePosition(id int64) {
	for i, pos := range p.Positions {
		if pos.ID == id {
			p.Positions = append(p.Positions[:i], p.Positions[i+1:]...)
			return
		}
	}
}
