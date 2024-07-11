package dfmanager

import (
	"github.com/go-gotop/kit/exchange"
)

type DataFeedRequest struct {
	ID           string
	Symbol       string
	StartTime    int64
	EndTime      int64
	FilePath     string
	Instrument   exchange.InstrumentType
	Event        func(data *exchange.TradeEvent)
	ErrorHandler func(err error)
}

type DataFeedManager interface {
	Name() string
	AddDataFeed(req *DataFeedRequest) error
	CloseDataFeed(id string) error
	DataFeedList() []string
	Shutdown() error
}
