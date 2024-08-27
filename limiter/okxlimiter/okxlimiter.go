package okxlimiter

import (
	"sync"

	"github.com/go-gotop/kit/exchange"
	"github.com/go-gotop/kit/limiter"
	"github.com/go-gotop/kit/rate"
	"github.com/redis/go-redis/v9"
)

const (
	Exchange = exchange.OkxExchange
)

func NewOkxLimiter(redisClient *redis.Client, opt ...limiter.Option) *OkxLimiter {
	o := &limiter.Options{
		PeriodLimitArray: []limiter.PeriodLimit{
			{
				WsConnectPeriod:           "1s",
				WsConnectTimes:            3,
				SpotCreateOrderPeriod:     "2s",
				SpotCreateOrderTimes:      60,
				FutureCreateOrderPeriod:   "2s",
				FutureCreateOrderTimes:    60,
				MarginNormalRequestPeriod: "2s",
				MarginNormalRequestTimes:  60,
			},
		},
	}
	for _, v := range opt {
		v(o)
	}

	ip, _ := limiter.GetOutBoundIP()

	ol := &OkxLimiter{
		ip:         ip,
		rdb:        redisClient,
		opts:       o,
		limiterMap: limiter.SetAllLimiters(*redisClient, Exchange, o.PeriodLimitArray),
		mutex:      sync.Mutex{},
	}

	return ol
}

type OkxLimiter struct {
	ip         string
	rdb        *redis.Client
	opts       *limiter.Options
	limiterMap map[string][]*rate.Limiter // 限流器
	mutex      sync.Mutex
}

func (o *OkxLimiter) WsAllow() bool {
	return limiter.LimiterAllow(o.limiterMap[limiter.WsConnectLimit], Exchange+"_"+o.ip)
}

func (o *OkxLimiter) SpotAllow(t *limiter.LimiterReq) bool {
	return limiter.LimiterAllow(o.limiterMap[limiter.SpotCreateOrderLimit], Exchange+"_"+limiter.SpotCreateOrderLimit+"_"+o.ip)
}

func (o *OkxLimiter) FutureAllow(t *limiter.LimiterReq) bool {
	return limiter.LimiterAllow(o.limiterMap[limiter.FutureCreateOrderLimit], Exchange+"_"+limiter.FutureCreateOrderLimit+"_"+o.ip)
}

func (o *OkxLimiter) MarginAllow(t *limiter.LimiterReq) bool {
	return limiter.LimiterAllow(o.limiterMap[limiter.MarginNormalRequestLimit], Exchange+"_"+limiter.MarginNormalRequestLimit+"_"+o.ip)
}
