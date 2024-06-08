package dfmock

type tradeEvent struct {
	TradeID   int64  `json:"trade_id"`
	Size      string `json:"size"`
	Price     string `json:"price"`
	Side      string `json:"side"`
	Symbol    string `json:"symbol"`
	TradeTime int64  `json:"traded_at"`
}
