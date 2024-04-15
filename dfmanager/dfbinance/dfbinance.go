package dfbinance

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-gotop/kit/dfmanager"
	"github.com/go-gotop/kit/exchange"
	"github.com/go-gotop/kit/limiter"
	"github.com/go-gotop/kit/websocket"
	"github.com/go-gotop/kit/wsmanager"
	"github.com/go-gotop/kit/wsmanager/manager"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/shopspring/decimal"
)

var _ dfmanager.DataFeedManager = (*df)(nil)

const (
	bnSpotWsEndpoint    = "wss://stream.binance.com:9443/ws"
	bnFuturesWsEndpoint = "wss://fstream.binance.com/ws"
)

func NewBinanceDataFeed(limiter limiter.Limiter, opts ...Option) dfmanager.DataFeedManager {
	// 默认配置
	o := &options{
		logger:          log.NewHelper(log.DefaultLogger),
		maxConnDuration: 24*time.Hour - 5*time.Minute,
	}

	for _, opt := range opts {
		opt(o)
	}

	return &df{
		name: "Binance",
		opts: o,
		wsm: manager.NewManager(
			manager.WithMaxConnDuration(o.maxConnDuration),
			manager.WithConnLimiter(limiter),
		),
	}
}

type df struct {
	name     string
	opts     *options
	wsm      wsmanager.WebsocketManager
	mux      sync.Mutex
}

func (d *df) Name() string {
	return d.name
}

func (d *df) AddDataFeed(req *dfmanager.DataFeedRequest) error {
	var (
		endpoint string
		symbol   string
		fn       func(message []byte) (*exchange.TradeEvent, error)
	)
	d.mux.Lock()
	defer d.mux.Unlock()

	symbol = strings.ToLower(exchange.ReverseBinanceSymbols[req.Symbol])
	conf := &wsmanager.WebsocketConfig{
		PingHandler: pingHandler,
		PongHandler: pongHandler,
	}
	switch req.Instrument {
	case exchange.InstrumentTypeSpot:
		endpoint = fmt.Sprintf("%s/%s@trade", bnSpotWsEndpoint, symbol)
		fn = spotToTradeEvent
	case exchange.InstrumentTypeFutures:
		endpoint = fmt.Sprintf("%s/%s@aggTrade", bnFuturesWsEndpoint, symbol)
		fn = futuresToTradeEvent
	}
	wsHandler := func(message []byte) {
		te, err := fn(message)
		if err != nil {
			req.ErrorHandler(err)
			return
		}
		req.Event(te)
	}
	_, err := d.addWebsocket(&websocket.WebsocketRequest{
		ID:             req.ID,
		Endpoint:       endpoint,
		MessageHandler: wsHandler,
		ErrorHandler:   req.ErrorHandler,
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

func (d *df) addWebsocket(req *websocket.WebsocketRequest, conf *wsmanager.WebsocketConfig) (string, error) {
	uniq, err := d.wsm.AddWebsocket(req, conf)
	if err != nil {
		return "", err
	}
	return uniq, nil
}

func pingHandler(appData string, conn websocket.WebSocketConn) error {
	return conn.WriteMessage(10, []byte(appData))
}

func pongHandler(appData string, conn websocket.WebSocketConn) error {
	return conn.WriteMessage(9, []byte(appData))
}

func spotToTradeEvent(message []byte) (*exchange.TradeEvent, error) {
	e := &binanceSpotTradeEvent{}
	err := json.Unmarshal(message, e)
	if err != nil {
		return nil, err
	}

	te := &exchange.TradeEvent{
		TradeID:  uint64(e.TradeID),
		Symbol:   e.Symbol,
		TradedAt: e.TradeTime,
	}
	size, err := decimal.NewFromString(e.Quantity)
	if err != nil {
		return nil, err
	}
	te.Size = size

	p, err := decimal.NewFromString(e.Price)
	if err != nil {
		return nil, err
	}
	te.Price = p
	if e.IsBuyerMaker {
		te.Side = true
	}
	return te, nil
}

func futuresToTradeEvent(message []byte) (*exchange.TradeEvent, error) {
	e := &binanceFuturesTradeEvent{}
	err := json.Unmarshal(message, e)
	if err != nil {
		return nil, err
	}
	te := &exchange.TradeEvent{
		TradeID:  uint64(e.AggregateTradeID),
		Symbol:   e.Symbol,
		TradedAt: e.TradeTime,
	}
	size, err := decimal.NewFromString(e.Quantity)
	if err != nil {
		return nil, err
	}
	te.Size = size

	p, err := decimal.NewFromString(e.Price)
	if err != nil {
		return nil, err
	}
	te.Price = p

	if e.Maker {
		te.Side = true
	}
	return te, nil
}
