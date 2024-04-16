package broker

import (
	"context"

	"github.com/go-gotop/kit/exchange"
	"github.com/shopspring/decimal"
)

const (
	OrderResultTopicType    string = "ORDER.RESULT"
	StrategySignalTopicType string = "STRATEGY.SIGNAL"
	StrategyStatusTopicType string = "STRATEGY.STATUS"
)

type Event interface {
	Topic() string

	Message() *Message
	RawMessage() interface{}

	Ack() error

	Error() error
}

type Handler func(ctx context.Context, evt Event) error

type CreateOrderEvent struct {
	Timestamp      int64
	ClientOrderID  string
	Symbol         string
	Side           exchange.SideType
	OrderType      exchange.OrderType
	PositionSide   exchange.PositionSide
	QuoteOrderSize decimal.Decimal
	Size           decimal.Decimal
	Price          decimal.Decimal
}

type OrderResultEvent struct {
	ClientOrderID     string
	OrderID           string
	FeeAsset          string
	TransactionTime   int64
	IsMaker           bool
	ExecutionType     exchange.OrderState
	State             exchange.OrderState
	Status            exchange.PositionStatus
	PositionSide      exchange.PositionSide
	Side              exchange.SideType
	Type              exchange.OrderType
	Volume            decimal.Decimal
	Price             decimal.Decimal
	LatestVolume      decimal.Decimal
	FilledVolume      decimal.Decimal
	LatestPrice       decimal.Decimal
	FeeCost           decimal.Decimal
	FilledQuoteVolume decimal.Decimal
	LatestQuoteVolume decimal.Decimal
	QuoteVolume       decimal.Decimal
	AvgPrice          decimal.Decimal
}

type StrategySignalEvent struct {
	Timestamp     int64
	ClientOrderID string
	TimeInForce   exchange.TimeInForce
	Side          exchange.SideType
	OrderType     exchange.OrderType
	PositionSide  exchange.PositionSide
	Symbol        exchange.Symbol
	Size          decimal.Decimal
	Price         decimal.Decimal
}

type StrategyStatusEvent struct {
	Symbol exchange.Symbol
	Status string
}
