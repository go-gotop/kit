package ofmanager

import (
	"github.com/go-gotop/kit/exchange"
)

type OrderFeedRequest struct {
	AccountId	 string
	APIKey     string
	SecretKey  string
	Instrument exchange.InstrumentType
	Event      func(data *exchange.OrderEvent)
}

type OrderFeedManager interface {
	Name() string
	AddOrderFeed(req *OrderFeedRequest) error
	CloseOrderFeed(id string) error
	OrderFeedList() []string
	Shutdown() error
}
