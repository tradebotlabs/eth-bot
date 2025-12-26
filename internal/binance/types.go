package binance

import (
	"time"
)

// Endpoints
const (
	BaseURLSpot       = "https://api.binance.com"
	BaseURLFutures    = "https://fapi.binance.com"
	BaseURLTestnet    = "https://testnet.binance.vision"

	WSBaseURLSpot     = "wss://stream.binance.com:9443/ws"
	WSBaseURLFutures  = "wss://fstream.binance.com/ws"
	WSBaseURLTestnet  = "wss://testnet.binance.vision/ws"
)

// API Endpoints
const (
	// Market Data
	EndpointPing         = "/api/v3/ping"
	EndpointTime         = "/api/v3/time"
	EndpointExchangeInfo = "/api/v3/exchangeInfo"
	EndpointDepth        = "/api/v3/depth"
	EndpointTrades       = "/api/v3/trades"
	EndpointKlines       = "/api/v3/klines"
	EndpointTicker24hr   = "/api/v3/ticker/24hr"
	EndpointTickerPrice  = "/api/v3/ticker/price"

	// Account
	EndpointAccount      = "/api/v3/account"
	EndpointMyTrades     = "/api/v3/myTrades"

	// Orders
	EndpointOrder        = "/api/v3/order"
	EndpointOpenOrders   = "/api/v3/openOrders"
	EndpointAllOrders    = "/api/v3/allOrders"

	// User Data Stream
	EndpointUserDataStream = "/api/v3/userDataStream"
)

// OrderSide represents buy or sell
type OrderSide string

const (
	SideBuy  OrderSide = "BUY"
	SideSell OrderSide = "SELL"
)

// OrderType represents order types
type OrderType string

const (
	OrderTypeMarket          OrderType = "MARKET"
	OrderTypeLimit           OrderType = "LIMIT"
	OrderTypeStopLoss        OrderType = "STOP_LOSS"
	OrderTypeStopLossLimit   OrderType = "STOP_LOSS_LIMIT"
	OrderTypeTakeProfit      OrderType = "TAKE_PROFIT"
	OrderTypeTakeProfitLimit OrderType = "TAKE_PROFIT_LIMIT"
	OrderTypeLimitMaker      OrderType = "LIMIT_MAKER"
)

// OrderStatus represents order status
type OrderStatus string

const (
	OrderStatusNew             OrderStatus = "NEW"
	OrderStatusPartiallyFilled OrderStatus = "PARTIALLY_FILLED"
	OrderStatusFilled          OrderStatus = "FILLED"
	OrderStatusCanceled        OrderStatus = "CANCELED"
	OrderStatusPendingCancel   OrderStatus = "PENDING_CANCEL"
	OrderStatusRejected        OrderStatus = "REJECTED"
	OrderStatusExpired         OrderStatus = "EXPIRED"
)

// TimeInForce represents time in force options
type TimeInForce string

const (
	TimeInForceGTC TimeInForce = "GTC" // Good Till Cancel
	TimeInForceIOC TimeInForce = "IOC" // Immediate or Cancel
	TimeInForceFOK TimeInForce = "FOK" // Fill or Kill
)

// Kline intervals
const (
	Interval1m  = "1m"
	Interval3m  = "3m"
	Interval5m  = "5m"
	Interval15m = "15m"
	Interval30m = "30m"
	Interval1h  = "1h"
	Interval2h  = "2h"
	Interval4h  = "4h"
	Interval6h  = "6h"
	Interval8h  = "8h"
	Interval12h = "12h"
	Interval1d  = "1d"
	Interval3d  = "3d"
	Interval1w  = "1w"
	Interval1M  = "1M"
)

// ServerTime represents Binance server time response
type ServerTime struct {
	ServerTime int64 `json:"serverTime"`
}

// ExchangeInfo represents exchange information
type ExchangeInfo struct {
	Timezone   string       `json:"timezone"`
	ServerTime int64        `json:"serverTime"`
	Symbols    []SymbolInfo `json:"symbols"`
}

