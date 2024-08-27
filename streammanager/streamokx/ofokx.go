package streamokx

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/go-gotop/kit/exchange"
	"github.com/go-gotop/kit/exchange/okexc"
	"github.com/go-gotop/kit/limiter"
	"github.com/go-gotop/kit/requests/okhttp"
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

var (
	ErrLimitExceed = errors.New("websocket request too frequent, please try again later")
)

const (
	okWsEndpoint = "wss://ws.okx.com:8443"
)

func NewOkxStream(cli *okhttp.Client, redisClient *redis.Client, limiter limiter.Limiter, t time.Duration, opts ...Option) streammanager.StreamManager {
	o := &options{
		logger:          log.NewHelper(log.DefaultLogger),
		maxConnDuration: t,
		connectCount:    2,
	}
	for _, opt := range opts {
		opt(o)
	}

	of := &of{
		name:    exchange.OkxExchange,
		opts:    o,
		rdb:     redisClient,
		client:  cli,
		limiter: limiter,
		wsm: manager.NewManager(
			manager.WithMaxConnDuration(o.maxConnDuration),
		),
		streamList: make([]streammanager.Stream, 0),
		exitChan:   make(chan struct{}),
	}

	go of.keepAlive()

	return of
}

type of struct {
	exitChan   chan struct{}
	name       string
	opts       *options
	rdb        *redis.Client // redis客户端
	client     *okhttp.Client
	limiter    limiter.Limiter
	streamList []streammanager.Stream
	wsm        wsmanager.WebsocketManager
	mux        sync.Mutex
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

	endpoint := okWsEndpoint + "/ws/v5/private"

	// 建立连接
	uuid := uuid.New().String()
	err := o.addWebsocket(&websocket.WebsocketRequest{
		Endpoint:       endpoint,
		ID:             uuid,
		MessageHandler: o.createWebsocketHandler(uuid, req, o.subscribe),
		ErrorHandler:   req.ErrorHandler,
	}, conf)
	if err != nil {
		return nil, err
	}

	// 登录
	err = o.login(uuid, req)
	if err != nil {
		return nil, err
	}

	o.streamList = append(o.streamList, streammanager.Stream{
		Exchange:   exchange.OkxExchange,
		Instrument: req.Instrument,
		UUID:       uuid,
		APIKey:     req.APIKey,
		AccountId:  req.AccountId,
	})

	return []string{uuid}, nil
}

func (o *of) CloseStream(accountId string, instrument exchange.InstrumentType, uuid string) error {
	o.mux.Lock()
	defer o.mux.Unlock()

	err := o.wsm.CloseWebsocket(uuid)
	if err != nil {
		return err
	}
	return nil
}

func (o *of) StreamList() []streammanager.Stream {
	o.mux.Lock()
	defer o.mux.Unlock()

	return o.streamList
}

func (o *of) Shutdown() error {
	o.mux.Lock()
	defer o.mux.Unlock()

	err := o.wsm.Shutdown()
	if err != nil {
		return err
	}
	return nil
}

func (o *of) login(uuid string, req *streammanager.StreamRequest) error {
	timestamp := time.Now().Unix()
	preSign := fmt.Sprintf("%dGET/users/self/verify", timestamp)

	mac := hmac.New(sha256.New, []byte(req.SecretKey))
	if _, err := mac.Write([]byte(preSign)); err != nil {
		return err
	}
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	loginMsg := &login{
		Op: "login",
		Args: []struct {
			ApiKey     string `json:"apiKey"`
			Passphrase string `json:"passphrase"`
			Timestamp  string `json:"timestamp"`
			Sign       string `json:"sign"`
		}{
			{
				ApiKey:     req.APIKey,
				Passphrase: req.Passphrase,
				Timestamp:  fmt.Sprintf("%d", timestamp),
				Sign:       signature,
			},
		},
	}

	msg, err := json.Marshal(loginMsg)
	if err != nil {
		return err
	}

	return o.wsm.GetWebsocket(uuid).WriteMessage(gwebsocket.TextMessage, msg)
}

