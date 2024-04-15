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
	State() exchange.SymbolStatus
	SetState(s exchange.SymbolStatus)
	SetConfig(config any) error
}
