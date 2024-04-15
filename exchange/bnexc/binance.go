package bnexc

import (
	"context"
	"errors"
	"net/http"
	"strings"

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

func (b *binance) Symbols(ctx context.Context, it exchange.InstrumentType) ([]exchange.Symbol, error) {
	if it == exchange.InstrumentTypeSpot {
		result, err := b.spotSymbol(ctx)
		if err != nil {
			return nil, err
		}
		data, err := bnSpotSymbolsToSymbols(result)
		if err != nil {
			return nil, err
		}
		return data, nil
	}
	result, err := b.futuresSymbol(ctx)
	if err != nil {
		return nil, err
	}
	data, err := bnFuturesSymbolsToSymbols(result)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (b *binance) futuresSymbol(ctx context.Context) (*bnFuturesExchangeInfo, error) {
	var res bnFuturesExchangeInfo
	r := &bnhttp.Request{
		Method:   http.MethodGet,
		Endpoint: "/fapi/v1/exchangeInfo",
		SecType:  bnhttp.SecTypeNone,
	}
	b.client.SetApiEndpoint(bnFuturesEndpoint)
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

func (b *binance) spotSymbol(ctx context.Context) (*bnSpotExchangeInfo, error) {
	var res bnSpotExchangeInfo
	r := &bnhttp.Request{
		Method:   http.MethodGet,
		Endpoint: "/api/v3/exchangeInfo",
		SecType:  bnhttp.SecTypeNone,
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

func (b *binance) CreateOrder(ctx context.Context, o *exchange.CreateOrderRequest) error {
	if o.Instrument == exchange.InstrumentTypeSpot {
		return b.createSpotOrder(ctx, o)
	}
	return b.createFuturesOrder(ctx, o)
}

func (b *binance) createSpotOrder(ctx context.Context, o *exchange.CreateOrderRequest) error {
	r := &bnhttp.Request{
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

// findFirstNonZeroDigitAfterDecimal 接受 interface{} 类型的参数，
// 返回小数点后第一个非零数字是第几位。
func findFirstNonZeroDigitAfterDecimal(value interface{}) (int, error) {
	var strValue string
	switch v := value.(type) {
	case string:
		strValue = v
	default:
		return 0, errors.New("unsupported type")
	}

	// Find the position of the decimal point
	dotIndex := strings.Index(strValue, ".")
	if dotIndex == -1 {
		return 0, nil
	}

	// Traverse the string after the decimal point to find the first non-zero digit
	for i := dotIndex + 1; i < len(strValue); i++ {
		if strValue[i] != '0' {
			// Calculate the position of the first non-zero digit after the decimal point
			return i - dotIndex, nil
		}
	}

	return 0, nil
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

func bnSpotSymbolsToSymbols(s *bnSpotExchangeInfo) ([]exchange.Symbol, error) {
	result := make([]exchange.Symbol, 0)
	for _, v := range s.Symbols {
		if s, exist := exchange.ReverseBinanceSymbols[v.Symbol]; exist {
			lotSizeFilter := v.LotSizeFilter()
			maxSize, err := decimal.NewFromString(lotSizeFilter.MaxQuantity)
			if err != nil {
				return nil, err
			}
			minSize, err := decimal.NewFromString(lotSizeFilter.MinQuantity)
			if err != nil {
				return nil, err
			}
			priceFilter := v.PriceFilter()
			minPrice, err := decimal.NewFromString(priceFilter.MinPrice)
			if err != nil {
				return nil, err
			}
			maxPrice, err := decimal.NewFromString(priceFilter.MaxPrice)
			if err != nil {
				return nil, err
			}
			pp, err := findFirstNonZeroDigitAfterDecimal(priceFilter.TickSize)
			if err != nil {
				return nil, err
			}
			sp, err := findFirstNonZeroDigitAfterDecimal(lotSizeFilter.MinQuantity)
			if err != nil {
				return nil, err
			}
			// 字段补全
			result = append(result, exchange.Symbol{
				SymbolName:     s,
				MaxSize:        maxSize,
				MinSize:        minSize,
				MinPrice:       minPrice,
				MaxPrice:       maxPrice,
				SizePrecision:  int32(sp),
				PricePrecision: int32(pp),
				Status:         exchange.SymbolStatusTrading,
				Exchange:       exchange.BinanceExchange,
				Instrument:     exchange.InstrumentTypeSpot,
				AssetName:      v.QuoteAsset,
			})
		}
	}
	return result, nil
}

func bnFuturesSymbolsToSymbols(s *bnFuturesExchangeInfo) ([]exchange.Symbol, error) {
	result := make([]exchange.Symbol, 0)
	for _, v := range s.Symbols {
		if s, exist := exchange.ReverseBinanceSymbols[v.Symbol]; exist {
			lotSizeFilter := v.LotSizeFilter()
			maxSize, err := decimal.NewFromString(lotSizeFilter.MaxQuantity)
			if err != nil {
				return nil, err
			}
			minSize, err := decimal.NewFromString(lotSizeFilter.MinQuantity)
			if err != nil {
				return nil, err
			}
			priceFilter := v.PriceFilter()
			minPrice, err := decimal.NewFromString(priceFilter.MinPrice)
			if err != nil {
				return nil, err
			}
			maxPrice, err := decimal.NewFromString(priceFilter.MaxPrice)
			if err != nil {
				return nil, err
			}
			pp, err := findFirstNonZeroDigitAfterDecimal(priceFilter.TickSize)
			if err != nil {
				return nil, err
			}
			sp, err := findFirstNonZeroDigitAfterDecimal(lotSizeFilter.MinQuantity)
			if err != nil {
				return nil, err
			}
			// TOFIX: 字段补全
			result = append(result, exchange.Symbol{
				SymbolName:     s,
				MaxSize:        maxSize,
				MinSize:        minSize,
				MinPrice:       minPrice,
				MaxPrice:       maxPrice,
				SizePrecision:  int32(sp),
				PricePrecision: int32(pp),
				Status:         exchange.SymbolStatusTrading,
				Exchange:       exchange.BinanceExchange,
				Instrument:     exchange.InstrumentTypeFutures,
				AssetName:      v.QuoteAsset,
			})
		}
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
		"timeInForce": o.TimeInForce,
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
