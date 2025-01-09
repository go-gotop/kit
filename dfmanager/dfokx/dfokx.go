package dfokx

import (
	"encoding/json"
	"errors"
	"fmt"
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

type wsInstTypeSub struct {
	Op   string `json:"op"`
	Args []struct {
		Channel  string `json:"channel"`
		InstType string `json:"instType"`
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
		streams:  make(map[string]dfmanager.Stream),
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
	streams  map[string]dfmanager.Stream
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

	d.streams[req.ID] = dfmanager.Stream{
		UUID:        req.ID,
		Instrument:  req.Instrument,
		Symbol:      req.Symbol,
		DataType:    "trade",
		IsConnected: true,
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

	d.streams[req.ID] = dfmanager.Stream{
		UUID:        req.ID,
		Instrument:  req.Instrument,
		Symbol:      req.Symbol,
		DataType:    "markprice",
		IsConnected: true,
	}

	return nil
}

func (d *df) AddMarketKlineDataFeed(req *dfmanager.KlineMarketRequest) error {
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

			te, err := toMarkKlineEvent(message, instrument)
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
		ErrorHandler:     d.errorMarkKlineHandler(req.ID, req),
		ConnectedHandler: d.connectedMarketKlineHandler(req),
	}, conf)
	if err != nil {
		return err
	}

	d.streams[req.ID] = dfmanager.Stream{
		UUID:        req.ID,
		Instrument:  req.Instrument,
		Symbol:      req.Symbol,
		DataType:    "markkline",
		IsConnected: true,
	}

	return nil
}

func (d *df) AddKlineDataFeed(req *dfmanager.KlineRequest) error {
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

			te, err := toKlineEvent(message, instrument)
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
		ErrorHandler:   d.errorKlineHandler(req.ID, req),
	}, conf)
	if err != nil {
		return err
	}

	d.streams[req.ID] = dfmanager.Stream{
		UUID:        req.ID,
		Instrument:  req.Instrument,
		Symbol:      req.Symbol,
		DataType:    "kline",
		IsConnected: true,
	}

	return nil
}

func (d *df) AddSymbolUpdateDataFeed(req *dfmanager.SymbolUpdateRequest) error {
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

			te, err := toSymbolUpdateEvent(message, instrument)
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
		ErrorHandler:     d.errorSymbolUpdateHandler(req.ID, req),
		ConnectedHandler: d.connectedSymbolUpdateHandler(req),
	}, conf)
	if err != nil {
		return err
	}

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

func (d *df) DataFeedList() []dfmanager.Stream {
	d.mux.RLock()
	defer d.mux.RUnlock()

	list := make([]dfmanager.Stream, 0, len(d.streams))
	for _, v := range d.streams {
		v.IsConnected = d.wsm.IsConnected(v.UUID)
		list = append(list, v)
	}
	return list
}

func (d *df) WriteMessage(id string, message []byte) error {
	conn := d.wsm.GetWebsocket(id)
	if conn == nil {
		return errors.New("websocket not found")
	}
	return conn.WriteMessage(gwebsocket.TextMessage, message)
}

