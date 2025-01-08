package bnexc

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-gotop/kit/exchange"
	"github.com/go-gotop/kit/requests/bnhttp"
	"github.com/shopspring/decimal"
)

const (
	bnSpotEndpoint            = "https://api.binance.com"
	bnFuturesEndpoint         = "https://fapi.binance.com"
	bnPortfolioMarginEndpoint = "https://papi.binance.com"
)

func NewBinance(cli *bnhttp.Client) exchange.Exchange {
	return &binance{
		client: cli,
	}
}

type binance struct {
	client *bnhttp.Client
}

func (b *binance) Name() string {
	return exchange.BinanceExchange
}

func (b *binance) GetDepth(ctx context.Context, req *exchange.GetDepthRequest) (exchange.GetDepthResponse, error) {
	r := &bnhttp.Request{
		Method:   http.MethodGet,
		Endpoint: "/api/v3/depth",
		SecType:  bnhttp.SecTypeNone,
	}
	r = r.SetParams(bnhttp.Params{"symbol": req.Symbol.OriginalSymbol, "limit": req.Limit})
	if req.InstrumentType == exchange.InstrumentTypeFutures {
		r.Endpoint = "/fapi/v1/depth"
		b.client.SetApiEndpoint(bnFuturesEndpoint)
	} else {
		b.client.SetApiEndpoint(bnSpotEndpoint)
	}
	data, err := b.client.CallAPI(ctx, r)
	if err != nil {
		return exchange.GetDepthResponse{}, err
	}

	var response bnDepthResponse

	err = bnhttp.Json.Unmarshal(data, &response)
	if err != nil {
		return exchange.GetDepthResponse{}, err
	}

	asks := make([][]decimal.Decimal, 0, len(response.Ask))
	for _, v := range response.Ask {
		price, err := decimal.NewFromString(v[0])
		if err != nil {
			return exchange.GetDepthResponse{}, err
		}
		quantity, err := decimal.NewFromString(v[1])
		if err != nil {
			return exchange.GetDepthResponse{}, err
		}
		asks = append(asks, []decimal.Decimal{price, quantity})
	}

	bids := make([][]decimal.Decimal, 0, len(response.Bid))
	for _, v := range response.Bid {
		price, err := decimal.NewFromString(v[0])
		if err != nil {
			return exchange.GetDepthResponse{}, err
		}
		quantity, err := decimal.NewFromString(v[1])
		if err != nil {
			return exchange.GetDepthResponse{}, err
		}
		bids = append(bids, []decimal.Decimal{price, quantity})
	}

	return exchange.GetDepthResponse{
		Asks: asks,
		Bids: bids,
		Ts:   response.Ts,
	}, nil
}

func (b *binance) SetLeverage(ctx context.Context, req *exchange.SetLeverageRequest) error {
	return nil
}

func (b *binance) GetMarkPriceKline(ctx context.Context, req *exchange.GetMarkPriceKlineRequest) ([]exchange.GetMarkPriceKlineResponse, error) {
	return nil, errors.New("not implemented")
}

