// Description: Binance OrderFeed Manager
// listenkey 对于现货和合约是不一样的，所以需要分开处理
// 一个账户类型只会对应一个 listenkey
// 调用generateListenKey，如果交易所现存有效的listenkey，则直接返回该listenkey，并延长有效期
// TOFIX:
// 1. listenkey 应该统一进行管理，而这里的kit包有服务引用的话，listenkey是相当于在其本地进行管理的，这里面可能有点问题，比如这里checkListenKey的时候，如果有多个服务引用，会导致listenkey的有效期不一致（使用redis可以解决）

package streambinance

import (
	"context"
	"encoding/json"
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
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
)

var _ streammanager.StreamManager = (*of)(nil)

var (
	ErrLimitExceed = errors.New("websocket request too frequent, please try again later")
)

const (
	bnSpotWsEndpoint            = "wss://stream.binance.com:9443/ws"
	bnFuturesWsEndpoint         = "wss://fstream.binance.com/ws"
	bnMarginWsEndpoint          = "wss://stream.binance.com:9443/ws"
	bnPortfolioMarginWsEndpoint = "wss://fstream.binance.com/pm/ws"
	bnSpotEndpoint              = "https://api.binance.com"
	bnFuturesEndpoint           = "https://fapi.binance.com"
	bnPortfolioMarginEndpoint   = "https://papi.binance.com"

	redisKeyPrefix = "binance_listenkey:"
)

// TODO: 限流器放在 ofbinance 做调用，不传入 wsmanager
func NewBinanceStream(cli *bnhttp.Client, redisClient *redis.Client, limiter limiter.Limiter, t time.Duration, opts ...Option) streammanager.StreamManager {
	// 默认配置
	o := &options{
		logger:               log.NewHelper(log.DefaultLogger),
		maxConnDuration:      t,
		listenKeyExpire:      58 * time.Minute,
		checkListenKeyPeriod: 5 * time.Second,
		connectCount:         1,
	}

	for _, opt := range opts {
		opt(o)
	}

	of := &of{
		name:          exchange.BinanceExchange,
		opts:          o,
		rdb:           redisClient,
		client:        cli,
		limiter:       limiter,
		listenKeySets: make(map[string]*listenKey),
		wsm: manager.NewManager(
			manager.WithMaxConnDuration(o.maxConnDuration),
			// manager.WithConnLimiter(limiter),
		),
		exitChan: make(chan struct{}),
	}

	of.initListenKeySetsFromRedis()

	go of.CheckListenKey()

	return of
}

type listenKey struct {
	AccountID        string                  `json:"account_id"`
	Key              string                  `json:"key"`
	APIKey           string                  `json:"api_key"`
	SecretKey        string                  `json:"secret_key"`
	CreatedTime      time.Time               `json:"created_time"`
	MarketType       exchange.MarketType     `json:"market_type"`
	UUIDList         []string                `json:"uuid_list"`
	IsUnifiedAccount bool                    `json:"is_unified_account"` // 统一账户, 默认 false

}

type of struct {
	exitChan      chan struct{}
	name          string
	opts          *options
	rdb           *redis.Client // redis客户端
	client        *bnhttp.Client
	limiter       limiter.Limiter
	wsm           wsmanager.WebsocketManager
	listenKeySets map[string]*listenKey // listenKey 集合, 合约一个，现货一个
	mux           sync.Mutex
}

func (o *of) Name() string {
	return o.name
}