func (o *of) subscribe(uuid string, req *streammanager.StreamRequest) error {
	subList := make([]string, 0)
	if req.Instrument == exchange.InstrumentTypeFutures {
		// 如果是合约类型，则添加永续和交割合约
		subList = append(subList, "SWAP")
		subList = append(subList, "FUTURES")
	} else {
		subList = append(subList, string(req.Instrument))
	}

	args := make([]struct {
		Channel  string `json:"channel"`
		InstType string `json:"instType"`
	}, 0)

	for _, inst := range subList {
		args = append(args, struct {
			Channel  string `json:"channel"`
			InstType string `json:"instType"`
		}{
			Channel:  "orders",
			InstType: inst,
		})
	}

	subMsg := &sub{
		Op:   "subscribe",
		Args: args,
	}

	msg, err := json.Marshal(subMsg)
	if err != nil {
		return err
	}

	err = o.wsm.GetWebsocket(uuid).WriteMessage(gwebsocket.TextMessage, msg)
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

func pingHandler(appData string, conn websocket.WebSocketConn) error {
	return conn.WriteMessage(gwebsocket.PongMessage, []byte(appData))
}

func pongHandler(appData string, conn websocket.WebSocketConn) error {
	return conn.WriteMessage(gwebsocket.PingMessage, []byte(appData))
}

func (o *of) createWebsocketHandler(uuid string, req *streammanager.StreamRequest, subhandler func(uuid string, req *streammanager.StreamRequest) error) func(message []byte) {
	return func(message []byte) {
		j, err := okhttp.NewJSON(message)
		if err != nil {
			o.opts.logger.Error("order new json error", err)
			return
		}

		event := j.Get("event").MustString()

		if event == "error" {
			req.ErrorHandler(errors.New(j.Get("msg").MustString()))
			return
		}

		if event == "login" {
			subhandler(uuid, req)
			return
		}

		if event != "" {
			return
		}

		tes, err := toOrderEvent(message, req.Instrument)
		if err != nil {
			if req.ErrorHandler != nil {
				req.ErrorHandler(err)
			}
			return
		}
		for _, te := range tes {
			req.OrderEvent(te)
		}
	}
}

func toOrderEvent(message []byte, instrument exchange.InstrumentType) ([]*exchange.OrderResultEvent, error) {
	event := &okWsOrderUpdateEvent{}

	err := okhttp.Json.Unmarshal(message, event)

	if err != nil {
		return nil, err
	}

	if event.Arg.Channel != "orders" {
		return nil, nil
	}

	// 如果是合约，则判断 instType 是否为 FUTURES 或 SWAP
	if instrument == exchange.InstrumentTypeFutures && (event.Arg.InstType != "FUTURES" && event.Arg.InstType != "SWAP") {
		return nil, nil
	} else if instrument != exchange.InstrumentTypeFutures && string(instrument) != event.Arg.InstType {
		// 其他直接判断 instType 是否与 instrument 相等
		return nil, nil
	}

	orderResultEvents := make([]*exchange.OrderResultEvent, 0)
	fmt.Printf("event: %v\n", event)
	for _, d := range event.Data {
		price, err := decimal.NewFromString(d.FillPx)
		if err != nil {
			price = decimal.Zero
		}
		volume, err := decimal.NewFromString(d.FillSz)
		if err != nil {
			volume = decimal.Zero
		}
		latestVolume, err := decimal.NewFromString(d.FillSz)
		if err != nil {
			return nil, err
		}
		filledVolume, err := decimal.NewFromString(d.AccFillSz)
		if err != nil {
			filledVolume = decimal.Zero
		}
		latestPrice, err := decimal.NewFromString(d.LastPx)
		if err != nil {
			latestPrice = decimal.Zero
		}
		feeCost, err := decimal.NewFromString(d.FillFee)
		if err != nil {
			// 有些市价成交单 FillFee不一定有值，所以如果没值就取该笔订单的累积手续费
			fee, err := decimal.NewFromString(d.Fee)
			if err != nil {
				feeCost = decimal.Zero
			} else {
				feeCost = fee
			}
		}
		filledQuoteVolume, err := decimal.NewFromString(d.FillNotionalUsd)
		if err != nil {
			filledQuoteVolume = decimal.Zero
		}
		avgPrice, err := decimal.NewFromString(d.AvgPx)
		if err != nil {
			avgPrice = decimal.Zero
		}

		executionType := "NEW"
		state := exchange.OrderStateNew
		switch d.State {
		case "partially_filled":
			executionType = "TRADE"
			state = exchange.OrderStatePartiallyFilled
		case "filled":
			executionType = "TRADE"
			state = exchange.OrderStateFilled
		case "canceled":
			executionType = "CANCELED"
			state = exchange.OrderStateCanceled
		case "rejected":
			executionType = "REJECTED"
			state = exchange.OrderStateRejected
		case "expired":
			executionType = "EXPIRED"
			state = exchange.OrderStateExpired
		}

		updateTime, err := strconv.ParseInt(d.UpdateTime, 10, 64)
		if err != nil {
			return nil, err
		}

		by := exchange.ByTaker
		if d.ExecType == "M" {
			by = exchange.ByMaker
		}

		ore := &exchange.OrderResultEvent{
			PositionSide:      okexc.OkxTPositionSide(d.PosSide),
			Exchange:          exchange.OkxExchange,
			Symbol:            d.InstID,
			ClientOrderID:     d.ClientOrderID,
			ExecutionType:     exchange.ExecutionState(executionType),
			State:             state,
			OrderID:           d.OrderID,
			TransactionTime:   updateTime,
			Side:              okexc.OkxTSide(d.Side),
			Type:              okexc.OkxTMarketType(d.OrderType),
			Instrument:        instrument,
			Volume:            volume,
			By:                by,
			Price:             price,
			LatestVolume:      latestVolume,
			FilledVolume:      filledVolume,
			LatestPrice:       latestPrice,
			FeeAsset:          d.FeeCcy,
			FeeCost:           feeCost,
			AvgPrice:          avgPrice,
			FilledQuoteVolume: filledQuoteVolume,
		}

		orderResultEvents = append(orderResultEvents, ore)
	}

	return orderResultEvents, nil
}

func (o *of) keepAlive() {
	for {
		select {
		case <-o.exitChan:
			return
		case <-time.After(10 * time.Second):
			o.mux.Lock()
			for _, stream := range o.streamList {
				err := o.wsm.GetWebsocket(stream.UUID).WriteMessage(gwebsocket.PingMessage, nil)
				if err != nil {
					o.opts.logger.Error("write ping message error", err)
				}
			}
			o.mux.Unlock()
		}
	}
}
