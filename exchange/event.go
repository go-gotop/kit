package exchange

import (
	"github.com/shopspring/decimal"
)

type OrderResultEvent struct {
	// AccountID 账户ID
	AccountID string
	// ID 交易ID
	TransactionID string
	// Exchange 交易所
	Exchange string
	// PositionID 仓位ID
	PositionID string
	// ClientOrderID 自定义客户端订单号
	ClientOrderID string
	// Symbol 交易对
	Symbol string
	// OrderID 交易所订单号
	OrderID string
	// FeeAsset 手续费资产
	FeeAsset string
	// TransactionTime 交易时间
	TransactionTime int64
	// By 是否是挂单方 MAKER, TAKER
	By string
	// CreatedBy 创建者 USER，SYSTEM
	CreatedBy string
	// Instrument 种类 SPOT, FUTURES
	Instrument InstrumentType
	// Status 订单状态: OpeningPosition, HoldingPosition, ClosingPosition, ClosedPosition
	Status PositionStatus
	// ExecutionType 本次订单执行类型:NEW, TRADE, CANCELED, REJECTED, EXPIRED
	ExecutionType ExecutionState
	// State 当前订单执行类型:NEW, PARTIALLY_FILLED, FILLED, CANCELED, REJECTED, EXPIRED
	State OrderState
	// PositionSide LONG，SHORT
	PositionSide PositionSide
	// SideType BUY，SELL
	Side SideType
	// OrderType LIMIT，MARKET
	Type OrderType
	// Volume 原交易数量
	Volume decimal.Decimal
	// Price 交易价格
	Price decimal.Decimal
	// LatestVolume 最新交易数量
	LatestVolume decimal.Decimal
	// FilledVolume 已成交数量
	FilledVolume decimal.Decimal
	// LatestPrice 最新交易价格
	LatestPrice decimal.Decimal
	// FeeCost 手续费
	FeeCost decimal.Decimal
	// FilledQuoteVolume 已成交金额
	FilledQuoteVolume decimal.Decimal
	// LatestQuoteVolume 最新成交金额
	LatestQuoteVolume decimal.Decimal
	// QuoteVolume 交易金额
	QuoteVolume decimal.Decimal
	// AvgPrice 平均成交价格
	AvgPrice decimal.Decimal
}

type StrategySignalEvent struct {
	// PositionID 仓位ID
	PositionID string
	// ID 交易ID
	TransactionID string
	// AccountID 账户ID
	AccountID string
	// Timestamp 当前时间戳
	Timestamp int64
	// ClientOrderID 自定义客户端订单号
	ClientOrderID string
	// TimeInForce GTC，IOC，FOK，GTX，GTD
	TimeInForce TimeInForce
	// SideType BUY，SELL
	Side SideType
	// OrderType LIMIT，MARKET
	OrderType OrderType
	// PositionSide LONG，SHORT
	PositionSide PositionSide
	// Instrument 种类 SPOT, FUTURES, MARGIN
	Instrument InstrumentType
	// Symbol 交易对
	Symbol Symbol
	// Size 头寸数量
	Size decimal.Decimal
	// Price 交易价格
	Price decimal.Decimal
	// CreatedBy 创建者 USER, SYSTEM
	CreatedBy string
}

type StrategyStatusEvent struct {
	// AccountID 账户ID
	AccountID string
	// ID 交易ID
	TransactionID string
	Symbol        Symbol
	// Status 策略状态: NEW, START, STOP, DELETE
	Status StrategyStatus
}

// 系统内交易账户变更事件
type AccountChangeEvent struct {
	// 账户id
	AccountID string
	// 账户类型: 经典 CLASSIC, 统一 UNIFIED
	AccountType string
	// 交易种类
	Instrument []InstrumentType
	// 交易所
	Exchange string
	// api_key
	APIKey string
	// secretKey
	SecretKey string
	// passphrase OKX的
	Passphrase string
	// 删除
	DelStatus int8
}

type TradeEvent struct {
	TradedAt   int64
	TradeID    string
	Symbol     string
	Exchange   string
	Size       decimal.Decimal
	Price      decimal.Decimal
	Side       SideType
	Instrument InstrumentType
}

type MarkPriceEvent struct {
	Time                 int64
	Symbol               string
	MarkPrice            decimal.Decimal
	IndexPrice           decimal.Decimal
	EstimatedSettlePrice decimal.Decimal
	LastFundingRate      decimal.Decimal
	NextFundingTime      int64
	IsSettlement         bool // 是否结算,mockExchange专用
}

type KlineEvent struct {
	Symbol                   string
	OpenTime                 int64
	Open                     decimal.Decimal
	High                     decimal.Decimal
	Low                      decimal.Decimal
	Close                    decimal.Decimal
	Volume                   decimal.Decimal
	CloseTime                int64
	QuoteAssetVolume         decimal.Decimal
	NumberOfTrades           int64
	TakerBuyBaseAssetVolume  decimal.Decimal
	TakerBuyQuoteAssetVolume decimal.Decimal
}

type KlineMarketEvent struct {
	Symbol   string
	OpenTime int64
	Open     decimal.Decimal
	High     decimal.Decimal
	Low      decimal.Decimal
	Close    decimal.Decimal
	Confirm  string // 0 代表 K 线未完结，1 代表 K 线已完结。
}

type SymbolUpdateEvent struct {
	// 种类
	InstrumentType InstrumentType
	// 原始交易对名称
	OriginalSymbol string
	// 原始交易对资产名称
	OriginalAsset string
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
	// 状态 live:交易中，suspend:暂停交易，expired:已下线，preopen:预开放, test:测试
	State string 
}

type AccountUpdateEvent struct {
	Asset   string
	Balance decimal.Decimal
}

type FrameErrorEvent struct {
	Error         string
	PositionID    string
	TransactionID string
	AccountID     string
	Timestamp     int64
	ClientOrderID string
}

// WebSocket 流错误事件
type StreamErrorEvent struct {
	Error     error
	AccountID string
}