func (o *of) AddStream(req *streammanager.StreamRequest) ([]string, error) {
	o.mux.Lock()
	defer o.mux.Unlock()

	if !o.limiter.WsAllow() {
		return nil, ErrLimitExceed
	}

	conf := &wsmanager.WebsocketConfig{
		PingHandler: pingHandler,
		PongHandler: pongHandler,
	}
	// 生成 listenKey
	key, err := o.generateListenKey(req)

	if err != nil {
		return nil, err
	}

	generateTime := time.Now()

	// 拼接 listenKey 到请求地址
	var endpoint string
	if req.IsUnifiedAccount {
		endpoint = fmt.Sprintf("%s/%s", bnPortfolioMarginWsEndpoint, key)
	} else {
		endpoint = fmt.Sprintf("%s/%s", bnSpotWsEndpoint, key)
		if req.MarketType == exchange.MarketTypeFuturesUSDMargined || req.MarketType == exchange.MarketTypePerpetualUSDMargined {
			endpoint = fmt.Sprintf("%s/%s", bnFuturesWsEndpoint, key)
		} else if req.MarketType == exchange.MarketTypeMargin {
			endpoint = fmt.Sprintf("%s/%s", bnMarginWsEndpoint, key)
		}
	}

	//构建连接池  //配置连接数量 默认2 要自定义连接时间
	for i := 0; i < o.opts.connectCount; i++ {
		uuid := uuid.New().String() // 一个链接的uuid，因为一个账户可能存在多条链接，所以不能用账户ID做标识
		err = o.addWebsocket(&websocket.WebsocketRequest{
			Endpoint:       endpoint,
			ID:             uuid,
			MessageHandler: o.createWebsocketHandler(req, o.rdb),
			ErrorHandler:   req.ErrorHandler,
		}, conf)
		if err != nil {
			return nil, err
		}

		// 判断账户id是否存在listenkey，存在则不用再次添加，只添加uuid, 更新createTime 和 listenkey
		if _, ok := o.listenKeySets[req.AccountId+string(req.MarketType)]; ok {
			o.listenKeySets[req.AccountId+string(req.MarketType)].UUIDList = append(o.listenKeySets[req.AccountId+string(req.MarketType)].UUIDList, uuid)
			o.listenKeySets[req.AccountId+string(req.MarketType)].CreatedTime = generateTime
			o.listenKeySets[req.AccountId+string(req.MarketType)].Key = key
			err := o.saveListenKeySet(req.AccountId, string(req.MarketType), o.listenKeySets[req.AccountId+string(req.MarketType)])
			if err != nil {
				return nil, err
			}
		} else {
			lk := &listenKey{
				AccountID:        req.AccountId,
				Key:              key,
				CreatedTime:      generateTime,
				MarketType:       req.MarketType,
				APIKey:           req.APIKey,
				SecretKey:        req.SecretKey,
				UUIDList:         []string{uuid},
				IsUnifiedAccount: req.IsUnifiedAccount,
			}
			o.listenKeySets[req.AccountId+string(req.MarketType)] = lk

			err = o.saveListenKeySet(req.AccountId, string(req.MarketType), lk)
			if err != nil {
				return nil, err
			}
		}
	}

	return o.listenKeySets[req.AccountId+string(req.MarketType)].UUIDList, nil
}

func (o *of) CloseStream(accountId string, marketType exchange.MarketType, uuid string) error {
	o.mux.Lock()
	defer o.mux.Unlock()

	lk, err := o.getListenKeySet(accountId, string(marketType))

	if err != nil {
		return err
	}

	// 删除uuid
	for i, v := range lk.UUIDList {
		if v == uuid {
			lk.UUIDList = append(lk.UUIDList[:i], lk.UUIDList[i+1:]...)
			break
		}
	}

	// 关闭链接
	err = o.wsm.CloseWebsocket(uuid)
	if err != nil {
		return err
	}

	if len(lk.UUIDList) > 0 {
		o.listenKeySets[accountId+string(marketType)] = lk

		err := o.saveListenKeySet(accountId, string(marketType), lk)
		if err != nil {
			return err
		}
		return nil
	}

	// 如果UUIDList为空，则删除listenKey
	delete(o.listenKeySets, lk.AccountID+string(lk.MarketType))
	err = o.deleteListenKeySet(lk.AccountID, string(marketType))
	if err != nil {
		return err
	}
	// TOFIX: 对于交易所来说一个账户下面只存在一个有效的listenkey，如果这里走close，会导致其他服务的listenkey也失效
	// err = o.closeListenKey(lk)
	// if err != nil {
	// 	return err
	// }

	return nil
}

func (o *of) StreamList() []streammanager.Stream {
	o.mux.Lock()
	defer o.mux.Unlock()

	o.initListenKeySetsFromRedis()

	list := make([]streammanager.Stream, 0, len(o.listenKeySets))
	for _, v := range o.listenKeySets {
		for _, uuid := range v.UUIDList {
			list = append(list, streammanager.Stream{
				UUID:        uuid,
				AccountId:   v.AccountID,
				APIKey:      v.APIKey,
				Exchange:    o.name,
				MarketType:  v.MarketType,
				IsConnected: o.wsm.IsConnected(uuid),
			})
		}
	}
	return list
}

