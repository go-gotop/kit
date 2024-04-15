package bnmanager

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-gotop/kit/exchange"
	"github.com/go-gotop/kit/excmanager"
	"github.com/go-gotop/kit/limiter"
	"github.com/go-gotop/kit/limiter/bnlimiter"
	"github.com/go-gotop/kit/requests/bnhttp"
	"github.com/go-gotop/kit/websocket"
	"github.com/go-gotop/kit/wsmanager"
	"github.com/go-gotop/kit/wsmanager/manager"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/shopspring/decimal"
)

const (
	bnSpotWsEndpoint    = "wss://stream.binance.com:9443/ws"
	bnFuturesWsEndpoint = "wss://fstream.binance.com/ws"
)

type listenKey struct {
	uniq           string
	key            string
	createTime     time.Time
	instrumentType exchange.InstrumentType
}

type BnManager struct {
	opts          *options
	exitChan      chan struct{}
	mux           sync.Mutex
	client        *bnhttp.Client
	wsm           wsmanager.WebsocketManager
	listenKeySets map[string]*listenKey // listenKey 集合, 合约一个，现货一个
}

func NewBnManager(cli *bnhttp.Client, opts ...Option) *BnManager {
	o := &options{
		logger:               log.NewHelper(log.DefaultLogger),
		maxConn:              1000,
		maxConnDuration:      24*time.Hour - 5*time.Minute,
		listenKeyExpire:      58 * time.Minute,
		checkListenKeyPeriod: 1 * time.Minute,
	}
	for _, opt := range opts {
		opt(o)
	}
	limiter := bnlimiter.NewBinanceLimiter(
		limiter.WithPeriodLimitArray([]limiter.PeriodLimit{
			{
				WsConnectPeriod:         "5m",
				WsConnectTimes:          300,
				SpotCreateOrderPeriod:   "10s",
				SpotCreateOrderTimes:    100,
				FutureCreateOrderPeriod: "10s",
				FutureCreateOrderTimes:  300,
				SpotNormalRequestPeriod: "5m",
				SpotNormalRequestTimes:  61000,
			},
			{
				FutureCreateOrderPeriod: "1m",
				FutureCreateOrderTimes:  1200,
			},
		}),
	)
	b := &BnManager{
		opts:          o,
		exitChan:      make(chan struct{}),
		client:        cli,
		listenKeySets: make(map[string]*listenKey),
		wsm: manager.NewManager(
			manager.WithMaxConn(o.maxConn),
			manager.WithMaxConnDuration(o.maxConnDuration),
			manager.WithConnLimiter(limiter),
			manager.WithCheckReConn(true),
		),
	}
	return b
}

