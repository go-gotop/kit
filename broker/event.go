package broker

import (
	"context"
)

const (
	OrderResultTopicType    string = "ORDER.RESULT"
	StrategySignalTopicType string = "STRATEGY.SIGNAL"
	StrategyStatusTopicType string = "STRATEGY.STATUS"
)

type Event interface {
	Topic() string

	Message() *Message
	RawMessage() interface{}

	Ack() error

	Error() error
}

type Handler func(ctx context.Context, evt Event) error