func (o *of) Shutdown() error {
	o.mux.Lock()
	defer o.mux.Unlock()
	close(o.exitChan)

	// TOFIX: 对于交易所来说一个账户下面只存在一个有效的listenkey，如果这里走close，会导致其他服务的listenkey也失效
	// for _, lk := range o.listenKeySets {
	// 	err := o.closeListenKey(lk)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	err := o.wsm.Shutdown()
	if err != nil {
		return err
	}

	return nil
}

func (o *of) createWebsocketHandler(req *streammanager.StreamRequest, rcli *redis.Client) func(message []byte) {
	return func(message []byte) {
		j, err := bnhttp.NewJSON(message)
		if err != nil {
			o.opts.logger.Error("order new json error", err)
			return
		}
		switch j.Get("e").MustString() {
		// 现货杠杠订单更新 | 统一账户杠杆订单更新
		case "executionReport":
			if req.MarketType == exchange.MarketTypeFuturesUSDMargined || req.MarketType == exchange.MarketTypePerpetualUSDMargined {
				return
			}
			o.opts.logger.Debugf("Binance WS订单事件: %s", string(message))
			if req.IsUnifiedAccount {
				event := &bnUniSpotWsOrderUpdateEvent{}

				err = bnhttp.Json.Unmarshal(message, event)
				if err != nil {
					o.opts.logger.Error("order unmarshal error", err)
					return
				}
				// if !o.onlyProcessing(event.ClientOrderId+event.Status, rcli) {
				// 	return
				// }
				oe, err := swoueUniToOrderEvent(event)
				if err != nil {
					o.opts.logger.Error("order to order event error", err)
					return
				}
				if req.OrderEvent != nil {
					req.OrderEvent(oe)
				}
			} else {
				event := &bnSpotWsOrderUpdateEvent{}

				err = bnhttp.Json.Unmarshal(message, event)
				if err != nil {
					o.opts.logger.Error("order unmarshal error", err)
					return
				}
				// if !o.onlyProcessing(event.ClientOrderId+event.Status, rcli) {
				// 	return
				// }
				oe, err := swoueToOrderEvent(event)
				if err != nil {
					o.opts.logger.Error("order to order event error", err)
					return
				}
				if req.OrderEvent != nil {
					req.OrderEvent(oe)
				}
			}

		// 合约订单更新  ｜ 统一账户合约订单更新
		case "ORDER_TRADE_UPDATE":
			if req.MarketType == exchange.MarketTypeSpot || req.MarketType == exchange.MarketTypeMargin {
				return
			}
			o.opts.logger.Debugf("Binance WS订单事件: %s", string(message))
			event := &bnFuturesWsUserDataEvent{}
			err = bnhttp.Json.Unmarshal(message, event)
			if err != nil {
				o.opts.logger.Error("order unmarshal error", err)
				return
			}
			// if !o.onlyProcessing(event.OrderTradeUpdate.ClientOrderID+event.OrderTradeUpdate.Status, rcli) {
			// 	return
			// }
			oe, err := fwoueToOrderEvent(&event.OrderTradeUpdate)
			if err != nil {
				o.opts.logger.Error("order to order event error", err)
				return
			}
			if req.OrderEvent != nil {
				req.OrderEvent(oe)
			}
		// 合约余额和持仓更新  ｜ 统一账户合约余额和持仓更新
		case "ACCOUNT_UPDATE":
			event := &bnFuturesWsAccountUpdateEvent{}
			err = bnhttp.Json.Unmarshal(message, event)
			if err != nil {
				o.opts.logger.Error("account unmarshal error", err)
				return
			}
			au, err := fwaueToAccountUpdateEvent(event)
			if err != nil {
				o.opts.logger.Error("account to account event error", err)
				return
			}
			if req.AccountEvent != nil {
				req.AccountEvent(au)
			}
		// 现货账户杠杠更新 ｜ 统一账户杠杆更新
		case "outboundAccountPosition":
			event := &bnSpotWsAccountUpdateEvent{}
			err = bnhttp.Json.Unmarshal(message, event)
			if err != nil {
				o.opts.logger.Error("account unmarshal error", err)
				return
			}
			au, err := swaueToAccountUpdateEvent(event)
			if err != nil {
				o.opts.logger.Error("account to account event error", err)
				return
			}
			if req.AccountEvent != nil {
				req.AccountEvent(au)
			}

		// listenKey 过期
		case "listenKeyExpired":
			event := &bnListenKeyExpiredEvent{}
			err = bnhttp.Json.Unmarshal(message, event)
			if err != nil {
				o.opts.logger.Error("listenKey unmarshal error", err)
				return
			}

			// 关闭accountId下所有连接
			for _, lk := range o.listenKeySets {
				if lk.AccountID+string(lk.MarketType) == req.AccountId+string(req.MarketType) {
					for _, uuid := range lk.UUIDList {
						o.wsm.CloseWebsocket(uuid)
					}
				}
			}
			// 删除 listenKey
			delete(o.listenKeySets, req.AccountId+string(req.MarketType))

			o.deleteListenKeySet(req.AccountId, string(req.MarketType))

			// 推送事件
			if req.ErrorEvent != nil {
				req.ErrorEvent(&exchange.StreamErrorEvent{
					AccountID: req.AccountId,
					Error:     exchange.ErrListenKeyExpired,
				})
			}
		}
	}
}

