package dfokx

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-gotop/kit/dfmanager"
	"github.com/go-gotop/kit/exchange"
	"github.com/go-gotop/kit/limiter"
	"github.com/go-gotop/kit/requests/okhttp"
	"github.com/go-gotop/kit/websocket"
	"github.com/go-gotop/kit/wsmanager"
	"github.com/go-gotop/kit/wsmanager/manager"
	"github.com/go-kratos/kratos/v2/log"
	gwebsocket "github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
)

const (
	okWsEndpoint = "wss://ws.okx.com:8443"
)

var (
	ErrLimitExceed = errors.New("websocket request too frequent, please try again later")
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
	d.mux.Lock()
	defer d.mux.Unlock()

	if !d.limiter.WsAllow() {
		return ErrLimitExceed
	}

	conf := &wsmanager.WebsocketConfig{
		PingHandler: pingHandler,
		PongHandler: pongHandler,
	}

	endpoint := okWsEndpoint + "/ws/v5/business"
	wsHandler := func(instrument exchange.InstrumentType) func(message []byte) {
		return func(message []byte) {
			j, err := okhttp.NewJSON(message)
			if err != nil {
				d.opts.logger.Error("order new json error", err)
				return
			}
			if j.Get("e").MustString() == "error" {
				req.ErrorHandler(errors.New(j.Get("msg").MustString()))
				return
			}

			if j.Get("e").MustString() != "" {
				return
			}

			te, err := toTradeEvent(message, instrument)
			if err != nil {
				if req.ErrorHandler != nil {
					req.ErrorHandler(err)
				}
				return
			}
			req.Event(te)
		}
	}

	err := d.addWebsocket(&websocket.WebsocketRequest{
		ID:             req.ID,
		Endpoint:       endpoint,
		MessageHandler: wsHandler(req.Instrument),
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
				Channel: "trades-all",
				InstID:  req.Symbol,
			},
		},
	}

	str, err := json.Marshal(sub)
	if err != nil {
		return err
	}

	time.Sleep(2 * time.Second)

	err = d.wsm.GetWebsocket(req.ID).WriteMessage(gwebsocket.TextMessage, str)
	if err != nil {
		return err
	}

	return nil
}

func (d *df) AddMarketPriceDataFeed(req *dfmanager.MarkPriceRequest) error {
	return nil
}

func (d *df) AddKlineDataFeed(req *dfmanager.KlineRequest) error {
	return nil
}

func (d *df) CloseDataFeed(id string) error {
	d.mux.Lock()
	defer d.mux.Unlock()

	err := d.wsm.CloseWebsocket(id)
	if err != nil {
		return err
	}

	return nil
}

func (d *df) DataFeedList() []string {
	mapList := d.wsm.GetWebsockets()
	list := make([]string, 0, len(mapList))
	for k := range mapList {
		list = append(list, k)
	}
	return list
}

func (d *df) Shutdown() error {
	err := d.wsm.Shutdown()
	if err != nil {
		return err
	}
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

func toTradeEvent(message []byte, instrument exchange.InstrumentType) (*exchange.TradeEvent, error) {
	e := &okxTradeAllEvent{}
	err := json.Unmarshal(message, e)
	if err != nil {
		return nil, err
	}

	if len(e.Data) == 0 {
		return nil, errors.New("data is empty")
	}

	trade := e.Data[0]

	te := &exchange.TradeEvent{
		TradeID:    trade.TradeID,
		Symbol:     e.Arg.InstID,
		Exchange:   exchange.OkxExchange,
		Instrument: instrument,
	}

	tradeTime, err := strconv.ParseInt(trade.TradeTime, 10, 64)
	if err != nil {
		return nil, err
	}
	te.TradedAt = tradeTime

	size, err := decimal.NewFromString(trade.Quantity)
	if err != nil {
		return nil, err
	}
	te.Size = size

	p, err := decimal.NewFromString(trade.Price)
	if err != nil {
		return nil, err
	}
	te.Price = p

	te.Side = exchange.SideType(strings.ToUpper(trade.Side))

	return te, nil
}