// SymbolInfo represents symbol trading rules
type SymbolInfo struct {
	Symbol              string        `json:"symbol"`
	Status              string        `json:"status"`
	BaseAsset           string        `json:"baseAsset"`
	BaseAssetPrecision  int           `json:"baseAssetPrecision"`
	QuoteAsset          string        `json:"quoteAsset"`
	QuoteAssetPrecision int           `json:"quoteAssetPrecision"`
	OrderTypes          []string      `json:"orderTypes"`
	Filters             []FilterInfo  `json:"filters"`

	// Parsed filter values (populated by GetSymbolInfo)
	MinPrice          float64 `json:"-"`
	MaxPrice          float64 `json:"-"`
	TickSize          float64 `json:"-"`
	MinQty            float64 `json:"-"`
	MaxQty            float64 `json:"-"`
	StepSize          float64 `json:"-"`
	MinNotional       float64 `json:"-"`
	PricePrecision    int     `json:"-"`
	QuantityPrecision int     `json:"-"`
}

// FilterInfo represents trading filters
type FilterInfo struct {
	FilterType  string `json:"filterType"`
	MinPrice    string `json:"minPrice,omitempty"`
	MaxPrice    string `json:"maxPrice,omitempty"`
	TickSize    string `json:"tickSize,omitempty"`
	MinQty      string `json:"minQty,omitempty"`
	MaxQty      string `json:"maxQty,omitempty"`
	StepSize    string `json:"stepSize,omitempty"`
	MinNotional string `json:"minNotional,omitempty"`
}

// Kline represents candlestick data
type Kline struct {
	OpenTime                 int64
	Open                     string
	High                     string
	Low                      string
	Close                    string
	Volume                   string
	CloseTime                int64
	QuoteAssetVolume         string
	NumberOfTrades           int64
	TakerBuyBaseAssetVolume  string
	TakerBuyQuoteAssetVolume string
}

// KlineEvent represents a WebSocket kline event
type KlineEvent struct {
	EventType string    `json:"e"`
	EventTime int64     `json:"E"`
	Symbol    string    `json:"s"`
	Kline     KlineData `json:"k"`
}

// KlineData represents kline data in WebSocket message
type KlineData struct {
	StartTime    int64  `json:"t"`
	CloseTime    int64  `json:"T"`
	Symbol       string `json:"s"`
	Interval     string `json:"i"`
	FirstTradeID int64  `json:"f"`
	LastTradeID  int64  `json:"L"`
	Open         string `json:"o"`
	Close        string `json:"c"`
	High         string `json:"h"`
	Low          string `json:"l"`
	Volume       string `json:"v"`
	NumberTrades int64  `json:"n"`
	IsClosed     bool   `json:"x"`
	QuoteVolume  string `json:"q"`
	TakerBuyVol  string `json:"V"`
	TakerBuyQuote string `json:"Q"`
}

// Ticker24hr represents 24hr price change statistics
type Ticker24hr struct {
	Symbol             string `json:"symbol"`
	PriceChange        string `json:"priceChange"`
	PriceChangePercent string `json:"priceChangePercent"`
	WeightedAvgPrice   string `json:"weightedAvgPrice"`
	PrevClosePrice     string `json:"prevClosePrice"`
	LastPrice          string `json:"lastPrice"`
	LastQty            string `json:"lastQty"`
	BidPrice           string `json:"bidPrice"`
	BidQty             string `json:"bidQty"`
	AskPrice           string `json:"askPrice"`
	AskQty             string `json:"askQty"`
	OpenPrice          string `json:"openPrice"`
	HighPrice          string `json:"highPrice"`
	LowPrice           string `json:"lowPrice"`
	Volume             string `json:"volume"`
	QuoteVolume        string `json:"quoteVolume"`
	OpenTime           int64  `json:"openTime"`
	CloseTime          int64  `json:"closeTime"`
	FirstID            int64  `json:"firstId"`
	LastID             int64  `json:"lastId"`
	Count              int64  `json:"count"`
}

// TickerPrice represents price ticker
type TickerPrice struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

// Depth represents order book depth
type Depth struct {
	LastUpdateID int64      `json:"lastUpdateId"`
	Bids         [][]string `json:"bids"`
	Asks         [][]string `json:"asks"`
}

// Account represents account information
type Account struct {
	MakerCommission  int64     `json:"makerCommission"`
	TakerCommission  int64     `json:"takerCommission"`
	BuyerCommission  int64     `json:"buyerCommission"`
	SellerCommission int64     `json:"sellerCommission"`
	CanTrade         bool      `json:"canTrade"`
	CanWithdraw      bool      `json:"canWithdraw"`
	CanDeposit       bool      `json:"canDeposit"`
	UpdateTime       int64     `json:"updateTime"`
	Balances         []Balance `json:"balances"`
}

// Balance represents asset balance
type Balance struct {
	Asset  string  `json:"asset"`
	Free   float64 `json:"-"`
	Locked float64 `json:"-"`
	FreeStr   string `json:"free"`
	LockedStr string `json:"locked"`
}

