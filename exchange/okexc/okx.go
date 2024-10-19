package okexc

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

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
	r := &okhttp.Request{
		APIKey:     req.APIKey,
		SecretKey:  req.SecretKey,
		Passphrase: req.Passphrase,
		Method:     "GET",
		Endpoint:   "/api/v5/account/balance",
		SecType:    okhttp.SecTypeSigned,
	}

	o.client.SetApiEndpoint(okEndpoint)

	data, err := o.client.CallAPI(ctx, r)
	if err != nil {
		return nil, err
	}

	var response AssetsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("error parsing response data: %v", err)
	}

	if response.Code != "0" {
		return nil, fmt.Errorf("operation failed, code: %s, message: %s", response.Code, response.Msg)
	}

	var assets []exchange.Asset
	fmt.Printf("%+v\n", response)
	for _, item := range response.Data {
		if item.Details == nil || len(item.Details) == 0 {
			continue
		}

		for _, detail := range item.Details {
			free, err := decimal.NewFromString(detail.Free)
			if err != nil {
				fmt.Println(err)
				continue
			}
			locked, err := decimal.NewFromString(detail.Locked)
			if err != nil {
				fmt.Println(err)
				continue
			}
			asset := exchange.Asset{
				Exchange:  exchange.OkxExchange,
				AssetName: detail.AssetsName,
				Free:      free,
				Locked:    locked,
			}
			assets = append(assets, asset)
		}
	}

	return assets, nil
}

func (o *okx) GetMarkPriceKline(ctx context.Context, req *exchange.GetMarkPriceKlineRequest) ([]exchange.GetMarkPriceKlineResponse, error) {
	r := &okhttp.Request{
		Method:   "GET",
		Endpoint: "/api/v5/market/mark-price-candles",
	}

	o.client.SetApiEndpoint(okEndpoint)
	fmt.Println(req)
	params := okhttp.Params{
		"instId": req.Symbol,
		"bar":    req.Period,
	}

	if req.Start > 0 {
		params["before"] = req.Start
	}

	if req.End > 0 {
		params["after"] = req.End
	}

	r.SetParams(params)
	data, err := o.client.CallAPI(ctx, r)
	if err != nil {
		return nil, err
	}
	var response MarkPriceKlineResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("error parsing response data: %v", err)
	}

	if response.Code != "0" {
		return nil, fmt.Errorf("operation failed, code: %s, message: %s", response.Code, response.Msg)
	}

	var klines []exchange.GetMarkPriceKlineResponse
	for i := len(response.Data) - 1; i >= 0; i-- {
		item := response.Data[i]
		openTime, err := strconv.ParseInt(item[0], 10, 64)
		if err != nil {
			return nil, err
		}
		open, err := decimal.NewFromString(item[1])
		if err != nil {
			return nil, err
		}
		high, err := decimal.NewFromString(item[2])
		if err != nil {
			return nil, err
		}
		low, err := decimal.NewFromString(item[3])
		if err != nil {
			return nil, err
		}
		close, err := decimal.NewFromString(item[4])
		if err != nil {
			return nil, err
		}
		kline := exchange.GetMarkPriceKlineResponse{
			OpenTime: openTime,
			Open:     open,
			High:     high,
			Low:      low,
			Close:    close,
			Confirm:  item[5],
		}
		klines = append(klines, kline)
	}
	return klines, nil
}

func (o *okx) GetAccountConfig(ctx context.Context, req *exchange.GetAccountConfigRequest) (exchange.GetAccountConfigResponse, error) {
	r := &okhttp.Request{
		APIKey:     req.APIKey,
		SecretKey:  req.SecretKey,
		Passphrase: req.Passphrase,
		Method:     "GET",
		Endpoint:   "/api/v5/account/config",
		SecType:    okhttp.SecTypeSigned,
	}

	o.client.SetApiEndpoint(okEndpoint)

	data, err := o.client.CallAPI(ctx, r)
	if err != nil {
		return exchange.GetAccountConfigResponse{}, err
	}

	var response AccountConfigResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return exchange.GetAccountConfigResponse{}, fmt.Errorf("error parsing response data: %v", err)
	}

	if response.Code != "0" {
		return exchange.GetAccountConfigResponse{}, fmt.Errorf("operation failed, code: %s, message: %s", response.Code, response.Msg)
	}

	if response.Data == nil || len(response.Data) == 0 {
		return exchange.GetAccountConfigResponse{}, nil
	}

	return exchange.GetAccountConfigResponse{
		Uid:        response.Data[0].Uid,
		AcctLv:     response.Data[0].AcctLv,
		PosMod:     response.Data[0].PosMod,
		AutoBorrow: response.Data[0].AutoBorrow,
	}, nil
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

	var responseData CreateOrderResponse
	if err := json.Unmarshal(data, &responseData); err != nil {
		return fmt.Errorf("error parsing response data: %v", err)
	}

	fmt.Println("创建订单结果:", responseData)

	// 检查 code 值
	if responseData.Code != "0" || len(responseData.Data) == 0 || responseData.Data[0].SCode != "0" {
		// 处理错误或特定条件
		msg := responseData.Msg
		code := responseData.Code
		if len(responseData.Data) > 0 {
			msg = responseData.Data[0].SMsg
			code = responseData.Data[0].SCode
		}
		return fmt.Errorf("operation failed, code: %s, message: %s", code, msg)
	}
	return nil
}

