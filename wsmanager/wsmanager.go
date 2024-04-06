package wsmanager

import (
	"github.com/go-gotop/kit/websocket"
)

type InstrumentType string // 合约类型 ：永续合约、现货
type LinkType string       // 链接类型 ：市场、账户

var (
	SpotInstrumentType   InstrumentType = "spot"
	FutureInstrumentType InstrumentType = "future"
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
	Reconnect(uniq string) error
	Shutdown()
}

// ExchangeWsManager 是 交易所websocket 管理接口
type ExchangeWsManager interface {
	AddMarketWebSocket(req *websocket.WebsocketRequest, instrumentType InstrumentType) (string, error)
	AddAccountWebSocket(req *websocket.WebsocketRequest, instrumentType InstrumentType) (string, error)
	CloseWeSocket(uniq string) error
	GetWebsocket(uniq string) websocket.Websocket
	Shutdown()
}
