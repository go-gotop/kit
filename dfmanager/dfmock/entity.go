package dfmock

type tradeEvent struct {
	TradeID   int64  `json:"TradeID"`
	Size      string `json:"Size"`
	Price     string `json:"Price"`
	Side      string `json:"Side"`
	Symbol    string `json:"Symbol"`
	TradeTime int64  `json:"TradeTime"`
}
