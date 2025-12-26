package storage

import (
	"encoding/json"
	"time"
)

// Candle represents OHLCV candlestick data
type Candle struct {
	ID        int64     `db:"id" json:"id,omitempty"`
	Symbol    string    `db:"symbol" json:"symbol"`
	Timeframe string    `db:"timeframe" json:"timeframe"`
	OpenTime  time.Time `db:"open_time" json:"open_time"`
	CloseTime time.Time `db:"close_time" json:"close_time"`
	Open      float64   `db:"open" json:"open"`
	High      float64   `db:"high" json:"high"`
	Low       float64   `db:"low" json:"low"`
	Close     float64   `db:"close" json:"close"`
	Volume    float64   `db:"volume" json:"volume"`
	Trades    int       `db:"trades" json:"trades"`
	IsClosed  bool      `db:"is_closed" json:"is_closed"`
}

// NewCandle creates a new Candle with the given parameters
func NewCandle(symbol, timeframe string, openTime time.Time, open, high, low, close, volume float64) *Candle {
	return &Candle{
		Symbol:    symbol,
		Timeframe: timeframe,
		OpenTime:  openTime,
		Open:      open,
		High:      high,
		Low:       low,
		Close:     close,
		Volume:    volume,
		IsClosed:  false,
	}
}

// Update updates the candle with new tick data
func (c *Candle) Update(high, low, close, volume float64, trades int) {
	if high > c.High {
		c.High = high
	}
	if low < c.Low {
		c.Low = low
	}
	c.Close = close
	c.Volume = volume
	c.Trades = trades
}

// BodySize returns the absolute size of the candle body
func (c *Candle) BodySize() float64 {
	body := c.Close - c.Open
	if body < 0 {
		return -body
	}
	return body
}

// Range returns the full range of the candle (high - low)
func (c *Candle) Range() float64 {
	return c.High - c.Low
}

// IsBullish returns true if the candle closed higher than it opened
func (c *Candle) IsBullish() bool {
	return c.Close > c.Open
}

// IsBearish returns true if the candle closed lower than it opened
func (c *Candle) IsBearish() bool {
	return c.Close < c.Open
}

// IsDoji returns true if the candle is a doji (small body)
func (c *Candle) IsDoji(threshold float64) bool {
	if c.Range() == 0 {
		return true
	}
	return c.BodySize()/c.Range() < threshold
}

// UpperWick returns the size of the upper wick
func (c *Candle) UpperWick() float64 {
	if c.IsBullish() {
		return c.High - c.Close
	}
	return c.High - c.Open
}

// LowerWick returns the size of the lower wick
func (c *Candle) LowerWick() float64 {
	if c.IsBullish() {
		return c.Open - c.Low
	}
	return c.Close - c.Low
}

// MidPrice returns the midpoint price
func (c *Candle) MidPrice() float64 {
	return (c.High + c.Low) / 2
}

// TypicalPrice returns the typical price (HLC/3)
func (c *Candle) TypicalPrice() float64 {
	return (c.High + c.Low + c.Close) / 3
}

// VWAP returns volume-weighted average price approximation
func (c *Candle) VWAP() float64 {
	return c.TypicalPrice()
}

// TrueRange calculates true range (requires previous candle)
func (c *Candle) TrueRange(prevClose float64) float64 {
	tr1 := c.High - c.Low
	tr2 := abs(c.High - prevClose)
	tr3 := abs(c.Low - prevClose)
	return max(tr1, max(tr2, tr3))
}

// Clone creates a deep copy of the candle
func (c *Candle) Clone() *Candle {
	return &Candle{
		ID:        c.ID,
		Symbol:    c.Symbol,
		Timeframe: c.Timeframe,
		OpenTime:  c.OpenTime,
		CloseTime: c.CloseTime,
		Open:      c.Open,
		High:      c.High,
		Low:       c.Low,
		Close:     c.Close,
		Volume:    c.Volume,
		Trades:    c.Trades,
		IsClosed:  c.IsClosed,
	}
}

// ToJSON converts the candle to JSON bytes
func (c *Candle) ToJSON() ([]byte, error) {
	return json.Marshal(c)
}

// CandleFromJSON creates a candle from JSON bytes
func CandleFromJSON(data []byte) (*Candle, error) {
	var candle Candle
	if err := json.Unmarshal(data, &candle); err != nil {
		return nil, err
	}
	return &candle, nil
}

// Trade represents an executed trade
type Trade struct {
	ID              int64     `db:"id" json:"id"`
	OrderID         string    `db:"order_id" json:"order_id"`
	Symbol          string    `db:"symbol" json:"symbol"`
	Side            string    `db:"side" json:"side"`
	Type            string    `db:"type" json:"type"`
	Quantity        float64   `db:"quantity" json:"quantity"`
	Price           float64   `db:"price" json:"price"`
	Commission      float64   `db:"commission" json:"commission"`
	CommissionAsset string    `db:"commission_asset" json:"commission_asset"`
	ExecutedAt      time.Time `db:"executed_at" json:"executed_at"`
	Strategy        string    `db:"strategy" json:"strategy"`
	SignalStrength  float64   `db:"signal_strength" json:"signal_strength"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
}

// Position represents an open or closed trading position
type Position struct {
	ID            int64      `db:"id" json:"id"`
	Symbol        string     `db:"symbol" json:"symbol"`
	Side          string     `db:"side" json:"side"`
	EntryPrice    float64    `db:"entry_price" json:"entry_price"`
	Quantity      float64    `db:"quantity" json:"quantity"`
	CurrentPrice  float64    `db:"current_price" json:"current_price"`
	UnrealizedPnL float64    `db:"unrealized_pnl" json:"unrealized_pnl"`
	RealizedPnL   float64    `db:"realized_pnl" json:"realized_pnl"`
	StopLoss      float64    `db:"stop_loss" json:"stop_loss"`
	TakeProfit    float64    `db:"take_profit" json:"take_profit"`
	Strategy      string     `db:"strategy" json:"strategy"`
	Status        string     `db:"status" json:"status"`
	OpenedAt      time.Time  `db:"opened_at" json:"opened_at"`
	ClosedAt      *time.Time `db:"closed_at" json:"closed_at,omitempty"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
}

