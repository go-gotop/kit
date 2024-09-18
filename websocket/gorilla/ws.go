package gorilla

import (
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-gotop/kit/websocket"
	gwebsocket "github.com/gorilla/websocket"
)

func NewGorillaWebsocket(conn websocket.WebSocketConn, config *websocket.WebsocketConfig) *GorillaWebsocket {
	g := &GorillaWebsocket{
		conn:    conn,
		config:  config,
		closeCh: make(chan struct{}),
		doneCh:  make(chan struct{}),
	}
	return g
}

// GorillaWebsocket 是 Websocket 接口的实现
type GorillaWebsocket struct {
	messageCount uint64
	isConnected  bool
	conn         websocket.WebSocketConn
	config       *websocket.WebsocketConfig
	req          *websocket.WebsocketRequest
	closeCh      chan struct{}
	doneCh       chan struct{}
	closeOnce    sync.Once
	doneOnce     sync.Once
	connectTime  time.Time
}

func (w *GorillaWebsocket) Connect(req *websocket.WebsocketRequest) error {
	if err := w.conn.Dial(req.Endpoint, nil); err != nil {
		close(w.doneCh)
		return err
	}
	w.configure()
	go w.readMessages(req)
	w.req = req
	w.isConnected = true
	w.connectTime = time.Now()
	w.messageCount = 0
	if req.ConnectedHandler != nil {
		req.ConnectedHandler(req.ID, w.conn)
	}

	return nil
}

func (w *GorillaWebsocket) configure() {
	if w.config.PingHandler != nil {
		w.conn.SetPingHandler(w.config.PingHandler)
	}
	if w.config.PongHandler != nil {
		w.conn.SetPongHandler(w.config.PongHandler)
	}
	// 应用其他配置...
}

func (w *GorillaWebsocket) readMessages(req *websocket.WebsocketRequest) {
	defer w.doneOnce.Do(func() {
		close(w.doneCh)
	}) // 确保此方法退出时标记doneCh为已完成
	for {
		select {
		case <-w.closeCh: // 如果收到关闭信号，则立即退出循环
			return
		default:
			_, message, err := w.conn.ReadMessage()
			if err != nil {
				// 当遇到错误时，首先检查是否因为连接已关闭
				select {
				case <-w.closeCh: // 如果已经收到关闭信号，则不处理错误
				default:
					// 读取消息时发生错误，标识连接已断开
					w.isConnected = false
					if w.conn != nil && req != nil && req.ErrorHandler != nil { // 增加对 req 和 ErrorHandler 的检查
						req.ErrorHandler(err)
					}
				}
				return // 退出循环
			}
			req.MessageHandler(message) // 处理接收到的消息
			atomic.AddUint64(&w.messageCount, 1)
		}
	}
}

func (w *GorillaWebsocket) ID() string {
	return w.req.ID
}

func (w *GorillaWebsocket) Disconnect() error {
	var err error
	w.closeOnce.Do(func() {
		close(w.closeCh) // 通知读协程退出
		if w.conn != nil {
			err = w.conn.Close() // 关闭WebSocket连接
		}
	})
	w.isConnected = false
	<-w.doneCh // 确保读协程已经结束
	return err
}

func (w *GorillaWebsocket) Reconnect() error {
	w.Disconnect()
	// 等待读循环完全停止
	<-w.doneCh

	// 重置通道，准备新的连接周期
	w.closeCh = make(chan struct{})
	w.doneCh = make(chan struct{})
	w.closeOnce = sync.Once{} // 重置sync.Once，以便再次使用
	w.doneOnce = sync.Once{}

	// 重新建立连接
	return w.Connect(w.req)
}

func (w *GorillaWebsocket) IsConnected() bool {
	return w.isConnected
}

func (w *GorillaWebsocket) WriteMessage(messageType int, data []byte) error {
	return w.conn.WriteMessage(messageType, data)
}

func (w *GorillaWebsocket) GetCurrentRate() int {
	elapsed := time.Since(w.connectTime).Seconds()
	if elapsed == 0 {
		return 0
	}
	// 使用atomic.LoadUint64确保读取的原子性
	count := atomic.LoadUint64(&w.messageCount)
	rate := float64(count) / elapsed
	return int(rate) // 返回每秒消息数
}

func (w *GorillaWebsocket) ConnectionDuration() time.Duration {
	return time.Since(w.connectTime)
}
