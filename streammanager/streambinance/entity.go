package streambinance

type bnSpotWsOrderUpdateEvent struct {
	Symbol                  string `json:"s"`
	ClientOrderId           string `json:"c"`
	Side                    string `json:"S"`
	Type                    string `json:"o"`
	TimeInForce             string `json:"f"`
	Volume                  string `json:"q"`
	Price                   string `json:"p"`
	StopPrice               string `json:"P"`
	TrailingDelta           int64  `json:"d"` // Trailing Delta
	IceBergVolume           string `json:"F"`
	OrderListId             int64  `json:"g"` // for OCO
	OrigCustomOrderId       string `json:"C"` // customized order ID for the original order
	ExecutionType           string `json:"x"` // execution type for this event NEW/TRADE...
	Status                  string `json:"X"` // order status
	RejectReason            string `json:"r"`
	Id                      int64  `json:"i"` // order id
	LatestVolume            string `json:"l"` // quantity for the latest trade
	FilledVolume            string `json:"z"`
	LatestPrice             string `json:"L"` // price for the latest trade
	FeeAsset                string `json:"N"`
	FeeCost                 string `json:"n"`
	TransactionTime         int64  `json:"T"`
	TradeId                 int64  `json:"t"`
	IsInOrderBook           bool   `json:"w"` // is the order in the order book?
	IsMaker                 bool   `json:"m"` // is this order maker?
	CreateTime              int64  `json:"O"`
	FilledQuoteVolume       string `json:"Z"` // the quote volume that already filled
	LatestQuoteVolume       string `json:"Y"` // the quote volume for the latest trade
	QuoteVolume             string `json:"Q"`
	TrailingTime            int64  `json:"D"` // Trailing Time
	StrategyId              int64  `json:"j"` // Strategy ID
	StrategyType            int64  `json:"J"` // Strategy Type
	WorkingTime             int64  `json:"W"` // Working Time
	SelfTradePreventionMode string `json:"V"`
}

type bnFuturesWsUserDataEvent struct {
	Event            string                      `json:"e"`
	Time             int64                       `json:"E"`
	TransactionTime  int64                       `json:"T"`
	OrderTradeUpdate bnFuturesWsOrderUpdateEvent `json:"o"`
}

// bnFuturesWsOrderUpdateEvent define order trade update
type bnFuturesWsOrderUpdateEvent struct {
	Symbol               string `json:"s"`
	ClientOrderID        string `json:"c"`
	Side                 string `json:"S"`
	Type                 string `json:"o"`
	TimeInForce          string `json:"f"`
	OriginalQty          string `json:"q"`
	OriginalPrice        string `json:"p"`
	AveragePrice         string `json:"ap"`
	StopPrice            string `json:"sp"`
	ExecutionType        string `json:"x"`
	Status               string `json:"X"`
	ID                   int64  `json:"i"`
	LastFilledQty        string `json:"l"`
	AccumulatedFilledQty string `json:"z"`
	LastFilledPrice      string `json:"L"`
	CommissionAsset      string `json:"N"`
	Commission           string `json:"n"`
	TradeTime            int64  `json:"T"`
	TradeID              int64  `json:"t"`
	BidsNotional         string `json:"b"`
	AsksNotional         string `json:"a"`
	IsMaker              bool   `json:"m"`
	IsReduceOnly         bool   `json:"R"`
	WorkingType          string `json:"wt"`
	OriginalType         string `json:"ot"`
	PositionSide         string `json:"ps"`
	IsClosingPosition    bool   `json:"cp"`
	ActivationPrice      string `json:"AP"`
	CallbackRate         string `json:"cr"`
	RealizedPnL          string `json:"rp"`
}

// 现货账户更新事件
type bnSpotWsAccountUpdateEvent struct {
	Event          string            `json:"e"`
	Time           int64             `json:"E"`
	LastUpdateTime int64             `json:"u"`
	Balances       []bnSpotWsBalance `json:"B"`
}

type bnSpotWsBalance struct {
	Asset  string `json:"a"` // 资产名称
	Free   string `json:"f"` // 可用余额
	Locked string `json:"l"` // 冻结余额
}

// 合约账户更新事件
type bnFuturesWsAccountUpdateEvent struct {
	Event           string                 `json:"e"`
	Time            int64                  `json:"E"`
	TransactionTime int64                  `json:"T"` // 撮合时间
	EventDetail     bnFuturesWsEventDetail `json:"a"`
}

type bnFuturesWsEventDetail struct {
	Reason   string                `json:"m"` // 事件类型
	Balances []bnFunturesWsBalance `json:"B"` // 账户余额

}

type bnFunturesWsBalance struct {
	Asset              string `json:"a"`  // 资产名称
	Balance            string `json:"wb"` // 账户余额
	CrossWalletBalance string `json:"cw"` // 全仓余额(除去逐仓仓位保证金的钱包余额)
	BalanceChange      string `json:"bc"` // 余额变化（除去盈亏与交易手续费以外的钱包余额改变量）
}
