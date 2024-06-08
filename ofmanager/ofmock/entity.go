package ofmock

type wsOrderUpdateEvent struct {
	Id                string `json:"id"`
	Exchange          string `json:"exchange"`
	ClientOrderID     string `json:"client_order_id"`
	Symbol            string `json:"symbol"`
	OrderID           string `json:"order_id"`
	FeeAsset          string `json:"fee_asset"`
	TransactionTime   int64  `json:"transaction_time"`
	Instrument        string `json:"instrument"`
	ExecutionType     string `json:"execution_type"`
	State             string `json:"state"`
	Status            string `json:"status"`
	PositionSide      string `json:"position_side"`
	Side              string `json:"side"`
	Type              string `json:"type"`
	Volume            string `json:"volume"`
	Price             string `json:"price"`
	LatestVolume      string `json:"latest_volume"`
	FilledVolume      string `json:"filled_volume"`
	LatestPrice       string `json:"latest_price"`
	FeeCost           string `json:"fee_cost"`
	FilledQuoteVolume string `json:"filled_quote_volume"`
	LatestQuoteVolume string `json:"latest_quote_volume"`
	QuoteVolume       string `json:"quote_volume"`
	AvgPrice          string `json:"avg_price"`
}