// SimpleTicker represents simple ticker with last price
type SimpleTicker struct {
	Symbol    string  `json:"symbol"`
	LastPrice float64 `json:"lastPrice"`
}

// ListenKeyResponse represents user data stream listen key response
type ListenKeyResponse struct {
	ListenKey string `json:"listenKey"`
}

// OrderResponse represents full order response with fills
type OrderResponse struct {
	Symbol              string      `json:"symbol"`
	OrderID             int64       `json:"orderId"`
	OrderListID         int64       `json:"orderListId"`
	ClientOrderID       string      `json:"clientOrderId"`
	TransactTime        int64       `json:"transactTime"`
	Price               float64     `json:"price,string"`
	OrigQty             float64     `json:"origQty,string"`
	ExecutedQty         float64     `json:"executedQty,string"`
	CummulativeQuoteQty float64     `json:"cummulativeQuoteQty,string"`
	Status              string      `json:"status"`
	TimeInForce         string      `json:"timeInForce"`
	Type                string      `json:"type"`
	Side                string      `json:"side"`
	Fills               []OrderFill `json:"fills"`
}

// OrderFill represents a fill in order response
type OrderFill struct {
	Price           float64 `json:"price,string"`
	Qty             float64 `json:"qty,string"`
	Commission      float64 `json:"commission,string"`
	CommissionAsset string  `json:"commissionAsset"`
	TradeID         int64   `json:"tradeId"`
}

// Order represents order response
type Order struct {
	Symbol              string      `json:"symbol"`
	OrderID             int64       `json:"orderId"`
	OrderListID         int64       `json:"orderListId"`
	ClientOrderID       string      `json:"clientOrderId"`
	TransactTime        int64       `json:"transactTime"`
	Price               string      `json:"price"`
	OrigQty             string      `json:"origQty"`
	ExecutedQty         string      `json:"executedQty"`
	CummulativeQuoteQty string      `json:"cummulativeQuoteQty"`
	Status              OrderStatus `json:"status"`
	TimeInForce         TimeInForce `json:"timeInForce"`
	Type                OrderType   `json:"type"`
	Side                OrderSide   `json:"side"`
	StopPrice           string      `json:"stopPrice,omitempty"`
	IcebergQty          string      `json:"icebergQty,omitempty"`
	Time                int64       `json:"time"`
	UpdateTime          int64       `json:"updateTime"`
	IsWorking           bool        `json:"isWorking"`
}

// Trade represents a trade execution
type Trade struct {
	ID              int64  `json:"id"`
	Symbol          string `json:"symbol"`
	OrderID         int64  `json:"orderId"`
	Price           string `json:"price"`
	Qty             string `json:"qty"`
	QuoteQty        string `json:"quoteQty"`
	Commission      string `json:"commission"`
	CommissionAsset string `json:"commissionAsset"`
	Time            int64  `json:"time"`
	IsBuyer         bool   `json:"isBuyer"`
	IsMaker         bool   `json:"isMaker"`
	IsBestMatch     bool   `json:"isBestMatch"`
}

// OrderRequest represents order creation request
type OrderRequest struct {
	Symbol           string
	Side             OrderSide
	Type             OrderType
	TimeInForce      TimeInForce
	Quantity         float64
	Price            float64
	StopPrice        float64
	NewClientOrderID string
}

// CancelOrderRequest represents order cancellation request
type CancelOrderRequest struct {
	Symbol            string
	OrderID           int64
	OrigClientOrderID string
}

// WSMessage represents generic WebSocket message
type WSMessage struct {
	Stream string          `json:"stream"`
	Data   interface{}     `json:"data"`
}

// WSSubscription represents WebSocket subscription
type WSSubscription struct {
	Method string   `json:"method"`
	Params []string `json:"params"`
	ID     int      `json:"id"`
}

// TradeEvent represents WebSocket trade event
type TradeEvent struct {
	EventType    string `json:"e"`
	EventTime    int64  `json:"E"`
	Symbol       string `json:"s"`
	TradeID      int64  `json:"t"`
	Price        string `json:"p"`
	Quantity     string `json:"q"`
	BuyerOrderID int64  `json:"b"`
	SellerOrderID int64 `json:"a"`
	TradeTime    int64  `json:"T"`
	IsBuyerMaker bool   `json:"m"`
}

