package wsmanager

import (
	"github.com/go-gotop/kit/websocket"
)

type WebsocketConfig struct {
	PingHandler func(appData string, conn websocket.WebSocketConn) error
	PongHandler func(appData string, conn websocket.WebSocketConn) error
}

// WebsocketManager 是 websocket 管理接口
type WebsocketManager interface {
	AddWebsocket(req *websocket.WebsocketRequest, conf *WebsocketConfig) (string, error)
	CloseWebsocket(uniq string) error
	GetWebsocket(uniq string) websocket.Websocket
	GetWebsockets() map[string]websocket.Websocket
	IsConnected(uniq string) bool
	Reconnect(uniq string) error
	Shutdown() error
}
