package ofmanager

import (
	"github.com/go-gotop/kit/exchange"
)

type OrderFeedRequest struct {
	AccountId  string
	APIKey     string
	SecretKey  string
	Instrument exchange.InstrumentType
	Event      func(evt *exchange.OrderResultEvent)
}

type AccountFeedRequest struct {
	AccountId  string
	APIKey     string
	SecretKey  string
	Instrument exchange.InstrumentType
	Event      func(evt *exchange.OrderResultEvent)
}

type OrderFeed struct {
	UUID       string
	AccountId  string
	APIKey     string
	Exchange   string
	Instrument exchange.InstrumentType
}

type OrderFeedManager interface {
	Name() string
	AddOrderFeed(req *OrderFeedRequest) (string, error)
	CloseOrderFeed(accountId string, uuid string) error
	OrderFeedList() []OrderFeed
	Shutdown() error
}
