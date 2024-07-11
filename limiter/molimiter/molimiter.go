// molimiter 用于mock交易所的限流器，回测暂时不需要限流，所以直接返回true
// 添加这个限流器是为了使用端代码的统一性，并且方便后续扩展
package molimiter

import (
	"context"
	"sync"
	"time"

	"github.com/go-gotop/kit/exchange"
	"github.com/go-gotop/kit/limiter"
	"github.com/go-gotop/kit/rate"
	"github.com/redis/go-redis/v9"
)

const (
	Exchange = exchange.MockExchange
)

// map 保存的限流器
func NewMockLimiter(redisClient *redis.Client, opts ...limiter.Option) *MockLimiter {
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

	ip, _ := limiter.GetOutBoundIP()

	bl := &MockLimiter{
		ip:         ip,
		rdb:        redisClient,
		opts:       o,
		limiterMap: limiter.SetAllLimiters(*redisClient, Exchange, o.PeriodLimitArray),

		spotWeight:          limiter.WeightType(initRedisInt(Exchange+"_"+limiter.SpotWeight+"_"+ip, *redisClient)),
		futureWeight:        limiter.WeightType(initRedisInt(Exchange+"_"+limiter.FutureWeight+"_"+ip, *redisClient)),
		spotLastResetTime:   initRedisTime(Exchange+"_"+limiter.SpotLastRestTime+"_"+ip, *redisClient),
		futureLastResetTime: initRedisTime(Exchange+"_"+limiter.FutureLastRestTime+"_"+ip, *redisClient),
		spotMutex:           sync.Mutex{},
		futureMutex:         sync.Mutex{},
	}

	return bl
}

