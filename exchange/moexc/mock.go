package moexc

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-gotop/kit/exchange"
	"github.com/go-gotop/kit/requests/mohttp"
	"github.com/shopspring/decimal"
)

const (
	mockExchangEndpoint = "http://192.168.1.105:8070"
)

type ApiResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func NewMockExchange(cli *mohttp.Client) exchange.Exchange {
	return &mockExchange{
		client: cli,
	}
}

type mockExchange struct {
	client *mohttp.Client
}

func (m *mockExchange) Name() string {
	return exchange.MockExchange
}

func (m *mockExchange) Assets(ctx context.Context, req *exchange.GetAssetsRequest) ([]exchange.Asset, error) {
	var response ApiResponse
	r := &mohttp.Request{
		Method:    http.MethodGet,
		Endpoint:  "/api/exchange/assets",
		SecType:   mohttp.SecTypeAPIKey,
		APIKey:    req.APIKey,
		SecretKey: req.SecretKey,
	}
	m.client.SetApiEndpoint(mockExchangEndpoint)
	r.SetParam("instrumentType", req.InstrumentType)
	data, err := m.client.CallAPI(ctx, r)
	if err != nil {
		return nil, err
	}
	err = mohttp.Json.Unmarshal(data, &response)
	if err != nil {
		return nil, err
	}

	// 断言response.Data为[]mockBalance类型
	dataSlice, ok := response.Data.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data type")
	}

	var balances []mockBalance
	for _, item := range dataSlice {
		balanceMap, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid balance data type")
		}
		balance := mockBalance{
			AssetName:  balanceMap["assetName"].(string),
			Exchange:   balanceMap["exchange"].(string),
			Instrument: balanceMap["instrument"].(string),
			Free:       balanceMap["free"].(string),
			Locked:     balanceMap["locked"].(string),
		}
		balances = append(balances, balance)
	}

	result := make([]exchange.Asset, 0)
	for _, v := range balances {
		free, err := decimal.NewFromString(v.Free)
		if err != nil {
			return nil, err
		}
		locked, err := decimal.NewFromString(v.Locked)
		if err != nil {
			return nil, err
		}
		result = append(result, exchange.Asset{
			AssetName:  v.AssetName,
			Exchange:   v.Exchange,
			Instrument: exchange.InstrumentType(v.Instrument),
			Free:       free,
			Locked:     locked,
		})

	}
	return result, nil
}

func (m *mockExchange) CreateOrder(ctx context.Context, o *exchange.CreateOrderRequest) error {
	r := &mohttp.Request{
		Method:    http.MethodPost,
		Endpoint:  "/api/exchange/order",
		APIKey:    o.APIKey,
		SecretKey: o.SecretKey,
		SecType:   mohttp.SecTypeAPIKey,
	}
	m.client.SetApiEndpoint(mockExchangEndpoint)
	params := mohttp.Params{
		"orderTime":     o.OrderTime,
		"symbol":        o.Symbol,
		"clientOrderId": o.ClientOrderID,
		"side":          o.Side,
		"orderType":     o.OrderType,
		"positionSide":  o.PositionSide,
		"timeInForce":   o.TimeInForce,
		"instrument":    o.Instrument,
		"size":          o.Size,
		"price":         o.Price,
	}
	r = r.SetFormParams(params)
	data, err := m.client.CallAPI(ctx, r)
	if err != nil {
		return err
	}
	res := &mockCreateOrderResponse{}
	err = mohttp.Json.Unmarshal(data, res)
	if err != nil {
		return err
	}
	return nil
}

func (m *mockExchange) CancelOrder(ctx context.Context, o *exchange.CancelOrderRequest) error {
	return nil
}

func (m *mockExchange) SearchOrder(ctx context.Context, o *exchange.SearchOrderRequest) (*exchange.SearchOrderResponse, error) {
	return nil, nil
}

func (m *mockExchange) SearchTrades(ctx context.Context, o *exchange.SearchTradesRequest) ([]*exchange.SearchTradesResponse, error) {
	return nil, nil
}
