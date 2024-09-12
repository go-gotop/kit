package okexc

import (
	"strings"

	"github.com/go-gotop/kit/exchange"
)

func OkxSide(side exchange.SideType) string {
	return strings.ToLower(string(side))
}

func OkxOrderType(orderType exchange.OrderType) string {
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
	InstType        string `json:"instType"`
	InstID          string `json:"instId"`
	OrderID         string `json:"ordId"`
	ClientOrderID   string `json:"clOrdId"`
	Px              string `json:"px"`              // 委托价格
	Sz              string `json:"sz"`              // 委托数量 币币/币币杠杆，以币为单位；交割/永续/期权 ，以张为单位
	OrderType       string `json:"ordType"`         // 订单类型
	Side            string `json:"side"`            // 订单方向
	PosSide         string `json:"posSide"`         // 持仓方向
	FillPx          string `json:"fillPx"`          // 最新成交价格
	FillSz          string `json:"fillSz"`          // 最新成交数量
	AccFillSz       string `json:"accFillSz"`       // 累计成交数量
	AvgPx           string `json:"avgPx"`           // 平均成交价格
	State           string `json:"state"`           // 订单状态
	Lever           string `json:"lever"`           // 杠杆倍数
	FeeCcy          string `json:"feeCcy"`          // 手续费币种
	Fee             string `json:"fee"`             // 订单交易累计的手续费与返佣
	RebateCcy       string `json:"rebateCcy"`       // 返佣币种
	Rebate          string `json:"rebate"`          // 返佣累积金额
	UpdateTime      string `json:"uTime"`           // 更新时间
	CreateTime      string `json:"cTime"`           // 创建时间
}
