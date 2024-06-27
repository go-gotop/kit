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
	Status        StrategyStatus
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
