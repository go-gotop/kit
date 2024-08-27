package dfokx

type okxTradeAllEvent struct {
	Arg  okxAllTradeArg    `json:"arg"`
	Data []okxTradeAllData `json:"data"`
}

type okxAllTradeArg struct {
	Channel string `json:"channel"`
	InstID  string `json:"instId"`
}

type okxTradeAllData struct {
	InstID    string `json:"instId"`
	TradeID   string  `json:"tradeId"`
	Price     string `json:"px"`
	Quantity  string `json:"sz"`
	Side      string `json:"side"`
	TradeTime string  `json:"ts"`
}
