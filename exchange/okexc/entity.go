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