// DepthEvent represents WebSocket depth event
type DepthEvent struct {
	EventType     string     `json:"e"`
	EventTime     int64      `json:"E"`
	Symbol        string     `json:"s"`
	FirstUpdateID int64      `json:"U"`
	FinalUpdateID int64      `json:"u"`
	Bids          [][]string `json:"b"`
	Asks          [][]string `json:"a"`
}

// MiniTickerEvent represents 24hr mini ticker event
type MiniTickerEvent struct {
	EventType   string `json:"e"`
	EventTime   int64  `json:"E"`
	Symbol      string `json:"s"`
	Close       string `json:"c"`
	Open        string `json:"o"`
	High        string `json:"h"`
	Low         string `json:"l"`
	BaseVolume  string `json:"v"`
	QuoteVolume string `json:"q"`
}

// UserDataEvent represents user data stream event
type UserDataEvent struct {
	EventType string `json:"e"`
	EventTime int64  `json:"E"`
}

// AccountUpdateEvent represents account update from user data stream
type AccountUpdateEvent struct {
	EventType  string           `json:"e"`
	EventTime  int64            `json:"E"`
	LastUpdate int64            `json:"u"`
	Balances   []BalanceUpdate  `json:"B"`
}

// BalanceUpdate represents balance update in account update
type BalanceUpdate struct {
	Asset  string `json:"a"`
	Free   string `json:"f"`
	Locked string `json:"l"`
}

// OrderUpdateEvent represents order update from user data stream
type OrderUpdateEvent struct {
	EventType          string      `json:"e"`
	EventTime          int64       `json:"E"`
	Symbol             string      `json:"s"`
	ClientOrderID      string      `json:"c"`
	Side               OrderSide   `json:"S"`
	OrderType          OrderType   `json:"o"`
	TimeInForce        TimeInForce `json:"f"`
	OrderQuantity      string      `json:"q"`
	OrderPrice         string      `json:"p"`
	StopPrice          string      `json:"P"`
	IcebergQty         string      `json:"F"`
	OrderListID        int64       `json:"g"`
	OrigClientOrderID  string      `json:"C"`
	ExecutionType      string      `json:"x"`
	OrderStatus        OrderStatus `json:"X"`
	RejectReason       string      `json:"r"`
	OrderID            int64       `json:"i"`
	LastExecutedQty    string      `json:"l"`
	CumFilledQty       string      `json:"z"`
	LastExecutedPrice  string      `json:"L"`
	Commission         string      `json:"n"`
	CommissionAsset    string      `json:"N"`
	TransactionTime    int64       `json:"T"`
	TradeID            int64       `json:"t"`
	IsOnBook           bool        `json:"w"`
	IsMaker            bool        `json:"m"`
	OrderCreationTime  int64       `json:"O"`
	CumQuoteQty        string      `json:"Z"`
	LastQuoteQty       string      `json:"Y"`
	QuoteOrderQty      string      `json:"Q"`
}

// APIError represents Binance API error
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
}

func (e *APIError) Error() string {
	return e.Message
}

// RateLimitInfo represents rate limit information
type RateLimitInfo struct {
	Type        string
	Interval    string
	IntervalNum int
	Limit       int
	Count       int
}

// IntervalToMilliseconds converts interval string to milliseconds
func IntervalToMilliseconds(interval string) int64 {
	switch interval {
	case Interval1m:
		return 60 * 1000
	case Interval3m:
		return 3 * 60 * 1000
	case Interval5m:
		return 5 * 60 * 1000
	case Interval15m:
		return 15 * 60 * 1000
	case Interval30m:
		return 30 * 60 * 1000
	case Interval1h:
		return 60 * 60 * 1000
	case Interval2h:
		return 2 * 60 * 60 * 1000
	case Interval4h:
		return 4 * 60 * 60 * 1000
	case Interval6h:
		return 6 * 60 * 60 * 1000
	case Interval8h:
		return 8 * 60 * 60 * 1000
	case Interval12h:
		return 12 * 60 * 60 * 1000
	case Interval1d:
		return 24 * 60 * 60 * 1000
	case Interval3d:
		return 3 * 24 * 60 * 60 * 1000
	case Interval1w:
		return 7 * 24 * 60 * 60 * 1000
	case Interval1M:
		return 30 * 24 * 60 * 60 * 1000
	default:
		return 60 * 1000
	}
}

// IntervalToDuration converts interval string to time.Duration
func IntervalToDuration(interval string) time.Duration {
	return time.Duration(IntervalToMilliseconds(interval)) * time.Millisecond
}
