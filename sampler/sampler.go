package sampler

import (
	"github.com/shopspring/decimal"
	"github.com/go-gotop/kit/exchange"
)

type PricePoint struct {
	Timestamp int64
	Price     decimal.Decimal
}

type AggregatedTrade struct {
	SellCount      uint64
	BuyCount       uint64
	Timestamp      int64
	OpenPrice      PricePoint
	ClosePrice     PricePoint
	HighestPrice   PricePoint
	LowestPrice    PricePoint
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
	if a.HighestPrice.Timestamp > a.LowestPrice.Timestamp {
		head = a.LowestPrice
		tail = a.HighestPrice
	}
	return
}

func (a *AggregatedTrade) IsUp() bool {
	head, tail := a.PriceRange()
	return tail.Price.GreaterThan(head.Price)
}

func (a *AggregatedTrade) Equal() bool {
	return a.HighestPrice.Price.Equal(a.LowestPrice.Price)
}

// Sampler is the interface that wraps the basic Sample method.
type Sampler interface {
	// Sample samples the data feed and returns the aggregated trade data
	Sample(te *exchange.TradeEvent) *AggregatedTrade
}
