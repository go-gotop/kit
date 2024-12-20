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
	Symbol       string
	Event        func(data *exchange.MarkPriceEvent)
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

type KlineMarketRequest struct {
	ID           string
	Symbol       string
	Period       string
	Instrument   exchange.InstrumentType
	Event        func(data *exchange.KlineMarketEvent)
	ErrorHandler func(err error)
}

type SymbolUpdateRequest struct {
	ID           string
	Instrument   exchange.InstrumentType
	Event        func(data []*exchange.SymbolUpdateEvent)
	ErrorHandler func(err error)
}

type Stream struct {
	UUID        string
	Instrument  exchange.InstrumentType
	DataType    string
	Symbol      string
	IsConnected bool
}

type DataFeedManager interface {
	Name() string
	AddDataFeed(req *DataFeedRequest) error
	AddMarketPriceDataFeed(req *MarkPriceRequest) error   // 全市场最新标记价格
	AddMarketKlineDataFeed(req *KlineMarketRequest) error // 全市场K线标记数据
	AddKlineDataFeed(req *KlineRequest) error
	AddSymbolUpdateDataFeed(req *SymbolUpdateRequest) error // 产品更新推送
	CloseDataFeed(id string) error
	DataFeedList() []Stream
	Shutdown() error
}
