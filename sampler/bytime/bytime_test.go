package bytime

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/go-gotop/kit/exchange"
	"github.com/shopspring/decimal"
)

func TestNewByTime(t *testing.T) {
	ms := int64(1000)
	s := NewByTime(ms)
	assert.NotNil(t, s)
}

func TestTimestampMod(t *testing.T) {
	tt := []struct {
		t int64
		m int64
	}{
		{
			t: 1000,
			m: 100,
		},
		{
			t: 1000,
			m: 1000,
		},
		{
			t: 1000,
			m: 10000,
		},
	}
	for _, tc := range tt {
		assert.Equal(t, tc.t%tc.m, timestampMod(tc.t, tc.m))
	}
}

func TestToPrice(t *testing.T) {
	te := &exchange.TradeEvent{
		TradedAt: 1000,
		TradeID:    "100",
		Symbol:     "btcusdt",
		Exchange:   "binance",
		Size:	decimal.NewFromInt(1),
		Price:	decimal.NewFromInt(1000),
		Side:	exchange.SideTypeBuy,
		Instrument: exchange.InstrumentTypeSpot,
	}
	pp := toPrice(te)
	assert.Equal(t, te.TradedAt, pp.Timestamp)
	assert.Equal(t, te.Price, pp.Price)
}