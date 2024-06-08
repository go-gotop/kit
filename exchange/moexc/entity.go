package moexc

type mockBalance struct {
	AssetName  string `json:"assetName"`
	Exchange   string `json:"exchange"`
	Instrument string `json:"instrument"`
	Free       string `json:"free"`
	Locked     string `json:"locked"`
}


type mockCreateOrderResponse struct {
	TransactTime int64  `json:"transactTime"`
	Symbol       string `json:"symbol"`
	ClientOrderID string `json:"clientOrderID"`
	OrderID      string `json:"orderID"`
	Side         string `json:"side"`
	State        string `json:"state"`
	PositionSide string `json:"positionSide"`
	Price        string `json:"price"`
	OriginalQuantity string `json:"originalQuantity"`
	ExecutedQuantity string `json:"executedQuantity"`
}
