package exchange

import (
	"github.com/shopspring/decimal"
)

type CreateOrderEvent struct {
	ID             string
	Timestamp      int64
	ClientOrderID  string
	Symbol         string
	Side           SideType
	OrderType      OrderType
	PositionSide   PositionSide
	QuoteOrderSize decimal.Decimal
	Size           decimal.Decimal
	Price          decimal.Decimal
}

type OrderResultEvent struct {
	// ID 事件ID
	ID                string
	// Exchange 交易所
	Exchange          string
	// ClientOrderID 自定义客户端订单号
	ClientOrderID     string
	// Symbol 交易对
	Symbol            string
	// OrderID 交易所订单号
	OrderID           string
	// FeeAsset 手续费资产
	FeeAsset          string
	// TransactionTime 交易时间
	TransactionTime   int64
	// IsTaker 是否是挂单方
	IsMaker           bool
	// Instrument 种类
	Instrument        InstrumentType
	// ExecutionType 本次订单执行类型:NEW，FILLED，CANCELED，REJECTED，EXPIRED
	ExecutionType	  OrderState
	// State 当前订单执行类型:NEW，FILLED，CANCELED，REJECTED，EXPIRED
	State             OrderState
	// PositionStatus 仓位状态: OpenPosition，HoldingPosition，ClosingPosition，PositionClosed
	Status            PositionStatus
	// PositionSide LONG，SHORT
	PositionSide      PositionSide
	// SideType BUY，SELL
	Side              SideType
	// OrderType LIMIT，MARKET
	Type              OrderType
	// Volume 原交易数量
	Volume            decimal.Decimal
	// Price 交易价格
	Price             decimal.Decimal
	// LatestVolume 最新交易数量
	LatestVolume      decimal.Decimal
	// FilledVolume 已成交数量
	FilledVolume      decimal.Decimal
	// LatestPrice 最新交易价格
	LatestPrice       decimal.Decimal
	// FeeCost 手续费
	FeeCost           decimal.Decimal
	// FilledQuoteVolume 已成交金额
	FilledQuoteVolume decimal.Decimal
	// LatestQuoteVolume 最新成交金额
	LatestQuoteVolume decimal.Decimal
	// QuoteVolume 交易金额
	QuoteVolume       decimal.Decimal
	// AvgPrice 平均成交价格
	AvgPrice          decimal.Decimal
}

type StrategySignalEvent struct {
	// ID 策略ID
	ID            string
	// AccountID 账户ID
	AccountID     string
	// Timestamp 当前时间戳
	Timestamp     int64
	// ClientOrderID 自定义客户端订单号
	ClientOrderID string
	// TimeInForce GTC，IOC，FOK，GTX，GTD
	TimeInForce   TimeInForce
	// SideType BUY，SELL
	Side          SideType
	// OrderType LIMIT，MARKET
	OrderType     OrderType
	// PositionSide LONG，SHORT
	PositionSide  PositionSide
	// Symbol 交易对
	Symbol        Symbol
	// Size 头寸数量
	Size          decimal.Decimal
	// Price 交易价格
	Price         decimal.Decimal
}

type StrategyStatusEvent struct {
	ID     string
	Symbol Symbol
	Status StrategyStatus
}

type TradeEvent struct {
	TradedAt int64
	TradeID  string
	Symbol   string
	Exchange string
	Size     decimal.Decimal
	Price    decimal.Decimal
	Side     SideType
	Instrument InstrumentType
}