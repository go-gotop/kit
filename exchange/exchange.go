package exchange

import (
	"context"
	"errors"

	"github.com/shopspring/decimal"
)

// PosMode ISOLATED 逐仓，CROSSED 全仓
type PosMode string

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

// MarketType 市场类型
// 可扩展为：
// 1. SPOT: 现货市场
// 2. MARGIN: 杠杆/保证金市场
// 3. FUTURES_USD_MARGINED: U本位期货（如 Binance USDT-M合约）
// 4. FUTURES_COIN_MARGINED: 币本位期货（如 Binance COIN-M合约）
// 5. PERPETUAL_USD_MARGINED: U本位永续合约
// 6. PERPETUAL_COIN_MARGINED: 币本位永续合约
// 7. OPTIONS: 期权市场
// 8. LEVERAGED_TOKENS: 杠杆代币
// 9. P2P: 点对点市场
// 10. ETF: ETF类产品市场
// 11. NFT: NFT数字藏品市场
type MarketType string

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
	OkxExchange      = "OKX"
	CoinBaseExchange = "COINBASE"
	MockExchange     = "MOCK"

	AccountTypeClassic = "CLASSIC"
	AccountTypeUnified = "UNIFIED"

	StrategyTypeGrid    = "GRID"
	StrategyTypeDynamic = "DYNAMIC"

	ByMaker = "MAKER"
	ByTaker = "TAKER"

	CreatedByUser   = "USER"
	CreatedBySystem = "SYSTEM"

	TransactionByUser   = "USER"
	TransactionBySystem = "SYSTEM"

	PosModeIsolated PosMode = "ISOLATED" // 逐仓
	PosModeCross    PosMode = "CROSS"    // 全仓

	MarketTypeSpot                 MarketType = "SPOT"                   // 现货
	MarketTypeFuturesUSDMargined   MarketType = "FUTURES_USD_MARGINED"   // 期货
	MarketTypePerpetualUSDMargined MarketType = "PERPETUAL_USD_MARGINED" // 永续
	MarketTypeMargin               MarketType = "MARGIN"                 // 杠杆

	MarginTypeMargin   MarginType = "MARGIN"
	MarginTypeIsolated MarginType = "ISOLATED"

	TransactionStatusTrading TransactionStatus = "RUNNING"
	TransactionStatusSuspend TransactionStatus = "SUSPENDED"
	TransactionStatusClose   TransactionStatus = "STOPPED"
	TransactionStatusFinish  TransactionStatus = "FINISHED"

	SideTypeBuy  SideType = "BUY"
	SideTypeSell SideType = "SELL"

	OrderTypeLimit              OrderType = "LIMIT"
	OrderTypeLimitMaker         OrderType = "LIMIT_MAKER" // 限价只做Marker单
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

	PositionStatusNew     PositionStatus = "NEW"
	PositionStatusOpening PositionStatus = "OPENING"
	PositionStatusHolding PositionStatus = "HOLDING"
	PositionStatusClosing PositionStatus = "CLOSING"
	PositionStatusClosed  PositionStatus = "CLOSED"

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

type GetDepthRequest struct {
	Symbol     Symbol
	Limit      uint8
	MarketType MarketType
}

type GetDepthResponse struct {
	Asks [][]decimal.Decimal
	Bids [][]decimal.Decimal
	Ts   int64
}

type GetAccountConfigRequest struct {
	APIKey     string
	SecretKey  string
	Passphrase string
}

type GetKlineRequest struct {
	Symbol     Symbol
	Start      int64
	End        int64
	Period     string
	Limit      uint8
	MarketType MarketType
}

type GetKlineResponse struct {
	Symbol      string
	OpenTime    int64
	Open        decimal.Decimal
	High        decimal.Decimal
	Low         decimal.Decimal
	Close       decimal.Decimal
	Volume      decimal.Decimal
	QuoteVolume decimal.Decimal
	Confirm     string // 0：未完结，1：已完结
}

type GetMarkPriceKlineRequest struct {
	Symbol string
	Start  int64
	End    int64
	Period string
}

type GetMarkPriceKlineResponse struct {
	Symbol   string
	OpenTime int64
	Open     decimal.Decimal
	High     decimal.Decimal
	Low      decimal.Decimal
	Close    decimal.Decimal
	Confirm  string // 0：未完结，1：已完结
}

type GetMaxSizeRequest struct {
	APIKey     string
	SecretKey  string
	Passphrase string
	InstIds    string
	TdMode     string
	Ccy        string
	Leverage   string
}

type GetMaxSizeResponse struct {
	InstId  string
	Ccy     string
	MaxBuy  decimal.Decimal
	MaxSell decimal.Decimal
}

