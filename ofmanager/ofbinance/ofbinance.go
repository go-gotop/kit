package ofbinance

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-gotop/kit/exchange"
	"github.com/go-gotop/kit/limiter"
	"github.com/go-gotop/kit/ofmanager"
	"github.com/go-gotop/kit/requests/bnhttp"
	"github.com/go-gotop/kit/websocket"
	"github.com/go-gotop/kit/wsmanager"
	"github.com/go-gotop/kit/wsmanager/manager"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/shopspring/decimal"
)

var _ ofmanager.OrderFeedManager = (*of)(nil)

var (
	ErrLimitExceed = errors.New("websocket request too frequent, please try again later")
)

const (
	bnSpotWsEndpoint    = "wss://stream.binance.com:9443/ws"
	bnFuturesWsEndpoint = "wss://fstream.binance.com/ws"
)

// TODO: 限流器放在 ofbinance 做调用，不传入 wsmanager
func NewBinanceOrderFeed(cli *bnhttp.Client, limiter limiter.Limiter, opts ...Option) ofmanager.OrderFeedManager {
	// 默认配置
	o := &options{
		logger:               log.NewHelper(log.DefaultLogger),
		maxConnDuration:      24*time.Hour - 5*time.Minute,
		listenKeyExpire:      58 * time.Minute,
		checkListenKeyPeriod: 1 * time.Minute,
	}

	for _, opt := range opts {
		opt(o)
	}

	of := &of{
		name:          "Binance",
		opts:          o,
		client:        cli,
		limiter:       limiter,
		listenKeySets: make(map[string]*listenKey),
		wsm: manager.NewManager(
			manager.WithMaxConnDuration(o.maxConnDuration),
			// manager.WithConnLimiter(limiter),
		),
		exitChan: make(chan struct{}),
	}

	of.CheckListenKey()

	return of
}

type listenKey struct {
	accountId      string
	uniq           string
	key            string
	createTime     time.Time
	instrumentType exchange.InstrumentType
	apikey         string
	secretkey      string
}

type of struct {
	exitChan      chan struct{}
	name          string
	opts          *options
	client        *bnhttp.Client
	limiter       limiter.Limiter
	wsm           wsmanager.WebsocketManager
	listenKeySets map[string]*listenKey // listenKey 集合, 合约一个，现货一个
	mux           sync.Mutex
}

func (o *of) Name() string {
	return o.name
}

func (o *of) AddOrderFeed(req *ofmanager.OrderFeedRequest) error {
	o.mux.Lock()
	defer o.mux.Unlock()

	if !o.limiter.WsAllow() {
		return ErrLimitExceed
	}

	conf := &wsmanager.WebsocketConfig{
		PingHandler: pingHandler,
		PongHandler: pongHandler,
	}
	// 生成 listenKey
	key, err := o.generateListenKey(req)

	if err != nil {
		return err
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
			o.opts.logger.Error("order new json error", err)
			return
		}
		switch j.Get("e").MustString() {
		case "executionReport":
			event := &bnSpotWsOrderUpdateEvent{}
			err = bnhttp.Json.Unmarshal(message, event)
			if err != nil {
				o.opts.logger.Error("order unmarshal error", err)
				return
			}
			oe, err := swoueToOrderEvent(event)
			if err != nil {
				o.opts.logger.Error("order to order event error", err)
				return
			}
			req.Event(oe)
		case "ORDER_TRADE_UPDATE":
			event := &bnFuturesWsUserDataEvent{}
			err = bnhttp.Json.Unmarshal(message, event)
			if err != nil {
				o.opts.logger.Error("order unmarshal error", err)
				return
			}
			oe, err := fwoueToOrderEvent(&event.OrderTradeUpdate)
			if err != nil {
				o.opts.logger.Error("order to order event error", err)
				return
			}
			req.Event(oe)
		}
	}
	uniq, err := o.addWebsocket(&websocket.WebsocketRequest{
		Endpoint:       endpoint,
		ID:             fmt.Sprintf("binance.order.%s", req.Instrument),
		MessageHandler: wsHandler,
	}, conf)

	o.listenKeySets[uniq] = &listenKey{
		accountId:      req.AccountId,
		uniq:           uniq,
		key:            key,
		createTime:     generateTime,
		instrumentType: req.Instrument,
		apikey:         req.APIKey,
		secretkey:      req.SecretKey,
	}

	if err != nil {
		return err
	}

	return nil
}

