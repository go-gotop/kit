package streammanager

import (
	"github.com/go-gotop/kit/exchange"
)

type StreamRequest struct {
	AccountId    string
	APIKey       string
	SecretKey    string
	Instrument   exchange.InstrumentType
	OrderEvent   func(evt *exchange.OrderResultEvent)
	AccountEvent func(evt []*exchange.AccountUpdateEvent)
}

type Stream struct {
	UUID       string
	AccountId  string
	APIKey     string
	Exchange   string
	Instrument exchange.InstrumentType
}

type StreamManager interface {
	Name() string
	AddStream(req *StreamRequest) (string, error)
	CloseStream(accountId string, uuid string) error
	StreamList() []Stream
	Shutdown() error
}
