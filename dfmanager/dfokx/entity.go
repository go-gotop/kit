package dfokx

type okxTradeAllEvent struct {
	Arg  okxAllTradeArg    `json:"arg"`
	Data []okxTradeAllData `json:"data"`
}

type okxMarkPriceEvent struct {
	Arg  okxAllTradeArg     `json:"arg"`
	Data []okxMarkPriceData `json:"data"`
}

type okxKlineEvent struct {
	Arg  okxAllTradeArg `json:"arg"`
	Data [][]string     `json:"data"`
}

type okxMarkKlineEvent struct {
	Arg  okxAllTradeArg `json:"arg"`
	Data [][]string     `json:"data"`
}

type okxSymbolUpdateEvent struct {
	Arg  okxInstrumentArg      `json:"arg"`
	Data []okxSymbolUpdateData `json:"data"`
}

type okxAllTradeArg struct {
	Channel string `json:"channel"`
	InstID  string `json:"instId"`
}

type okxInstrumentArg struct {
	Channel  string `json:"channel"`
	InstType string `json:"instType"`
}

type okxTradeAllData struct {
	InstID    string `json:"instId"`
	TradeID   string `json:"tradeId"`
	Price     string `json:"px"`
	Quantity  string `json:"sz"`
	Side      string `json:"side"`
	TradeTime string `json:"ts"`
}

type okxMarkPriceData struct {
	InstID    string `json:"instId"`
	InstType  string `json:"instType"`
	MarkPx    string `json:"markPx"`
	Timestamp string `json:"ts"`
}

type okxSymbolUpdateData struct {
	InstType string `json:"instType"`
	InstID   string `json:"instId"`
	BaseCcy  string `json:"baseCcy"`
	QuoteCcy string `json:"quoteCcy"`
	CtVal    string `json:"ctVal"`
	CtMult   string `json:"ctMult"`
	ListTime string `json:"listTime"`
	ExpTime  string `json:"expTime"`
	Lever    string `json:"lever"`
	TickSz   string `json:"tickSz"`
	LotSz    string `json:"lotSz"`
	MinSz    string `json:"minSz"`
	State    string `json:"state"`
	MaxLmtSz string `json:"maxLmtSz"`
	MaxMktSz string `json:"maxMktSz"`
}