func (b *binance) GetKline(ctx context.Context, req *exchange.GetKlineRequest) ([]exchange.GetKlineResponse, error) {
	var r *bnhttp.Request

	if req.InstrumentType == exchange.InstrumentTypeSpot {
		r = &bnhttp.Request{
			Method:   http.MethodGet,
			Endpoint: "/api/v3/klines",
			SecType:  bnhttp.SecTypeNone,
		}
		b.client.SetApiEndpoint(bnSpotEndpoint)
	} else if req.InstrumentType == exchange.InstrumentTypeFutures {
		r = &bnhttp.Request{
			Method:   http.MethodGet,
			Endpoint: "/fapi/v1/klines",
			SecType:  bnhttp.SecTypeNone,
		}
		b.client.SetApiEndpoint(bnFuturesEndpoint)
	}

	params := bnhttp.Params{
		"symbol":   req.Symbol.OriginalSymbol,
		"interval": req.Period,
	}

	if req.Start != 0 {
		params["startTime"] = req.Start
	}
	if req.End != 0 {
		params["endTime"] = req.End
	}
	if req.Limit != 0 {
		params["limit"] = req.Limit
	}

	r = r.SetParams(params)
	data, err := b.client.CallAPI(ctx, r)
	if err != nil {
		return nil, err
	}

	var klines [][]interface{}
	err = bnhttp.Json.Unmarshal(data, &klines)
	if err != nil {
		return nil, err
	}

	result := make([]exchange.GetKlineResponse, 0, len(klines))
	for _, k := range klines {
		open, err := decimal.NewFromString(k[1].(string))
		if err != nil {
			return nil, err
		}
		high, err := decimal.NewFromString(k[2].(string))
		if err != nil {
			return nil, err
		}
		low, err := decimal.NewFromString(k[3].(string))
		if err != nil {
			return nil, err
		}
		close, err := decimal.NewFromString(k[4].(string))
		if err != nil {
			return nil, err
		}
		volume, err := decimal.NewFromString(k[5].(string))
		if err != nil {
			return nil, err
		}
		quoteVolume, err := decimal.NewFromString(k[7].(string))
		if err != nil {
			return nil, err
		}

		result = append(result, exchange.GetKlineResponse{
			Symbol:      req.Symbol.UnifiedSymbol,
			OpenTime:    int64(k[0].(float64)),
			Open:        open,
			High:        high,
			Low:         low,
			Close:       close,
			Volume:      volume,
			QuoteVolume: quoteVolume,
			Confirm:     "1",
		})
	}
	return result, nil
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

func (b *binance) GetAccountConfig(ctx context.Context, req *exchange.GetAccountConfigRequest) (exchange.GetAccountConfigResponse, error) {
	return exchange.GetAccountConfigResponse{}, errors.New("not implemented")
}

func (b *binance) CreateOrder(ctx context.Context, o *exchange.CreateOrderRequest) error {
	if o.Instrument == exchange.InstrumentTypeSpot {
		return b.createSpotOrder(ctx, o)
	} else if o.Instrument == exchange.InstrumentTypeFutures {
		return b.createFuturesOrder(ctx, o)
	} else if o.Instrument == exchange.InstrumentTypeMargin {
		return b.createMarginOrder(ctx, o)
	}
	return exchange.ErrInstrumentTypeNotSupported
}

func (b *binance) SearchOrder(ctx context.Context, o *exchange.SearchOrderRequest) (*exchange.SearchOrderResponse, error) {
	if o.InstrumentType == exchange.InstrumentTypeSpot {
		return b.searchSpotOrder(ctx, o)
	}
	return b.searchFuturesOrder(ctx, o)
}

func (b *binance) SearchTrades(ctx context.Context, o *exchange.SearchTradesRequest) ([]*exchange.SearchTradesResponse, error) {
	if o.InstrumentType == exchange.InstrumentTypeSpot {
		return b.searchSpotTrades(ctx, o)
	}
	return b.searchFuturesTrades(ctx, o)
}

func (b *binance) GetFundingRate(ctx context.Context, req *exchange.GetFundingRate) ([]*exchange.GetFundingRateResponse, error) {
	b.client.SetApiEndpoint(bnFuturesEndpoint)
	if req.Symbol != "" {
		return b.getSingleFundingRate(ctx, req.Symbol)
	}
	return b.getAllFundingRates(ctx)
}

func (b *binance) GetMarginInterestRate(ctx context.Context, req *exchange.GetMarginInterestRateRequest) ([]*exchange.GetMarginInterestRateResponse, error) {
	r := &bnhttp.Request{
		APIKey:    req.APIKey,
		SecretKey: req.SecretKey,
		Method:    http.MethodGet,
		Endpoint:  "/sapi/v1/margin/next-hourly-interest-rate",
		SecType:   bnhttp.SecTypeSigned,
	}
	b.client.SetApiEndpoint(bnSpotEndpoint)
	r = r.SetParams(bnhttp.Params{"assets": req.Assets, "isIsolated": req.IsIsolated})
	data, err := b.client.CallAPI(ctx, r)
	if err != nil {
		return nil, err
	}
	var res []*bnMarginInterestRate
	err = bnhttp.Json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	result := make([]*exchange.GetMarginInterestRateResponse, 0, len(res))
	for _, v := range res {
		interestRate, err := decimal.NewFromString(v.NextHourlyInterestRate)
		if err != nil {
			return nil, err
		}
		result = append(result, &exchange.GetMarginInterestRateResponse{
			Asset:                  v.Asset,
			NextHourlyInterestRate: interestRate,
		})
	}
	return result, nil
}

func (b *binance) MarginBorrowOrRepay(ctx context.Context, req *exchange.MarginBorrowOrRepayRequest) error {
	r := &bnhttp.Request{
		APIKey:    req.APIKey,
		SecretKey: req.SecretKey,
		Method:    http.MethodPost,
		Endpoint:  "/sapi/v1/margin/borrow-repay",
		SecType:   bnhttp.SecTypeSigned,
	}

	b.client.SetApiEndpoint(bnSpotEndpoint)
	r = r.SetFormParams(bnhttp.Params{
		"asset":      req.Asset,
		"amount":     req.Amount,
		"isIsolated": req.IsIsolated,
		"symbol":     req.Symbol,
		"type":       req.Typ,
	})
	data, err := b.client.CallAPI(ctx, r)
	if err != nil {
		return err
	}

	type result struct {
		TranId int64 `json:"tranId"`
	}

	fmt.Printf("data: %s\n", string(data))

	res := &result{}
	err = bnhttp.Json.Unmarshal(data, res)
	if err != nil {
		return err
	}
	return nil
}

func (b *binance) GetMarginInventory(ctx context.Context, req *exchange.MarginInventoryRequest) (*exchange.MarginInventory, error) {
	r := &bnhttp.Request{
		APIKey:    req.APIKey,
		SecretKey: req.SecretKey,
		Method:    http.MethodGet,
		Endpoint:  "/sapi/v1/margin/available-inventory",
		SecType:   bnhttp.SecTypeSigned,
	}
	b.client.SetApiEndpoint(bnSpotEndpoint)
	r = r.SetParams(bnhttp.Params{"type": req.Typ})
	data, err := b.client.CallAPI(ctx, r)
	if err != nil {
		return nil, err
	}

	var res bnMarginInventory
	err = bnhttp.Json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	return &exchange.MarginInventory{
		Assets: res.Assets,
	}, nil
}

func (b *binance) GetMaxSize(ctx context.Context, req *exchange.GetMaxSizeRequest) ([]exchange.GetMaxSizeResponse, error) {
	return nil, nil
}

func (b *binance) GetPosition(ctx context.Context, req *exchange.GetPositionRequest) ([]*exchange.GetPositionResponse, error) {
	return nil, nil
}

func (b *binance) GetHistoryPosition(ctx context.Context, req *exchange.GetPositionHistoryRequest) error {
	return nil
}

func (b *binance) ConvertContractCoin(typ string, symbol exchange.Symbol, sz string, opTyp string) (string, error) {
	return "", errors.New("not implemented")
}

func (b *binance) getSingleFundingRate(ctx context.Context, symbol string) ([]*exchange.GetFundingRateResponse, error) {
	r := &bnhttp.Request{
		Method:   http.MethodGet,
		Endpoint: "/fapi/v1/premiumIndex",
		SecType:  bnhttp.SecTypeNone,
	}

	r = r.SetParams(bnhttp.Params{"symbol": symbol})
	data, err := b.client.CallAPI(ctx, r)
	if err != nil {
		return nil, err
	}
	var result bnPremiumIndex
	err = bnhttp.Json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}
	return []*exchange.GetFundingRateResponse{b.convertFundingRate(&result)}, nil
}

