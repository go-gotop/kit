package bnmanager

import (
	"context"
	"sync"
	"time"

	"github.com/go-gotop/kit/exchange"
	"github.com/go-gotop/kit/limiter/bnlimiter"
	"github.com/go-gotop/kit/requests/bnhttp"
	"github.com/go-gotop/kit/websocket"
	"github.com/go-gotop/kit/wsmanager"
	"github.com/go-gotop/kit/wsmanager/manager"
)

const (
	bnSpotEndpoint    = "https://api.binance.com"
	bnFuturesEndpoint = "https://fapi.binance.com"
)

var (
	exitChan = make(chan struct{})
)

type listenKey struct {
	key            string
	createTime     time.Time
	instrumentType exchange.InstrumentType
}

type BnManager struct {
	mux             sync.Mutex
	client          *bnhttp.Client
	wsm             wsmanager.WebsocketManager
	listenKeyExpire time.Duration
	listenKeySets   map[string]listenKey // listenKey 集合, 合约一个，现货一个

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
func (b *BnManager) AddMarketWebSocket(aReq *websocket.WebsocketRequest, instrumentType exchange.InstrumentType) (string, error) {
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
func (b *BnManager) AddAccountWebSocket(aReq *websocket.WebsocketRequest, instrumentType exchange.InstrumentType) (string, error) {
	b.mux.Lock()
	defer b.mux.Unlock()

	conf := &wsmanager.ExchangeWebsocketConfig{
		PingHandler: pingHandler,
		PongHandler: pongHandler,
	}
	// 生成 listenKey
	key, err := b.generateListenKey(instrumentType)
	generateTime := time.Now()

	if err != nil {
		return "", err
	}
	// 拼接 listenKey 到请求地址
	aReq.Endpoint += "/ws/" + key
	uniq, err := b.addWebsocket(aReq, conf)

	b.listenKeySets[uniq] = listenKey{
		key:            key,
		createTime:     generateTime,
		instrumentType: instrumentType,
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
		b.closeListenKey(lk.instrumentType)
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

func (b *BnManager) generateListenKey(instrumentType exchange.InstrumentType) (string, error) {
	r := &bnhttp.Request{
		Method:  "POST",
		SecType: bnhttp.SecTypeAPIKey,
	}

	if instrumentType == exchange.InstrumentTypeSpot {
		r.Endpoint = "/api/v3/userDataStream"
		b.client.SetApiEndpoint(bnSpotEndpoint)
	} else if instrumentType == exchange.InstrumentTypeFutures {
		r.Endpoint = "/fapi/v1/listenKey"
		b.client.SetApiEndpoint(bnFuturesEndpoint)
	}

	data, err := b.client.CallAPI(context.Background(), r)

	if err != nil {
		return "", err
	}

	var res struct {
		ListenKey string `json:"listenKey"`
	}

	err = bnhttp.Json.Unmarshal(data, &res)
	if err != nil {
		return "", err
	}

	return res.ListenKey, nil
}

func (b *BnManager) updateListenKey(instrumentType exchange.InstrumentType) error {
	r := &bnhttp.Request{
		Method:  "PUT",
		SecType: bnhttp.SecTypeAPIKey,
	}

	if instrumentType == exchange.InstrumentTypeSpot {
		r.Endpoint = "/api/v3/userDataStream"
		b.client.SetApiEndpoint(bnSpotEndpoint)
	} else if instrumentType == exchange.InstrumentTypeFutures {
		r.Endpoint = "/fapi/v1/listenKey"
		b.client.SetApiEndpoint(bnFuturesEndpoint)
	}

	_, err := b.client.CallAPI(context.Background(), r)
	if err != nil {
		return err
	}

	return nil
}

func (b *BnManager) closeListenKey(instrumentType exchange.InstrumentType) error {
	r := &bnhttp.Request{
		Method:  "DELETE",
		SecType: bnhttp.SecTypeAPIKey,
	}

	if instrumentType == exchange.InstrumentTypeSpot {
		r.Endpoint = "/api/v3/userDataStream"
		b.client.SetApiEndpoint(bnSpotEndpoint)
	} else if instrumentType == exchange.InstrumentTypeFutures {
		r.Endpoint = "/fapi/v1/listenKey"
		b.client.SetApiEndpoint(bnFuturesEndpoint)
	}

	_, err := b.client.CallAPI(context.Background(), r)
	if err != nil {
		return err
	}

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
					b.updateListenKey(lk.instrumentType)
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
