package okexc

import (
	"context"
	"fmt"

	"github.com/go-gotop/kit/exchange"
	"github.com/go-gotop/kit/requests/okhttp"
	"github.com/shopspring/decimal"
)

const (
	okEndpoint = "https://aws.okx.com"
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
	return exchange.OkxExchange
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

// typ：1-币转账 2-张转币; symbol: 交易对; sz：数量; opTyp: open（舍位），close（四舍五入）
func (o *okx) ConvertContractCoin(typ string, symbol exchange.Symbol, sz string, opTyp string) (string, error) {
	if opTyp == "" {
		opTyp = "open"
	}

	size, err := decimal.NewFromString(sz)
	if err != nil {
		return "", err
	}
	if typ == "1" {
		// 币转张, 数量除以张的面值
		if symbol.CtVal.IsZero() {
			return "", fmt.Errorf("invalid ctVal: %v", symbol.CtVal)
		}
		size = size.Div(symbol.CtVal)
		size = o.sizePrecision(size, symbol, opTyp)
		return size.String(), nil
	} else if typ == "2" {
		// 张转币，张的面值乘以数量
		size = size.Mul(symbol.CtVal)
		return size.String(), nil
	}
	return "", fmt.Errorf("invalid type: %v", typ)
}

func (o *okx) getMarketPrice(instId string, instType string) (decimal.Decimal, error) {
	r := &okhttp.Request{
		Method:   "GET",
		Endpoint: "/api/v5/public/mark-price",
	}
	r.SetParams(okhttp.Params{
		"instId":   instId,
		"instType": instType,
	})
	o.client.SetApiEndpoint(okEndpoint)
	data, err := o.client.CallAPI(context.Background(), r)
	if err != nil {
		fmt.Println(err)
	}
	var result struct {
		Code string `json:"code"`
		Data []struct {
			InstType  string `json:"instType"`
			InstID    string `json:"instId"`
			MarkPx    string `json:"markPx"`
			Timestamp string `json:"ts"`
		} `json:"data"`
		Msg string `json:"msg"`
	}
	err = okhttp.Json.Unmarshal(data, &result)
	if err != nil {
		return decimal.Zero, err
	}
	mkp, err := decimal.NewFromString(result.Data[0].MarkPx)
	if err != nil {
		return decimal.Zero, err
	}
	return mkp, nil
}

func (o *okx) toOrderParams(req *exchange.CreateOrderRequest) (okhttp.Params, error) {
	m := okhttp.Params{
		"instId":  req.Symbol,
		"side":    OkxSide(req.Side),
		"ordType": OkxOrderType(req.OrderType),
	}

	if req.Instrument == exchange.InstrumentTypeFutures {
		// 合约类型要将币转位张
		m["tdMode"] = OkxPosMode(exchange.PosModeCross) // 默认全仓
		opType := "open"
		if req.Side == exchange.SideTypeSell && req.PositionSide == exchange.PositionSideLong ||
			req.Side == exchange.SideTypeBuy && req.PositionSide == exchange.PositionSideShort {
			opType = "close"
		}

		sz, err := o.ConvertContractCoin("1", req.Symbol, fmt.Sprintf("%v", req.Size), opType)
		if err != nil {
			return nil, err
		}
		m["sz"] = sz

	} else if req.Instrument == exchange.InstrumentTypeSpot {
		m["tgtCcy"] = "base_ccy" // 指定size为交易币种
		m["tdMode"] = "cash"
		m["sz"] = fmt.Sprintf("%v", req.Size)
	} else if req.Instrument == exchange.InstrumentTypeMargin {
		// okx 杠杠买入时，size 为 计价货币，所以这里要转换
		m["tdMode"] = OkxPosMode(exchange.PosModeCross) // 默认全仓
		if req.Side == exchange.SideTypeSell {
			m["sz"] = fmt.Sprintf("%v", req.Size)
		} else {
			mkp, err := o.getMarketPrice(req.Symbol.OriginalSymbol, "MARGIN")
			if err != nil {
				return nil, err
			}
			m["sz"] = fmt.Sprintf("%v", req.Size.Mul(mkp))
		}

	}

	if req.Instrument == exchange.InstrumentTypeMargin {
		// TOFIX: 保证金模式
		m["ccy"] = "USDT"
	}

	if !req.Price.IsZero() {
		m["px"] = req.Price
	}

	if req.ClientOrderID != "" {
		m["clOrdId"] = req.ClientOrderID
	}

	if req.Instrument == exchange.InstrumentTypeFutures && req.PositionSide != "" {
		m["posSide"] = OkxPositionSide(req.PositionSide)
	}

	return m, nil
}

// size 精度处理
func (h *okx) sizePrecision(size decimal.Decimal, symbol exchange.Symbol, opType string) decimal.Decimal {
	orderQuantity := size
	if opType == "open" {
		// 向下取整到指定精度
		orderQuantity = orderQuantity.Truncate(symbol.SizePrecision)
	} else {
		// 四舍五入到指定精度
		orderQuantity = orderQuantity.Round(symbol.SizePrecision)
	}

	// 2. 限制最大值
	if orderQuantity.GreaterThan(symbol.MaxSize) {
		orderQuantity = symbol.MaxSize
	}

	// 3. 限制最小值
	if orderQuantity.LessThan(symbol.MinSize) {
		orderQuantity = symbol.MinSize
	}
	return orderQuantity
}
