package exchange

import (
	"context"
	"errors"

	"github.com/shopspring/decimal"
)

// SideType BUY, SELL
type SideType string

// OrderType LIMIT, MARKET
type OrderType string

// OrderState NEW, FILLED, CANCELED, REJECTED, EXPIRED
type OrderState string

// PositionSide LONG, SHORT
type PositionSide string

// PositionStatus OpeningPosition, HoldingPosition, ClosingPosition, ClosedPosition
type PositionStatus string

// InstrumentType SPOT，FUTURES
type InstrumentType string

// SymbolStatus SYMBOL_TRADING, SYMBOL_SUSPEND, SYMBOL_CLOSE, SYMBOL_FINISH
type SymbolStatus string

// TimeInForce GTC, IOC, FOK, GTX, GTD
type TimeInForce string

type StrategyStatus string

// Global enums
const (
	BinanceExchange  = "BINANCE"
	HuobiExchange    = "HUOBI"
	CoinBaseExchange = "COINBASE"
	MockExchange     = "MOCK"

	StrategyTypeGrid = "GRID"
	StrategyTypeDynamic = "DYNAMIC"
	
	ByMaker = "MAKER"
	ByTaker = "TAKER"

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

	StrategyStatusNew    StrategyStatus = "NEW"
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
	// 原标的物名称
	OriginalSymbol string
	// 统一标的物名称
	UnifiedSymbol string
	// 原资产名称
	OriginalAsset string
	// 统一资产名称
	UnifiedAsset string
	// 交易所
	Exchange string
	// 种类: SPOT, FUTURES
	Instrument InstrumentType
	// 状态: SYMBOL_TRADING, SYMBOL_SUSPEND, SYMBOL_CLOSE, SYMBOL_FINISH
	Status SymbolStatus
	// 最小头寸
	MinSize decimal.Decimal
	// 最大头寸
	MaxSize decimal.Decimal
	// 最小价格
	MinPrice decimal.Decimal
	// 最大价格
	MaxPrice decimal.Decimal
	// 价格精度
	PricePrecision int32
	// 头寸精度
	SizePrecision int32
}

// type Symbol struct {
// 	AssetName      string
// 	SymbolName     string
// 	Exchange       string
// 	AutoAllocation bool
// 	PricePrecision int32
// 	SizePrecision  int32
// 	MinSize        decimal.Decimal
// 	MaxSize        decimal.Decimal
// 	MinPrice       decimal.Decimal
// 	MaxPrice       decimal.Decimal
// 	Status         SymbolStatus
// 	Instrument     InstrumentType
// }

type Position struct {
	// 交易id
	TransactionID string
	// 客户端订单ID
	ClientOrderID string
	// 交易所
	Exchange string
	// 交易对
	Symbol string
	// 账户ID
	AccountID string
	// 头寸
	Size decimal.Decimal
	// 已成交数量
	ExecutedQuantity decimal.Decimal
	// 手续费
	FeeCost decimal.Decimal
	// 入场价格
	EntryPrice decimal.Decimal
	// 退出价格
	ExitPrice decimal.Decimal
	// 止损价格
	StopPrice decimal.Decimal
	// 平均价格
	AvgPrice decimal.Decimal
	// 订单状态
	State OrderState
	// 仓位状态
	Status PositionStatus
	// 种类
	Instrument InstrumentType
	// 仓位方向
	PositionSide PositionSide
}

type Order struct {
	// 帐号ID
	AccountID string
	// 客户端订单ID
	ClientOrderID string
	// 策略ID
	StrategyID string
	// 交易所
	Exchange string
	// 交易对
	Symbol string
	// 仓位侧
	Side SideType
	// 头寸
	Size decimal.Decimal
	// 价格
	Price decimal.Decimal
	// 金额
	Amount decimal.Decimal
	// 费用资产
	FeeAsset string
	// 费用
	Fee decimal.Decimal
	// 状态
	State OrderState
	// 种类
	Instrument InstrumentType
	// 仓位方向
	PositionSide PositionSide
	// 成交时间
	TransactionTime int64
}

type Exchange interface {
	Name() string
	Assets(ctx context.Context, it InstrumentType) ([]Asset, error)
	// Symbols(ctx context.Context, it InstrumentType) ([]Symbol, error)
	CreateOrder(ctx context.Context, o *CreateOrderRequest) error
	CancelOrder(ctx context.Context, o *CancelOrderRequest) error
}

