package dfmock

import (
	"encoding/json"
	"errors"
	"fmt"
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

var (
	ErrLimitExceed = errors.New("websocket request too frequent, please try again later")
)

func NewMockDataFeed(limiter limiter.Limiter, opts ...Option) dfmanager.DataFeedManager {
	// 默认配置
	o := &options{
		wsEndpoint:      "ws://192.168.100.3:8072/ws/data",
		logger:          log.NewHelper(log.DefaultLogger),
		maxConnDuration: 24*time.Hour - 5*time.Minute,
	}

	for _, opt := range opts {
		opt(o)
	}

	return &df{
		name:    exchange.MockExchange,
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

	symbol = req.Symbol
	conf := &wsmanager.WebsocketConfig{
		PingHandler: pingHandler,
		PongHandler: pongHandler,
	}
	endpoint = fmt.Sprintf("%s?streams=trade&instrument=%s&symbol=%s&startTime=%v&endTime=%v", d.opts.wsEndpoint, req.Instrument, symbol, req.StartTime, req.EndTime)

	fn = spotToTradeEvent
	if req.Instrument == exchange.InstrumentTypeFutures {
		fn = funturesToTradeEvent
	}

	wsHandler := func(message []byte) {
		te, err := fn(message)
		if err != nil {
			req.ErrorHandler(err)
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
	return errors.New("not implemented")
	// var (
	// 	endpoint string
	// 	fn       func(message []byte) ([]*exchange.MarkPriceEvent, error)
	// )
	// d.mux.Lock()
	// defer d.mux.Unlock()

	// if !d.limiter.WsAllow() {
	// 	return ErrLimitExceed
	// }

	// conf := &wsmanager.WebsocketConfig{
	// 	PingHandler: pingHandler,
	// 	PongHandler: pongHandler,
	// }
	// switch req.Instrument {
	// case exchange.InstrumentTypeFutures:
	// 	endpoint = fmt.Sprintf("%s?streams=fundingrate&startTime=%v&endTime=%v", d.opts.wsEndpoint, req.StartTime, req.EndTime)
	// 	fn = futuresMarkPriceToMarkPrice
	// }
	// wsHandler := func(message []byte) {
	// 	te, err := fn(message)
	// 	if err != nil {
	// 		if req.ErrorHandler != nil {
	// 			req.ErrorHandler(err)
	// 		}
	// 		return
	// 	}
	// 	req.Event(te)
	// }
	// err := d.addWebsocket(&websocket.WebsocketRequest{
	// 	ID:             req.ID,
	// 	Endpoint:       endpoint,
	// 	MessageHandler: wsHandler,
	// 	ErrorHandler:   req.ErrorHandler,
	// }, conf)
	// if err != nil {
	// 	return err
	// }
	// return nil
}

func (d *df) AddKlineDataFeed(req *dfmanager.KlineRequest) error {
	var (
		endpoint string
		fn       func(message []byte) (*exchange.KlineEvent, error)
	)
	d.mux.Lock()
	defer d.mux.Unlock()

	if !d.limiter.WsAllow() {
		return ErrLimitExceed
	}

	conf := &wsmanager.WebsocketConfig{
		PingHandler: pingHandler,
		PongHandler: pongHandler,
	}

	endpoint = fmt.Sprintf("%s?streams=kline&instrument=%v&symbol=%v&period=%v&startTime=%v&endTime=%v", d.opts.wsEndpoint, req.Instrument, req.Symbol, req.Period, req.StartTime, req.EndTime)
	fn = klineToEvent
	fmt.Printf("endpoint: %v\n", endpoint)
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
	return conn.WriteMessage(10, []byte(appData))
}

func pongHandler(appData string, conn websocket.WebSocketConn) error {
	return conn.WriteMessage(9, []byte(appData))
}

func spotToTradeEvent(message []byte) (*exchange.TradeEvent, error) {
	e := &tradeEvent{}
	err := json.Unmarshal(message, e)
	if err != nil {
		return nil, err
	}

	te := &exchange.TradeEvent{
		TradeID:    fmt.Sprintf("%d", e.TradeID),
		Symbol:     e.Symbol,
		TradedAt:   e.TradeTime,
		Exchange:   exchange.MockExchange,
		Instrument: exchange.InstrumentTypeSpot,
	}
	size, err := decimal.NewFromString(e.Size)
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
	return te, nil
}

func funturesToTradeEvent(message []byte) (*exchange.TradeEvent, error) {
	e := &tradeEvent{}
	err := json.Unmarshal(message, e)
	if err != nil {
		return nil, err
	}

	te := &exchange.TradeEvent{
		TradeID:    fmt.Sprintf("%d", e.TradeID),
		Symbol:     e.Symbol,
		TradedAt:   e.TradeTime,
		Exchange:   exchange.MockExchange,
		Instrument: exchange.InstrumentTypeFutures,
	}
	size, err := decimal.NewFromString(e.Size)
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
	return te, nil
}

func futuresMarkPriceToMarkPrice(message []byte) ([]*exchange.MarkPriceEvent, error) {
	var e []*funtureMarkPriceEvent
	err := json.Unmarshal(message, &e)
	if err != nil {
		return nil, err
	}
	var list []*exchange.MarkPriceEvent
	for _, v := range e {
		markPrice, err := decimal.NewFromString(v.MarkPrice)
		if err != nil {
			markPrice = decimal.Zero
		}
		indexPrice, err := decimal.NewFromString(v.IndexPrice)
		if err != nil {
			indexPrice = decimal.Zero
		}
		estimatedSettlePrice, err := decimal.NewFromString(v.EstimatedSettlePrice)
		if err != nil {
			estimatedSettlePrice = decimal.Zero
		}
		lastFundingRate, err := decimal.NewFromString(v.LastFundingRate)
		if err != nil {
			lastFundingRate = decimal.Zero
		}

		te := &exchange.MarkPriceEvent{
			Symbol:               v.Symbol,
			MarkPrice:            markPrice,
			IndexPrice:           indexPrice,
			EstimatedSettlePrice: estimatedSettlePrice,
			LastFundingRate:      lastFundingRate,
			NextFundingTime:      v.NextFundingTime,
			Time:                 v.Time,
			IsSettlement:         v.IsSettlement,
		}
		list = append(list, te)
	}

	return list, nil
}

func klineToEvent(message []byte) (*exchange.KlineEvent, error) {
	var e *klineEvent
	err := json.Unmarshal(message, &e)
	if err != nil {
		return nil, err
	}
	open, err := decimal.NewFromString(e.Open)
	if err != nil {
		return nil, err
	}
	high, err := decimal.NewFromString(e.High)
	if err != nil {
		return nil, err
	}
	low, err := decimal.NewFromString(e.Low)
	if err != nil {
		return nil, err
	}
	close, err := decimal.NewFromString(e.Close)
	if err != nil {
		return nil, err
	}
	volume, err := decimal.NewFromString(e.Volume)
	if err != nil {
		return nil, err
	}
	// quoteAssetVolume, err := decimal.NewFromString(e.QuoteAssetVolume)
	// if err != nil {
	// 	return nil, err
	// }
	// takerBuyBaseAssetVolume, err := decimal.NewFromString(e.TakerBuyBaseAssetVolume)
	// if err != nil {
	// 	return nil, err
	// }
	// takerBuyQuoteAssetVolume, err := decimal.NewFromString(e.TakerBuyQuoteAssetVolume)
	// if err != nil {
	// 	return nil, err
	// }

	te := &exchange.KlineEvent{
		Symbol:    e.Symbol,
		OpenTime:  e.OpenTime,
		Open:      open,
		High:      high,
		Low:       low,
		Close:     close,
		Volume:    volume,
		CloseTime: e.CloseTime,
		// QuoteAssetVolume:         quoteAssetVolume,
		// NumberOfTrades:           e.NumberOfTrades,
		// TakerBuyBaseAssetVolume:  takerBuyBaseAssetVolume,
		// TakerBuyQuoteAssetVolume: takerBuyQuoteAssetVolume,
	}
	return te, nil
}
