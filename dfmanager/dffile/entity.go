package dffile

import (
	"context"

	"github.com/go-gotop/kit/exchange"
	"github.com/shopspring/decimal"
)

type StreamRequest struct {
	Symbols       []string                               // 订阅的交易符号
	Event         func(event *exchange.TradeEvent) error // 事件处理回调
	FinishedEvent func() error                           // 数据流读取完成回调
	Ctx           context.Context                        // 上下文信息
}

type TradeData struct {
	TradeID  uint64 `csv:"trade_id,omitempty"`
	Size     string `csv:"size,omitempty"`
	Price    string `csv:"price,omitempty"`
	Side     string `csv:"side,omitempty"`
	Symbol   string `csv:"symbol,omitempty"`
	Quote    string `csv:"quote,omitempty"`
	TradedAt int64  `csv:"traded_at,omitempty"`
}

type TradeEvent struct {
	TradeID  uint64
	Size     decimal.Decimal
	Price    decimal.Decimal
	Side     string
	Symbol   string
	TradedAt int64
}
