package bnexc

import (
	"context"
	"net/http"

	"github.com/shopspring/decimal"
	"github.com/go-gotop/kit/exchange"
	"github.com/go-gotop/kit/requests/bnhttp"
)

const (
	bnSpotEndpoint      = "https://api.binance.com"
	bnFuturesEndpoint   = "https://fapi.binance.com"
)

func NewBinance(cli *bnhttp.Client) exchange.Exchange {
	return &binance{
		client: cli,
	}
}

type binance struct {
	client *bnhttp.Client
}

func (b *binance) Assets(ctx context.Context, it exchange.InstrumentType) ([]exchange.Asset, error) {
	if it == exchange.InstrumentTypeSpot {
		result, err := b.spotAssets(ctx)
		if err != nil {
			return nil, err
		}
		data, err := bnSpotAssetsToAssets(result)
		if err != nil {
			return nil, err
		}
		return data, nil
	}
	result, err := b.futuresAssets(ctx)
	if err != nil {
		return nil, err
	}
	data, err := bnFuturesAssetsToAssets(result)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (b *binance) spotAssets(ctx context.Context) (*bnSpotAccount, error) {
	var res bnSpotAccount
	r := &bnhttp.Request{
		Method:   http.MethodGet,
		Endpoint: "/api/v3/account",
		SecType:  bnhttp.SecTypeSigned,
	}
	b.client.SetApiEndpoint(bnSpotEndpoint)
	data, err := b.client.CallAPI(ctx, r)
	if err != nil {
		return nil, err
	}
	err = bnhttp.Json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (b *binance) futuresAssets(ctx context.Context) ([]*bnFuturesBalance, error) {
	r := &bnhttp.Request{
		Method:   http.MethodGet,
		Endpoint: "/fapi/v2/balance",
		SecType:  bnhttp.SecTypeSigned,
	}
	b.client.SetApiEndpoint(bnFuturesEndpoint)
	data, err := b.client.CallAPI(ctx, r)
	if err != nil {
		return nil, err
	}
	res := make([]*bnFuturesBalance, 0)
	err = bnhttp.Json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (b *binance) Name() string {
	return exchange.BinanceExchange
}

func (b *binance) CreateOrder(ctx context.Context, o *exchange.CreateOrderRequest) error {
	if o.Instrument == exchange.InstrumentTypeSpot {
		return b.createSpotOrder(ctx, o)
	}
	return b.createFuturesOrder(ctx, o)
}

func (b *binance) createSpotOrder(ctx context.Context, o *exchange.CreateOrderRequest) error {
	r := &bnhttp.Request{
		APIKey:  o.APIKey,
		SecretKey: o.SecretKey,
		Method:   http.MethodPost,
		Endpoint: "/api/v3/order",
		SecType:  bnhttp.SecTypeSigned,
	}
	b.client.SetApiEndpoint(bnSpotEndpoint)
	r = r.SetFormParams(toBnSpotOrderParams(o))
	data, err := b.client.CallAPI(ctx, r)
	if err != nil {
		return err
	}
	res := &bnSpotCreateOrderResponse{}
	err = bnhttp.Json.Unmarshal(data, res)
	if err != nil {
		return err
	}
	return nil
}

func (b *binance) createFuturesOrder(ctx context.Context, o *exchange.CreateOrderRequest) error {
	r := &bnhttp.Request{
		APIKey:  o.APIKey,
		SecretKey: o.SecretKey,
		Method:   http.MethodPost,
		Endpoint: "/fapi/v1/order",
		SecType:  bnhttp.SecTypeSigned,
	}
	r = r.SetFormParams(toBnFuturesOrderParams(o))
	data, err := b.client.CallAPI(ctx, r)
	if err != nil {
		return err
	}
	res := &bnFuturesOrderResponse{}
	err = bnhttp.Json.Unmarshal(data, res)
	if err != nil {
		return err
	}
	return nil
}

func (b *binance) CancelOrder(ctx context.Context, o *exchange.CancelOrderRequest) error {
	return nil
}

func bnFuturesAssetsToAssets(b []*bnFuturesBalance) ([]exchange.Asset, error) {
	result := make([]exchange.Asset, 0)
	for _, v := range b {
		balance, err := decimal.NewFromString(v.AvailableBalance)
		if err != nil {
			return nil, err
		}
		locked, err := decimal.NewFromString(v.Balance)
		if err != nil {
			return nil, err
		}
		result = append(result, exchange.Asset{
			AssetName:  v.Asset,
			Free:       balance,
			Locked:     locked,
			Exchange:   exchange.BinanceExchange,
			Instrument: exchange.InstrumentTypeFutures,
		})
	}
	return result, nil
}

func bnSpotAssetsToAssets(s *bnSpotAccount) ([]exchange.Asset, error) {
	result := make([]exchange.Asset, 0)
	for _, v := range s.Balances {
		free, err := decimal.NewFromString(v.Free)
		if err != nil {
			return nil, err
		}
		locked, err := decimal.NewFromString(v.Locked)
		if err != nil {
			return nil, err
		}
		result = append(result, exchange.Asset{
			AssetName:  v.Asset,
			Free:       free,
			Locked:     locked,
			Exchange:   exchange.BinanceExchange,
			Instrument: exchange.InstrumentTypeSpot,
		})
	}
	return result, nil
}

func toBnFuturesOrderParams(o *exchange.CreateOrderRequest) bnhttp.Params {
	m := bnhttp.Params{
		"symbol":           o.Symbol,
		"side":             o.Side,
		"type":             o.OrderType,
		"quantity":         o.Size,
		"newOrderRespType": "RESULT",
	}
	if o.OrderType == exchange.OrderTypeLimit {
		m["timeInForce"] = "GTC"
		m["newOrderRespType"] = "ACK"
	}
	if o.PositionSide != "" {
		m["positionSide"] = o.PositionSide
	}
	if !o.Price.IsZero() {
		m["price"] = o.Price.String()
	}
	if o.ClientOrderID != "" {
		m["newClientOrderId"] = o.ClientOrderID
	}
	return m
}

func toBnSpotOrderParams(o *exchange.CreateOrderRequest) bnhttp.Params {
	// TODO: 公共参数和每个交易所的参数之间的变换，这个得后面根据具体情况再来完善
	m := bnhttp.Params{
		"symbol":      o.Symbol,
		"side":        o.Side,
		"type":        o.OrderType,
	}
	if o.TimeInForce != "" {
		m["timeInForce"] = o.TimeInForce
	}
	if !o.Size.IsZero() {
		m["quantity"] = o.Size.String()
	}
	if !o.Price.IsZero() {
		m["price"] = o.Price.String()
	}
	if o.ClientOrderID != "" {
		m["newClientOrderId"] = o.ClientOrderID
	}
	return m
}
