package sampler

import (
	"github.com/shopspring/decimal"
)

type Sampler interface {
	Close() error
	Run() error
}

type AggregatedTrade struct {
	SellCount      uint64
	BuyCount       uint64
	Timestamp      int64
	HighestPrice   PricePoint
	LowestPrice    PricePoint
	CurrentPrice   PricePoint
	TotalBuySize   decimal.Decimal
	TotalSellSize  decimal.Decimal
	TotalBuyQuote  decimal.Decimal
	TotalSellQuote decimal.Decimal
}

func (a *AggregatedTrade) Difference() decimal.Decimal {
	head, tail := a.PriceRange()
	return tail.Price.Add(tail.Price.Sub(head.Price))
}

// PriceRange returns the prices at the highest and lowest points.
func (a *AggregatedTrade) PriceRange() (head PricePoint, tail PricePoint) {
	head = a.HighestPrice
	tail = a.LowestPrice
	if a.HighestPrice.Timestamp - a.LowestPrice.Timestamp > 0 {
		head = a.LowestPrice
		tail = a.HighestPrice
	}
	return
}

type Handler func(*AggregatedTrade)

type Middleware func(Handler) Handler

type PricePoint struct {
	Price     decimal.Decimal
	Timestamp int64
}

func Chain(m ...Middleware) Middleware {
	return func(next Handler) Handler {
		for i := len(m) - 1; i >= 0; i-- {
			next = m[i](next)
		}
		return next
	}
}