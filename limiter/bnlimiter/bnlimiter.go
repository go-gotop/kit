package bnlimiter

import (
	"context"
	"sync"
	"time"

	"github.com/go-gotop/kit/limiter"
	"github.com/go-gotop/kit/rate"
	"github.com/redis/go-redis/v9"
)

const (
	Exchange = "BINANCE"
)

// map 保存的限流器
func NewBinanceLimiter(accountId string, redisClient *redis.Client, opts ...limiter.Option) *BinanceLimiter {
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

	bl := &BinanceLimiter{
		rdb:        redisClient,
		opts:       o,
		accountId:  accountId,
		limiterMap: limiter.SetAllLimiters(accountId, *redisClient, Exchange, o.PeriodLimitArray),

		spotWeight:          limiter.WeightType(initRedisInt(Exchange+"_"+limiter.SpotWeight+"_"+accountId, *redisClient)),
		futureWeight:        limiter.WeightType(initRedisInt(Exchange+"_"+limiter.FutureWeight+"_"+accountId, *redisClient)),
		spotLastResetTime:   initRedisTime(Exchange+"_"+limiter.SpotLastRestTime+"_"+accountId, *redisClient),
		futureLastResetTime: initRedisTime(Exchange+"_"+limiter.FutureLastRestTime+"_"+accountId, *redisClient),
		spotMutex:           sync.Mutex{},
		futureMutex:         sync.Mutex{},
	}
	
	return bl
}

type BinanceLimiter struct {
	rdb       *redis.Client     // redis客户端
	opts      *limiter.Options // 配置
	accountId string

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

	b.spotLastResetTime = getRedisTime(Exchange+"_"+limiter.SpotLastRestTime+"_"+b.accountId, b.rdb)
	b.spotWeight = limiter.WeightType(getRedisInt(Exchange+"_"+limiter.SpotWeight+"_"+b.accountId, b.rdb))

	// 检查是否需要重置权重值
	if time.Since(b.spotLastResetTime) > time.Minute {
		b.spotWeight = 0
		b.spotLastResetTime = time.Now()
		b.rdb.Set(context.Background(), Exchange+"_"+limiter.SpotWeight+"_"+b.accountId, int64(b.spotWeight), time.Hour*24)
		b.rdb.Set(context.Background(), Exchange+"_"+limiter.SpotLastRestTime+"_"+b.accountId, b.spotLastResetTime.Format(time.RFC3339), time.Hour*24)
	}

	// 检查是否超过权重限制
	if b.spotWeight+wt > 6000 {
		return false
	}

	// 更新权重值
	b.spotWeight += wt
	b.rdb.Set(context.Background(), Exchange+"_"+limiter.SpotWeight+"_"+b.accountId, int64(b.spotWeight), time.Hour*24)

	return true
}

// 合约权重统计与判断
func (b *BinanceLimiter) allowFutureWeights(wt limiter.WeightType) bool {
	b.futureMutex.Lock()
	defer b.futureMutex.Unlock()

	b.futureLastResetTime = getRedisTime(Exchange+"_"+limiter.FutureLastRestTime+"_"+b.accountId, b.rdb)
	b.futureWeight = limiter.WeightType(getRedisInt(Exchange+"_"+limiter.FutureWeight+"_"+b.accountId, b.rdb))

	// 检查是否需要重置权重值
	if time.Since(b.futureLastResetTime) > time.Minute {
		b.futureWeight = 0
		b.futureLastResetTime = time.Now()
		b.rdb.Set(context.Background(), Exchange+"_"+limiter.FutureWeight+"_"+b.accountId, int64(b.futureWeight), time.Hour*24)
		b.rdb.Set(context.Background(), Exchange+"_"+limiter.FutureLastRestTime+"_"+b.accountId, b.futureLastResetTime.Format(time.RFC3339), time.Hour*24)
	}

	// 检查是否超过权重限制
	if b.futureWeight+wt > 2400 {
		return false
	}

	// 更新权重值
	b.futureWeight += wt
	b.rdb.Set(context.Background(), Exchange+"_"+limiter.FutureWeight+"_"+b.accountId, int64(b.futureWeight), time.Hour*24)
	return true
}

func initRedisTime(uniqId string, redisClient redis.Client) time.Time {
	val := time.Now()
	timeVal, err := redisClient.Get(context.Background(), uniqId).Result()
	if err == redis.Nil {
		val = time.Now()
		redisClient.Set(context.Background(), uniqId, val.Format(time.RFC3339), 0)
	} else if timeVal != "" {
		timeVal, err := time.Parse(time.RFC3339, timeVal)
		if err != nil {
			val = time.Now()
			redisClient.Set(context.Background(), uniqId, val, time.Hour*24)
		} else {
			val = timeVal
		}
	}
	return val
}

func initRedisInt(uniqId string, redisClient redis.Client) int64 {
	val := int64(0)
	intVal, err := redisClient.Get(context.Background(), uniqId).Int64()
	if err == redis.Nil {
		val = 0
		redisClient.Set(context.Background(), uniqId, val, time.Hour*24)
	} else if intVal != 0 {
		val = intVal
	}
	return val
}

func getRedisTime(uniqId string, redisClient *redis.Client) time.Time {
	val := time.Now()
	timeVal, err := redisClient.Get(context.Background(), uniqId).Result()
	if err != nil {
		return val
	}
	if timeVal != "" {
		timeVal, err := time.Parse(time.RFC3339, timeVal)
		if err != nil {
			val = time.Now()
		} else {
			val = timeVal
		}
	}
	// println("val:", val.Format(time.RFC3339))
	return val
}

func getRedisInt(uniqId string, redisClient *redis.Client) int64 {
	val := int64(0)
	intVal, err := redisClient.Get(context.Background(), uniqId).Int64()
	if err != nil {
		return val
	}
	if intVal != 0 {
		val = intVal
	}
	return val
}
