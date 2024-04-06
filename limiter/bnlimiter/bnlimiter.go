package bnlimiter

import (
	"sync"
	"time"

	"github.com/go-gotop/kit/limiter"
	"github.com/go-gotop/kit/rate"
)

// map 保存的限流器
const (
	wsConnectLimit         = "ws_connect"
	spotCreateOrderLimit   = "spot_create_order"
	spotNormalRequestLimit = "spot_normal_request"
	futureCreateOrderLimit = "future_create_order"
)

func NewBinanceLimiter(opts ...Option) *BinanceLimiter {
	o := &options{
		periodLimitArray: []PeriodLimit{
			{
				wsConnectPeriod:         "5m",
				wsConnectTimes:          300,
				spotCreateOrderPeriod:   "10s",
				spotCreateOrderTimes:    100,
				futureCreateOrderPeriod: "10s",
				futureCreateOrderTimes:  300,
				spotNormalRequestPeriod: "5m",
				spotNormalRequestTimes:  61000,
			},
			{
				futureCreateOrderPeriod: "1m",
				futureCreateOrderTimes:  1200,
			},
		},
		createSpotOrderWeights:   1,
		createOcoOrderWeights:    2,
		createFutureOrderWeights: 0,
		cancelSpotOrderWeights:   1,
		cancelFutureOrderWeights: 1,
		searchSpotOrderWeights:   1,
		searchFutureOrderWeights: 1,
		updateSpotOrderWeights:   1,
		updateFutureOrderWeights: 1,
		otherWeights:             1,
	}
	for _, opt := range opts {
		opt(o)
	}

	return &BinanceLimiter{
		opts:       o,
		limiterMap: SetAllLimiters(o.periodLimitArray),

		spotWeight:          0,
		futureWeight:        0,
		spotLastResetTime:   time.Now(),
		futureLastResetTime: time.Now(),
		spotMutex:           sync.Mutex{},
		futureMutex:         sync.Mutex{},
	}
}

type BinanceLimiter struct {
	opts *options // 配置

	limiterMap map[string][]*rate.Limiter // 限流器

	spotWeight          WeightType // 现货权重统计
	futureWeight        WeightType // 合约权重统计
	spotLastResetTime   time.Time  // 现货上次重置时间
	futureLastResetTime time.Time  // 合约上次重置时间
	spotMutex           sync.Mutex // 互斥锁
	futureMutex         sync.Mutex // 互斥锁
}

type LimiterGroup struct {
	SpotLimiter          limiter.Limiter
	FutureLimiter        limiter.Limiter
	NormalRequestLimiter limiter.Limiter
}

func (b *BinanceLimiter) WsAllow() bool {
	return limiterAllow(b.limiterMap[wsConnectLimit])
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
	return limiterAllow(b.limiterMap[spotCreateOrderLimit]) && b.allowSpotWeights(b.opts.createOcoOrderWeights)
}

// 允许创建现货订单
func (b *BinanceLimiter) allowCreateSpotOrder() bool {
	return limiterAllow(b.limiterMap[spotCreateOrderLimit]) && b.allowSpotWeights(b.opts.createSpotOrderWeights)
}

// 允许取消现货订单
func (b *BinanceLimiter) allowCancelSpotOrder() bool {
	return limiterAllow(b.limiterMap[spotNormalRequestLimit]) && b.allowSpotWeights(b.opts.cancelSpotOrderWeights)
}

// 允许查询现货订单
func (b *BinanceLimiter) allowSearchSpotOrder() bool {
	return limiterAllow(b.limiterMap[spotNormalRequestLimit]) && b.allowSpotWeights(b.opts.searchSpotOrderWeights)
}

// 允许现货其他普通请求
func (b *BinanceLimiter) allowSpotNormalRequest() bool {
	return limiterAllow(b.limiterMap[spotNormalRequestLimit]) && b.allowSpotWeights(b.opts.otherWeights)
}

// 允许创建合约订单
func (b *BinanceLimiter) allowCreateFutureOrder() bool {
	return limiterAllow(b.limiterMap[futureCreateOrderLimit]) && b.allowFutureWeights(b.opts.createFutureOrderWeights)
}

// 允许取消合约订单
func (b *BinanceLimiter) allCancelFutureOrder() bool {
	return b.allowFutureWeights(b.opts.cancelFutureOrderWeights)
}

// 允许查询合约订单
func (b *BinanceLimiter) allSearchFutureOrder() bool {
	return b.allowFutureWeights(b.opts.searchFutureOrderWeights)
}

// 允许合约其他普通请求
func (b *BinanceLimiter) allFutureNormalRequest() bool {
	return b.allowFutureWeights(b.opts.otherWeights)
}

// 现货权重统计与判断
func (b *BinanceLimiter) allowSpotWeights(wt WeightType) bool {
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
func (b *BinanceLimiter) allowFutureWeights(wt WeightType) bool {
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
