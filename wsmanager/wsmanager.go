package wsmanager

import (
	"github.com/go-gotop/kit/exchange"
	"github.com/go-gotop/kit/websocket"
)

type ExchangeWebsocketConfig struct {
	PingHandler func(appData string, conn websocket.WebSocketConn) error
	PongHandler func(appData string, conn websocket.WebSocketConn) error
}

// WebsocketManager 是 websocket 管理接口
type WebsocketManager interface {
	AddWebsocket(req *websocket.WebsocketRequest, conf *ExchangeWebsocketConfig) (string, error)
	CloseWebsocket(uniq string) error
	GetWebsocket(uniq string) websocket.Websocket
	IsConnected(uniq string) bool
	Reconnect(uniq string) error
	Shutdown()
}

// ExchangeWsManager 是 交易所websocket 管理接口
type ExchangeWsManager interface {
	AddMarketWebSocket(req *websocket.WebsocketRequest, instrumentType exchange.InstrumentType) (string, error)
	AddAccountWebSocket(req *websocket.WebsocketRequest, instrumentType exchange.InstrumentType) (string, error)
	CloseWeSocket(uniq string) error
	GetWebsocket(uniq string) websocket.Websocket
	IsConnected(uniq string) bool
	Shutdown()
}
