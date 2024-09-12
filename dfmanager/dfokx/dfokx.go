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

	df := &df{
		name:    exchange.BinanceExchange,
		opts:    o,
		limiter: limiter,
		wsm: manager.NewManager(
			manager.WithMaxConnDuration(o.maxConnDuration),
		),
		exitChan: make(chan struct{}),
	}

	go df.keepAlive()

	return df
}

type df struct {
	exitChan chan struct{}
	name     string
	opts     *options
	limiter  limiter.Limiter
	wsm      wsmanager.WebsocketManager
	mux      sync.RWMutex
}

func (d *df) Name() string {
	return d.name
}

func (d *df) AddDataFeed(req *dfmanager.DataFeedRequest) error {
	d.mux.Lock()
	defer d.mux.Unlock()

	if !d.limiter.WsAllow() {
		return manager.ErrLimitExceed
	}

	conf := &wsmanager.WebsocketConfig{}

	endpoint := okWsEndpoint + "/ws/v5/business"
	wsHandler := func(instrument exchange.InstrumentType) func(message []byte) {
		return func(message []byte) {
			if string(message) == "pong" {
				// 每隔20s发送ping过去，预期会收到pong
				return
			}
			j, err := okhttp.NewJSON(message)
			if err != nil {
				d.opts.logger.Error("new json error", err)
				return
			}
			if j.Get("event").MustString() == "error" {
				req.ErrorHandler(errors.New(j.Get("msg").MustString()))
				return
			}

			if j.Get("event").MustString() != "" {
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
		ID:               req.ID,
		Endpoint:         endpoint,
		MessageHandler:   wsHandler(req.Instrument),
		ErrorHandler:     d.errorHandler(req.ID, req),
		ConnectedHandler: d.connectedTradeAllHandler(req),
	}, conf)
	if err != nil {
		return err
	}

	return nil
}

func (d *df) AddMarketPriceDataFeed(req *dfmanager.MarkPriceRequest) error {
	d.mux.Lock()
	defer d.mux.Unlock()

	if !d.limiter.WsAllow() {
		return manager.ErrLimitExceed
	}

	conf := &wsmanager.WebsocketConfig{}

	endpoint := okWsEndpoint + "/ws/v5/public"
	wsHandler := func(instrument exchange.InstrumentType) func(message []byte) {
		return func(message []byte) {
			if string(message) == "pong" {
				// 每隔20s发送ping过去，预期会收到pong
				return
			}
			j, err := okhttp.NewJSON(message)
			if err != nil {
				d.opts.logger.Error("new json error", err)
				return
			}
			if j.Get("event").MustString() == "error" {
				req.ErrorHandler(errors.New(j.Get("msg").MustString()))
				return
			}

			if j.Get("event").MustString() != "" {
				return
			}

			te, err := toMarkPriceEvent(message, instrument)
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
		ID:               req.ID,
		Endpoint:         endpoint,
		MessageHandler:   wsHandler(req.Instrument),
		ErrorHandler:     d.errorMarkPriceHandler(req.ID, req),
		ConnectedHandler: d.connectedMarketPriceHandler(req),
	}, conf)
	if err != nil {
		return err
	}

	return nil
}

func (d *df) AddKlineDataFeed(req *dfmanager.KlineRequest) error {
	return nil
}

func (d *df) CloseDataFeed(id string) error {
	d.mux.RLock()
	defer d.mux.RUnlock()

	err := d.wsm.CloseWebsocket(id)
	if err != nil {
		return err
	}

	return nil
}

func (d *df) DataFeedList() []string {
	d.mux.RLock()
	defer d.mux.RUnlock()

	mapList := d.wsm.GetWebsockets()
	list := make([]string, 0, len(mapList))
	for k := range mapList {
		list = append(list, k)
	}
	return list
}

// 连接成功后订阅交易数据
func (d *df) connectedTradeAllHandler(req *dfmanager.DataFeedRequest) func(id string, conn websocket.WebSocketConn) {
	return func(id string, conn websocket.WebSocketConn) {
		// ws := d.wsm.GetWebsocket(id)
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
			if req.ErrorHandler != nil {
				req.ErrorHandler(err)
			}
		}

		err = conn.WriteMessage(gwebsocket.TextMessage, str)
		if err != nil {
			if req.ErrorHandler != nil {
				req.ErrorHandler(err)
			}
		}
	}
}

