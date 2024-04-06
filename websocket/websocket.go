package websocket

import (
	"time"
	"net/http"
)

type WebSocketConn interface {
	Dial(endpoint string, requestHeader http.Header) error
	ReadMessage() (messageType int, p []byte, err error)
	WriteMessage(messageType int, data []byte) error
	SetPingHandler(h func(appData string) error)
	SetPongHandler(h func(appData string) error)
	Close() error
}

// WebsocketConfig 结构体定义了WebSocket实例的配置选项
type WebsocketConfig struct {
	PingHandler func(appData string) error
	PongHandler func(appData string) error
}

type WebsocketRequest struct {
	// Endpoint 是Websocket服务器的地址
	Endpoint string

	// ID 是Websocket连接的唯一标识符
	ID string

	// MessageHandler 是Websocket消息处理函数
	MessageHandler func([]byte) error

	// ErrorHandler 是Websocket错误处理函数
	ErrorHandler func(id string, err error)
}

// Websocket 接口定义了基本的连接管理操作
type Websocket interface {
	// Connect 方法用于建立Websocket连接
	// req 参数是连接请求的相关信息
	Connect(req *WebsocketRequest) error

	// Disconnect 方法用于关闭Websocket连接
	Disconnect() error

	// Reconnect 方法用于重新建立Websocket连接
	Reconnect() error
	// IsConnected 方法用于检查Websocket连接是否处于活跃状态
	// 返回 true 表示连接是活跃的，false 表示连接已经关闭或尚未建立
	IsConnected() bool

	// GetCurrentRate 方法用于获取当前的通讯速率
	// 返回值是每秒传输的字节数
	GetCurrentRate() int

	// GetConnectionDuration 方法用于获取当前连接的持续时间
	ConnectionDuration() time.Duration
}