func (o *okx) CancelOrder(ctx context.Context, req *exchange.CancelOrderRequest) error {
	return nil
}

func (o *okx) SearchOrder(ctx context.Context, req *exchange.SearchOrderRequest) (*exchange.SearchOrderResponse, error) {
	r := &okhttp.Request{
		APIKey:     req.APIKey,
		SecretKey:  req.SecretKey,
		Passphrase: req.Passphrase,
		Method:     "GET",
		Endpoint:   "/api/v5/trade/order",
		SecType:    okhttp.SecTypeSigned,
	}

	o.client.SetApiEndpoint(okEndpoint)

	params := okhttp.Params{
		"instId":  req.Symbol.OriginalSymbol,
		"clOrdId": req.ClientOrderID,
	}

	// var err error

	r.SetParams(params)
	data, err := o.client.CallAPI(ctx, r)
	if err != nil {
		return nil, err
	}

	type result struct {
		Code string      `json:"code"`
		Data []OrderInfo `json:"data"`
		Msg  string      `json:"msg"`
	}

	orderInfoRes := result{}

	err = okhttp.Json.Unmarshal(data, &orderInfoRes)
	if err != nil {
		return nil, err
	}

	if orderInfoRes.Code != "0" {
		return nil, fmt.Errorf("error: %v", orderInfoRes.Msg)
	}

	if len(orderInfoRes.Data) == 0 {
		return nil, fmt.Errorf("order not found")
	}

	orderInfo := orderInfoRes.Data[0]

	state := exchange.OrderStateNew
	switch orderInfo.State {
	case "partially_filled":
		state = exchange.OrderStatePartiallyFilled
	case "filled":
		state = exchange.OrderStateFilled
	case "canceled":
		state = exchange.OrderStateCanceled
	case "rejected":
		state = exchange.OrderStateRejected
	case "expired":
		state = exchange.OrderStateExpired
	}

	avgPrice, err := decimal.NewFromString(orderInfo.AvgPx)
	if err != nil {
		return nil, err
	}

	if orderInfo.InstType == "FUTURES" || orderInfo.InstType == "SWAP" {
		// 合约类型要将张转位币
		orderInfo.AccFillSz, err = o.ConvertContractCoin("2", req.Symbol, orderInfo.AccFillSz, "close")
		if err != nil {
			return nil, err
		}
	}
	filledVolume, err := decimal.NewFromString(orderInfo.AccFillSz)
	if err != nil {
		return nil, err
	}

	px, err := decimal.NewFromString(orderInfo.Px)
	if err != nil {
		px = decimal.Zero
	}

	fee, err := decimal.NewFromString(orderInfo.Fee)
	if err != nil {
		fee = decimal.Zero
	}

	updateTime, err := strconv.ParseInt(orderInfo.UpdateTime, 10, 64)
	if err != nil {
		return nil, err
	}

	createdTime, err := strconv.ParseInt(orderInfo.CreateTime, 10, 64)
	if err != nil {
		return nil, err
	}

	return &exchange.SearchOrderResponse{
		OrderID:           orderInfo.OrderID,
		ClientOrderID:     orderInfo.ClientOrderID,
		State:             state,
		Symbol:            orderInfo.InstID,
		AvgPrice:          avgPrice,
		Volume:            filledVolume,
		Price:             px,
		FilledQuoteVolume: filledVolume.Mul(avgPrice),
		FilledVolume:      filledVolume,
		FeeCost:           fee,
		FeeAsset:          orderInfo.FeeCcy,
		Side:              exchange.SideType(OkxTSide(orderInfo.Side)),
		PositionSide:      exchange.PositionSide(OkxTPositionSide(orderInfo.PosSide)),
		OrderType:         exchange.OrderType(OkxTOrderType(orderInfo.OrderType)),
		CreatedTime:       createdTime,
		UpdateTime:        updateTime,
	}, nil
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

func (o *okx) GetPosition(ctx context.Context, req *exchange.GetPositionRequest) ([]*exchange.GetPositionResponse, error) {
	r := &okhttp.Request{
		APIKey:     req.APIKey,
		SecretKey:  req.SecretKey,
		Passphrase: req.Passphrase,
		Method:     "GET",
		Endpoint:   "/api/v5/account/positions",
		SecType:    okhttp.SecTypeSigned,
	}

	o.client.SetApiEndpoint(okEndpoint)
	data, err := o.client.CallAPI(ctx, r)
	if err != nil {
		return nil, err
	}

	response := PositionsResponse{}
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("error parsing response data: %v", err)
	}

	if response.Code != "0" {
		return nil, fmt.Errorf("operation failed, code: %s, message: %s", response.Code, response.Msg)
	}

	var positions []*exchange.GetPositionResponse
	for _, item := range response.Data {

		avgPx, err := decimal.NewFromString(item.AvgPx)
		if err != nil {
			return nil, err
		}

		fee, err := decimal.NewFromString(item.Fee)
		if err != nil {
			fee = decimal.Zero
		}

		fundingFee, err := decimal.NewFromString(item.FundingFee)
		if err != nil {
			fundingFee = decimal.Zero
		}

		pos, err := decimal.NewFromString(item.Pos)
		if err != nil {
			pos = decimal.Zero
		}

		upl, err := decimal.NewFromString(item.Upl)
		if err != nil {
			upl = decimal.Zero
		}

		pnl, err := decimal.NewFromString(item.Pnl)
		if err != nil {
			pnl = decimal.Zero
		}

		realizedPnl, err := decimal.NewFromString(item.RealizedPnl)
		if err != nil {
			realizedPnl = decimal.Zero
		}

		liqPx, err := decimal.NewFromString(item.LiqPx)
		if err != nil {
			liqPx = decimal.Zero
		}

		margin, err := decimal.NewFromString(item.Margin)
		if err != nil {
			margin = decimal.Zero
		}

		interest, err := decimal.NewFromString(item.Interest)
		if err != nil {
			interest = decimal.Zero
		}

		liab, err := decimal.NewFromString(item.Liab)
		if err != nil {
			liab = decimal.Zero
		}

		bePx, err := decimal.NewFromString(item.BePx)
		if err != nil {
			bePx = decimal.Zero
		}

		ctime, err := strconv.ParseInt(item.CTime, 10, 64)
		if err != nil {
			return nil, err
		}

		utime, err := strconv.ParseInt(item.UTime, 10, 64)
		if err != nil {
			return nil, err
		}

		positionSide, err := o.posSideToPositionSide(item.InstType, item.PosSide, item.Pos, item.PosCcy, item.Ccy)
		if err != nil {
			return nil, err
		}

		instrumentType := exchange.InstrumentTypeFutures
		if item.InstType == "MARGIN" {
			instrumentType = exchange.InstrumentTypeMargin
		}

		position := exchange.GetPositionResponse{
			Symbol:         item.InstID,
			InstrumentType: instrumentType,
			AvgPrice:       avgPx,
			Fee:            fee,
			FundingFee:     fundingFee,
			Size:           pos,
			Upl:            upl,
			Pnl:            pnl,
			RealizedPnl:    realizedPnl,
			Lever:          item.Lever,
			LiqPx:          liqPx,
			Margin:         margin,
			Liab:           liab,
			Interest:       interest,
			BePx:           bePx,
			PositionSide:   positionSide,
			CreateTime:     ctime,
			UpdateTime:     utime,
		}
		positions = append(positions, &position)
	}

	return positions, nil
}

