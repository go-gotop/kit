package bytime

import (
	"github.com/go-gotop/kit/exchange"
	"github.com/go-gotop/kit/sampler"
)

func NewByTime(ms int64) sampler.Sampler {
	return &millisecond{
		ms: ms,
	}
}

func timestampMod(t int64, m int64) int64 {
	return t % m
}

func toPrice(te *exchange.TradeEvent) sampler.PricePoint {
	return sampler.PricePoint{
		Timestamp: te.TradedAt,
		Price:     te.Price,
	}
}

func toAgg(te *exchange.TradeEvent, ms int64) *sampler.AggregatedTrade {
	agg := &sampler.AggregatedTrade{}
	agg.Timestamp = te.TradedAt
	agg.HighestPrice = toPrice(te)
	agg.LowestPrice = toPrice(te)
	agg.OpenPrice = toPrice(te)
	// 当前逐笔数据的时间戳减掉余数
	agg.Timestamp = te.TradedAt - timestampMod(te.TradedAt, ms)
	if te.Side == exchange.SideTypeBuy {
		agg.TotalBuyQuote = te.Price.Mul(te.Size)
		agg.TotalBuySize = te.Size
		agg.BuyCount = 1
	} else {
		agg.TotalSellQuote = te.Price.Mul(te.Size)
		agg.TotalSellSize = te.Size
		agg.SellCount = 1
	}
	return agg
}

type millisecond struct {
	ms    int64
	agg    *sampler.AggregatedTrade
}

func (m *millisecond) Sample(te *exchange.TradeEvent) (agg *sampler.AggregatedTrade) {
	if m.agg == nil {
		m.agg = toAgg(te, m.ms)
	} else {
		if te.TradedAt >= m.agg.Timestamp+m.ms {
			agg = m.agg
			m.agg = toAgg(te, m.ms)
		} else {
			m.aggregate(te)
		}
	}
	return
}

func (m *millisecond) aggregate(te *exchange.TradeEvent) {
	m.agg.ClosePrice = toPrice(te)
	if te.Side == exchange.SideTypeBuy {
		m.agg.BuyCount++
		m.agg.TotalBuyQuote = m.agg.TotalBuyQuote.Add(te.Price.Mul(te.Size))
		m.agg.TotalBuySize = m.agg.TotalBuySize.Add(te.Size)
	} else {
		m.agg.SellCount++
		m.agg.TotalSellQuote = m.agg.TotalSellQuote.Add(te.Price.Mul(te.Size))
		m.agg.TotalSellSize = m.agg.TotalSellSize.Add(te.Size)
	}
	if te.Price.GreaterThan(m.agg.HighestPrice.Price) {
		m.agg.HighestPrice = toPrice(te)
	}
	if te.Price.LessThan(m.agg.LowestPrice.Price) {
		m.agg.LowestPrice = toPrice(te)
	}
}
