package bnmanager

import (
	"sync"
	"time"

	"github.com/go-gotop/kit/exchange"
	"github.com/go-gotop/kit/limiter/bnlimiter"
	"github.com/go-gotop/kit/requests/bnhttp"
	"github.com/go-gotop/kit/websocket"
	"github.com/go-gotop/kit/wsmanager"
	"github.com/go-gotop/kit/wsmanager/manager"
)

var (
	exitChan = make(chan struct{})
)

type listenKey struct {
	key        string
	createTime time.Time
}

type BnManager struct {
	client          *bnhttp.Client
	exchange        exchange.Exchange
	listenKeyExpire time.Duration
	wsm             wsmanager.WebsocketManager
	listenKeySets   map[string]listenKey // listenKey 集合
	mux             sync.Mutex
}

func NewBnManager(cli *bnhttp.Client) *BnManager {
	limiter := bnlimiter.NewBinanceLimiter()
	b := &BnManager{
		client:          cli,
		listenKeyExpire: 60 * time.Minute, // listenkey 60分钟过期，这里设置59分钟
		wsm: manager.NewManager(
			manager.WithMaxConn(1000),
			manager.WithMaxConnDuration(24*time.Hour-5*time.Minute),
			manager.WithConnLimiter(limiter),
			manager.WithCheckReConn(true),
		),
	}
	b.checkListenKey()
	return b
}

// 添加市场行情推送 websocket 连接
func (b *BnManager) AddMarketWebSocket(aReq *websocket.WebsocketRequest, instrumentType wsmanager.InstrumentType) (string, error) {
	b.mux.Lock()
	defer b.mux.Unlock()

	conf := &wsmanager.ExchangeWebsocketConfig{
		PingHandler: pingHandler,
		PongHandler: pongHandler,
	}

	uniq, err := b.addWebsocket(aReq, conf)
	if err != nil {
		return "", err
	}

	return uniq, nil
}

// 添加账户信息推送 websocket 连接
func (b *BnManager) AddAccountWebSocket(aReq *websocket.WebsocketRequest, instrumentType wsmanager.InstrumentType) (string, error) {
	b.mux.Lock()
	defer b.mux.Unlock()

	conf := &wsmanager.ExchangeWebsocketConfig{
		PingHandler: pingHandler,
		PongHandler: pongHandler,
	}
	// 生成 listenKey
	key, err := b.generateListenKey()
	generateTime := time.Now()

	if err != nil {
		return "", err
	}
	// 拼接 listenKey 到请求地址
	aReq.Endpoint += "/ws/" + key
	uniq, err := b.addWebsocket(aReq, conf)

	b.listenKeySets[uniq] = listenKey{
		key:        key,
		createTime: generateTime,
	}

	if err != nil {
		return "", err
	}

	return uniq, nil
}

// 关闭websocket，删除 listenKey
func (b *BnManager) CloseWebSocket(uniq string) error {
	err := b.wsm.CloseWebsocket(uniq)
	if err != nil {
		return err
	}
	b.mux.Lock()
	defer b.mux.Unlock()

	if lk, ok := b.listenKeySets[uniq]; ok {
		delete(b.listenKeySets, uniq)
		b.closeListenKey(lk.key)
	}
	return nil
}

func (b *BnManager) GetWebSocket(uniq string) websocket.Websocket {
	return b.wsm.GetWebsocket(uniq)
}

func (b *BnManager) Shutdown() {
	b.wsm.Shutdown()
	close(exitChan)
}

func (b *BnManager) addWebsocket(req *websocket.WebsocketRequest, conf *wsmanager.ExchangeWebsocketConfig) (string, error) {
	uniq, err := b.wsm.AddWebsocket(req, conf)
	if err != nil {
		return "", err
	}

	return uniq, nil
}

func (b *BnManager) generateListenKey(instrumentType wsmanager.InstrumentType) (string, error) {
	// r := &bnhttp.Request{
	// 	Method: "POST",
	// 	Endpoint: "/api/v3/userDataStream",
	// 	SecType: bnhttp.SecTypeAPIKey,
	// }
	return "", nil
}

func (b *BnManager) updateListenKey(instrumentType wsmanager.InstrumentType) error {
	return nil
}

func (b *BnManager) closeListenKey(instrumentType wsmanager.InstrumentType, key string) error {
	return nil
}

// 检查 listenKey 是否过期
func (b *BnManager) checkListenKey() {
	for {
		select {
		case <-exitChan:
			return
		default:
			b.mux.Lock()
			for _, lk := range b.listenKeySets {
				if time.Since(lk.createTime) >= b.listenKeyExpire-1*time.Minute {
					b.updateListenKey()
				}
			}
			b.mux.Unlock()
			time.Sleep(1 * time.Minute)
		}
	}
}

func pingHandler(appData string, conn websocket.WebSocketConn) error {
	return conn.WriteMessage(10, []byte(appData))
}

func pongHandler(appData string, conn websocket.WebSocketConn) error {
	return conn.WriteMessage(9, []byte(appData))
}
