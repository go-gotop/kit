package testing

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-gotop/kit/websocket"
	"github.com/go-gotop/kit/wsmanager"
	"github.com/go-gotop/kit/wsmanager/manager"
	"github.com/google/uuid"
	gwebsocket "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

const (
	bnSpotWsEndpoint    = "wss://stream.binance.com:9443/ws"
	bnFuturesWsEndpoint = "wss://fstream.binance.com/ws"
)

func messageHandler(message []byte) {
	// do something
	fmt.Println(string(message))
}

func errorHandle(err error) {
	// do something
	fmt.Println(err)
}

func pingHandler(appData string, conn websocket.WebSocketConn) error {
	fmt.Printf("pingHandler: %s\n", appData)
	return conn.WriteMessage(gwebsocket.PongMessage, []byte(appData))
}

func pongHandler(appData string, conn websocket.WebSocketConn) error {
	fmt.Printf("pongHandler: %s\n", appData)
	return conn.WriteMessage(gwebsocket.PingMessage, []byte(appData))
}

// 测试基础功能
func TestNewWsManger(t *testing.T) {
	m := manager.NewManager()
	assert.NotNil(t, m)

	uuid := uuid.New().String()

	m.AddWebsocket(&websocket.WebsocketRequest{
		Endpoint:       bnSpotWsEndpoint + "/btcusdt@trade",
		ID:             uuid,
		MessageHandler: messageHandler,
		ErrorHandler:   errorHandle,
	}, &wsmanager.WebsocketConfig{
		PingHandler: pingHandler,
		PongHandler: pongHandler,
	})

	time.Sleep(20 * time.Minute)

	err := m.CloseWebsocket(uuid)
	assert.Nil(t, err)

	err = m.Shutdown()
	assert.Nil(t, err)
}

// 测试最大连接数
func TestMaxConn(t *testing.T) {
	m := manager.NewManager(
		manager.WithMaxConn(2),
	)
	assert.NotNil(t, m)

	for i := 0; i < 3; i++ {
		uuid := uuid.New().String()
		err := m.AddWebsocket(&websocket.WebsocketRequest{
			Endpoint:       bnSpotWsEndpoint + "/btcusdt@trade",
			ID:             uuid,
			MessageHandler: messageHandler,
		}, nil)
		if i == 2 {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
	}

	time.Sleep(30 * time.Second)

	err := m.Shutdown()
	assert.Nil(t, err)
}

// 测试断开重连
func TestReConn(t *testing.T) {
	m := manager.NewManager(
		manager.WithCheckReConn(true),
	)
	assert.NotNil(t, m)

	uuid := uuid.New().String()

	m.AddWebsocket(&websocket.WebsocketRequest{
		Endpoint:       bnSpotWsEndpoint + "/btcusdt@trade",
		ID:             uuid,
		MessageHandler: messageHandler,
	}, nil)

	wsConn := m.GetWebsocket(uuid)
	wsConn.Disconnect()

	isConnect := m.IsConnected(uuid)
	assert.False(t, isConnect)

	time.Sleep(10 * time.Second)

	isConnect = m.IsConnected(uuid)
	assert.True(t, isConnect)
}

// 测试超过最大连接限制是否重新
func TestReConnMaxConn(t *testing.T) {
	m := manager.NewManager(
		manager.WithMaxConn(2),
		manager.WithCheckReConn(true),
		manager.WithMaxConnDuration(30 * time.Second),
	)
	assert.NotNil(t, m)

	for i := 0; i < 1; i++ {
		uuid := uuid.New().String()
		err := m.AddWebsocket(&websocket.WebsocketRequest{
			Endpoint:       bnSpotWsEndpoint + "/btcusdt@trade",
			ID:             uuid,
			MessageHandler: messageHandler,
		}, nil)
		if i == 2 {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
	}

	time.Sleep(40 * time.Second)
}
