package strategy

import (
	"github.com/go-gotop/kit/exchange"
)

type Strategy interface {
	ID() string
	Start() error
	Type() string
	ErrHandler(err error)
	Next(data *exchange.TradeEvent) error
	State() exchange.TransactionStatus
	SetState(s exchange.TransactionStatus)
	SetConfig(config any) error
}