type GetAccountConfigResponse struct {
	Uid        string
	AcctLv     string // 账户模式
	PosMod     string // 持仓方式
	AutoBorrow bool   // 是否自动借币
}

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

type GetPositionRequest struct {
	APIKey     string
	SecretKey  string
	Passphrase string
	Symbol     string
}

type GetPositionHistoryRequest struct {
	APIKey     string
	SecretKey  string
	Passphrase string
	Start      string
	End        string
}

type SetLeverageRequest struct {
	APIKey     string
	SecretKey  string
	Passphrase string
	Mode       string
	Lever      string
	Symbol     string
	MarketType MarketType
}

type GetPositionResponse struct {
	Symbol       string
	MarketType   MarketType
	AvgPrice     decimal.Decimal // 开仓均价
	Fee          decimal.Decimal
	FundingFee   decimal.Decimal
	PositionSide PositionSide
	Size         decimal.Decimal // 仓位数量
	Upl          decimal.Decimal // 未实现盈亏
	Pnl          decimal.Decimal // 平仓订单累积收益额
	RealizedPnl  decimal.Decimal // 已实现盈亏
	Lever        string          // 杠杆倍数
	LiqPx        decimal.Decimal // 强平价格
	Margin       decimal.Decimal // 保证金率
	Liab         decimal.Decimal // 仓位的负债
	Interest     decimal.Decimal // 仓位的利息
	BePx         decimal.Decimal // 盈亏平衡价
	CreateTime   int64           // 创建时间
	UpdateTime   int64           // 更新时间

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
	APIKey     string
	SecretKey  string
	Passphrase string
	MarketType MarketType
}

type CreateOrderRequest struct {
	APIKey           string
	SecretKey        string
	Passphrase       string // 秘钥 密码 (okex)
	OrderTime        int64
	Symbol           Symbol
	CtVal            decimal.Decimal // 合约面值； 合约张数 = 合约数量 / 合约面值
	ClientOrderID    string
	Side             SideType
	OrderType        OrderType
	PositionSide     PositionSide
	TimeInForce      TimeInForce
	MarketType       MarketType
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
	APIKey        string
	SecretKey     string
	Passphrase    string
	ClientOrderID string
	MarketType    MarketType
	Symbol        Symbol
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
	APIKey     string
	SecretKey  string
	Symbol     string
	OrderID    string
	MarketType MarketType
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
	Passphrase    string
	ClientOrderID string
	Symbol        string
	MarketType    MarketType
}

type CancelOrderResponse struct {
}

// 获取标的物杠杆配置
type GetLeverageRequest struct {
	APIKey     string
	SecretKey  string
	Passphrase string
	Symbol     string
	MarketType MarketType
}

type GetLeverageResponse struct {
	Symbol     string
	Leverage   string
	MarginType string
}

type Asset struct {
	AssetName  string
	Exchange   string
	MarketType MarketType
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
	MarketType MarketType
	// 状态: ENABLED, DISABLED
	Status string
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
	// 合约面值
	CtVal decimal.Decimal
	// 合约乘数
	CtMult decimal.Decimal
	// 上线时间
	ListTime int64
	// 下线时间
	ExpTime int64
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
	// okx 合约张币转换
	ConvertContractCoin(typ string, symbol Symbol, sz string, opTyp string) (string, error)
	// 获取当前持仓
	GetPosition(ctx context.Context, req *GetPositionRequest) ([]*GetPositionResponse, error)
	// 获取历史持仓
	GetHistoryPosition(ctx context.Context, req *GetPositionHistoryRequest) error
	// 批量设置杠杠
	SetLeverage(ctx context.Context, req *SetLeverageRequest) error
	// 获取标的物杠杆配置
	GetLeverage(ctx context.Context, req *GetLeverageRequest) (GetLeverageResponse, error)
	// 获取账户配置
	GetAccountConfig(ctx context.Context, req *GetAccountConfigRequest) (GetAccountConfigResponse, error)
	// 获取最大下单量
	GetMaxSize(ctx context.Context, req *GetMaxSizeRequest) ([]GetMaxSizeResponse, error)
	// 下载标记价格K线数据
	GetMarkPriceKline(ctx context.Context, req *GetMarkPriceKlineRequest) ([]GetMarkPriceKlineResponse, error)
	// 获取K线数据
	GetKline(ctx context.Context, req *GetKlineRequest) ([]GetKlineResponse, error)
	// 获取产品深度
	GetDepth(ctx context.Context, req *GetDepthRequest) (GetDepthResponse, error)
	// 获取最新价格
	GetTickerPrice(ctx context.Context, symbol string, marketType MarketType) (decimal.Decimal, error)
}
