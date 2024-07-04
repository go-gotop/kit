// Description: Binance OrderFeed Manager
// listenkey 对于现货和合约是不一样的，所以需要分开处理
// 一个账户类型只会对应一个 listenkey
// 调用generateListenKey，如果交易所现存有效的listenkey，则直接返回该listenkey，并延长有效期

package streambinance

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-gotop/kit/exchange"
	"github.com/go-gotop/kit/limiter"
	"github.com/go-gotop/kit/requests/bnhttp"
	"github.com/go-gotop/kit/streammanager"
	"github.com/go-gotop/kit/websocket"
	"github.com/go-gotop/kit/wsmanager"
	"github.com/go-gotop/kit/wsmanager/manager"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	gwebsocket "github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
)

var _ streammanager.StreamManager = (*of)(nil)

var (
	ErrLimitExceed = errors.New("websocket request too frequent, please try again later")
)

const (
	bnSpotWsEndpoint    = "wss://stream.binance.com:9443/ws"
	bnFuturesWsEndpoint = "wss://fstream.binance.com/ws"
	bnSpotEndpoint      = "https://api.binance.com"
	bnFuturesEndpoint   = "https://fapi.binance.com"
)

// TODO: 限流器放在 ofbinance 做调用，不传入 wsmanager
func NewBinanceStream(cli *bnhttp.Client, limiter limiter.Limiter, opts ...Option) streammanager.StreamManager {
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
		name:          exchange.BinanceExchange,
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
	uuidList    []string
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

func (o *of) AddStream(req *streammanager.StreamRequest) (string, error) {
	o.mux.Lock()
	defer o.mux.Unlock()

	if !o.limiter.WsAllow() {
		return "", ErrLimitExceed
	}

	conf := &wsmanager.WebsocketConfig{
		PingHandler: pingHandler,
		PongHandler: pongHandler,
	}
	// 生成 listenKey
	key, err := o.generateListenKey(req)

	if err != nil {
		return "", err
	}

	generateTime := time.Now()
	uuid := uuid.New().String() // 一个链接的uuid，因为一个账户可能存在多条链接，所以不能用账户ID做标识

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
	err = o.addWebsocket(&websocket.WebsocketRequest{
		Endpoint:       endpoint,
		ID:             uuid,
		MessageHandler: wsHandler,
	}, conf)

	if err != nil {
		return "", err
	}
	// 判断账户id是否存在listenkey，存在则不用再次添加，只添加uuid
	if _, ok := o.listenKeySets[req.AccountId]; ok {
		o.listenKeySets[req.AccountId].uuidList = append(o.listenKeySets[req.AccountId].uuidList, uuid)
		return uuid, nil
	}
	o.listenKeySets[req.AccountId] = &listenKey{
		AccountID:   req.AccountId,
		Key:         key,
		CreatedTime: generateTime,
		Instrument:  req.Instrument,
		APIKey:      req.APIKey,
		SecretKey:   req.SecretKey,
		uuidList:    []string{uuid},
	}

	return uuid, nil
}

func (o *of) CloseStream(accountId string, uuid string) error {
	o.mux.Lock()
	defer o.mux.Unlock()

	lk, ok := o.listenKeySets[accountId]
	if !ok {
		return fmt.Errorf("listenKey not found")
	}

	// 删除uuid
	for i, v := range lk.uuidList {
		if v == uuid {
			lk.uuidList = append(lk.uuidList[:i], lk.uuidList[i+1:]...)
			break
		}
	}

	// 关闭链接
	err := o.wsm.CloseWebsocket(uuid)
	if err != nil {
		return err
	}

	if len(lk.uuidList) > 0 {
		return nil
	}

	// 如果uuidList为空，则删除listenKey
	err = o.closeListenKey(lk)
	if err != nil {
		return err
	}

	return nil
}

func (o *of) StreamList() []streammanager.Stream {
	o.mux.Lock()
	defer o.mux.Unlock()

	list := make([]streammanager.Stream, 0, len(o.listenKeySets))
	for _, v := range o.listenKeySets {
		for _, uuid := range v.uuidList {
			list = append(list, streammanager.Stream{
				UUID:       uuid,
				AccountId:  v.AccountID,
				APIKey:     v.APIKey,
				Exchange:   o.name,
				Instrument: v.Instrument,
			})
		}
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

func (o *of) generateListenKey(req *streammanager.StreamRequest) (string, error) {
	for _, lk := range o.listenKeySets {
		if lk.AccountID == req.AccountId && lk.Instrument == req.Instrument {
			return lk.Key, nil
		}
	}

	r := &bnhttp.Request{
		APIKey:    req.APIKey,
		SecretKey: req.SecretKey,
		Method:    http.MethodPost,
		SecType:   bnhttp.SecTypeAPIKey,
	}

	if req.Instrument == exchange.InstrumentTypeFutures {
		r.Endpoint = "/fapi/v1/listenKey"
		o.client.SetApiEndpoint(bnFuturesEndpoint)
	} else {
		r.Endpoint = "/api/v3/userDataStream"
		o.client.SetApiEndpoint(bnSpotEndpoint)
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
		APIKey:    lk.APIKey,
		SecretKey: lk.SecretKey,
		Method:    http.MethodPut,
		SecType:   bnhttp.SecTypeAPIKey,
	}

	if lk.Instrument == exchange.InstrumentTypeSpot {
		r.Endpoint = "/api/v3/userDataStream"
		r.SetFormParam("listenKey", lk.Key)
		o.client.SetApiEndpoint(bnSpotEndpoint)
	} else if lk.Instrument == exchange.InstrumentTypeFutures {
		r.Endpoint = "/fapi/v1/listenKey"
		o.client.SetApiEndpoint(bnFuturesEndpoint)
	}

	_, err := o.client.CallAPI(context.Background(), r)
	if err != nil {
		return err
	}

	// 更新listenKey createTime
	lk.CreatedTime = time.Now()

	return nil
}

func (o *of) closeListenKey(lk *listenKey) error {
	r := &bnhttp.Request{
		Method:    http.MethodDelete,
		SecType:   bnhttp.SecTypeAPIKey,
		APIKey:    lk.APIKey,
		SecretKey: lk.SecretKey,
	}

	if lk.Instrument == exchange.InstrumentTypeFutures {
		r.Endpoint = "/fapi/v1/listenKey"
		o.client.SetApiEndpoint(bnFuturesEndpoint)
	} else {
		r.Endpoint = "/api/v3/userDataStream"
		o.client.SetApiEndpoint(bnSpotEndpoint)
	}

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
	fmt.Printf("pingHandler: %s\n", appData)
	return conn.WriteMessage(gwebsocket.PongMessage, []byte(appData))
}

func pongHandler(appData string, conn websocket.WebSocketConn) error {
	fmt.Printf("pongHandler: %s\n", appData)
	return conn.WriteMessage(gwebsocket.PingMessage, []byte(appData))
}

func swoueToOrderEvent(event *bnSpotWsOrderUpdateEvent) (*exchange.OrderResultEvent, error) {
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
	avgPrice := decimal.Zero
	if filledQuoteVolume.GreaterThan(decimal.Zero) && filledVolume.GreaterThan(decimal.Zero) {
		avgPrice = filledQuoteVolume.Div(filledVolume)
	}
	ore := &exchange.OrderResultEvent{
		PositionSide:    ps,
		Exchange:        exchange.BinanceExchange,
		Symbol:          event.Symbol,
		ClientOrderID:   event.ClientOrderId,
		ExecutionType:   exchange.ExecutionState(event.ExecutionType),
		State:           exchange.OrderState(event.Status),
		OrderID:         fmt.Sprintf("%d", event.Id),
		TransactionTime: event.TransactionTime,
		Side:            exchange.SideType(event.Side),
		Type:            exchange.OrderType(event.Type),
		Instrument:      exchange.InstrumentTypeSpot,
		Volume:          volume,
		By:              exchange.ByTaker,
		Price:           price,
		LatestVolume:    latestVolume,
		FilledVolume:    filledVolume,
		LatestPrice:     latestPrice,
		FeeAsset:        event.FeeAsset,
		FeeCost:         feeCost,
		AvgPrice:        avgPrice,
	}
	if event.IsMaker {
		ore.By = exchange.ByMaker
	}
	return ore, nil
}

func fwoueToOrderEvent(event *bnFuturesWsOrderUpdateEvent) (*exchange.OrderResultEvent, error) {
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
	ore := &exchange.OrderResultEvent{
		PositionSide:    ps,
		Exchange:        exchange.BinanceExchange,
		Symbol:          event.Symbol,
		ClientOrderID:   event.ClientOrderID,
		ExecutionType:   exchange.ExecutionState(event.ExecutionType),
		State:           exchange.OrderState(event.Status),
		OrderID:         fmt.Sprintf("%d", event.ID),
		TransactionTime: event.TradeTime,
		By:              exchange.ByTaker,
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
	}
	if event.IsMaker {
		ore.By = exchange.ByMaker
	}
	return ore, nil
}
