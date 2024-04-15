package excmanager

import (
	"github.com/go-gotop/kit/exchange"
	"github.com/go-gotop/kit/websocket"
)

type DataFeedRequest struct {
	ID             string
	Symbol         string
	Instrument     exchange.InstrumentType
	Event func(data *exchange.TradeEvent)
	ErrorHandler   func(err error)
}

type OrderFeedRequest struct {
	Instrument exchange.InstrumentType
	Event 	func(data *exchange.OrderEvent)
}

// ExchangeWsManager 是 交易所websocket 管理接口
type ExchangeWsManager interface {
	DataFeed(req *DataFeedRequest) (string, error)
	OrderFeed(req *OrderFeedRequest) (string, error)
	CloseWeSocket(uniq string) error
	GetWebsocket(uniq string) websocket.Websocket
	IsConnected(uniq string) bool
	Shutdown()
}