// 添加市场行情推送 websocket 连接
func (b *BnManager) DataFeed(req *excmanager.DataFeedRequest) error {
	var (
		endpoint string
		symbol   string
		fn       func(message []byte) (*exchange.TradeEvent, error)
	)
	b.mux.Lock()
	defer b.mux.Unlock()

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
	_, err := b.addWebsocket(&websocket.WebsocketRequest{
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

// 添加账户信息推送 websocket 连接
func (b *BnManager) OrderFeed(req *excmanager.OrderFeedRequest) (string, error) {
	b.mux.Lock()
	defer b.mux.Unlock()

	conf := &wsmanager.WebsocketConfig{
		PingHandler: pingHandler,
		PongHandler: pongHandler,
	}
	// 生成 listenKey
	key, err := b.generateListenKey(req.Instrument)

	if err != nil {
		return "", err
	}
	generateTime := time.Now()
	// 拼接 listenKey 到请求地址
	endpoint := fmt.Sprintf("%s/%s", bnSpotWsEndpoint, key)
	if req.Instrument == exchange.InstrumentTypeFutures {
		endpoint = fmt.Sprintf("%s/%s", bnFuturesWsEndpoint, key)
	}
	wsHandler := func(message []byte) {
		j, err := bnhttp.NewJSON(message)
		if err != nil {
			b.opts.logger.Error("order new json error", err)
			return
		}
		switch j.Get("e").MustString() {
		case "executionReport":
			event := &bnSpotWsOrderUpdateEvent{}
			err = bnhttp.Json.Unmarshal(message, event)
			if err != nil {
				b.opts.logger.Error("order unmarshal error", err)
				return
			}
			oe, err := swoueToOrderEvent(event)
			if err != nil {
				b.opts.logger.Error("order to order event error", err)
				return
			}
			req.Event(oe)
		case "ORDER_TRADE_UPDATE":
			event := &bnFuturesWsUserDataEvent{}
			err = bnhttp.Json.Unmarshal(message, event)
			if err != nil {
				b.opts.logger.Error("order unmarshal error", err)
				return
			}
			oe, err := fwoueToOrderEvent(&event.OrderTradeUpdate)
			if err != nil {
				b.opts.logger.Error("order to order event error", err)
				return
			}
			req.Event(oe)
		}
	}
	uniq, err := b.addWebsocket(&websocket.WebsocketRequest{
		Endpoint:       endpoint,
		ID:             fmt.Sprintf("binance.order.%s", req.Instrument),
		MessageHandler: wsHandler,
	}, conf)

	b.listenKeySets[uniq] = &listenKey{
		uniq:           uniq,
		key:            key,
		createTime:     generateTime,
		instrumentType: req.Instrument,
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
		b.closeListenKey(lk)
	}
	return nil
}

func (b *BnManager) GetWebSocket(uniq string) websocket.Websocket {
	return b.wsm.GetWebsocket(uniq)
}

func (b *BnManager) IsConnected(uniq string) bool {
	return b.wsm.IsConnected(uniq)
}

func (b *BnManager) Shutdown() {
	b.wsm.Shutdown()
	close(b.exitChan)
}

func (b *BnManager) addWebsocket(req *websocket.WebsocketRequest, conf *wsmanager.WebsocketConfig) (string, error) {
	uniq, err := b.wsm.AddWebsocket(req, conf)
	if err != nil {
		return "", err
	}
	return uniq, nil
}

func (b *BnManager) generateListenKey(instrumentType exchange.InstrumentType) (string, error) {
	r := &bnhttp.Request{
		Method:  http.MethodPost,
		SecType: bnhttp.SecTypeAPIKey,
	}

	if instrumentType == exchange.InstrumentTypeSpot {
		r.Endpoint = "/api/v3/userDataStream"
	} else if instrumentType == exchange.InstrumentTypeFutures {
		r.Endpoint = "/fapi/v1/listenKey"
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

func (b *BnManager) updateListenKey(lk *listenKey) error {
	r := &bnhttp.Request{
		Method:  http.MethodPut,
		SecType: bnhttp.SecTypeAPIKey,
	}

	if lk.instrumentType == exchange.InstrumentTypeSpot {
		r.Endpoint = "/api/v3/userDataStream"
		r.SetFormParam("listenKey", lk.key)
	} else if lk.instrumentType == exchange.InstrumentTypeFutures {
		r.Endpoint = "/fapi/v1/listenKey"
	}

	_, err := b.client.CallAPI(context.Background(), r)
	if err != nil {
		return err
	}

	// 更新listenKey createTime
	lk.createTime = time.Now()

	return nil
}

func (b *BnManager) closeListenKey(lk *listenKey) error {
	r := &bnhttp.Request{
		Method:  http.MethodDelete,
		SecType: bnhttp.SecTypeAPIKey,
	}

	if lk.instrumentType == exchange.InstrumentTypeSpot {
		r.Endpoint = "/api/v3/userDataStream"
	} else if lk.instrumentType == exchange.InstrumentTypeFutures {
		r.Endpoint = "/fapi/v1/listenKey"
	}

	_, err := b.client.CallAPI(context.Background(), r)
	if err != nil {
		return err
	}

	// 删除 listenKey
	delete(b.listenKeySets, lk.uniq)

	return nil
}

// 检查 listenKey 是否过期
func (b *BnManager) CheckListenKey() {
	for {
		select {
		case <-b.exitChan:
			return
		default:
			b.mux.Lock()
			for _, lk := range b.listenKeySets {
				if time.Since(lk.createTime) >= b.opts.listenKeyExpire {
					b.updateListenKey(lk)
				}
			}
			b.mux.Unlock()
			time.Sleep(b.opts.checkListenKeyPeriod)
		}
	}
}

func pingHandler(appData string, conn websocket.WebSocketConn) error {
	return conn.WriteMessage(10, []byte(appData))
}

func pongHandler(appData string, conn websocket.WebSocketConn) error {
	return conn.WriteMessage(9, []byte(appData))
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

func swoueToOrderEvent(event *bnSpotWsOrderUpdateEvent) (*exchange.OrderEvent, error) {
	price, err := decimal.NewFromString(event.Price)
	if err != nil {
		return nil, err
	}
	volume, err := decimal.NewFromString(event.Volume)
	if err != nil {
		return nil, err
	}
	latestVolume, err := decimal.NewFromString(event.LatestVolume)
	if err != nil {
		return nil, err
	}
	filledVolume, err := decimal.NewFromString(event.FilledVolume)
	if err != nil {
		return nil, err
	}
	latestPrice, err := decimal.NewFromString(event.LatestPrice)
	if err != nil {
		return nil, err
	}
	feeCost, err := decimal.NewFromString(event.FeeCost)
	if err != nil {
		return nil, err
	}
	filledQuoteVolume, err := decimal.NewFromString(event.FilledQuoteVolume)
	if err != nil {
		return nil, err
	}
	ps := exchange.PositionSideLong
	if event.Side == "SELL" {
		ps = exchange.PositionSideShort
	}
	avgPrice := decimal.Zero
	if filledQuoteVolume.GreaterThan(decimal.Zero) && filledVolume.GreaterThan(decimal.Zero) {
		avgPrice = filledQuoteVolume.Div(filledVolume)
	}
	return &exchange.OrderEvent{
		PositionSide:    ps,
		Exchange:        exchange.BinanceExchange,
		Symbol:          event.Symbol,
		ClientOrderID:   event.ClientOrderId,
		ExecutionType:   event.ExecutionType,
		Status:          event.Status,
		OrderID:         fmt.Sprintf("%d", event.Id),
		TransactionTime: event.TransactionTime,
		IsMaker:         event.IsMaker,
		Side:            exchange.SideType(event.Side),
		Type:            exchange.OrderType(event.Type),
		Instrument:      exchange.InstrumentTypeSpot,
		Volume:          volume,
		Price:           price,
		LatestVolume:    latestVolume,
		FilledVolume:    filledVolume,
		LatestPrice:     latestPrice,
		FeeAsset:        event.FeeAsset,
		FeeCost:         feeCost,
		AvgPrice:        avgPrice,
	}, nil
}

func fwoueToOrderEvent(event *bnFuturesWsOrderUpdateEvent) (*exchange.OrderEvent, error) {
	price, err := decimal.NewFromString(event.OriginalPrice)
	if err != nil {
		return nil, err
	}
	volume, err := decimal.NewFromString(event.OriginalQty)
	if err != nil {
		return nil, err
	}
	latestVolume, err := decimal.NewFromString(event.LastFilledQty)
	if err != nil {
		return nil, err
	}
	filledVolume, err := decimal.NewFromString(event.AccumulatedFilledQty)
	if err != nil {
		return nil, err
	}
	latestPrice, err := decimal.NewFromString(event.LastFilledPrice)
	if err != nil {
		return nil, err
	}
	feeCost, err := decimal.NewFromString(event.Commission)
	if err != nil {
		return nil, err
	}
	avg, err := decimal.NewFromString(event.AveragePrice)
	if err != nil {
		return nil, err
	}
	ps := exchange.PositionSideLong
	if event.PositionSide == "SHORT" {
		ps = exchange.PositionSideShort
	}
	return &exchange.OrderEvent{
		PositionSide:    ps,
		Exchange:        exchange.BinanceExchange,
		Symbol:          event.Symbol,
		ClientOrderID:   event.ClientOrderID,
		ExecutionType:   event.ExecutionType,
		Status:          event.Status,
		OrderID:         fmt.Sprintf("%d", event.ID),
		TransactionTime: event.TradeTime,
		IsMaker:         event.IsMaker,
		Side:            exchange.SideType(event.Side),
		Type:            exchange.OrderType(event.Type),
		Instrument:      exchange.InstrumentTypeFutures,
		Volume:          volume,
		Price:           price,
		LatestVolume:    latestVolume,
		FilledVolume:    filledVolume,
		LatestPrice:     latestPrice,
		FeeAsset:        event.CommissionAsset,
		FeeCost:         feeCost,
		AvgPrice:        avg,
	}, nil
}
