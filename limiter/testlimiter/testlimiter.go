package testlimiter

import (
	"github.com/go-gotop/kit/limiter"
	"github.com/go-gotop/kit/rate"
	"github.com/redis/go-redis/v9"
)

func NewTestLimiter(accountId string, redisClient redis.Client, opts ...limiter.Option) *TestLimiter {
	o := &limiter.Options{
		PeriodLimitArray: []limiter.PeriodLimit{
			{
				WsConnectPeriod:         "5m",
				WsConnectTimes:          300,
				SpotCreateOrderPeriod:   "10s",
				SpotCreateOrderTimes:    100,
				FutureCreateOrderPeriod: "10s",
				FutureCreateOrderTimes:  300,
				SpotNormalRequestPeriod: "5m",
				SpotNormalRequestTimes:  61000,
			},
			{
				FutureCreateOrderPeriod: "1m",
				FutureCreateOrderTimes:  1200,
			},
		},
		CreateSpotOrderWeights:   1,
		CreateOcoOrderWeights:    2,
		CreateFutureOrderWeights: 0,
		CancelSpotOrderWeights:   1,
		CancelFutureOrderWeights: 1,
		SearchSpotOrderWeights:   1,
		SearchFutureOrderWeights: 1,
		UpdateSpotOrderWeights:   1,
		UpdateFutureOrderWeights: 1,
		OtherWeights:             1,
	}
	for _, opt := range opts {
		opt(o)
	}

	return &TestLimiter{
		opts:       o,
		limiterMap: limiter.SetAllLimiters(accountId, redisClient, "test", o.PeriodLimitArray),
	}
}

type TestLimiter struct {
	opts *limiter.Options

	limiterMap map[string][]*rate.Limiter
}

func (t *TestLimiter) WsAllow() bool {
	return limiter.LimiterAllow(t.limiterMap[limiter.WsConnectLimit])
}

func (t *TestLimiter) SpotAllow(limiterType limiter.LimitType) bool {
	return true
}

func (t *TestLimiter) FutureAllow(limiterType limiter.LimitType) bool {
	return true
}