// 连接成功后订阅交易数据
func (d *df) connectedMarketPriceHandler(req *dfmanager.MarkPriceRequest) func(id string, conn websocket.WebSocketConn) {
	return func(id string, conn websocket.WebSocketConn) {
		// ws := d.wsm.GetWebsocket(id)
		sub := wsSub{
			Op: "subscribe",
			Args: []struct {
				Channel string `json:"channel"`
				InstID  string `json:"instId"`
			}{
				{
					Channel: "mark-price",
					InstID:  req.Symbol,
				},
			},
		}

		str, err := json.Marshal(sub)
		if err != nil {
			if req.ErrorHandler != nil {
				req.ErrorHandler(err)
			}
		}

		err = conn.WriteMessage(gwebsocket.TextMessage, str)
		if err != nil {
			if req.ErrorHandler != nil {
				req.ErrorHandler(err)
			}
		}
	}
}

func (d *df) errorHandler(id string, req *dfmanager.DataFeedRequest) func(err error) {
	return func(err error) {
		if req.ErrorHandler != nil {
			if strings.Contains(err.Error(), "close 4004") {
				go d.wsm.Reconnect(id)
				req.ErrorHandler(manager.ErrServerClosedConn)
			} else if err == manager.ErrServerClosedConn {
				go d.wsm.Reconnect(id)
				req.ErrorHandler(err)
			} else {
				req.ErrorHandler(err)
			}
		}
		if !d.wsm.GetWebsocket(id).IsConnected() {
			// 开启一个计时器，10秒后再次检查连接状态，如果连接已经关闭，则删除连接
			time.AfterFunc(10*time.Second, func() {
				if !d.wsm.GetWebsocket(id).IsConnected() {
					req.ErrorHandler(manager.ErrReconnectFailed)
					d.wsm.CloseWebsocket(id)
				}
			})
		}
	}
}

func (d *df) errorMarkPriceHandler(id string, req *dfmanager.MarkPriceRequest) func(err error) {
	return func(err error) {
		if req.ErrorHandler != nil {
			if strings.Contains(err.Error(), "close 4004") {
				go d.wsm.Reconnect(id)
				req.ErrorHandler(manager.ErrServerClosedConn)
			} else if err == manager.ErrServerClosedConn {
				go d.wsm.Reconnect(id)
				req.ErrorHandler(err)
			} else {
				req.ErrorHandler(err)
			}
		}
		if !d.wsm.GetWebsocket(id).IsConnected() {
			// 开启一个计时器，10秒后再次检查连接状态，如果连接已经关闭，则删除连接
			time.AfterFunc(10*time.Second, func() {
				if !d.wsm.GetWebsocket(id).IsConnected() {
					d.wsm.CloseWebsocket(id)
				}
			})
		}
	}
}

func (d *df) Shutdown() error {
	d.mux.RLock()
	defer d.mux.RUnlock()

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

func (d *df) keepAlive() {
	for {
		select {
		case <-d.exitChan:
			return
		case <-time.After(20 * time.Second):
			d.mux.RLock()
			for _, ws := range d.wsm.GetWebsockets() {
				err := ws.WriteMessage(gwebsocket.TextMessage, []byte("ping"))
				if err != nil {
					d.opts.logger.Error("write ping message error", err)
				}
			}
			d.mux.RUnlock()
		}
	}
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

func toMarkPriceEvent(message []byte, instrument exchange.InstrumentType) (*exchange.MarkPriceEvent, error) {
	e := &okxMarkPriceEvent{}
	err := json.Unmarshal(message, e)
	if err != nil {
		return nil, err
	}

	if len(e.Data) == 0 {
		return nil, errors.New("data is empty")
	}

	trade := e.Data[0]

	// 字符串转成int64
	ts, err := strconv.ParseInt(trade.Timestamp, 10, 64)
	if err != nil {
		return nil, err
	}

	mk, err := decimal.NewFromString(trade.MarkPx)
	if err != nil {
		return nil, err
	}

	te := &exchange.MarkPriceEvent{
		Symbol:    trade.InstID,
		Time:      ts,
		MarkPrice: mk,
	}
	return te, nil
}
