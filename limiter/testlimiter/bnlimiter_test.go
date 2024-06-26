package testlimiter

import (
	"log"
	"sync"
	"testing"
	"time"

	"github.com/go-gotop/kit/limiter"
	"github.com/go-gotop/kit/limiter/bnlimiter"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

// 测试并发下单，验证多线程下单，最终超过限制的次数是否符合预期
func TestRoutineCreateSpotOrderAllow(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379", // Redis 服务器地址
		// Password: "123456",
		DB: 0, // 使用的数据库编号
	})
	b := bnlimiter.NewBinanceLimiter(rdb)
	// 使用 WaitGroup 等待所有 goroutine 结束
	var wg sync.WaitGroup
	// 设置并发数量
	concurrency := 3
	wg.Add(concurrency)
	cycle := 35        // 每个 goroutine 执行次数
	notAllowedNum := 0 // 记录不允许的次数
	allowed := false
	// 并发调用 SpotAllow 函数
	for i := 0; i < concurrency; i++ {
		go func(index int) {
			// 减少 WaitGroup 计数
			defer wg.Done()
			for j := 0; j < cycle; j++ {
				// 执行 SpotAllow 函数
				req := &limiter.LimiterReq{
					AccountId:   "manbasji",
					LimiterType: limiter.CreateOrderLimit,
				}
				if i == 1 {
					// 一个 goroutine 执行 Oco 下单
					allowed = b.SpotAllow(req)
				} else {
					allowed = b.SpotAllow(req)
				}
				if !allowed {
					notAllowedNum++
				}
			}
		}(i)
	}
	// 等待所有 goroutine 结束
	wg.Wait()
	log.Printf("notAllowedNum: %v", notAllowedNum)
	assert.Equal(t, concurrency*cycle-100, notAllowedNum, "Expected spot order creation within rate limit")
	defer rdb.Close()
}

// 测试下单次数
func TestCreateSpotOrderAllow(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Redis 服务器地址
		DB:   0,                // 使用的数据库编号
	})
	b := bnlimiter.NewBinanceLimiter(rdb)
	req := &limiter.LimiterReq{
		AccountId:   "manbasji",
		LimiterType: limiter.CreateOrderLimit,
	}
	for i := 1; i <= 101; i++ {
		allowed := b.SpotAllow(req)
		if i < 101 {
			assert.True(t, allowed, "Expected spot order creation within rate limit at %dth attempt", i)
		} else {
			assert.False(t, allowed, "Expected spot order creation beyond rate limit at %dth attempt", i)
		}
	}
	defer rdb.Close()
}

// 测试权重
func TestCancelFutureOrderAllow(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Redis 服务器地址
		DB:   0,                // 使用的数据库编号
	})
	b := bnlimiter.NewBinanceLimiter(rdb)
	req := &limiter.LimiterReq{
		AccountId:   "manbasji",
		LimiterType: limiter.CancelOrderLimit,
	}
	// 测试通过情况：10秒内取消现货订单不超过100次
	for i := 0; i < 2400; i++ {
		allowed := b.FutureAllow(req)
		assert.True(t, allowed, "Expected spot order cancellation within rate limit at %dth attempt", i+1)
	}
	assert.False(t, b.FutureAllow(req), "Expected spot order cancellation beyond rate limit at 2401th attempt")
	time.Sleep(time.Minute)
	assert.True(t, b.FutureAllow(req), "Expected spot order cancellation within rate limit after 60 seconds")
	defer rdb.Close()
}

// 测试创建合约订单
func TestCreateFutureOrderAllow(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Redis 服务器地址
		DB:   0,                // 使用的数据库编号
	})
	b := bnlimiter.NewBinanceLimiter(rdb)
	req := &limiter.LimiterReq{
		AccountId:   "manbasji",
		LimiterType: limiter.CreateOrderLimit,
	}
	for i := 1; i <= 301; i++ {
		allowed := b.FutureAllow(req)
		if i < 301 {
			assert.True(t, allowed, "Expected future order creation within rate limit at %dth attempt", i)
		} else {
			assert.False(t, allowed, "Expected future order creation beyond rate limit at %dth attempt", i)
		}
	}
	time.Sleep(time.Second * 10)
	allowed := b.FutureAllow(req)
	assert.True(t, allowed, "Expected future order creation within rate limit after 10 seconds")
	defer rdb.Close()
}

// 测试创建合约订单 10s 最多300次，1min 内最多1200次
func TestCreateFutureOrderAllow1(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Redis 服务器地址
		DB:   0,                // 使用的数据库编号
	})
	b := bnlimiter.NewBinanceLimiter(rdb)
	initTime := time.Now()
	req := &limiter.LimiterReq{
		AccountId:   "manbasji",
		LimiterType: limiter.CreateOrderLimit,
	}
	for g := 1; g <= 4; g++ {
		for i := 1; i <= 301; i++ {
			allowed := b.FutureAllow(req)
			if i < 301 {
				assert.True(t, allowed, "Expected future order creation within rate limit at %dth attempt", i)
			} else {
				assert.False(t, allowed, "Expected future order creation beyond rate limit at %dth attempt", i)
			}
		}
		time.Sleep(10 * time.Second)
	}
	secondTime := time.Now()
	// 超过1200次
	allowed := b.FutureAllow(req)
	assert.False(t, allowed, "Expected future order creation beyond rate limit at %dth attempt", 1201)
	time.Sleep(time.Minute - secondTime.Sub(initTime))
	assert.True(t, b.FutureAllow(req), "Expected future order creation beyond rate limit at %dth attempt", 1201)
	defer rdb.Close()
}
