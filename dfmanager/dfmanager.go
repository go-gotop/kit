package dfmanager

import (
	"github.com/go-gotop/kit/exchange"
)

type DataFeedRequest struct {
	ID           string
	Symbol       string
	StartTime    int64
	EndTime      int64
	Instrument   exchange.InstrumentType
	Event        func(data *exchange.TradeEvent)
	ErrorHandler func(err error)
}

type MarkPriceRequest struct {
	ID           string
	Instrument   exchange.InstrumentType
	StartTime    int64
	EndTime      int64
	Event        func(data []*exchange.MarkPriceEvent)
	ErrorHandler func(err error)
}

type KlineRequest struct {
	ID           string
	Symbol       string
	Period       string
	StartTime    int64
	EndTime      int64
	Instrument   exchange.InstrumentType
	Event        func(data *exchange.KlineEvent)
	ErrorHandler func(err error)
}

type DataFeedManager interface {
	Name() string
	AddDataFeed(req *DataFeedRequest) error
	AddMarketPriceDataFeed(req *MarkPriceRequest) error // 全市场最新标记价格
	AddKlineDataFeed(req *KlineRequest) error
	CloseDataFeed(id string) error
	DataFeedList() []string
	Shutdown() error
}
