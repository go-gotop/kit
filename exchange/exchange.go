package exchange

import (
	"context"
	"errors"

	"github.com/shopspring/decimal"
)

// SideType BUY, SELL
type SideType string

// STOP 止损限价单,STOP_MARKET 止损市价单,TAKE_PROFIT 止盈限价单,TAKE_PROFIT_MARKET 止盈市价单,TRAILING_STOP_MARKET 跟踪止损单
type OrderType string

// NEW, TRADE, CANCELED, REJECTED, EXPIRED
type ExecutionState string

// OrderState NEW, PARTIALLY_FILLED, FILLED, CANCELED, REJECTED, EXPIRED
type OrderState string

// PositionSide LONG, SHORT
type PositionSide string

// PositionStatus OpeningPosition, HoldingPosition, ClosingPosition, ClosedPosition
type PositionStatus string

// InstrumentType SPOT，FUTURES, MARGIN
type InstrumentType string

// MarginType MARGIN,ISOLATED 全仓，逐仓
type MarginType string

// TransactionStatus TRANSACTION_TRADING, TRANSACTION_SUSPEND, TRANSACTION_CLOSE, TRANSACTION_FINISH
type TransactionStatus string

// TimeInForce GTC, IOC, FOK, GTX, GTD
type TimeInForce string

// StrategyStatus NEW, START, STOP, DELETE
type StrategyStatus string

// StrategySide LONG, SHORT, BOTH
type StrategySide string

// Global enums
const (
	BinanceExchange  = "BINANCE"
	HuobiExchange    = "HUOBI"
	CoinBaseExchange = "COINBASE"
	MockExchange     = "MOCK"

	StrategyTypeGrid    = "GRID"
	StrategyTypeDynamic = "DYNAMIC"

	ByMaker = "MAKER"
	ByTaker = "TAKER"

	CreatedByUser   = "USER"
	CreatedBySystem = "SYSTEM"

	TransactionByUser   = "USER"
	TransactionBySystem = "SYSTEM"

	InstrumentTypeSpot    InstrumentType = "SPOT"
	InstrumentTypeFutures InstrumentType = "FUTURES"
	InstrumentTypeMargin  InstrumentType = "MARGIN"

	MarginTypeMargin   MarginType = "MARGIN"
	MarginTypeIsolated MarginType = "ISOLATED"

	TransactionStatusTrading TransactionStatus = "TRANSACTION_TRADING"
	TransactionStatusSuspend TransactionStatus = "TRANSACTION_SUSPEND"
	TransactionStatusClose   TransactionStatus = "TRANSACTION_CLOSE"
	TransactionStatusFinish  TransactionStatus = "TRANSACTION_FINISH"

	SideTypeBuy  SideType = "BUY"
	SideTypeSell SideType = "SELL"

	OrderTypeLimit              OrderType = "LIMIT"
	OrderTypeMarket             OrderType = "MARKET"
	OrderTypeStop               OrderType = "STOP"
	OrderTypeStopMarket         OrderType = "STOP_MARKET"
	OrderTypeTakeProfit         OrderType = "TAKE_PROFIT"
	OrderTypeTakeProfitMarket   OrderType = "TAKE_PROFIT_MARKET"
	OrderTypeTrailingStopMarket OrderType = "TRAILING_STOP_MARKET"

	OrderStateTrade           OrderState = "TRADE"
	OrderStateNew             OrderState = "NEW"
	OrderStateFilled          OrderState = "FILLED"
	OrderStateCanceled        OrderState = "CANCELED"
	OrderStateRejected        OrderState = "REJECTED"
	OrderStateExpired         OrderState = "EXPIRED"
	OrderStateClose           OrderState = "CLOSE"
	OrderStatePartiallyFilled OrderState = "PARTIALLY_FILLED"
	// 标识系统异常订单
	OrderStateUnusual OrderState = "UNUSUAL"

	PositionStatusNew     PositionStatus = "NEW_POSITION"
	PositionStatusOpening PositionStatus = "OPENING_POSITION"
	PositionStatusHolding PositionStatus = "HOlDING_POSITION"
	PositionStatusClosing PositionStatus = "CLOSING_POSITION"
	PositionStatusClosed  PositionStatus = "CLOSED_P0SITION"

	PositionSideLong  PositionSide = "LONG"
	PositionSideShort PositionSide = "SHORT"

	StrategySideLong  StrategySide = "LONG"
	StrategySideShort StrategySide = "SHORT"
	StrategySideBoth  StrategySide = "BOTH"

	StrategyStatusNew    StrategyStatus = "NEW"
	StrategyStatusStart  StrategyStatus = "START"
	StrategyStatusStop   StrategyStatus = "STOP"
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
	// ErrInstrumentTypeNotSupported 不支持的交易类型
	ErrInstrumentTypeNotSupported = errors.New("instrument type not supported")
	// ErrRateLimitExceeded 访问限制
	ErrRateLimitExceeded = errors.New("rate limit exceeded, IP ban imminent")
	// ErrListenKeyExpired Stream listenKey 过期（适用binance）
	ErrListenKeyExpired = errors.New("listen key expired")
)