type MockLimiter struct {
	ip   string           // 出口ip，作为rediskey进行保存，binance交易所除了下单针对apikey限制外，其他是针对ip的
	rdb  *redis.Client    // redis客户端
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

func (b *MockLimiter) WsAllow() bool {
	return true
	// return limiter.LimiterAllow(b.limiterMap[limiter.WsConnectLimit], Exchange+"_"+b.ip)
}

// SpotAllow checks if the request is allowed for spot trading
func (b *MockLimiter) SpotAllow(t *limiter.LimiterReq) bool {
	return true
	// switch t.LimiterType {
	// case limiter.CreateOcoOrderLimit:
	// 	return b.allowCreateOcoOrder(Exchange + "_" + limiter.SpotCreateOrderLimit + "_" + t.AccountId)
	// case limiter.CreateOrderLimit:
	// 	return b.allowCreateSpotOrder(Exchange + "_" + limiter.SpotCreateOrderLimit + "_" + t.AccountId)
	// case limiter.CancelOrderLimit:
	// 	return b.allowCancelSpotOrder(Exchange + "_" + limiter.SpotNormalRequestLimit + "_" + b.ip)
	// case limiter.SearchOrderLimit:
	// 	return b.allowSearchSpotOrder(Exchange + "_" + limiter.SpotNormalRequestLimit + "_" + b.ip)
	// case limiter.NormalRequestLimit:
	// 	return b.allowSpotNormalRequest(Exchange + "_" + limiter.SpotNormalRequestLimit + "_" + b.ip)
	// default:
	// 	return true
	// }
}

// FutureAllow checks if the request is allowed for future trading
func (b *MockLimiter) FutureAllow(t *limiter.LimiterReq) bool {
	return true
	// switch t.LimiterType {
	// case limiter.CreateOrderLimit:
	// 	return b.allowCreateFutureOrder(Exchange + "_" + limiter.FutureCreateOrderLimit + "_" + t.AccountId)
	// case limiter.CancelOrderLimit:
	// 	return b.allCancelFutureOrder()
	// case limiter.SearchOrderLimit:
	// 	return b.allSearchFutureOrder()
	// case limiter.NormalRequestLimit:
	// 	return b.allFutureNormalRequest()
	// default:
	// 	return true
	// }
}

// 允许创建现货oco订单
func (b *MockLimiter) allowCreateOcoOrder(uniq string) bool {
	return limiter.LimiterAllow(b.limiterMap[limiter.SpotCreateOrderLimit], uniq) && b.allowSpotWeights(b.opts.CreateOcoOrderWeights)
}

// 允许创建现货订单
func (b *MockLimiter) allowCreateSpotOrder(uniq string) bool {
	return limiter.LimiterAllow(b.limiterMap[limiter.SpotCreateOrderLimit], uniq) && b.allowSpotWeights(b.opts.CreateSpotOrderWeights)
}

// 允许取消现货订单
func (b *MockLimiter) allowCancelSpotOrder(uniq string) bool {
	return limiter.LimiterAllow(b.limiterMap[limiter.SpotNormalRequestLimit], uniq) && b.allowSpotWeights(b.opts.CancelSpotOrderWeights)
}

// 允许查询现货订单
func (b *MockLimiter) allowSearchSpotOrder(uniq string) bool {
	return limiter.LimiterAllow(b.limiterMap[limiter.SpotNormalRequestLimit], uniq) && b.allowSpotWeights(b.opts.SearchSpotOrderWeights)
}

// 允许现货其他普通请求
func (b *MockLimiter) allowSpotNormalRequest(uniq string) bool {
	return limiter.LimiterAllow(b.limiterMap[limiter.SpotNormalRequestLimit], uniq) && b.allowSpotWeights(b.opts.OtherWeights)
}

// 允许创建合约订单
func (b *MockLimiter) allowCreateFutureOrder(uniq string) bool {
	return limiter.LimiterAllow(b.limiterMap[limiter.FutureCreateOrderLimit], uniq) && b.allowFutureWeights(b.opts.CreateFutureOrderWeights)
}

// 允许取消合约订单
func (b *MockLimiter) allCancelFutureOrder() bool {
	return b.allowFutureWeights(b.opts.CancelFutureOrderWeights)
}

// 允许查询合约订单
func (b *MockLimiter) allSearchFutureOrder() bool {
	return b.allowFutureWeights(b.opts.SearchFutureOrderWeights)
}

// 允许合约其他普通请求
func (b *MockLimiter) allFutureNormalRequest() bool {
	return b.allowFutureWeights(b.opts.OtherWeights)
}

// 现货权重统计与判断
func (b *MockLimiter) allowSpotWeights(wt limiter.WeightType) bool {
	b.spotMutex.Lock()
	defer b.spotMutex.Unlock()

	b.spotLastResetTime = getRedisTime(Exchange+"_"+limiter.SpotLastRestTime+"_"+b.ip, b.rdb)
	b.spotWeight = limiter.WeightType(getRedisInt(Exchange+"_"+limiter.SpotWeight+"_"+b.ip, b.rdb))

	// 检查是否需要重置权重值
	if time.Since(b.spotLastResetTime) > time.Minute {
		b.spotWeight = 0
		b.spotLastResetTime = time.Now()
		b.rdb.Set(context.Background(), Exchange+"_"+limiter.SpotWeight+"_"+b.ip, int64(b.spotWeight), time.Hour*24)
		b.rdb.Set(context.Background(), Exchange+"_"+limiter.SpotLastRestTime+"_"+b.ip, b.spotLastResetTime.Format(time.RFC3339), time.Hour*24)
	}

	// 检查是否超过权重限制
	if b.spotWeight+wt > 6000 {
		return false
	}

	// 更新权重值
	b.spotWeight += wt
	b.rdb.Set(context.Background(), Exchange+"_"+limiter.SpotWeight+"_"+b.ip, int64(b.spotWeight), time.Hour*24)

	return true
}

// 合约权重统计与判断
func (b *MockLimiter) allowFutureWeights(wt limiter.WeightType) bool {
	b.futureMutex.Lock()
	defer b.futureMutex.Unlock()

	b.futureLastResetTime = getRedisTime(Exchange+"_"+limiter.FutureLastRestTime+"_"+b.ip, b.rdb)
	b.futureWeight = limiter.WeightType(getRedisInt(Exchange+"_"+limiter.FutureWeight+"_"+b.ip, b.rdb))

	// 检查是否需要重置权重值
	if time.Since(b.futureLastResetTime) > time.Minute {
		b.futureWeight = 0
		b.futureLastResetTime = time.Now()
		b.rdb.Set(context.Background(), Exchange+"_"+limiter.FutureWeight+"_"+b.ip, int64(b.futureWeight), time.Hour*24)
		b.rdb.Set(context.Background(), Exchange+"_"+limiter.FutureLastRestTime+"_"+b.ip, b.futureLastResetTime.Format(time.RFC3339), time.Hour*24)
	}

	// 检查是否超过权重限制
	if b.futureWeight+wt > 2400 {
		return false
	}

	// 更新权重值
	b.futureWeight += wt
	b.rdb.Set(context.Background(), Exchange+"_"+limiter.FutureWeight+"_"+b.ip, int64(b.futureWeight), time.Hour*24)
	return true
}

func initRedisTime(uniq string, redisClient redis.Client) time.Time {
	val := time.Now()
	timeVal, err := redisClient.Get(context.Background(), uniq).Result()
	if err == redis.Nil {
		val = time.Now()
		redisClient.Set(context.Background(), uniq, val.Format(time.RFC3339), 0)
	} else if timeVal != "" {
		timeVal, err := time.Parse(time.RFC3339, timeVal)
		if err != nil {
			val = time.Now()
			redisClient.Set(context.Background(), uniq, val, time.Hour*24)
		} else {
			val = timeVal
		}
	}
	return val
}

func initRedisInt(uniq string, redisClient redis.Client) int64 {
	val := int64(0)
	intVal, err := redisClient.Get(context.Background(), uniq).Int64()
	if err == redis.Nil {
		val = 0
		redisClient.Set(context.Background(), uniq, val, time.Hour*24)
	} else if intVal != 0 {
		val = intVal
	}
	return val
}

func getRedisTime(uniq string, redisClient *redis.Client) time.Time {
	val := time.Now()
	timeVal, err := redisClient.Get(context.Background(), uniq).Result()
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

func getRedisInt(uniq string, redisClient *redis.Client) int64 {
	val := int64(0)
	intVal, err := redisClient.Get(context.Background(), uniq).Int64()
	if err != nil {
		return val
	}
	if intVal != 0 {
		val = intVal
	}
	return val
}
