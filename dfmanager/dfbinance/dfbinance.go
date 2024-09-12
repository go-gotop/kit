package dfbinance

import (
	"encoding/json"
	"errors"
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
	gwebsocket "github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
)

var _ dfmanager.DataFeedManager = (*df)(nil)

const (
	bnSpotWsEndpoint         = "wss://stream.binance.com:9443/ws"
	bnFuturesWsEndpoint      = "wss://fstream.binance.com/ws"
	bnFunturesStreamEndpoint = "wss://fstream.binance.com/stream"
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
	var (
		endpoint string
		symbol   string
		fn       func(message []byte) (*exchange.TradeEvent, error)
	)
	d.mux.Lock()
	defer d.mux.Unlock()

	if !d.limiter.WsAllow() {
		return manager.ErrLimitExceed
	}

	symbol = strings.ToLower(req.Symbol)
	conf := &wsmanager.WebsocketConfig{
		PingHandler: pingHandler,
		PongHandler: pongHandler,
	}
	switch req.Instrument {
	case exchange.InstrumentTypeSpot:
		endpoint = fmt.Sprintf("%s/%s@trade", bnSpotWsEndpoint, symbol)
		fn = spotToTradeEvent
	case exchange.InstrumentTypeMargin:
		endpoint = fmt.Sprintf("%s/%s@trade", bnSpotWsEndpoint, symbol)
		fn = marginToTradeEvent
	case exchange.InstrumentTypeFutures:
		endpoint = fmt.Sprintf("%s/%s@aggTrade", bnFuturesWsEndpoint, symbol)
		fn = futuresToTradeEvent
	}
	wsHandler := func(message []byte) {
		te, err := fn(message)
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
	return nil
}

func (d *df) AddMarketPriceDataFeed(req *dfmanager.MarkPriceRequest) error {
	var (
		endpoint string
		fn       func(message []byte) (*exchange.MarkPriceEvent, error)
	)
	d.mux.Lock()
	defer d.mux.Unlock()

	if !d.limiter.WsAllow() {
		return manager.ErrLimitExceed
	}

	conf := &wsmanager.WebsocketConfig{
		PingHandler: pingHandler,
		PongHandler: pongHandler,
	}
	switch req.Instrument {
	case exchange.InstrumentTypeFutures:
		symbol := strings.ToLower(req.Symbol)
		endpoint = fmt.Sprintf("%s?streams=%s@markPrice@1s", bnFunturesStreamEndpoint, symbol)
		fn = futuresMarkPriceToMarkPrice
	}
	wsHandler := func(message []byte) {
		te, err := fn(message)
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
	return nil
}

func (d *df) AddKlineDataFeed(req *dfmanager.KlineRequest) error {
	return errors.New("not implemented")
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

func spotToTradeEvent(message []byte) (*exchange.TradeEvent, error) {
	e := &binanceSpotTradeEvent{}
	err := json.Unmarshal(message, e)
	if err != nil {
		return nil, err
	}

	te := &exchange.TradeEvent{
		TradeID:    fmt.Sprintf("%d", e.TradeID),
		Symbol:     e.Symbol,
		TradedAt:   e.TradeTime,
		Exchange:   exchange.BinanceExchange,
		Instrument: exchange.InstrumentTypeSpot,
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
	te.Side = exchange.SideTypeBuy
	if e.IsBuyerMaker {
		te.Side = exchange.SideTypeSell
	}
	return te, nil
}

func marginToTradeEvent(message []byte) (*exchange.TradeEvent, error) {
	e := &binanceSpotTradeEvent{}
	err := json.Unmarshal(message, e)
	if err != nil {
		return nil, err
	}

	te := &exchange.TradeEvent{
		TradeID:    fmt.Sprintf("%d", e.TradeID),
		Symbol:     e.Symbol,
		TradedAt:   e.TradeTime,
		Exchange:   exchange.BinanceExchange,
		Instrument: exchange.InstrumentTypeMargin,
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
	te.Side = exchange.SideTypeBuy
	if e.IsBuyerMaker {
		te.Side = exchange.SideTypeSell
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
		TradeID:    fmt.Sprintf("%d", e.AggregateTradeID),
		Symbol:     e.Symbol,
		TradedAt:   e.TradeTime,
		Exchange:   exchange.BinanceExchange,
		Instrument: exchange.InstrumentTypeFutures,
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
	te.Side = exchange.SideTypeBuy
	if e.Maker {
		te.Side = exchange.SideTypeSell
	}
	return te, nil
}

func futuresMarkPriceToMarkPrice(message []byte) (*exchange.MarkPriceEvent, error) {
	var e binanceFuturesMarkPriceSingleStream
	err := json.Unmarshal(message, &e)
	if err != nil {
		return nil, err
	}
	data := e.Data
	markPrice, err := decimal.NewFromString(data.MarkPrice)
	if err != nil {
		markPrice = decimal.Zero
	}
	indexPrice, err := decimal.NewFromString(data.IndexPrice)
	if err != nil {
		indexPrice = decimal.Zero
	}
	estimatedSettlePrice, err := decimal.NewFromString(data.EstimatedSettlePrice)
	if err != nil {
		estimatedSettlePrice = decimal.Zero
	}
	lastFundingRate, err := decimal.NewFromString(data.LastFundingRate)
	if err != nil {
		lastFundingRate = decimal.Zero
	}

	te := &exchange.MarkPriceEvent{
		Symbol:               data.Symbol,
		MarkPrice:            markPrice,
		IndexPrice:           indexPrice,
		EstimatedSettlePrice: estimatedSettlePrice,
		LastFundingRate:      lastFundingRate,
		NextFundingTime:      data.NextFundingTime,
		Time:                 data.Time,
	}

	return te, nil

}
