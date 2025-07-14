package okexc

import (
	"strings"

	"github.com/go-gotop/kit/exchange"
)

func OkxSide(side exchange.SideType) string {
	return strings.ToLower(string(side))
}

func OkxOrderType(orderType exchange.OrderType) string {
	if orderType == exchange.OrderTypeLimitMaker {
		return "post_only"
	}
	return strings.ToLower(string(orderType))
}

func OkxPositionSide(positionSide exchange.PositionSide) string {
	return strings.ToLower(string(positionSide))
}

func OkxPosMode(posMode exchange.PosMode) string {
	return strings.ToLower(string(posMode))
}

func OkxTSide(side string) exchange.SideType {
	return exchange.SideType(strings.ToUpper(side))
}

func OkxTOrderType(orderType string) exchange.OrderType {
	return exchange.OrderType(strings.ToUpper(orderType))
}

func OkxTPositionSide(positionSide string) exchange.PositionSide {
	return exchange.PositionSide(strings.ToUpper(positionSide))
}

func OkxTPosMode(posMode string) exchange.PosMode {
	return exchange.PosMode(strings.ToUpper(posMode))
}

func OkxTMarketType(marketType string) exchange.OrderType {
	return exchange.OrderType(strings.ToUpper(marketType))
}

type OrderInfo struct {
	InstType      string `json:"instType"`
	InstID        string `json:"instId"`
	OrderID       string `json:"ordId"`
	ClientOrderID string `json:"clOrdId"`
	Px            string `json:"px"`        // 委托价格
	Sz            string `json:"sz"`        // 委托数量 币币/币币杠杆，以币为单位；交割/永续/期权 ，以张为单位
	OrderType     string `json:"ordType"`   // 订单类型
	Side          string `json:"side"`      // 订单方向
	PosSide       string `json:"posSide"`   // 持仓方向
	FillPx        string `json:"fillPx"`    // 最新成交价格
	FillSz        string `json:"fillSz"`    // 最新成交数量
	AccFillSz     string `json:"accFillSz"` // 累计成交数量
	AvgPx         string `json:"avgPx"`     // 平均成交价格
	State         string `json:"state"`     // 订单状态
	Lever         string `json:"lever"`     // 杠杆倍数
	FeeCcy        string `json:"feeCcy"`    // 手续费币种
	Fee           string `json:"fee"`       // 订单交易累计的手续费与返佣
	RebateCcy     string `json:"rebateCcy"` // 返佣币种
	Rebate        string `json:"rebate"`    // 返佣累积金额
	UpdateTime    string `json:"uTime"`     // 更新时间
	CreateTime    string `json:"cTime"`     // 创建时间
}

type CreateOrderResponse struct {
	Code string `json:"code"`
	Data []struct {
		ClOrdId string `json:"clOrdId"`
		OrdId   string `json:"ordId"`
		SCode   string `json:"sCode"`
		SMsg    string `json:"sMsg"`
		Tag     string `json:"tag"`
		Ts      string `json:"ts"`
	} `json:"data"`
	InTime  string `json:"inTime"`
	Msg     string `json:"msg"`
	OutTime string `json:"outTime"`
}

type AssetsResponse struct {
	Code string `json:"code"`
	Data []struct {
		Details []struct {
			AssetsName string `json:"ccy"`
			Free       string `json:"availBal"`
			Locked     string `json:"frozenBal"`
		} `json:"details"`
	} `json:"data"`
	Msg string `json:"msg"`
}

type AccountConfigResponse struct {
	Code string `json:"code"`
	Data []struct {
		Uid        string `json:"uid"`
		AcctLv     string `json:"acctLv"`
		PosMod     string `json:"posMode"`
		AutoBorrow bool   `json:"autoLoan"`
	} `json:"data"`
	Msg string `json:"msg"`
}

type PositionsResponse struct {
	Code string `json:"code"`
	Data []struct {
		InstType    string `json:"instType"`
		InstID      string `json:"instId"`
		Ccy         string `json:"ccy"`    // 占用保证金币种
		PosCcy      string `json:"posCcy"` // 仓位币种
		MgnMode     string `json:"mgnMode"`
		PosID       string `json:"posId"`
		PosSide     string `json:"posSide"`
		Pos         string `json:"pos"`
		AvailPos    string `json:"availPos"`
		AvgPx       string `json:"avgPx"`
		Upl         string `json:"upl"`
		Lever       string `json:"lever"`
		LiqPx       string `json:"liqPx"`
		MarkPx      string `json:"markPx"`
		Imr         string `json:"imr"`
		Margin      string `json:"margin"`
		Liab        string `json:"liab"`        // 仓位的负债
		Interest    string `json:"interest"`    // 仓位的利息
		BePx        string `json:"bePx"`        // 盈亏平衡价
		RealizedPnl string `json:"realizedPnl"` // 已实现盈亏
		Pnl         string `json:"pnl"`         // 平仓订单累积收益额
		Fee         string `json:"fee"`         // 仓位交易累计的手续费与返佣
		FundingFee  string `json:"fundingFee"`  // 累积资金费用
		CTime       string `json:"cTime"`       // 创建时间
		UTime       string `json:"uTime"`       // 更新时间
	} `json:"data"`
	Msg string `json:"msg"`
}

type LeverageResponse struct {
	Code string `json:"code"`
	Data []struct {
		Lever string `json:"lever"`
	} `json:"data"`
	Msg string `json:"msg"`
}

type LeverageInfoResponse struct {
	Code string `json:"code"`
	Data []struct {
		InstID  string `json:"instId"`
		Lever   string `json:"lever"`
		MgnMode string `json:"mgnMode"`
		PosSide string `json:"posSide"`
	} `json:"data"`
	Msg string `json:"msg"`
}

type MaxSizeResponse struct {
	Code string `json:"code"`
	Data []struct {
		InstID  string `json:"instId"`
		Ccy     string `json:"ccy"`
		MaxBuy  string `json:"maxBuy"`
		MaxSell string `json:"maxSell"`
	} `json:"data"`
	Msg string `json:"msg"`
}

type MarkPriceKlineResponse struct {
	Code string     `json:"code"`
	Data [][]string `json:"data"`
	Msg  string     `json:"msg"`
}

type KlineResponse struct {
	Code string     `json:"code"`
	Data [][]string `json:"data"`
	Msg  string     `json:"msg"`
}

type DepthResponse struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data []struct {
		Asks [][]string `json:"asks"`
		Bids [][]string `json:"bids"`
		Ts   string     `json:"ts"`
	} `json:"data"`
}

type CancelOrderResponse struct {
	Code string `json:"code"`
	Data []struct {
		ClOrdId string `json:"clOrdId"`
		OrdId   string `json:"ordId"`
		SCode   string `json:"sCode"`
		SMsg    string `json:"sMsg"`
		Ts      string `json:"ts"`
	} `json:"data"`
	Msg string `json:"msg"`
}

type TickerPriceResponse struct {
	Code string `json:"code"`
	Data []struct {
		InstID string `json:"instId"`
		Last   string `json:"last"`
	} `json:"data"`
	Msg string `json:"msg"`
}