// UpdatePrice updates the position's current price and unrealized P&L
func (p *Position) UpdatePrice(price float64) {
	p.CurrentPrice = price
	if p.Side == "long" {
		p.UnrealizedPnL = (price - p.EntryPrice) * p.Quantity
	} else {
		p.UnrealizedPnL = (p.EntryPrice - price) * p.Quantity
	}
}

// IsLong returns true if this is a long position
func (p *Position) IsLong() bool {
	return p.Side == "long"
}

// IsShort returns true if this is a short position
func (p *Position) IsShort() bool {
	return p.Side == "short"
}

// IsOpen returns true if the position is still open
func (p *Position) IsOpen() bool {
	return p.Status == "open"
}

// ShouldStopLoss returns true if current price triggers stop loss
func (p *Position) ShouldStopLoss() bool {
	if p.StopLoss == 0 {
		return false
	}
	if p.IsLong() {
		return p.CurrentPrice <= p.StopLoss
	}
	return p.CurrentPrice >= p.StopLoss
}

// ShouldTakeProfit returns true if current price triggers take profit
func (p *Position) ShouldTakeProfit() bool {
	if p.TakeProfit == 0 {
		return false
	}
	if p.IsLong() {
		return p.CurrentPrice >= p.TakeProfit
	}
	return p.CurrentPrice <= p.TakeProfit
}

// Order represents a trading order
type Order struct {
	ID             int64      `db:"id" json:"id"`
	OrderID        string     `db:"order_id" json:"order_id"`
	ClientOrderID  string     `db:"client_order_id" json:"client_order_id"`
	Symbol         string     `db:"symbol" json:"symbol"`
	Side           string     `db:"side" json:"side"`
	Type           string     `db:"type" json:"type"`
	Quantity       float64    `db:"quantity" json:"quantity"`
	Price          float64    `db:"price" json:"price,omitempty"`
	StopPrice      float64    `db:"stop_price" json:"stop_price,omitempty"`
	Status         string     `db:"status" json:"status"`
	FilledQuantity float64    `db:"filled_quantity" json:"filled_quantity"`
	AvgFillPrice   float64    `db:"avg_fill_price" json:"avg_fill_price,omitempty"`
	Strategy       string     `db:"strategy" json:"strategy"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at" json:"updated_at"`
}

// AccountSnapshot represents a point-in-time account state
type AccountSnapshot struct {
	ID               int64     `db:"id" json:"id"`
	TotalEquity      float64   `db:"total_equity" json:"total_equity"`
	AvailableBalance float64   `db:"available_balance" json:"available_balance"`
	UnrealizedPnL    float64   `db:"unrealized_pnl" json:"unrealized_pnl"`
	DailyPnL         float64   `db:"daily_pnl" json:"daily_pnl"`
	OpenPositions    int       `db:"open_positions" json:"open_positions"`
	SnapshotTime     time.Time `db:"snapshot_time" json:"snapshot_time"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
}

// StrategyPerformance represents daily performance metrics for a strategy
type StrategyPerformance struct {
	ID          int64     `db:"id" json:"id"`
	Strategy    string    `db:"strategy" json:"strategy"`
	Date        time.Time `db:"date" json:"date"`
	Trades      int       `db:"trades" json:"trades"`
	Wins        int       `db:"wins" json:"wins"`
	Losses      int       `db:"losses" json:"losses"`
	GrossProfit float64   `db:"gross_profit" json:"gross_profit"`
	GrossLoss   float64   `db:"gross_loss" json:"gross_loss"`
	NetPnL      float64   `db:"net_pnl" json:"net_pnl"`
	MaxDrawdown float64   `db:"max_drawdown" json:"max_drawdown"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

// WinRate returns the win rate as a percentage
func (sp *StrategyPerformance) WinRate() float64 {
	if sp.Trades == 0 {
		return 0
	}
	return float64(sp.Wins) / float64(sp.Trades) * 100
}

// ProfitFactor returns the profit factor
func (sp *StrategyPerformance) ProfitFactor() float64 {
	if sp.GrossLoss == 0 {
		if sp.GrossProfit > 0 {
			return float64(1000) // Capped at 1000 for display
		}
		return 0
	}
	return sp.GrossProfit / abs(sp.GrossLoss)
}

// Alert represents a system alert
type Alert struct {
	ID           int64     `db:"id" json:"id"`
	Type         string    `db:"type" json:"type"`
	Severity     string    `db:"severity" json:"severity"`
	Message      string    `db:"message" json:"message"`
	Data         string    `db:"data" json:"data,omitempty"`
	Acknowledged bool      `db:"acknowledged" json:"acknowledged"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

// Helper functions
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
