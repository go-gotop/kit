package okexc

import (
	"strings"

	"github.com/go-gotop/kit/exchange"
)

func okxSide(side exchange.SideType) string {
	return strings.ToLower(string(side))
}

func okxOrderType(orderType exchange.OrderType) string {
	return strings.ToLower(string(orderType))
}

func okxPositionSide(positionSide exchange.PositionSide) string {
	return strings.ToLower(string(positionSide))
}

func okxPosMode(posMode exchange.PosMode) string {
	return strings.ToLower(string(posMode))
}
