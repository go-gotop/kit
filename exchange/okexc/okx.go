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

	params, err := o.toOrderParams(req)
	if err != nil {
		return err
	}

	r = r.SetJSONBody(params)
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

func (o *okx) convertContractCoin(typ string, instId string, sz string, opTyp string) (string, error) {
	// typ 1:币转账 2:张转币
	// instId 产品ID
	// sz 数量
	// opTyp open or close
	r := &okhttp.Request{
		Method:   "GET",
		Endpoint: "/api/v5/public/convert-contract-coin",
	}
	r.SetParams(okhttp.Params{
		"type":   typ,
		"instId": instId,
		"sz":     sz,
		"opType": opTyp,
	})
	o.client.SetApiEndpoint(okEndpoint)
	data, err := o.client.CallAPI(context.Background(), r)
	if err != nil {
		fmt.Println(err)
	}

	var result struct {
		Code string `json:"code"`
		Data []struct {
			Typ    string `json:"type"`
			InstID string `json:"instId"`
			Px     string `json:"px"`
			Sz     string `json:"sz"`
			Unit   string `json:"unit"`
		} `json:"data"`
		Msg string `json:"msg"`
	}
	err = okhttp.Json.Unmarshal(data, &result)
	if err != nil {
		return "", err
	}
	return result.Data[0].Sz, nil
}

func (o *okx) toOrderParams(req *exchange.CreateOrderRequest) (okhttp.Params, error) {
	m := okhttp.Params{
		"instId":  req.Symbol,
		"tdMode":  OkxPosMode(exchange.PosModeCross), // 默认全仓
		"side":    OkxSide(req.Side),
		"ordType": OkxOrderType(req.OrderType),
	}

	if req.Instrument == exchange.InstrumentTypeFutures {
		// 合约类型要将币转位张
		opType := "open"
		if req.Side == exchange.SideTypeSell && req.PositionSide == exchange.PositionSideLong ||
			req.Side == exchange.SideTypeBuy && req.PositionSide == exchange.PositionSideShort {
			opType = "close"
		}

		sz, err := o.convertContractCoin("1", req.Symbol, fmt.Sprintf("%v", req.Size), opType)
		if err != nil {
			return nil, err
		}
		m["sz"] = sz

	} else {
		m["sz"] = fmt.Sprintf("%v", req.Size)
	}

	if !req.Price.IsZero() {
		m["px"] = req.Price
	}

	if req.ClientOrderID != "" {
		m["clOrdId"] = req.ClientOrderID
	}

	if req.PositionSide != "" {
		m["posSide"] = OkxPositionSide(req.PositionSide)
	}

	return m, nil
}