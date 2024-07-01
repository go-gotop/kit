package ofmock

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
	"github.com/go-gotop/kit/requests/mohttp"
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
	wsEndpoint          = "ws://127.0.0.1:8082/ws/order"
	mockExchangEndpoint = "http://127.0.0.1:8000"
)

// TODO: 限流器放在 ofbinance 做调用，不传入 wsmanager
func NewMockOrderFeed(cli *mohttp.Client, limiter limiter.Limiter, opts ...Option) ofmanager.OrderFeedManager {
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
		name:          exchange.MockExchange,
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

	go of.CheckListenKey()

	return of
}

type listenKey struct {
	AccountID   string
	Key         string
	APIKey      string
	SecretKey   string
	CreatedTime time.Time
	Instrument  exchange.InstrumentType
}

type of struct {
	exitChan      chan struct{}
	name          string
	opts          *options
	client        *mohttp.Client
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
	endpoint := fmt.Sprintf("%s?listenKey=%s", wsEndpoint, key)

	wsHandler := func(message []byte) {
		event := &wsOrderUpdateEvent{}
		err = mohttp.Json.Unmarshal(message, event)
		if err != nil {
			o.opts.logger.Error("order new json error", err)
			return
		}
		oe, err := swoueToOrderEvent(event)
		if err != nil {
			o.opts.logger.Error("order to order event error", err)
			return
		}
		req.Event(oe)
	}
	err = o.addWebsocket(&websocket.WebsocketRequest{
		Endpoint:       endpoint,
		ID:             req.AccountId,
		MessageHandler: wsHandler,
	}, conf)

	if err != nil {
		return err
	}
	o.listenKeySets[req.AccountId] = &listenKey{
		AccountID:   req.AccountId,
		Key:         key,
		CreatedTime: generateTime,
		Instrument:  req.Instrument,
		APIKey:      req.APIKey,
		SecretKey:   req.SecretKey,
	}

	return nil
}

func (o *of) CloseOrderFeed(id string) error {
	o.mux.Lock()
	defer o.mux.Unlock()

	lk, ok := o.listenKeySets[id]
	if !ok {
		return errors.New("listenKey not found")
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

func (o *of) OrderFeedList() []ofmanager.OrderFeed {
	o.mux.Lock()
	defer o.mux.Unlock()

	list := make([]ofmanager.OrderFeed, 0, len(o.listenKeySets))
	for _, v := range o.listenKeySets {
		list = append(list, ofmanager.OrderFeed{
			AccountId:  v.AccountID,
			APIKey:     v.APIKey,
			Exchange:   o.name,
			Instrument: v.Instrument,
		})
	}
	return list
}

func (o *of) Shutdown() error {
	o.mux.Lock()
	defer o.mux.Unlock()
	close(o.exitChan)

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

func (o *of) addWebsocket(req *websocket.WebsocketRequest, conf *wsmanager.WebsocketConfig) error {
	err := o.wsm.AddWebsocket(req, conf)
	if err != nil {
		return err
	}
	return nil
}

func (o *of) generateListenKey(req *ofmanager.OrderFeedRequest) (string, error) {
	for _, lk := range o.listenKeySets {
		if lk.AccountID == req.AccountId && lk.Instrument == req.Instrument {
			return lk.Key, nil
		}
	}

	r := &mohttp.Request{
		APIKey:    req.APIKey,
		SecretKey: req.SecretKey,
		Method:    http.MethodPost,
		SecType:   mohttp.SecTypeAPIKey,
	}

	r.Endpoint = "/api/exchange/listenkey"
	r.SetFormParam("instrumentType", req.Instrument)
	o.client.SetApiEndpoint(mockExchangEndpoint)
	data, err := o.client.CallAPI(context.Background(), r)
	if err != nil {
		return "", err
	}
	var res struct {
		Code    int    `json:"code"`
		Data    string `json:"data"`
		Message string `json:"message"`
	}

	err = mohttp.Json.Unmarshal(data, &res)
	if err != nil {
		return "", err
	}

	return res.Data, nil
}

func (o *of) updateListenKey(lk *listenKey) error {
	r := &mohttp.Request{
		APIKey:    lk.APIKey,
		SecretKey: lk.SecretKey,
		Method:    http.MethodPut,
		SecType:   mohttp.SecTypeAPIKey,
	}

	r.Endpoint = "/api/exchange/listenkey"

	_, err := o.client.CallAPI(context.Background(), r)
	if err != nil {
		return err
	}

	// 更新listenKey createTime
	lk.CreatedTime = time.Now()

	return nil
}

func (o *of) closeListenKey(lk *listenKey) error {
	r := &mohttp.Request{
		Method:    http.MethodDelete,
		SecType:   mohttp.SecTypeAPIKey,
		APIKey:    lk.APIKey,
		SecretKey: lk.SecretKey,
	}

	r.Endpoint = "/api/exchange/listenkey"

	_, err := o.client.CallAPI(context.Background(), r)
	if err != nil {
		return err
	}
	o.mux.Lock()
	// 删除 listenKey
	delete(o.listenKeySets, lk.AccountID)
	o.mux.Unlock()
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
				if time.Since(lk.CreatedTime) >= o.opts.listenKeyExpire {
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

func swoueToOrderEvent(event *wsOrderUpdateEvent) (*exchange.OrderResultEvent, error) {
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
	ore := &exchange.OrderResultEvent{
		PositionSide:    ps,
		Exchange:        exchange.BinanceExchange,
		Symbol:          event.Symbol,
		ClientOrderID:   event.ClientOrderID,
		ExecutionType:   exchange.ExecutionState(event.ExecutionType),
		State:           exchange.OrderState(event.Status),
		OrderID:         event.ID,
		TransactionTime: event.TransactionTime,
		By:              exchange.ByTaker, // 默认为吃单  mock 交易所目前只有市价单
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
	}
	// if event.IsMaker {
	// 	ore.By = exchange.ByMaker
	// }
	return ore, nil
}
