package okexc

import (
	"context"
	"fmt"

	"github.com/go-gotop/kit/exchange"
	"github.com/go-gotop/kit/requests/okhttp"
)

const (
	okEndpoint = "https://www.okx.com"
)

func NewOkx(cli *okhttp.Client) exchange.Exchange {
	return &okx{
		client: cli,
	}
}

type okx struct {
	client *okhttp.Client
}

func (o *okx) Name() string {
	return "okx"
}

func (o *okx) Assets(ctx context.Context, req *exchange.GetAssetsRequest) ([]exchange.Asset, error) {
	return nil, nil
}

func (o *okx) CreateOrder(ctx context.Context, req *exchange.CreateOrderRequest) error {
	r := &okhttp.Request{
		APIKey:     req.APIKey,
		SecretKey:  req.SecretKey,
		Method:     "POST",
		Endpoint:   "/api/v5/trade/order",
		SecType:    okhttp.SecTypeSigned,
		Passphrase: req.Passphrase,
	}

	o.client.SetApiEndpoint(okEndpoint)

	r = r.SetJSONBody(toOrderParams(req))
	data, err := o.client.CallAPI(ctx, r)
	if err != nil {
		return err
	}

	fmt.Println(string(data))

	return nil
}

func (o *okx) CancelOrder(ctx context.Context, req *exchange.CancelOrderRequest) error {
	return nil
}

func (o *okx) SearchOrder(ctx context.Context, req *exchange.SearchOrderRequest) (*exchange.SearchOrderResponse, error) {
	return nil, nil
}

func (o *okx) SearchTrades(ctx context.Context, req *exchange.SearchTradesRequest) ([]*exchange.SearchTradesResponse, error) {
	return nil, nil
}

func (o *okx) GetFundingRate(ctx context.Context, req *exchange.GetFundingRate) ([]*exchange.GetFundingRateResponse, error) {
	return nil, nil
}

func (o *okx) GetMarginInterestRate(ctx context.Context, req *exchange.GetMarginInterestRateRequest) ([]*exchange.GetMarginInterestRateResponse, error) {
	return nil, nil
}

func (o *okx) MarginBorrowOrRepay(ctx context.Context, req *exchange.MarginBorrowOrRepayRequest) error {
	return nil
}

func (o *okx) GetMarginInventory(ctx context.Context, req *exchange.MarginInventoryRequest) (*exchange.MarginInventory, error) {
	return nil, nil
}

func toOrderParams(o *exchange.CreateOrderRequest) okhttp.Params {
	m := okhttp.Params{
		"instId":  o.Symbol,
		"tdMode":  okxPosMode(exchange.PosModeCrossed), // 默认全仓
		"side":    okxSide(o.Side),
		"ordType": okxOrderType(o.OrderType),
	}

	if o.Instrument == exchange.InstrumentTypeFutures {
		m["sz"] = fmt.Sprintf("%v", o.Size.Mul(o.CtVal))
	} else {
		m["sz"] = fmt.Sprintf("%v", o.Size)
	}

	if !o.Price.IsZero() {
		m["px"] = o.Price
	}

	if o.ClientOrderID != "" {
		m["clOrdId"] = o.ClientOrderID
	}

	if o.PositionSide != "" {
		m["posSide"] = okxPositionSide(o.PositionSide)
	}

	return m
}
