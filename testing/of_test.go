package testing

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-gotop/kit/exchange"
	"github.com/go-gotop/kit/limiter/bnlimiter"
	"github.com/go-gotop/kit/ofmanager"
	"github.com/go-gotop/kit/ofmanager/ofbinance"
	"github.com/go-gotop/kit/requests/bnhttp"
)

const (
	APIKey    = "5gFpsxaWLRk0GG5NJY5DvcCcWboVlkG14WzPxu6d12BgyoMRxNwvNXIN5wIfTGPd"
	SecretKey = "FUhJMPMFeaORvOaHqN2bBhWC1Qshf5ssZTL4v471BuIs66N0bSF4qWCpK00KfjQ4"
)

func orderResultEvent(evt *exchange.OrderResultEvent) {
	fmt.Printf("OrderResultEvent: %v\n", evt)
}

func newHttpClient() *bnhttp.Client {
	return bnhttp.NewClient()
}

func TestNewOwf(t *testing.T) {
	limiter := bnlimiter.NewBinanceLimiter(newRedis())

	of := ofbinance.NewBinanceOrderFeed(newHttpClient(), limiter)

	err := of.AddOrderFeed(&ofmanager.OrderFeedRequest{
		AccountId:  "123456",
		APIKey:     APIKey,
		SecretKey:  SecretKey,
		Instrument: exchange.InstrumentTypeSpot,
		Event:      orderResultEvent,
	})

	if err != nil {
		t.Error(err)
	}

	time.Sleep(20 * time.Minute)
}