type MarginInventoryRequest struct {
	APIKey    string
	SecretKey string
	Typ       MarginType
}

type MarginInventory struct {
	Assets map[string]string
}

type MarginBorrowOrRepayRequest struct {
	APIKey     string
	SecretKey  string
	Asset      string
	IsIsolated bool   // 是否逐仓，默认false
	Symbol     string // 逐仓交易对，配合逐仓使用
	Amount     decimal.Decimal
	Typ        string // BORROW, REPAY
}

type GetMarginInterestRateRequest struct {
	APIKey     string
	SecretKey  string
	Assets     string // 支持多资产查询，以逗号分隔，最多支持20个资产
	IsIsolated bool   // 是否逐仓
}

type GetMarginInterestRateResponse struct {
	Asset                  string
	NextHourlyInterestRate decimal.Decimal
}

type GetFundingRate struct {
	Symbol string
}

type GetFundingRateResponse struct {
	Symbol               string
	MarkPrice            decimal.Decimal // 标记价格
	IndexPrice           decimal.Decimal // 指数价格
	EstimatedSettlePrice decimal.Decimal // 预估结算价，仅在交割开始前最后一小时有意义
	LastFundingRate      decimal.Decimal // 最近更新的资金费率
	NextFundingTime      int64           // 下一个资金费时间
	InterestRate         decimal.Decimal // 标的资产基础利率
	Time                 int64           // 更新时间
}

type GetAssetsRequest struct {
	APIKey         string
	SecretKey      string
	InstrumentType InstrumentType
}

type CreateOrderRequest struct {
	APIKey           string
	SecretKey        string
	OrderTime        int64
	Symbol           string
	ClientOrderID    string
	Side             SideType
	OrderType        OrderType
	PositionSide     PositionSide
	TimeInForce      TimeInForce
	Instrument       InstrumentType
	Size             decimal.Decimal
	Price            decimal.Decimal
	IsUnifiedAccount bool // 统一账户, 默认 false
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

type SearchOrderRequest struct {
	APIKey         string
	SecretKey      string
	ClientOrderID  string
	InstrumentType InstrumentType
	Symbol         string
}

type SearchOrderResponse struct {
	ClientOrderID     string
	OrderID           string
	State             OrderState
	Symbol            string
	AvgPrice          decimal.Decimal
	Volume            decimal.Decimal
	Price             decimal.Decimal
	FilledQuoteVolume decimal.Decimal
	FilledVolume      decimal.Decimal
	FeeCost           decimal.Decimal
	FeeAsset          string
	Side              SideType
	PositionSide      PositionSide
	TimeInForce       TimeInForce
	OrderType         OrderType
	By                string
	CreatedTime       int64
	UpdateTime        int64
}

// 账户成交历史
type SearchTradesRequest struct {
	APIKey         string
	SecretKey      string
	Symbol         string
	OrderID        string
	InstrumentType InstrumentType
}

type SearchTradesResponse struct {
	Symbol   string
	ID       string
	OrderID  string
	Price    decimal.Decimal
	Volume   decimal.Decimal
	FeeCost  decimal.Decimal
	FeeAsset string
	Time     int64
	By       string
}

type CancelOrderRequest struct {
	APIKey        string
	SecretKey     string
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
	// 状态: TRANSACTION_TRADING, TRANSACTION_SUSPEND, TRANSACTION_CLOSE, TRANSACTION_FINISH
	Status TransactionStatus
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

//go:generate mockgen -destination=../exchange/mocks/exchange.go -package=mkexchange . Exchange
type Exchange interface {
	Name() string
	Assets(ctx context.Context, req *GetAssetsRequest) ([]Asset, error)
	// Symbols(ctx context.Context, it InstrumentType) ([]Symbol, error)
	CreateOrder(ctx context.Context, o *CreateOrderRequest) error
	CancelOrder(ctx context.Context, o *CancelOrderRequest) error
	SearchOrder(ctx context.Context, o *SearchOrderRequest) (*SearchOrderResponse, error)
	// 查询成交记录
	SearchTrades(ctx context.Context, o *SearchTradesRequest) ([]*SearchTradesResponse, error)
	// 获取资金费率
	GetFundingRate(ctx context.Context, req *GetFundingRate) ([]*GetFundingRateResponse, error)
	// 获取杠杠资产小时利率
	GetMarginInterestRate(ctx context.Context, req *GetMarginInterestRateRequest) ([]*GetMarginInterestRateResponse, error)
	// 杠杠借贷Or还款
	MarginBorrowOrRepay(ctx context.Context, req *MarginBorrowOrRepayRequest) error
	// 获取杠杠可用放贷库存
	GetMarginInventory(ctx context.Context, req *MarginInventoryRequest) (*MarginInventory, error)
}
