package testing

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-gotop/kit/dfmanager"
	"github.com/go-gotop/kit/dfmanager/dfbinance"
	"github.com/go-gotop/kit/exchange"
	"github.com/go-gotop/kit/limiter/bnlimiter"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func messageEvent(data *exchange.TradeEvent) {
	fmt.Printf("TradeEvent: %v\n", data)
}

func errEvent(err error) {
	fmt.Printf("Error: %v\n", err)
}

func newRedis() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379", // Redis 服务器地址
		// Password: "123456",
		DB: 0, // 使用的数据库编号
	})
	return rdb
}

func TestNewBinanceDataFeed(t *testing.T) {
	// 默认配置

	limiter := bnlimiter.NewBinanceLimiter(newRedis())

	df := dfbinance.NewBinanceDataFeed(limiter)
	assert.NotNil(t, df)
	uuid := uuid.New().String()
	err := df.AddDataFeed(&dfmanager.DataFeedRequest{
		ID:           uuid,
		Symbol:       "BTCUSDT",
		Instrument:   exchange.InstrumentTypeSpot,
		Event:        messageEvent,
		ErrorHandler: errEvent,
	})

	assert.Nil(t, err)

	time.Sleep(30 * time.Second)

}

// 测试closeDatafeed
func TestCloseDataFeed(t *testing.T) {
	// 默认配置

	limiter := bnlimiter.NewBinanceLimiter(newRedis())

	df := dfbinance.NewBinanceDataFeed(limiter)
	assert.NotNil(t, df)
	uuid := uuid.New().String()
	err := df.AddDataFeed(&dfmanager.DataFeedRequest{
		ID:           uuid,
		Symbol:       "BTCUSDT",
		Instrument:   exchange.InstrumentTypeSpot,
		Event:        messageEvent,
		ErrorHandler: errEvent,
	})

	assert.Nil(t, err)
	time.Sleep(30 * time.Second)
	err = df.CloseDataFeed(uuid)
	assert.Nil(t, err)

	time.Sleep(10 * time.Second)

	l := df.DataFeedList()
	fmt.Println(l)
	assert.Equal(t, 0, len(l))

}
