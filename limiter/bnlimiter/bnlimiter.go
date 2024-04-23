package bnlimiter

import (
	"sync"
	"time"

	"github.com/go-gotop/kit/limiter"
	"github.com/go-gotop/kit/rate"
	"github.com/go-redis/redis"
)

// map 保存的限流器
func NewBinanceLimiter(accountId string, redisClient redis.Client, opts ...limiter.Option) *BinanceLimiter {
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

	return &BinanceLimiter{
		opts:       o,
		limiterMap: limiter.SetAllLimiters(o.PeriodLimitArray),

		spotWeight:          0,
		futureWeight:        0,
		spotLastResetTime:   time.Now(),
		futureLastResetTime: time.Now(),
		spotMutex:           sync.Mutex{},
		futureMutex:         sync.Mutex{},
	}
}

type BinanceLimiter struct {
	opts *limiter.Options // 配置

	limiterMap map[string][]*rate.Limiter // 限流器

	spotWeight          limiter.WeightType // 现货权重统计
	futureWeight        limiter.WeightType // 合约权重统计
	spotLastResetTime   time.Time          // 现货上次重置时间
	futureLastResetTime time.Time          // 合约上次重置时间
	spotMutex           sync.Mutex         // 互斥锁
	futureMutex         sync.Mutex         // 互斥锁
}

type LimiterGroup struct {
	SpotLimiter          limiter.Limiter
	FutureLimiter        limiter.Limiter
	NormalRequestLimiter limiter.Limiter
}

func (b *BinanceLimiter) WsAllow() bool {
	return limiter.LimiterAllow(b.limiterMap[limiter.WsConnectLimit])
}

// SpotAllow checks if the request is allowed for spot trading
func (b *BinanceLimiter) SpotAllow(t limiter.LimitType) bool {
	switch t {
	case limiter.CreateOcoOrderLimit:
		return b.allowCreateOcoOrder()
	case limiter.CreateOrderLimit:
		return b.allowCreateSpotOrder()
	case limiter.CancelOrderLimit:
		return b.allowCancelSpotOrder()
	case limiter.SearchOrderLimit:
		return b.allowSearchSpotOrder()
	case limiter.NormalRequestLimit:
		return b.allowSpotNormalRequest()
	default:
		return true
	}
}

// FutureAllow checks if the request is allowed for future trading
func (b *BinanceLimiter) FutureAllow(t limiter.LimitType) bool {
	switch t {
	case limiter.CreateOrderLimit:
		return b.allowCreateFutureOrder()
	case limiter.CancelOrderLimit:
		return b.allCancelFutureOrder()
	case limiter.SearchOrderLimit:
		return b.allSearchFutureOrder()
	case limiter.NormalRequestLimit:
		return b.allFutureNormalRequest()
	default:
		return true
	}
}

// 允许创建现货oco订单
func (b *BinanceLimiter) allowCreateOcoOrder() bool {
	return limiter.LimiterAllow(b.limiterMap[limiter.SpotCreateOrderLimit]) && b.allowSpotWeights(b.opts.CreateOcoOrderWeights)
}

// 允许创建现货订单
func (b *BinanceLimiter) allowCreateSpotOrder() bool {
	return limiter.LimiterAllow(b.limiterMap[limiter.SpotCreateOrderLimit]) && b.allowSpotWeights(b.opts.CreateSpotOrderWeights)
}

// 允许取消现货订单
func (b *BinanceLimiter) allowCancelSpotOrder() bool {
	return limiter.LimiterAllow(b.limiterMap[limiter.SpotNormalRequestLimit]) && b.allowSpotWeights(b.opts.CancelSpotOrderWeights)
}

// 允许查询现货订单
func (b *BinanceLimiter) allowSearchSpotOrder() bool {
	return limiter.LimiterAllow(b.limiterMap[limiter.SpotNormalRequestLimit]) && b.allowSpotWeights(b.opts.SearchSpotOrderWeights)
}

// 允许现货其他普通请求
func (b *BinanceLimiter) allowSpotNormalRequest() bool {
	return limiter.LimiterAllow(b.limiterMap[limiter.SpotNormalRequestLimit]) && b.allowSpotWeights(b.opts.OtherWeights)
}

// 允许创建合约订单
func (b *BinanceLimiter) allowCreateFutureOrder() bool {
	return limiter.LimiterAllow(b.limiterMap[limiter.FutureCreateOrderLimit]) && b.allowFutureWeights(b.opts.CreateFutureOrderWeights)
}

// 允许取消合约订单
func (b *BinanceLimiter) allCancelFutureOrder() bool {
	return b.allowFutureWeights(b.opts.CancelFutureOrderWeights)
}

// 允许查询合约订单
func (b *BinanceLimiter) allSearchFutureOrder() bool {
	return b.allowFutureWeights(b.opts.SearchFutureOrderWeights)
}

// 允许合约其他普通请求
func (b *BinanceLimiter) allFutureNormalRequest() bool {
	return b.allowFutureWeights(b.opts.OtherWeights)
}

// 现货权重统计与判断
func (b *BinanceLimiter) allowSpotWeights(wt limiter.WeightType) bool {
	b.spotMutex.Lock()
	defer b.spotMutex.Unlock()

	// 检查是否需要重置权重值
	if time.Since(b.spotLastResetTime) > time.Minute {
		b.spotWeight = 0
		b.spotLastResetTime = time.Now()
	}

	// 检查是否超过权重限制
	if b.spotWeight+wt > 6000 {
		return false
	}

	// 更新权重值
	b.spotWeight += wt

	return true
}

// 合约权重统计与判断
func (b *BinanceLimiter) allowFutureWeights(wt limiter.WeightType) bool {
	b.futureMutex.Lock()
	defer b.futureMutex.Unlock()

	// 检查是否需要重置权重值
	if time.Since(b.futureLastResetTime) > time.Minute {
		b.futureWeight = 0
		b.futureLastResetTime = time.Now()
	}

	// 检查是否超过权重限制
	if b.futureWeight+wt > 2400 {
		return false
	}

	// 更新权重值
	b.futureWeight += wt

	return true
}
