package test

import (
	"log"
	"testing"
	"time"

	"github.com/go-gotop/kit/limiter"
	"github.com/go-gotop/kit/limiter/testlimiter"
	"github.com/go-gotop/kit/websocket"
	"github.com/go-gotop/kit/wsmanager"
	"github.com/go-gotop/kit/wsmanager/manager"
	"github.com/stretchr/testify/assert"
)

// 测试websocket 连接状态
func TestWebSocketsConnected(t *testing.T) {
	limiter := testlimiter.NewTestLimiter(
		limiter.WithPeriodLimitArray([]limiter.PeriodLimit{
			{
				WsConnectPeriod: "1s",
				WsConnectTimes:  10,
			},
		},
		))
	websocketManager := manager.NewManager(
		manager.WithMaxConn(10),
		manager.WithMaxConnDuration(2*time.Minute),
		manager.WithConnLimiter(limiter),
		manager.WithCheckReConn(true),
	)

	wsReq := websocket.WebsocketRequest{
		Endpoint:       "wss://testnet.binance.vision/ws",
		ID:             "test",
		MessageHandler: handleMessage,
	}

	config := wsmanager.ExchangeWebsocketConfig{
		PingHandler: pingHandler,
		PongHandler: pongHandler,
	}
	uniq, err := websocketManager.AddWebsocket(&wsReq, &config)
	if err != nil {
		log.Printf("err: %v", err)
	}
	isConnected := websocketManager.IsConnected(uniq)
	assert.True(t, isConnected, "Expected websocket connection to be established")
}

// 测试websocket最大连接数
func TestWebSocketsMaxConn(t *testing.T) {
	limiter := testlimiter.NewTestLimiter(
		limiter.WithPeriodLimitArray([]limiter.PeriodLimit{
			{
				WsConnectPeriod: "1s",
				WsConnectTimes:  10,
			},
		},
		))
	websocketManager := manager.NewManager(
		manager.WithMaxConn(10),
		manager.WithMaxConnDuration(2*time.Minute),
		manager.WithConnLimiter(limiter),
		manager.WithCheckReConn(true),
	)

	wsReq := websocket.WebsocketRequest{
		Endpoint:       "wss://testnet.binance.vision/ws",
		ID:             "test",
		MessageHandler: handleMessage,
	}

	config := wsmanager.ExchangeWebsocketConfig{
		PingHandler: pingHandler,
		PongHandler: pongHandler,
	}
	for i := 0; i < 10; i++ {
		_, err := websocketManager.AddWebsocket(&wsReq, &config)
		if err != nil {
			log.Printf("err: %v", err)
		}
	}
	_, err := websocketManager.AddWebsocket(&wsReq, &config)
	assert.Error(t, err, "Expected error when websocket connection exceeds maximum connections")
}

// 测试websocket连接频率限制
func TestWebSocketsConnLimiter(t *testing.T) {
	limiter := testlimiter.NewTestLimiter(
		limiter.WithPeriodLimitArray([]limiter.PeriodLimit{
			{
				WsConnectPeriod: "1s",
				WsConnectTimes:  10,
			},
		},
		))
	websocketManager := manager.NewManager(
		manager.WithMaxConn(10),
		manager.WithMaxConnDuration(2*time.Minute),
		manager.WithConnLimiter(limiter),
		manager.WithCheckReConn(true),
	)

	wsReq := websocket.WebsocketRequest{
		Endpoint:       "wss://testnet.binance.vision/ws",
		ID:             "test",
		MessageHandler: handleMessage,
	}

	config := wsmanager.ExchangeWebsocketConfig{
		PingHandler: pingHandler,
		PongHandler: pongHandler,
	}
	for i := 0; i < 11; i++ {
		if i <= 10 {
			_, err := websocketManager.AddWebsocket(&wsReq, &config)
			if err != nil {
				log.Printf("err: %v", err)
			}
		} else {
			_, err := websocketManager.AddWebsocket(&wsReq, &config)
			assert.Error(t, err, "Expected error when websocket connection exceeds maximum connections")
		}
	}
}

// 测试websocket shutdown
func TestWebSocketsShutdown(t *testing.T) {
	limiter := testlimiter.NewTestLimiter(
		limiter.WithPeriodLimitArray([]limiter.PeriodLimit{
			{
				WsConnectPeriod: "1s",
				WsConnectTimes:  10,
			},
		},
		))
	websocketManager := manager.NewManager(
		manager.WithMaxConn(10),
		manager.WithMaxConnDuration(2*time.Minute),
		manager.WithConnLimiter(limiter),
		manager.WithCheckReConn(true),
	)

	wsReq := websocket.WebsocketRequest{
		Endpoint:       "wss://testnet.binance.vision/ws",
		ID:             "test",
		MessageHandler: handleMessage,
	}

	config := wsmanager.ExchangeWebsocketConfig{
		PingHandler: pingHandler,
		PongHandler: pongHandler,
	}
	uniq, err := websocketManager.AddWebsocket(&wsReq, &config)
	if err != nil {
		log.Printf("err: %v", err)
	}
	isConnected := websocketManager.IsConnected(uniq)
	assert.True(t, isConnected, "Expected websocket connection to be established")

	websocketManager.Shutdown()
	isConnected = websocketManager.IsConnected(uniq)
	assert.False(t, isConnected, "Expected websocket connection to be shutdown")
}

// 测试websocket重连机制
func TestWebSocketsReconnect(t *testing.T) {
	limiter := testlimiter.NewTestLimiter(
		limiter.WithPeriodLimitArray([]limiter.PeriodLimit{
			{
				WsConnectPeriod: "1s",
				WsConnectTimes:  10,
			},
		},
		))
	websocketManager := manager.NewManager(
		manager.WithMaxConn(10),
		manager.WithMaxConnDuration(1*time.Minute),
		manager.WithConnLimiter(limiter),
		manager.WithCheckReConn(true),
	)

	wsReq := websocket.WebsocketRequest{
		Endpoint:       "wss://testnet.binance.vision/ws/btcusdt@kline_1s",
		ID:             "test",
		MessageHandler: handleMessage,
	}

	config := wsmanager.ExchangeWebsocketConfig{
		PingHandler: pingHandler,
		PongHandler: pongHandler,
	}
	uniq, err := websocketManager.AddWebsocket(&wsReq, &config)
	if err != nil {
		log.Printf("err: %v", err)
	}
	isConnected := websocketManager.IsConnected(uniq)
	assert.True(t, isConnected, "Expected websocket connection to be established")

	time.Sleep(1 * time.Minute)
}

func handleMessage(message []byte) error {
	log.Printf("message: %v", string(message))
	return nil
}

func pingHandler(appData string, conn websocket.WebSocketConn) error {
	return nil
}

func pongHandler(appData string, conn websocket.WebSocketConn) error {
	return nil
}