func (o *of) CloseOrderFeed(id string) error {
	o.mux.Lock()
	defer o.mux.Unlock()

	lk, ok := o.listenKeySets[id]
	if !ok {
		return fmt.Errorf("listenKey not found")
	}

	err := o.closeListenKey(lk)
	if err != nil {
		return err
	}

	err = o.wsm.CloseWebsocket(id)
	if err != nil {
		return err
	}

	return nil
}

func (o *of) OrderFeedList() []string {
	o.mux.Lock()
	defer o.mux.Unlock()

	list := make([]string, 0, len(o.listenKeySets))
	for k := range o.listenKeySets {
		list = append(list, k)
	}
	return list
}

func (o *of) Shutdown() error {
	close(o.exitChan)
	o.mux.Lock()
	defer o.mux.Unlock()

	for _, lk := range o.listenKeySets {
		err := o.closeListenKey(lk)
		if err != nil {
			return err
		}
	}

	err := o.wsm.Shutdown()
	if err != nil {
		return err
	}

	return nil
}

func (o *of) addWebsocket(req *websocket.WebsocketRequest, conf *wsmanager.WebsocketConfig) (string, error) {
	uniq, err := o.wsm.AddWebsocket(req, conf)
	if err != nil {
		return "", err
	}
	return uniq, nil
}

func (o *of) generateListenKey(req *ofmanager.OrderFeedRequest) (string, error) {

	for _, lk := range o.listenKeySets {
		if lk.accountId == req.AccountId && lk.instrumentType == req.Instrument {
			return lk.key, nil
		}
	}

	r := &bnhttp.Request{
		APIKey:    req.APIKey,
		SecretKey: req.SecretKey,
		Method:    http.MethodPost,
		SecType:   bnhttp.SecTypeAPIKey,
	}

	if req.Instrument == exchange.InstrumentTypeSpot {
		r.Endpoint = "/api/v3/userDataStream"
	} else if req.Instrument == exchange.InstrumentTypeFutures {
		r.Endpoint = "/fapi/v1/listenKey"
	}

	data, err := o.client.CallAPI(context.Background(), r)
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

func (o *of) updateListenKey(lk *listenKey) error {
	r := &bnhttp.Request{
		APIKey:    lk.apikey,
		SecretKey: lk.secretkey,
		Method:    http.MethodPut,
		SecType:   bnhttp.SecTypeAPIKey,
	}

	if lk.instrumentType == exchange.InstrumentTypeSpot {
		r.Endpoint = "/api/v3/userDataStream"
		r.SetFormParam("listenKey", lk.key)
	} else if lk.instrumentType == exchange.InstrumentTypeFutures {
		r.Endpoint = "/fapi/v1/listenKey"
	}

	_, err := o.client.CallAPI(context.Background(), r)
	if err != nil {
		return err
	}

	// 更新listenKey createTime
	lk.createTime = time.Now()

	return nil
}

func (o *of) closeListenKey(lk *listenKey) error {
	r := &bnhttp.Request{
		Method:    http.MethodDelete,
		SecType:   bnhttp.SecTypeAPIKey,
		APIKey:    lk.apikey,
		SecretKey: lk.secretkey,
	}

	if lk.instrumentType == exchange.InstrumentTypeSpot {
		r.Endpoint = "/api/v3/userDataStream"
	} else if lk.instrumentType == exchange.InstrumentTypeFutures {
		r.Endpoint = "/fapi/v1/listenKey"
	}

	_, err := o.client.CallAPI(context.Background(), r)
	if err != nil {
		return err
	}

	// 删除 listenKey
	delete(o.listenKeySets, lk.uniq)

	return nil
}

// 检查 listenKey 是否过期
func (o *of) CheckListenKey() {
	for {
		select {
		case <-o.exitChan:
			return
		default:
			o.mux.Lock()
			for _, lk := range o.listenKeySets {
				if time.Since(lk.createTime) >= o.opts.listenKeyExpire {
					o.updateListenKey(lk)
				}
			}
			o.mux.Unlock()
			time.Sleep(o.opts.checkListenKeyPeriod)
		}
	}
}

func pingHandler(appData string, conn websocket.WebSocketConn) error {
	return conn.WriteMessage(10, []byte(appData))
}

func pongHandler(appData string, conn websocket.WebSocketConn) error {
	return conn.WriteMessage(9, []byte(appData))
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