func (o *okx) GetHistoryPosition(ctx context.Context, req *exchange.GetPositionHistoryRequest) error {
	r := &okhttp.Request{
		APIKey:     req.APIKey,
		SecretKey:  req.SecretKey,
		Passphrase: req.Passphrase,
		Method:     "GET",
		Endpoint:   "/api/v5/account/positions-history",
		SecType:    okhttp.SecTypeSigned,
	}
	r.SetParam("before", "1725942111000")
	o.client.SetApiEndpoint(okEndpoint)
	data, err := o.client.CallAPI(ctx, r)
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func (o *okx) SetLeverage(ctx context.Context, req *exchange.SetLeverageRequest) error {
	r := &okhttp.Request{
		APIKey:     req.APIKey,
		SecretKey:  req.SecretKey,
		Passphrase: req.Passphrase,
		Method:     "POST",
		Endpoint:   "/api/v5/account/set-leverage",
		SecType:    okhttp.SecTypeSigned,
	}
	params := okhttp.Params{
		"instId":  req.Symbol,
		"lever":   req.Lever,
		"mgnMode": req.Mode,
	}
	r.SetJSONBody(params)
	o.client.SetApiEndpoint(okEndpoint)
	data, err := o.client.CallAPI(ctx, r)
	if err != nil {
		return err
	}

	var response LeverageResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return fmt.Errorf("error parsing response data: %v", err)
	}

	if response.Code != "0" {
		return fmt.Errorf("operation failed, code: %s, message: %s", response.Code, response.Msg)
	}

	if response.Data == nil || len(response.Data) == 0 {
		return fmt.Errorf("operation failed, code: %s, message: %s", response.Code, response.Msg)
	}

	if response.Data[0].Lever != req.Lever {
		return fmt.Errorf("operation failed, code: %s, message: %s", response.Code, response.Msg)
	}
	return nil
}