// 连接成功后订阅交易数据
func (d *df) connectedTradeAllHandler(req *dfmanager.DataFeedRequest) func(id string, conn websocket.WebSocketConn) {
	return func(id string, conn websocket.WebSocketConn) {
		// ws := d.wsm.GetWebsocket(id)
		fmt.Println("tradeall链接成功回调:", req.Symbol)
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
		fmt.Println("markprice链接成功回调:", req.Symbol)
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

func (d *df) connectedMarketKlineHandler(req *dfmanager.KlineMarketRequest) func(id string, conn websocket.WebSocketConn) {
	return func(id string, conn websocket.WebSocketConn) {
		// ws := d.wsm.GetWebsocket(id)
		fmt.Println("markkline链接成功回调:", req.Symbol)
		sub := wsSub{
			Op: "subscribe",
			Args: []struct {
				Channel string `json:"channel"`
				InstID  string `json:"instId"`
			}{
				{
					Channel: "mark-price-candle" + req.Period,
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

func (d *df) connectedSymbolUpdateHandler(req *dfmanager.SymbolUpdateRequest) func(id string, conn websocket.WebSocketConn) {
	return func(id string, conn websocket.WebSocketConn) {
		// ws := d.wsm.GetWebsocket(id)
		fmt.Println("symbolupdate链接成功回调:", req.Instrument)
		sub := wsInstTypeSub{
			Op: "subscribe",
			Args: []struct {
				Channel  string `json:"channel"`
				InstType string `json:"instType"`
			}{
				{
					Channel:  "instruments",
					InstType: string(req.Instrument),
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
			fmt.Println("连接错误回调:", req.Symbol, err)
			req.ErrorHandler(err)
		}
		go d.wsm.Reconnect(id)
		// 开启一个计时器，10秒后再次检查连接状态，如果连接已经关闭，则删除连接
		time.AfterFunc(10*time.Second, func() {
			if !d.wsm.GetWebsocket(id).IsConnected() {
				if req.ErrorHandler != nil {
					req.ErrorHandler(manager.ErrReconnectFailed)
				}
				d.wsm.CloseWebsocket(id)
			}
		})
	}
}

func (d *df) errorMarkPriceHandler(id string, req *dfmanager.MarkPriceRequest) func(err error) {
	return func(err error) {
		if req.ErrorHandler != nil {
			req.ErrorHandler(err)
		}
		go d.wsm.Reconnect(id)
		// 开启一个计时器，10秒后再次检查连接状态，如果连接已经关闭，则删除连接
		time.AfterFunc(10*time.Second, func() {
			if !d.wsm.GetWebsocket(id).IsConnected() {
				if req.ErrorHandler != nil {
					req.ErrorHandler(manager.ErrReconnectFailed)
				}
				d.wsm.CloseWebsocket(id)
			}
		})
	}
}

func (d *df) errorMarkKlineHandler(id string, req *dfmanager.KlineMarketRequest) func(err error) {
	return func(err error) {
		if req.ErrorHandler != nil {
			req.ErrorHandler(err)
		}
		go d.wsm.Reconnect(id)
		// 开启一个计时器，10秒后再次检查连接状态，如果连接已经关闭，则删除连接
		time.AfterFunc(10*time.Second, func() {
			if !d.wsm.GetWebsocket(id).IsConnected() {
				if req.ErrorHandler != nil {
					req.ErrorHandler(manager.ErrReconnectFailed)
				}
				d.wsm.CloseWebsocket(id)
			}
		})
	}
}

func (d *df) errorKlineHandler(id string, req *dfmanager.KlineRequest) func(err error) {
	return func(err error) {
		if req.ErrorHandler != nil {
			req.ErrorHandler(err)
		}
		go d.wsm.Reconnect(id)
		// 开启一个计时器，10秒后再次检查连接状态，如果连接已经关闭，则删除连接
		time.AfterFunc(10*time.Second, func() {
			if !d.wsm.GetWebsocket(id).IsConnected() {
				if req.ErrorHandler != nil {
					req.ErrorHandler(manager.ErrReconnectFailed)
				}
				d.wsm.CloseWebsocket(id)
			}
		})
	}
}

func (d *df) errorSymbolUpdateHandler(id string, req *dfmanager.SymbolUpdateRequest) func(err error) {
	return func(err error) {
		if req.ErrorHandler != nil {
			req.ErrorHandler(err)
		}
		go d.wsm.Reconnect(id)
		// 开启一个计时器，10秒后再次检查连接状态，如果连接已经关闭，则删除连接
		time.AfterFunc(10*time.Second, func() {
			if !d.wsm.GetWebsocket(id).IsConnected() {
				if req.ErrorHandler != nil {
					req.ErrorHandler(manager.ErrReconnectFailed)
				}
				d.wsm.CloseWebsocket(id)
			}
		})
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
		case <-time.After(10 * time.Second):
			for _, ws := range d.wsm.GetWebsockets() {
				d.mux.RLock()
				if ws != nil {
					err := ws.WriteMessage(gwebsocket.TextMessage, []byte("ping"))
					if err != nil {
						d.opts.logger.Error("write ping message error", err)
						ws.Reconnect()
					}
					d.mux.RUnlock()
				}
			}
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

func toMarkKlineEvent(message []byte, instrument exchange.InstrumentType) (*exchange.KlineMarketEvent, error) {
	e := &okxMarkKlineEvent{}
	err := json.Unmarshal(message, e)
	if err != nil {
		return nil, err
	}

	if len(e.Data) == 0 {
		return nil, errors.New("data is empty")
	}

	trade := e.Data[0]

	// 字符串转成int64
	ts, err := strconv.ParseInt(trade[0], 10, 64)
	if err != nil {
		return nil, err
	}

	open, err := decimal.NewFromString(trade[1])
	if err != nil {
		return nil, err
	}

	high, err := decimal.NewFromString(trade[2])
	if err != nil {
		return nil, err
	}

	low, err := decimal.NewFromString(trade[3])
	if err != nil {
		return nil, err
	}

	close, err := decimal.NewFromString(trade[4])
	if err != nil {
		return nil, err
	}

	te := &exchange.KlineMarketEvent{
		Symbol:   e.Arg.InstID,
		OpenTime: ts,
		Open:     open,
		High:     high,
		Low:      low,
		Close:    close,
		Confirm:  trade[5],
	}
	return te, nil
}

func toKlineEvent(message []byte, instrument exchange.InstrumentType) (*exchange.KlineEvent, error) {
	e := &okxMarkKlineEvent{}
	err := json.Unmarshal(message, e)
	if err != nil {
		return nil, err
	}

	if len(e.Data) == 0 {
		return nil, errors.New("data is empty")
	}

	trade := e.Data[0]

	// 字符串转成int64
	ts, err := strconv.ParseInt(trade[0], 10, 64)
	if err != nil {
		return nil, err
	}

	open, err := decimal.NewFromString(trade[1])
	if err != nil {
		return nil, err
	}

	high, err := decimal.NewFromString(trade[2])
	if err != nil {
		return nil, err
	}

	low, err := decimal.NewFromString(trade[3])
	if err != nil {
		return nil, err
	}

	close, err := decimal.NewFromString(trade[4])
	if err != nil {
		return nil, err
	}

	volume, err := decimal.NewFromString(trade[5])
	if err != nil {
		return nil, err
	}

	te := &exchange.KlineEvent{
		Symbol:         e.Arg.InstID,
		OpenTime:       ts,
		Open:           open,
		High:           high,
		Low:            low,
		Close:          close,
		Volume:         volume,
		InstrumentType: instrument,
		Confirm:        trade[6],
	}
	return te, nil
}

func toSymbolUpdateEvent(message []byte, instrument exchange.InstrumentType) ([]*exchange.SymbolUpdateEvent, error) {
	e := &okxSymbolUpdateEvent{}
	err := json.Unmarshal(message, e)
	if err != nil {
		return nil, err
	}

	if len(e.Data) == 0 {
		return nil, errors.New("data is empty")
	}

	result := make([]*exchange.SymbolUpdateEvent, 0, len(e.Data))

	for _, v := range e.Data {
		// instrumentType := exchange.InstrumentType(v.InstType)
		minsz, err := decimal.NewFromString(v.MinSz)
		if err != nil {
			return nil, err
		}
		ctval, err := decimal.NewFromString(v.CtVal)
		if err != nil {
			return nil, err
		}
		ctmult, err := decimal.NewFromString(v.CtMult)
		if err != nil {
			return nil, err
		}
		maxLmtSz, err := decimal.NewFromString(v.MaxLmtSz)
		if err != nil {
			maxLmtSz = decimal.Zero
		}
		maxsz := maxLmtSz
		maxMktSz, err := decimal.NewFromString(v.MaxMktSz)
		if err != nil {
			maxMktSz = decimal.Zero
		}

		if maxsz.IsZero() {
			maxsz = maxMktSz
		}

		listTime, err := strconv.ParseInt(v.ListTime, 10, 64)
		if err != nil {
			listTime = 0
		}

		expTime, err := strconv.ParseInt(v.ExpTime, 10, 64)
		if err != nil {
			expTime = 0
		}

		pricePrecision, err := findFirstNonZeroDigitAfterDecimal(v.TickSz)
		if err != nil {
			return nil, err
		}

		sizePrecision, err := findFirstNonZeroDigitAfterDecimal(v.LotSz)
		if err != nil {
			return nil, err
		}

		te := &exchange.SymbolUpdateEvent{
			InstrumentType: instrument,
			OriginalSymbol: v.InstID,
			OriginalAsset:  v.BaseCcy,
			MinSize:        minsz,
			MaxSize:        maxsz,
			PricePrecision: pricePrecision,
			SizePrecision:  sizePrecision,
			CtVal:          ctval,
			CtMult:         ctmult,
			ListTime:       listTime,
			ExpTime:        expTime,
			State:          v.State,
		}

		result = append(result, te)
	}

	return result, nil
}

func findFirstNonZeroDigitAfterDecimal(value string) (int32, error) {
	// Check if the input is a valid string type
	if value == "" {
		return 0, errors.New("unsupported type")
	}

	// Find the position of the decimal point
	dotIndex := strings.Index(value, ".")
	if dotIndex == -1 {
		return 0, nil // No decimal point, so no digits after it
	}

	// Traverse the string after the decimal point to find the first non-zero digit
	for i := dotIndex + 1; i < len(value); i++ {
		if value[i] != '0' {
			// Calculate the position of the first non-zero digit after the decimal point
			return int32(i - dotIndex), nil
		}
	}

	// No non-zero digits found after the decimal point
	return 0, nil
}
