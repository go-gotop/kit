package manager

import (
	"errors"
	"sync"
	"time"

	"github.com/go-gotop/kit/websocket"
	"github.com/go-gotop/kit/websocket/gorilla"
	"github.com/go-gotop/kit/wsmanager"
)

var (
	// 错误定义
	ErrMaxConnReached = errors.New("max connection reached")
	ErrWSNotFound     = errors.New("websocket not found")
)

type Manager struct {
	exitChan         chan struct{}                  // 退出通道
	config           *connConfig                    // 连接配置
	currentConnCount int                            // 当前连接数
	mux              sync.Mutex                     // 互斥锁
	wsSets           map[string]websocket.Websocket // websocket 集合
}

func NewManager(opts ...ConnConfig) *Manager {
	config := &connConfig{
		maxConn:         1000,
		maxConnDuration: 24 * time.Hour,
		// connLimiter:     nil,
		isCheckReConn: true,
	}

	for _, opt := range opts {
		opt(config)
	}

	m := &Manager{
		exitChan:         make(chan struct{}),
		config:           config,
		currentConnCount: 0,
		mux:              sync.Mutex{},
		wsSets:           make(map[string]websocket.Websocket),
	}

	if config.isCheckReConn {
		go m.checkConnection()
	}

	return m
}

func (b *Manager) AddWebsocket(req *websocket.WebsocketRequest, conf *wsmanager.WebsocketConfig) error {
	b.mux.Lock()
	defer b.mux.Unlock()

	// 最大连接数限制
	if b.currentConnCount >= b.config.maxConn {
		return ErrMaxConnReached
	}

	conn := gorilla.NewGorillaWebSocketConn()

	pingh := func(appData string) error {
		return nil
	}

	pongh := func(appData string) error {
		return nil
	}

	// ping pong 处理函数
	if conf != nil && conf.PingHandler != nil {
		pingh = func(appData string) error {
			return conf.PingHandler(appData, conn)
		}
	}

	if conf != nil && conf.PongHandler != nil {
		pongh = func(appData string) error {
			return conf.PongHandler(appData, conn)
		}
	}

	ws := gorilla.NewGorillaWebsocket(conn, &websocket.WebsocketConfig{
		PingHandler: pingh,
		PongHandler: pongh,
	})

	err := ws.Connect(req)
	if err != nil {
		return err
	}

	b.currentConnCount++
	b.wsSets[req.ID] = ws
	return nil
}

func (b *Manager) CloseWebsocket(uniq string) error {
	b.mux.Lock()
	defer b.mux.Unlock()

	ws := b.wsSets[uniq]
	if ws == nil {
		return ErrWSNotFound
	}
	delete(b.wsSets, uniq)
	ws.Disconnect()
	b.currentConnCount--
	return nil
}

func (b *Manager) GetWebsocket(uniq string) websocket.Websocket {
	return b.wsSets[uniq]
}

func (b *Manager) GetWebsockets() map[string]websocket.Websocket {
	return b.wsSets
}

func (b *Manager) IsConnected(uniq string) bool {
	ws := b.wsSets[uniq]
	if ws == nil {
		return false
	}
	return ws.IsConnected()
}

func (b *Manager) Reconnect(uniq string) error {
	ws := b.wsSets[uniq]
	if ws == nil {
		return ErrWSNotFound
	}
	err := ws.Reconnect()
	if err != nil {
		return err
	}
	return nil
}

func (b *Manager) Shutdown() error {
	close(b.exitChan)
	b.mux.Lock()
	defer b.mux.Unlock()

	var err error

	for key, ws := range b.wsSets {
		err = ws.Disconnect()
		if err == nil {
			delete(b.wsSets, key) // 删除映射中的连接
			b.currentConnCount--
		}
	}
	if err != nil {
		return err
	}

	return nil
}

func (b *Manager) checkConnection() {
	for {
		select {
		case <-b.exitChan:
			return
		default:
			b.mux.Lock()
			for _, ws := range b.wsSets {
				if !ws.IsConnected() ||
					ws.ConnectionDuration() > b.config.maxConnDuration {
					if err := ws.Reconnect(); err != nil {
						b.config.logger.Errorf("reconnect websocket error: %s", err)
					} else {
						b.config.logger.Infof("reconnect websocket success")
					}
				}
			}
			b.mux.Unlock()
			time.Sleep(1 * time.Second)
		}
	}
}