func (o *of) onlyProcessing(uid string, rcli *redis.Client) bool {
	nx := rcli.SetNX(context.Background(), uid, 1, 10*time.Minute)
	if nx.Err() != nil {
		o.opts.logger.Errorf("onlyProcessing setnx redis error: %v", nx.Err())
		return false
	}
	return nx.Val()
}

func (o *of) addWebsocket(req *websocket.WebsocketRequest, conf *wsmanager.WebsocketConfig) error {
	err := o.wsm.AddWebsocket(req, conf)
	if err != nil {
		return err
	}
	return nil
}

func (o *of) generateListenKey(req *streammanager.StreamRequest) (string, error) {
	r := &bnhttp.Request{
		APIKey:    req.APIKey,
		SecretKey: req.SecretKey,
		Method:    http.MethodPost,
		SecType:   bnhttp.SecTypeAPIKey,
	}

	if !req.IsUnifiedAccount {
		if req.MarketType == exchange.MarketTypeFuturesUSDMargined || req.MarketType == exchange.MarketTypePerpetualUSDMargined {
			r.Endpoint = "/fapi/v1/listenKey"
			o.client.SetApiEndpoint(bnFuturesEndpoint)
		} else if req.MarketType == exchange.MarketTypeSpot {
			r.Endpoint = "/api/v3/userDataStream"
			o.client.SetApiEndpoint(bnSpotEndpoint)
		} else if req.MarketType == exchange.MarketTypeMargin {
			r.Endpoint = "/sapi/v1/userDataStream"
			o.client.SetApiEndpoint(bnSpotEndpoint)
		}
	} else {
		r.Endpoint = "/papi/v1/listenKey"
		o.client.SetApiEndpoint(bnPortfolioMarginEndpoint)
	}

	data, err := o.client.CallAPI(context.Background(), r)
	if err != nil {
		return "", err
	}

	var res struct {
		ListenKey string `json:"listenKey"`
	}
	fmt.Printf("data: %s\n", string(data))
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

	if !lk.IsUnifiedAccount {
		if lk.MarketType == exchange.MarketTypeSpot {
			r.Endpoint = "/api/v3/userDataStream"
			r.SetFormParam("listenKey", lk.Key)
			o.client.SetApiEndpoint(bnSpotEndpoint)
		} else if lk.MarketType == exchange.MarketTypeFuturesUSDMargined || lk.MarketType == exchange.MarketTypePerpetualUSDMargined {
			r.Endpoint = "/fapi/v1/listenKey"
			o.client.SetApiEndpoint(bnFuturesEndpoint)
		} else if lk.MarketType == exchange.MarketTypeMargin {
			r.Endpoint = "/sapi/v1/userDataStream"
			r.SetFormParam("listenKey", lk.Key)
			o.client.SetApiEndpoint(bnSpotEndpoint)
		}
	} else {
		r.Endpoint = "/papi/v1/listenKey"
		r.SetFormParam("listenKey", lk.Key)
		o.client.SetApiEndpoint(bnPortfolioMarginEndpoint)
	}

	_, err := o.client.CallAPI(context.Background(), r)
	if err != nil {
		return err
	}

	// 更新listenKey createTime
	lk.CreatedTime = time.Now()
	err = o.saveListenKeySet(lk.AccountID, string(lk.MarketType), lk)
	if err != nil {
		return err
	}

	return nil
}

