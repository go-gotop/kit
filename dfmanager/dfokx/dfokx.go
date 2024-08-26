package dfokx

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/go-gotop/kit/dfmanager"
	"github.com/go-gotop/kit/exchange"
	"github.com/go-gotop/kit/limiter"
	"github.com/go-gotop/kit/websocket"
	"github.com/go-gotop/kit/wsmanager"
	"github.com/go-gotop/kit/wsmanager/manager"
	"github.com/go-kratos/kratos/v2/log"
	gwebsocket "github.com/gorilla/websocket"
)

const (
	okWsEndpoint = "wss://ws.okx.com:8443"
)

type wsSub struct {
	Op   string `json:"op"`
	Args []struct {
		Channel string `json:"channel"`
		InstID  string `json:"instId"`
	} `json:"args"`
}

func NewOkxDataFeed(limiter limiter.Limiter, opts ...Option) dfmanager.DataFeedManager {
	// 默认配置
	o := &options{
		logger:          log.NewHelper(log.DefaultLogger),
		maxConnDuration: 24*time.Hour - 5*time.Minute,
	}

	for _, opt := range opts {
		opt(o)
	}

	return &df{
		name:    exchange.BinanceExchange,
		opts:    o,
		limiter: limiter,
		wsm: manager.NewManager(
			manager.WithMaxConnDuration(o.maxConnDuration),
		),
	}
}

type df struct {
	name    string
	opts    *options
	limiter limiter.Limiter
	wsm     wsmanager.WebsocketManager
	mux     sync.Mutex
}

func (d *df) Name() string {
	return d.name
}

func (d *df) AddDataFeed(req *dfmanager.DataFeedRequest) error {

	conf := &wsmanager.WebsocketConfig{
		PingHandler: pingHandler,
		PongHandler: pongHandler,
	}

	endpoint := okWsEndpoint + "/ws/v5/business"
	wsHandler := func(message []byte) {
		te, err := toTradeEvent(message)
		if err != nil {
			if req.ErrorHandler != nil {
				req.ErrorHandler(err)
			}
			return
		}
		req.Event(te)
	}
	err := d.addWebsocket(&websocket.WebsocketRequest{
		ID:             req.ID,
		Endpoint:       endpoint,
		MessageHandler: wsHandler,
		ErrorHandler:   req.ErrorHandler,
	}, conf)
	if err != nil {
		return err
	}

	sub := wsSub{
		Op: "subscribe",
		Args: []struct {
			Channel string `json:"channel"`
			InstID  string `json:"instId"`
		}{
			{
				Channel: "trades",
				InstID:  req.Symbol,
			},
		},
	}

	str, err := json.Marshal(sub)
	if err != nil {
		return err
	}

	d.wsm.GetWebsocket(req.ID).WriteMessage(gwebsocket.BinaryMessage, []byte(str))

	return nil
}

func (d *df) AddMarketPriceDataFeed(req *dfmanager.MarkPriceRequest) error {
	return nil
}

func (d *df) AddKlineDataFeed(req *dfmanager.KlineRequest) error {
	return nil
}

func (d *df) CloseDataFeed(id string) error {
	return nil
}

func (d *df) DataFeedList() []string {
	return nil
}

func (d *df) Shutdown() error {
	return nil
}

func (d *df) addWebsocket(req *websocket.WebsocketRequest, conf *wsmanager.WebsocketConfig) error {
	err := d.wsm.AddWebsocket(req, conf)
	if err != nil {
		return err
	}
	return nil
}

func pingHandler(appData string, conn websocket.WebSocketConn) error {
	return conn.WriteMessage(gwebsocket.PongMessage, []byte(appData))
}

func pongHandler(appData string, conn websocket.WebSocketConn) error {
	return conn.WriteMessage(gwebsocket.PingMessage, []byte(appData))
}

func toTradeEvent(message []byte) (*exchange.TradeEvent, error) {
	return nil, nil
}