func (b *binance) getAllFundingRates(ctx context.Context) ([]*exchange.GetFundingRateResponse, error) {
	r := &bnhttp.Request{
		Method:   http.MethodGet,
		Endpoint: "/fapi/v1/premiumIndex",
		SecType:  bnhttp.SecTypeNone,
	}
	data, err := b.client.CallAPI(ctx, r)
	if err != nil {
		return nil, err
	}
	var results []*bnPremiumIndex
	err = bnhttp.Json.Unmarshal(data, &results)
	if err != nil {
		return nil, err
	}
	res := make([]*exchange.GetFundingRateResponse, 0, len(results))
	for _, v := range results {
		res = append(res, b.convertFundingRate(v))
	}
	return res, nil
}

func (b *binance) convertFundingRate(data *bnPremiumIndex) *exchange.GetFundingRateResponse {
	markPrice, err := decimal.NewFromString(data.MarkPrice)
	if err != nil {
		return nil
	}
	indexPrice, err := decimal.NewFromString(data.IndexPrice)
	if err != nil {
		return nil
	}
	estimatedSettlePrice, err := decimal.NewFromString(data.EstimatedSettlePrice)
	if err != nil {
		return nil
	}
	lastFundingRate, err := decimal.NewFromString(data.LastFundingRate)
	if err != nil {
		return nil
	}
	interestRate, err := decimal.NewFromString(data.InterestRate)
	if err != nil {
		return nil
	}
	return &exchange.GetFundingRateResponse{
		Symbol:               data.Symbol,
		MarkPrice:            markPrice,
		IndexPrice:           indexPrice,
		EstimatedSettlePrice: estimatedSettlePrice,
		LastFundingRate:      lastFundingRate,
		NextFundingTime:      data.NextFundingTime,
		InterestRate:         interestRate,
		Time:                 data.Time,
	}
}

