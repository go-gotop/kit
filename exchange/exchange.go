package exchange

import (
	"context"
	"errors"

	"github.com/shopspring/decimal"
)

// SideType BUY，SELL
type SideType string

// OrderType LIMIT，MARKET
type OrderType string

// OrderState NEW，FILLED，CANCELED，REJECTED，EXPIRED
type OrderState string

// PositionSide LONG，SHORT
type PositionSide string

// PositionStatus OpenPosition，HoldingPosition，ClosingPosition，PositionClosed
type PositionStatus string

// InstrumentType SPOT，FUTURES
type InstrumentType string

// SymbolStatus SYMBOL_TRADING，SYMBOL_SUSPEND，SYMBOL_CLOSE，SYMBOL_FINISH
type SymbolStatus string

// TimeInForce GTC，IOC，FOK，GTX，GTD
type TimeInForce string

type StrategyStatus string

// Global enums
const (
	BinanceExchange  = "BINANCE"
	HuobiExchange    = "HUOBI"
	CoinBaseExchange = "COINBASE"
	MockExchange     = "MOCK"

	InstrumentTypeSpot    InstrumentType = "SPOT"
	InstrumentTypeFutures InstrumentType = "FUTURES"

	SymbolStatusTrading SymbolStatus = "SYMBOL_TRADING"
	SymbolStatusSuspend SymbolStatus = "SYMBOL_SUSPEND"
	SymbolStatusClose   SymbolStatus = "SYMBOL_CLOSE"
	SymbolStatusFinish  SymbolStatus = "SYMBOL_FINISH"

	SideTypeBuy  SideType = "BUY"
	SideTypeSell SideType = "SELL"

	OrderTypeLimit  OrderType = "LIMIT"
	OrderTypeMarket OrderType = "MARKET"

	OrderStateTrade           OrderState = "TRADE"
	OrderStateNew             OrderState = "NEW"
	OrderStateFilled          OrderState = "FILLED"
	OrderStateCanceled        OrderState = "CANCELED"
	OrderStateRejected        OrderState = "REJECTED"
	OrderStateExpired         OrderState = "EXPIRED"
	OrderStateClose           OrderState = "CLOSE"
	OrderStatePartiallyFilled OrderState = "PARTIALLY_FILLED"

	PositionStatusNew     PositionStatus = "NEW_POSITION"
	PositionStatusOpening PositionStatus = "OPENING_POSITION"
	PositionStatusHolding PositionStatus = "HOlDING_POSITION"
	PositionStatusClosing PositionStatus = "CLOSING_POSITION"
	PositionStatusClosed  PositionStatus = "CLOSED_P0SITION"

	PositionSideLong  PositionSide = "LONG"
	PositionSideShort PositionSide = "SHORT"

	StrategyStatusNew   StrategyStatus = "NEW"
	StrategyStatusDelete StrategyStatus = "DELETE"

	// Good Till Cancel 成交为止, 一直有效直到被取消
	TimeInForceGTC TimeInForce = "GTC"
	// Immediate or Cancel 无法立即成交(吃单)的部分就撤销
	TimeInForceIOC TimeInForce = "IOC"
	// Fill or Kill 无法全部立即成交就撤销
	TimeInForceFOK TimeInForce = "FOK"
	// GTX - Good Till Crossing 无法成为挂单方就撤销
	TimeInForceGTX TimeInForce = "GTX"
	// GTD - Good Till Date 在特定时间之前有效，到期自动撤销
	TimeInForceGTD TimeInForce = "GTD"
)

var (
	// ErrOrderNotFound 订单未找到
	ErrOrderNotFound = errors.New("order not found")
	// ErrOrderAlreadyExists 订单已存在
	ErrOrderAlreadyExists = errors.New("order already exists")
	// ErrOrderNotEnoughBalance 余额不足
	ErrOrderNotEnoughBalance = errors.New("order not enough balance")
	// ErrOrderNotEnoughMargin 保证金不足
	ErrOrderNotEnoughMargin = errors.New("order not enough margin")
	// ErrCreateOrderLimitExceeded 下单限制
	ErrCreateOrderLimitExceeded = errors.New("create order limit exceeded")
	// ErrRateLimitExceeded 访问限制
	ErrRateLimitExceeded = errors.New("rate limit exceeded, IP ban imminent")
)

type CreateOrderRequest struct {
	OrderTime     int64
	Symbol        string
	ClientOrderID string
	Side          SideType
	OrderType     OrderType
	PositionSide  PositionSide
	TimeInForce   TimeInForce
	Instrument    InstrumentType
	Size          decimal.Decimal
	Price         decimal.Decimal
}

type CreateOrderResponse struct {
	TransactTime     int64
	Symbol           string
	ClientOrderID    string
	OrderID          string
	Side             SideType
	State            OrderState
	PositionSide     PositionSide
	Price            decimal.Decimal
	OriginalQuantity decimal.Decimal
	ExecutedQuantity decimal.Decimal
}

type CancelOrderRequest struct {
	ClientOrderID string
	Symbol        string
}

type CancelOrderResponse struct {
}

type Asset struct {
	AssetName  string
	Exchange   string
	Instrument InstrumentType
	Free       decimal.Decimal
	Locked     decimal.Decimal
}

type Symbol struct {
	AssetName      string
	SymbolName     string
	Exchange       string
	AutoAllocation bool
	PricePrecision int32
	SizePrecision  int32
	MinSize        decimal.Decimal
	MaxSize        decimal.Decimal
	MinPrice       decimal.Decimal
	MaxPrice       decimal.Decimal
	Status         SymbolStatus
	Instrument     InstrumentType
}

type TradeEvent struct {
	TradeID  uint64
	Symbol   string
	TradedAt int64
	Side     bool
	Size     decimal.Decimal
	Price    decimal.Decimal
}

type OrderEvent struct {
	Exchange        string
	Symbol          string
	ClientOrderID   string
	ExecutionType   string
	Status          string
	OrderID         string
	FeeAsset        string
	TransactionTime int64
	IsMaker         bool
	Side            SideType
	Type            OrderType
	Instrument      InstrumentType
	PositionSide    PositionSide
	Volume          decimal.Decimal
	Price           decimal.Decimal
	LatestVolume    decimal.Decimal
	FilledVolume    decimal.Decimal
	LatestPrice     decimal.Decimal
	FeeCost         decimal.Decimal
	AvgPrice        decimal.Decimal
}

type Position struct {
	ClientOrderID string
	Status        PositionStatus
	// FILLED, PARTIALLY_FILLED, CANCELED
	State            OrderState
	Price            decimal.Decimal
	OriginalQuantity decimal.Decimal
	ExecutedQuantity decimal.Decimal
	FeeCost          decimal.Decimal
}

type Exchange interface {
	Name() string
	Assets(ctx context.Context, it InstrumentType) ([]Asset, error)
	Symbols(ctx context.Context, it InstrumentType) ([]Symbol, error)
	CreateOrder(ctx context.Context, o *CreateOrderRequest) error
	CancelOrder(ctx context.Context, o *CancelOrderRequest) error
}
