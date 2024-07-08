package bnexc

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-gotop/kit/exchange"
	"github.com/go-gotop/kit/requests/bnhttp"
	"github.com/shopspring/decimal"
)

const (
	bnSpotEndpoint    = "https://api.binance.com"
	bnFuturesEndpoint = "https://fapi.binance.com"
)

func NewBinance(cli *bnhttp.Client) exchange.Exchange {
	return &binance{
		client: cli,
	}
}

type binance struct {
	client *bnhttp.Client
}

func (b *binance) Assets(ctx context.Context, req *exchange.GetAssetsRequest) ([]exchange.Asset, error) {
	if req.InstrumentType == exchange.InstrumentTypeSpot {
		result, err := b.spotAssets(ctx, req)
		if err != nil {
			return nil, err
		}
		data, err := bnSpotAssetsToAssets(result)
		if err != nil {
			return nil, err
		}
		return data, nil
	}
	result, err := b.futuresAssets(ctx, req)
	if err != nil {
		return nil, err
	}
	data, err := bnFuturesAssetsToAssets(result)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (b *binance) spotAssets(ctx context.Context, req *exchange.GetAssetsRequest) (*bnSpotAccount, error) {
	var res bnSpotAccount
	r := &bnhttp.Request{
		Method:    http.MethodGet,
		Endpoint:  "/api/v3/account",
		APIKey:    req.APIKey,
		SecretKey: req.SecretKey,
		SecType:   bnhttp.SecTypeSigned,
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

func (b *binance) futuresAssets(ctx context.Context, req *exchange.GetAssetsRequest) ([]*bnFuturesBalance, error) {
	r := &bnhttp.Request{
		Method:    http.MethodGet,
		Endpoint:  "/fapi/v2/balance",
		APIKey:    req.APIKey,
		SecretKey: req.SecretKey,
		SecType:   bnhttp.SecTypeSigned,
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
		APIKey:    o.APIKey,
		SecretKey: o.SecretKey,
		Method:    http.MethodPost,
		Endpoint:  "/api/v3/order",
		SecType:   bnhttp.SecTypeSigned,
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
		APIKey:    o.APIKey,
		SecretKey: o.SecretKey,
		Method:    http.MethodPost,
		Endpoint:  "/fapi/v1/order",
		SecType:   bnhttp.SecTypeSigned,
	}
	b.client.SetApiEndpoint(bnFuturesEndpoint)
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

func (b *binance) SearchOrder(ctx context.Context, o *exchange.SearchOrderRequest) (*exchange.SearchOrderResponse, error) {
	if o.InstrumentType == exchange.InstrumentTypeSpot {
		return b.searchSpotOrder(ctx, o)
	}
	return b.searchFuturesOrder(ctx, o)
}

func (b *binance) searchSpotOrder(ctx context.Context, o *exchange.SearchOrderRequest) (*exchange.SearchOrderResponse, error) {
	r := &bnhttp.Request{
		APIKey:    o.APIKey,
		SecretKey: o.SecretKey,
		Method:    http.MethodGet,
		Endpoint:  "/api/v3/order",
		SecType:   bnhttp.SecTypeSigned,
	}
	b.client.SetApiEndpoint(bnSpotEndpoint)
	params := bnhttp.Params{
		"symbol":            o.Symbol,
		"origClientOrderId": o.ClientOrderID,
		"timestamp":         time.Now().UnixNano() / 1e6,
	}
	r = r.SetParams(params)
	data, err := b.client.CallAPI(ctx, r)
	if err != nil {
		return nil, err
	}
	res := &bnSpotSearchOrderReponse{}
	err = bnhttp.Json.Unmarshal(data, res)
	if err != nil {
		return nil, err
	}
	volume, err := decimal.NewFromString(res.Volume)
	if err != nil {
		return nil, err
	}
	price, err := decimal.NewFromString(res.Price)
	if err != nil {
		return nil, err
	}
	filledQuoteVolume, err := decimal.NewFromString(res.FilledQuoteVolume)
	if err != nil {
		return nil, err
	}
	filledVolume, err := decimal.NewFromString(res.FilledVolume)
	if err != nil {
		return nil, err
	}
	avgPrice := filledQuoteVolume.Div(filledVolume)

	result := &exchange.SearchOrderResponse{
		ClientOrderID:     res.ClientOrderID,
		OrderID:           fmt.Sprintf("%d", res.OrderID),
		State:             exchange.OrderState(res.Status),
		Symbol:            res.Symbol,
		AvgPrice:          avgPrice,
		Volume:            volume,
		Price:             price,
		FilledQuoteVolume: filledQuoteVolume,
		FilledVolume:      filledVolume,
		Side:              exchange.SideType(res.Side),
		PositionSide:      exchange.PositionSideLong,
		TimeInForce:       exchange.TimeInForce(res.TimeInForce),
		OrderType:         exchange.OrderType(res.OrderType),
		CreatedTime:       res.CreatedTime,
		UpdateTime:        res.UpdateTime,
	}

	// 获取成交记录，统计手续费
	trades, err := b.SearchTrades(ctx, &exchange.SearchTradesRequest{
		APIKey:         o.APIKey,
		SecretKey:      o.SecretKey,
		Symbol:         o.Symbol,
		OrderID:        result.OrderID,
		InstrumentType: exchange.InstrumentTypeSpot,
	})
	if err != nil {
		return nil, err
	}
	if len(trades) == 0 {
		return result, nil
	}
	feeCost := decimal.Zero
	for _, v := range trades {
		feeCost = feeCost.Add(v.FeeCost)
	}

	result.FeeCost = feeCost
	result.FeeAsset = trades[0].FeeAsset
	result.By = trades[0].By

	return result, nil
}

func (b *binance) searchFuturesOrder(ctx context.Context, o *exchange.SearchOrderRequest) (*exchange.SearchOrderResponse, error) {
	r := &bnhttp.Request{
		APIKey:    o.APIKey,
		SecretKey: o.SecretKey,
		Method:    http.MethodGet,
		Endpoint:  "/fapi/v1/order",
		SecType:   bnhttp.SecTypeSigned,
	}
	b.client.SetApiEndpoint(bnSpotEndpoint)
	params := bnhttp.Params{
		"symbol":            o.Symbol,
		"origClientOrderId": o.ClientOrderID,
		"timestamp":         time.Now().UnixNano() / 1e6,
	}
	r = r.SetParams(params)
	data, err := b.client.CallAPI(ctx, r)
	if err != nil {
		return nil, err
	}
	res := &bnFuturesSearchOrderResponse{}
	err = bnhttp.Json.Unmarshal(data, res)
	if err != nil {
		return nil, err
	}
	volume, err := decimal.NewFromString(res.Volume)
	if err != nil {
		return nil, err
	}
	price, err := decimal.NewFromString(res.Price)
	if err != nil {
		return nil, err
	}
	filledQuoteVolume, err := decimal.NewFromString(res.FilledQuoteVolume)
	if err != nil {
		return nil, err
	}
	filledVolume, err := decimal.NewFromString(res.FilledVolume)
	if err != nil {
		return nil, err
	}
	avgPrice, err := decimal.NewFromString(res.AvgPrice)
	if err != nil {
		return nil, err
	}

	result := &exchange.SearchOrderResponse{
		ClientOrderID:     res.ClientOrderID,
		OrderID:           fmt.Sprintf("%d", res.OrderID),
		State:             exchange.OrderState(res.Status),
		Symbol:            res.Symbol,
		AvgPrice:          avgPrice,
		Volume:            volume,
		Price:             price,
		FilledQuoteVolume: filledQuoteVolume,
		FilledVolume:      filledVolume,
		Side:              exchange.SideType(res.Side),
		PositionSide:      exchange.PositionSideLong,
		TimeInForce:       exchange.TimeInForce(res.TimeInForce),
		OrderType:         exchange.OrderType(res.OrderType),
		CreatedTime:       res.CreatedTime,
		UpdateTime:        res.UpdateTime,
	}

	// 获取成交记录，统计手续费
	trades, err := b.SearchTrades(ctx, &exchange.SearchTradesRequest{
		APIKey:         o.APIKey,
		SecretKey:      o.SecretKey,
		Symbol:         o.Symbol,
		OrderID:        result.OrderID,
		InstrumentType: exchange.InstrumentTypeSpot,
	})
	if err != nil {
		return nil, err
	}
	if len(trades) == 0 {
		return result, nil
	}
	feeCost := decimal.Zero
	for _, v := range trades {
		feeCost = feeCost.Add(v.FeeCost)
	}

	result.FeeCost = feeCost
	result.FeeAsset = trades[0].FeeAsset
	result.By = trades[0].By

	return result, nil
}

func (b *binance) SearchTrades(ctx context.Context, o *exchange.SearchTradesRequest) ([]*exchange.SearchTradesResponse, error) {
	if o.InstrumentType == exchange.InstrumentTypeSpot {
		return b.searchSpotTrades(ctx, o)
	}
	return b.searchFuturesTrades(ctx, o)
}

func (b *binance) searchSpotTrades(ctx context.Context, o *exchange.SearchTradesRequest) ([]*exchange.SearchTradesResponse, error) {
	r := &bnhttp.Request{
		APIKey:    o.APIKey,
		SecretKey: o.SecretKey,
		Method:    http.MethodGet,
		Endpoint:  "/api/v3/myTrades",
		SecType:   bnhttp.SecTypeSigned,
	}
	b.client.SetApiEndpoint(bnSpotEndpoint)
	params := bnhttp.Params{
		"symbol":  o.Symbol,
		"orderId": o.OrderID,
	}
	r = r.SetParams(params)
	data, err := b.client.CallAPI(ctx, r)
	if err != nil {
		return nil, err
	}
	res := make([]*bnSpotTrades, 0)
	err = bnhttp.Json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	result := make([]*exchange.SearchTradesResponse, 0)
	for _, v := range res {
		quantity, err := decimal.NewFromString(v.Quantity)
		if err != nil {
			return nil, err
		}
		price, err := decimal.NewFromString(v.Price)
		if err != nil {
			return nil, err
		}
		feeCost, err := decimal.NewFromString(v.Commission)
		if err != nil {
			return nil, err
		}
		by := exchange.ByTaker
		if v.IsMaker {
			by = exchange.ByMaker
		}
		result = append(result, &exchange.SearchTradesResponse{
			Symbol:   v.Symbol,
			ID:       fmt.Sprintf("%d", v.ID),
			OrderID:  fmt.Sprintf("%d", v.OrderID),
			Price:    price,
			Volume:   quantity,
			FeeCost:  feeCost,
			FeeAsset: v.CommissionAsset,
			Time:     v.Time,
			By:       by,
		})
	}
	return result, nil
}

func (b *binance) searchFuturesTrades(ctx context.Context, o *exchange.SearchTradesRequest) ([]*exchange.SearchTradesResponse, error) {
	r := &bnhttp.Request{
		APIKey:    o.APIKey,
		SecretKey: o.SecretKey,
		Method:    http.MethodGet,
		Endpoint:  "/fapi/v1/userTrades",
		SecType:   bnhttp.SecTypeSigned,
	}
	b.client.SetApiEndpoint(bnFuturesEndpoint)
	params := bnhttp.Params{
		"symbol":  o.Symbol,
		"orderId": o.OrderID,
	}
	r = r.SetParams(params)
	data, err := b.client.CallAPI(ctx, r)
	if err != nil {
		return nil, err
	}
	res := make([]*bnFuturesTrades, 0)
	err = bnhttp.Json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	result := make([]*exchange.SearchTradesResponse, 0)
	for _, v := range res {
		quantity, err := decimal.NewFromString(v.Quantity)
		if err != nil {
			return nil, err
		}
		price, err := decimal.NewFromString(v.Price)
		if err != nil {
			return nil, err
		}
		feeCost, err := decimal.NewFromString(v.Commission)
		if err != nil {
			return nil, err
		}
		by := exchange.ByTaker
		if v.IsMaker {
			by = exchange.ByMaker
		}
		result = append(result, &exchange.SearchTradesResponse{
			Symbol:   v.Symbol,
			ID:       fmt.Sprintf("%d", v.ID),
			OrderID:  fmt.Sprintf("%d", v.OrderID),
			Price:    price,
			Volume:   quantity,
			FeeCost:  feeCost,
			FeeAsset: v.CommissionAsset,
			Time:     v.Time,
			By:       by,
		})
	}

	return result, nil
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
		"symbol": o.Symbol,
		"side":   o.Side,
		"type":   o.OrderType,
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
