package moexc

import (
	"context"
	"net/http"

	"github.com/go-gotop/kit/exchange"
	"github.com/go-gotop/kit/requests/mohttp"
	"github.com/shopspring/decimal"
)

const (
	mockExchangEndpoint = ""
)

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

func (m *mockExchange) Assets(ctx context.Context, it exchange.InstrumentType) ([]exchange.Asset, error) {
	var res []mockBalance
	r := &mohttp.Request{
		Method:   http.MethodGet,
		Endpoint: "/api/exchange/assets",
	}
	m.client.SetApiEndpoint(mockExchangEndpoint)
	data, err := m.client.CallAPI(ctx, r)
	if err != nil {
		return nil, err
	}
	err = mohttp.Json.Unmarshal(data, &res)
	if err != nil {
		return nil, err
	}
	result := make([]exchange.Asset, 0)
	for _, v := range res {
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
		Method:   http.MethodPost,
		Endpoint: "/api/exchange/order",
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