// func (o *of) closeListenKey(lk *listenKey) error {
// 	r := &bnhttp.Request{
// 		Method:    http.MethodDelete,
// 		SecType:   bnhttp.SecTypeAPIKey,
// 		APIKey:    lk.APIKey,
// 		SecretKey: lk.SecretKey,
// 	}

// 	if lk.Instrument == exchange.InstrumentTypeFutures {
// 		r.Endpoint = "/fapi/v1/listenKey"
// 		o.client.SetApiEndpoint(bnFuturesEndpoint)
// 	} else {
// 		r.Endpoint = "/api/v3/userDataStream"
// 		o.client.SetApiEndpoint(bnSpotEndpoint)
// 	}

// 	_, err := o.client.CallAPI(context.Background(), r)
// 	if err != nil {
// 		return err
// 	}
// 	o.mux.Lock()
// 	// 删除 listenKey
// 	delete(o.listenKeySets, lk.AccountID)
// 	o.mux.Unlock()
// 	return nil
// }

// 检查 listenKey 是否过期
func (o *of) CheckListenKey() {
	for {
		select {
		case <-o.exitChan:
			return
		default:
			o.mux.Lock()
			o.initListenKeySetsFromRedis()
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

func (o *of) initListenKeySetsFromRedis() error {
	keys, err := o.rdb.Keys(context.Background(), redisKeyPrefix+"*").Result()
	if err != nil {
		return err
	}

	for _, key := range keys {
		lk, err := o.getListenKeySet(key[len(redisKeyPrefix):], "")
		if err != nil {
			o.opts.logger.Error("Failed to get listenKey from redis", err)
			continue
		}
		if lk != nil {
			o.listenKeySets[lk.AccountID+string(lk.MarketType)] = lk
		}
	}

	return nil
}

func (o *of) saveListenKeySet(accountId string, marketType string, lk *listenKey) error {
	data, err := json.Marshal(lk)
	if err != nil {
		return err
	}

	return o.rdb.Set(context.Background(), redisKeyPrefix+accountId+marketType, data, o.opts.listenKeyExpire).Err()
}

func (o *of) getListenKeySet(accountId string, marketType string) (*listenKey, error) {
	data, err := o.rdb.Get(context.Background(), redisKeyPrefix+accountId+marketType).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var lk listenKey
	err = json.Unmarshal([]byte(data), &lk)
	if err != nil {
		return nil, err
	}

	return &lk, nil
}

func (o *of) deleteListenKeySet(accountId string, marketType string) error {
	return o.rdb.Del(context.Background(), redisKeyPrefix+accountId+marketType).Err()
}

func pingHandler(appData string, conn websocket.WebSocketConn) error {
	return conn.WriteMessage(gwebsocket.PongMessage, []byte(appData))
}

func pongHandler(appData string, conn websocket.WebSocketConn) error {
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
		MarketType:      exchange.MarketTypeSpot,
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

func swoueUniToOrderEvent(event *bnUniSpotWsOrderUpdateEvent) (*exchange.OrderResultEvent, error) {
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
		MarketType:      exchange.MarketTypeSpot,
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
		MarketType:      exchange.MarketTypePerpetualUSDMargined,
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

func swaueToAccountUpdateEvent(event *bnSpotWsAccountUpdateEvent) ([]*exchange.AccountUpdateEvent, error) {
	result := make([]*exchange.AccountUpdateEvent, 0)

	for _, b := range event.Balances {
		amount, err := decimal.NewFromString(b.Free)
		if err != nil {
			return nil, err
		}
		au := &exchange.AccountUpdateEvent{
			Asset:   b.Asset,
			Balance: amount,
		}
		result = append(result, au)
	}
	return result, nil
}

func fwaueToAccountUpdateEvent(event *bnFuturesWsAccountUpdateEvent) ([]*exchange.AccountUpdateEvent, error) {
	result := make([]*exchange.AccountUpdateEvent, 0)
	balance := event.EventDetail.Balances
	for _, b := range balance {
		amount, err := decimal.NewFromString(b.CrossWalletBalance)
		if err != nil {
			return nil, err
		}
		au := &exchange.AccountUpdateEvent{
			Asset:   b.Asset,
			Balance: amount,
		}
		result = append(result, au)
	}
	return result, nil
}
