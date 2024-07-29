package dfmock

type tradeEvent struct {
	TradeID   int64  `json:"TradeID"`
	Size      string `json:"Size"`
	Price     string `json:"Price"`
	Side      string `json:"Side"`
	Symbol    string `json:"Symbol"`
	TradeTime int64  `json:"TradedAt"`
}

type funtureMarkPriceEvent struct {
	Time                 int64  `json:"time"`
	Symbol               string `json:"symbol"`
	MarkPrice            string `json:"markPrice"`
	IndexPrice           string `json:"indexPrice"`           // 现货指数价格
	EstimatedSettlePrice string `json:"estimatedSettlePrice"` // 预估结算价
	LastFundingRate      string `json:"lastFundingRate"`      // 资金费率
	NextFundingTime      int64  `json:"nextFundingTime"`      // 下个资金时间
	IsSettlement         bool   `json:"isSettlement"`         // 是否结算
}
