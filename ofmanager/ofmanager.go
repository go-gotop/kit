package ofmanager

import (
	"github.com/go-gotop/kit/exchange"
)

type OrderFeedRequest struct {
	Instrument exchange.InstrumentType
	Event 	func(data *exchange.OrderEvent)
}

type OrderFeedManager interface {
	Name() string
	Close() error
	AddOrderFeed(req *OrderFeedRequest) error
}