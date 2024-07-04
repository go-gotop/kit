package streammock

type wsOrderUpdateEvent struct {
	ID                string `json:"ID"`
	Exchange          string `json:"Exchange"`
	ClientOrderID     string `json:"ClientOrderID"`
	Symbol            string `json:"Symbol"`
	OrderID           string `json:"OrderID"`
	FeeAsset          string `json:"FeeAsset"`
	TransactionTime   int64  `json:"TransactionTime"`
	Instrument        string `json:"Instrument"`
	ExecutionType     string `json:"ExecutionType"`
	State             string `json:"State"`
	Status            string `json:"Status"`
	PositionSide      string `json:"PositionSide"`
	Side              string `json:"Side"`
	Type              string `json:"Type"`
	Volume            string `json:"Volume"`
	Price             string `json:"Price"`
	LatestVolume      string `json:"LatestVolume"`
	FilledVolume      string `json:"FilledVolume"`
	LatestPrice       string `json:"LatestPrice"`
	FeeCost           string `json:"FeeCost"`
	FilledQuoteVolume string `json:"FilledQuoteVolume"`
	LatestQuoteVolume string `json:"LatestQuoteVolume"`
	QuoteVolume       string `json:"QuoteVolume"`
	AvgPrice          string `json:"AvgPrice"`
}