func (b *binance) CancelOrder(ctx context.Context, o *exchange.CancelOrderRequest) error {
	return nil
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

// TOFIX:创建杠杠订单默认自动借款和还款，后期按需要把该参数抽离出来sideEffectType
func (b *binance) createMarginOrder(ctx context.Context, o *exchange.CreateOrderRequest) error {
	r := &bnhttp.Request{
		APIKey:    o.APIKey,
		SecretKey: o.SecretKey,
		Method:    http.MethodPost,
		Endpoint:  "/sapi/v1/margin/order",
		SecType:   bnhttp.SecTypeSigned,
	}
	if o.IsUnifiedAccount {
		r.Endpoint = "/papi/v1/margin/order"
		b.client.SetApiEndpoint(bnPortfolioMarginEndpoint)
	} else {
		b.client.SetApiEndpoint(bnSpotEndpoint)
	}
	r = r.SetFormParams(toBnMarginOrderParams(o))
	data, err := b.client.CallAPI(ctx, r)
	if err != nil {
		return err
	}
	res := &bnMarginCreateOrderResponse{}
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
	if o.IsUnifiedAccount {
		r.Endpoint = "/papi/v1/um/order"
		b.client.SetApiEndpoint(bnPortfolioMarginEndpoint)
	} else {
		b.client.SetApiEndpoint(bnFuturesEndpoint)
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
		"symbol":            o.Symbol.OriginalSymbol,
		"origClientOrderId": o.ClientOrderID,
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
		TimeInForce:       exchange.TimeInForce(res.TimeInForce),
		OrderType:         exchange.OrderType(res.OrderType),
		CreatedTime:       res.CreatedTime,
		UpdateTime:        res.UpdateTime,
	}

	// 获取成交记录，统计手续费
	trades, err := b.SearchTrades(ctx, &exchange.SearchTradesRequest{
		APIKey:         o.APIKey,
		SecretKey:      o.SecretKey,
		Symbol:         o.Symbol.OriginalSymbol,
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
		"symbol":            o.Symbol.OriginalSymbol,
		"origClientOrderId": o.ClientOrderID,
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
		Symbol:         o.Symbol.OriginalSymbol,
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
		"symbol":           o.Symbol.OriginalSymbol,
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
		"symbol": o.Symbol.OriginalSymbol,
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

func toBnMarginOrderParams(o *exchange.CreateOrderRequest) bnhttp.Params {
	m := bnhttp.Params{
		"symbol": o.Symbol.OriginalSymbol,
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
	// 杠杠默认自动借款和还款
	m["sideEffectType"] = "AUTO_BORROW_REPAY"
	return m
}
