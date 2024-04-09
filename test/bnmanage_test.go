package test

import (
	"log"
	"testing"
	"time"

	"github.com/go-gotop/kit/exchange"
	"github.com/go-gotop/kit/requests/bnhttp"
	"github.com/go-gotop/kit/websocket"
	"github.com/go-gotop/kit/wsmanager/bnmanager"
	"github.com/stretchr/testify/assert"
)

// 测试添加bn websocket连接
func TestAddBnWebsocket(t *testing.T) {
	client := bnhttp.NewClient(
		bnhttp.APIKey("HqeCVqbbOP9kFcJy7gmudcXEcZUR2Z1IXePu3L9Jp1K85aSEMH86nqwOH6pgPxQD"),
		bnhttp.SecretKey("caXcvwIhOvcWrcZgI9SHb5ifL4a5bgS5NXOO3iDthURgNnJtWE5xtPRfcBzsgHYg"),
		bnhttp.BaseUrl("https://testnet.binance.vision"),
	)

	bm := bnmanager.NewBnManager(client)

	aReq_market := &websocket.WebsocketRequest{
		Endpoint: "wss://testnet.binance.vision/ws/btcusdt@kline_1s",
		ID:       "test_market",
		MessageHandler: func(data []byte) error {
			log.Printf("data: %v", string(data))
			return nil
		},
	}

	aReq_account := &websocket.WebsocketRequest{
		Endpoint: "wss://testnet.binance.vision/ws",
		ID:       "test_account",
		MessageHandler: func(data []byte) error {
			log.Printf("data: %v", string(data))
			return nil
		},
	}

	unqi_market, err := bm.AddMarketWebSocket(aReq_market, exchange.InstrumentTypeSpot)
	if err != nil {
		log.Printf("err: %v", err)
	}
	uniq_account, err := bm.AddAccountWebSocket(aReq_account, exchange.InstrumentTypeSpot)
	if err != nil {
		log.Printf("err: %v", err)
	}
	isConnected_market := bm.IsConnected(unqi_market)
	isConnected_account := bm.IsConnected(uniq_account)
	assert.True(t, isConnected_market, "Expected websocket connection to be established")
	assert.True(t, isConnected_account, "Expected websocket connection to be established")
	time.Sleep(10 * time.Second)
}

// 测试listenkey 检查机制
func TestListenKeyCheck(t *testing.T) {

	client := bnhttp.NewClient(
		bnhttp.APIKey("HqeCVqbbOP9kFcJy7gmudcXEcZUR2Z1IXePu3L9Jp1K85aSEMH86nqwOH6pgPxQD"),
		bnhttp.SecretKey("caXcvwIhOvcWrcZgI9SHb5ifL4a5bgS5NXOO3iDthURgNnJtWE5xtPRfcBzsgHYg"),
		bnhttp.BaseUrl("https://testnet.binance.vision"),
	)

	bm := bnmanager.NewBnManager(client,
		bnmanager.WithMaxConn(10),
		bnmanager.WithMaxConnDuration(30*time.Second),
		bnmanager.WithListenKeyExpire(10*time.Second),
		bnmanager.WithCheckListenKeyPeriod(5*time.Second))

	aReq_market := &websocket.WebsocketRequest{
		Endpoint: "wss://testnet.binance.vision/ws",
		ID:       "test_market",
		MessageHandler: func(data []byte) error {
			log.Printf("data: %v", string(data))
			return nil
		},
	}

	unqi_market, err := bm.AddAccountWebSocket(aReq_market, exchange.InstrumentTypeSpot)
	if err != nil {
		log.Printf("err: %v", err)
	}

	isConnected_market := bm.IsConnected(unqi_market)

	assert.True(t, isConnected_market, "Expected websocket connection to be established")

	for i := 0; i < 40; i++ {
		time.Sleep(1 * time.Second)
		log.Printf("i: %v", i)
	}

	bm.CloseWebSocket(unqi_market)
	isConnected_market = bm.IsConnected(unqi_market)
	assert.False(t, isConnected_market, "Expected websocket connection to be closed")
}

// 测试主动close之后不会重连
func TestCloseWebsocket(t *testing.T) {
	client := bnhttp.NewClient(
		bnhttp.APIKey("HqeCVqbbOP9kFcJy7gmudcXEcZUR2Z1IXePu3L9Jp1K85aSEMH86nqwOH6pgPxQD"),
		bnhttp.SecretKey("caXcvwIhOvcWrcZgI9SHb5ifL4a5bgS5NXOO3iDthURgNnJtWE5xtPRfcBzsgHYg"),
		bnhttp.BaseUrl("https://testnet.binance.vision"),
	)

	bm := bnmanager.NewBnManager(client,
		bnmanager.WithMaxConn(10),
		bnmanager.WithMaxConnDuration(30*time.Second),
		bnmanager.WithListenKeyExpire(10*time.Second),
		bnmanager.WithCheckListenKeyPeriod(5*time.Second))

	aReq_market := &websocket.WebsocketRequest{
		Endpoint: "wss://testnet.binance.vision/ws",
		ID:       "test_market",
		MessageHandler: func(data []byte) error {
			log.Printf("data: %v", string(data))
			return nil
		},
	}

	unqi_market, err := bm.AddAccountWebSocket(aReq_market, exchange.InstrumentTypeSpot)
	if err != nil {
		log.Printf("err: %v", err)
	}

	isConnected_market := bm.IsConnected(unqi_market)

	assert.True(t, isConnected_market, "Expected websocket connection to be established")
	err = bm.CloseWebSocket(unqi_market)
	if err != nil {
		log.Printf("err: %v", err)
	}
	time.Sleep(2 * time.Second) // 延迟2秒，看是否会自动重连
	isConnected_market = bm.IsConnected(unqi_market)
	assert.False(t, isConnected_market, "Expected websocket connection to be closed")
}
