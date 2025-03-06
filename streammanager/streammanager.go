package streammanager

import (
	"github.com/go-gotop/kit/exchange"
)

type StreamRequest struct {
	AccountId        string
	APIKey           string
	SecretKey        string
	Passphrase       string
	MarketType       exchange.MarketType
	OrderEvent       func(evt *exchange.OrderResultEvent)
	AccountEvent     func(evt []*exchange.AccountUpdateEvent)
	ErrorEvent       func(evt *exchange.StreamErrorEvent) // 连接正常，用于交易所推送的异常事件回调
	ErrorHandler     func(err error)                      // 连接异常的回调
	IsUnifiedAccount bool                                 // 统一账户, 默认 false
}

type Stream struct {
	UUID       string
	AccountId  string
	APIKey     string
	Exchange   string
	MarketType exchange.MarketType
	IsConnected bool
}

type StreamManager interface {
	Name() string
	AddStream(req *StreamRequest) ([]string, error)
	CloseStream(accountId string, marketType exchange.MarketType, uuid string) error
	StreamList() []Stream
	Shutdown() error
}
