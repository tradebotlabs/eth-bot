package execution

import (
	"time"

	"github.com/eth-trading/internal/strategy"
)

// ExecutionMode represents trading mode
type ExecutionMode int

const (
	ModePaper ExecutionMode = iota
	ModeLive
)

func (m ExecutionMode) String() string {
	switch m {
	case ModePaper:
		return "PAPER"
	case ModeLive:
		return "LIVE"
	default:
		return "UNKNOWN"
	}
}

// OrderType represents order type
type OrderType string

const (
	OrderTypeMarket     OrderType = "MARKET"
	OrderTypeLimit      OrderType = "LIMIT"
	OrderTypeStopLoss   OrderType = "STOP_LOSS"
	OrderTypeTakeProfit OrderType = "TAKE_PROFIT"
)

// OrderSide represents order side
type OrderSide string

const (
	OrderSideBuy  OrderSide = "BUY"
	OrderSideSell OrderSide = "SELL"
)

// OrderStatus represents order status
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusOpen      OrderStatus = "OPEN"
	OrderStatusFilled    OrderStatus = "FILLED"
	OrderStatusPartial   OrderStatus = "PARTIAL"
	OrderStatusCanceled  OrderStatus = "CANCELED"
	OrderStatusRejected  OrderStatus = "REJECTED"
	OrderStatusExpired   OrderStatus = "EXPIRED"
)

// Order represents a trading order
type Order struct {
	ID              string
	ClientID        string
	Symbol          string
	Side            OrderSide
	Type            OrderType
	Quantity        float64
	Price           float64
	StopPrice       float64
	Status          OrderStatus
	FilledQuantity  float64
	AvgFillPrice    float64
	Commission      float64
	CommissionAsset string
	Strategy        string
	Signal          *strategy.Signal
	CreatedAt       time.Time
	UpdatedAt       time.Time
	FilledAt        time.Time
}

// Position represents an open position
type Position struct {
	ID               int64
	Symbol           string
	Side             PositionSide
	Quantity         float64
	EntryPrice       float64
	CurrentPrice     float64
	StopLoss         float64
	TakeProfit       float64
	UnrealizedPnL    float64
	UnrealizedPnLPct float64
	RealizedPnL      float64
	Commission       float64
	Strategy         string
	OpenTime         time.Time
	UpdatedAt        time.Time
	Orders           []string // Order IDs associated with position
}

// PositionSide represents position side
type PositionSide string

const (
	PositionSideLong  PositionSide = "LONG"
	PositionSideShort PositionSide = "SHORT"
)

// Trade represents an executed trade
type Trade struct {
	ID              string
	OrderID         string
	PositionID      int64
	Symbol          string
	Side            OrderSide
	Quantity        float64
	Price           float64
	Commission      float64
	CommissionAsset string
	RealizedPnL     float64
	Strategy        string
	ExecutedAt      time.Time
}

// ExecutionResult represents result of order execution
type ExecutionResult struct {
	Success     bool
	Order       *Order
	Trade       *Trade
	Position    *Position
	Error       error
	Message     string
	Latency     time.Duration
}

// Executor interface for order execution
type Executor interface {
	// GetMode returns execution mode
	GetMode() ExecutionMode

	// PlaceOrder places a new order
	PlaceOrder(order *Order) (*ExecutionResult, error)

	// CancelOrder cancels an existing order
	CancelOrder(orderID string) error

	// GetOrder returns order by ID
	GetOrder(orderID string) (*Order, error)

	// GetOpenOrders returns all open orders
	GetOpenOrders(symbol string) ([]*Order, error)

	// GetPosition returns position by symbol
	GetPosition(symbol string) (*Position, error)

	// GetPositions returns all open positions
	GetPositions() ([]*Position, error)

	// ClosePosition closes a position
	ClosePosition(positionID int64) (*ExecutionResult, error)

	// UpdateStopLoss updates position stop loss
	UpdateStopLoss(positionID int64, stopLoss float64) error

	// UpdateTakeProfit updates position take profit
	UpdateTakeProfit(positionID int64, takeProfit float64) error

	// GetBalance returns account balance
	GetBalance(asset string) (free, locked float64, err error)

	// GetEquity returns total equity
	GetEquity() (float64, error)

	// Sync synchronizes state with exchange (for live)
	Sync() error
}

// ExecutorConfig holds executor configuration
type ExecutorConfig struct {
	Mode              ExecutionMode
	Symbol            string

	// Paper trading
	InitialBalance    float64
	Commission        float64 // Commission rate (e.g., 0.001 = 0.1%)
	Slippage          float64 // Slippage rate

	// Live trading
	APIKey            string
	SecretKey         string
	Testnet           bool

	// General
	MaxRetries        int
	RetryDelay        time.Duration
}

// DefaultExecutorConfig returns default configuration
func DefaultExecutorConfig() *ExecutorConfig {
	return &ExecutorConfig{
		Mode:           ModePaper,
		InitialBalance: 100000,
		Commission:     0.001,
		Slippage:       0.0005,
		MaxRetries:     3,
		RetryDelay:     time.Second,
	}
}

// AccountSummary holds account summary
type AccountSummary struct {
	Mode            ExecutionMode
	Equity          float64
	AvailableBalance float64
	UsedMargin      float64
	UnrealizedPnL   float64
	RealizedPnL     float64
	TotalCommission float64
	OpenPositions   int
	TotalTrades     int
	WinRate         float64
	ProfitFactor    float64
}

// TradeStats holds trading statistics
type TradeStats struct {
	TotalTrades     int
	WinningTrades   int
	LosingTrades    int
	WinRate         float64
	AvgWin          float64
	AvgLoss         float64
	LargestWin      float64
	LargestLoss     float64
	GrossProfit     float64
	GrossLoss       float64
	NetProfit       float64
	ProfitFactor    float64
	ExpectancyRatio float64
	AvgHoldTime     time.Duration
}

// FillEvent represents an order fill event
type FillEvent struct {
	OrderID    string
	TradeID    string
	Symbol     string
	Side       OrderSide
	Quantity   float64
	Price      float64
	Commission float64
	Timestamp  time.Time
}

// PositionEvent represents a position event
type PositionEvent struct {
	Type       PositionEventType
	Position   *Position
	Trade      *Trade
	Timestamp  time.Time
}

// PositionEventType represents position event types
type PositionEventType int

const (
	PositionEventOpened PositionEventType = iota
	PositionEventClosed
	PositionEventUpdated
	PositionEventStopLossHit
	PositionEventTakeProfitHit
)

func (p PositionEventType) String() string {
	switch p {
	case PositionEventOpened:
		return "OPENED"
	case PositionEventClosed:
		return "CLOSED"
	case PositionEventUpdated:
		return "UPDATED"
	case PositionEventStopLossHit:
		return "STOP_LOSS"
	case PositionEventTakeProfitHit:
		return "TAKE_PROFIT"
	default:
		return "UNKNOWN"
	}
}