func (o *okx) GetMaxSize(ctx context.Context, req *exchange.GetMaxSizeRequest) ([]exchange.GetMaxSizeResponse, error) {
	r := &okhttp.Request{
		APIKey:     req.APIKey,
		SecretKey:  req.SecretKey,
		Passphrase: req.Passphrase,
		Method:     "GET",
		Endpoint:   "/api/v5/account/max-size",
		SecType:    okhttp.SecTypeSigned,
	}

	o.client.SetApiEndpoint(okEndpoint)
	httpParams := okhttp.Params{
		"instId":   req.InstIds,
		"ccy":      req.Ccy,
		"tdMode":   req.TdMode,
		"leverage": req.Leverage,
	}
	r.SetParams(httpParams)
	data, err := o.client.CallAPI(ctx, r)
	if err != nil {
		return nil, err
	}

	var response MaxSizeResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("error parsing response data: %v", err)
	}

	if response.Code != "0" {
		return nil, fmt.Errorf("operation failed, code: %s, message: %s", response.Code, response.Msg)
	}

	if response.Data == nil || len(response.Data) == 0 {
		return nil, nil
	}

	var maxSize []exchange.GetMaxSizeResponse
	for _, item := range response.Data {
		maxBuy, err := decimal.NewFromString(item.MaxBuy)
		if err != nil {
			return nil, err
		}

		maxSell, err := decimal.NewFromString(item.MaxSell)
		if err != nil {
			return nil, err
		}
		maxSize = append(maxSize, exchange.GetMaxSizeResponse{
			InstId:  item.InstID,
			Ccy:     item.Ccy,
			MaxBuy:  maxBuy,
			MaxSell: maxSell,
		})
	}

	return maxSize, nil
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
		"instId":  req.Symbol.OriginalSymbol,
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

func (h *okx) posSideToPositionSide(instType string, posSide string, pos string, posCcy string, ccy string) (exchange.PositionSide, error) {
	// 持仓方向
	// long：开平仓模式开多，pos为正
	// short：开平仓模式开空，pos为正
	// net：买卖模式（交割/永续/期权：pos为正代表开多，pos为负代表开空。币币杠杆时，pos均为正，posCcy为交易货币时，代表开多；posCcy为计价货币时，代表开空。）
	size, err := decimal.NewFromString(pos)
	if err != nil {
		return exchange.PositionSide(""), err
	}
	if posSide == "long" {
		// 开多
		return exchange.PositionSideLong, nil
	} else if posSide == "short" {
		// 开空
		return exchange.PositionSideShort, nil
	} else if posSide == "net" {
		if instType == "MARGIN" {
			if posCcy == ccy {
				// 保证金默认是计价货币 usdt，如果仓位币种是 usdt，代表开空
				return exchange.PositionSideShort, nil
			} else {
				return exchange.PositionSideLong, nil
			}
		} else {
			if size.GreaterThan(decimal.Zero) {
				return exchange.PositionSideLong, nil
			} else {
				return exchange.PositionSideShort, nil
			}
		}
	}
	return exchange.PositionSide(""), fmt.Errorf("invalid posSide: %v", posSide)
}
