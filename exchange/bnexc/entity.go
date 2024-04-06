package bnexc

type bnSpotAccount struct {
	MakerCommission  int64                 `json:"makerCommission"`
	TakerCommission  int64                 `json:"takerCommission"`
	BuyerCommission  int64                 `json:"buyerCommission"`
	SellerCommission int64                 `json:"sellerCommission"`
	CommissionRates  bnSpotCommissionRates `json:"commissionRates"`
	CanTrade         bool                  `json:"canTrade"`
	CanWithdraw      bool                  `json:"canWithdraw"`
	CanDeposit       bool                  `json:"canDeposit"`
	UpdateTime       uint64                `json:"updateTime"`
	AccountType      string                `json:"accountType"`
	Balances         []bnSpotBalance       `json:"balances"`
	Permissions      []string              `json:"permissions"`
}

type bnFuturesBalance struct {
	AccountAlias       string `json:"accountAlias"`
	Asset              string `json:"asset"`
	Balance            string `json:"balance"`
	CrossWalletBalance string `json:"crossWalletBalance"`
	CrossUnPnl         string `json:"crossUnPnl"`
	AvailableBalance   string `json:"availableBalance"`
	MaxWithdrawAmount  string `json:"maxWithdrawAmount"`
}

type bnSpotBalance struct {
	Asset  string `json:"asset"`
	Free   string `json:"free"`
	Locked string `json:"locked"`
}

type bnSpotCommissionRates struct {
	Maker  string `json:"maker"`
	Taker  string `json:"taker"`
	Buyer  string `json:"buyer"`
	Seller string `json:"seller"`
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

type binanceSpotTradeEvent struct {
	Event         string `json:"e"`
	Time          int64  `json:"E"`
	Symbol        string `json:"s"`
	TradeID       int64  `json:"t"`
	Price         string `json:"p"`
	Quantity      string `json:"q"`
	BuyerOrderID  int64  `json:"b"`
	SellerOrderID int64  `json:"a"`
	TradeTime     int64  `json:"T"`
	IsBuyerMaker  bool   `json:"m"`
	Placeholder   bool   `json:"M"`
}

type binanceFuturesTradeEvent struct {
	Event            string `json:"e"`
	Time             int64  `json:"E"`
	Symbol           string `json:"s"`
	AggregateTradeID int64  `json:"a"`
	Price            string `json:"p"`
	Quantity         string `json:"q"`
	FirstTradeID     int64  `json:"f"`
	LastTradeID      int64  `json:"l"`
	TradeTime        int64  `json:"T"`
	Maker            bool   `json:"m"`
}

type bnFuturesExchangeInfo struct {
	Timezone        string               `json:"timezone"`
	ServerTime      int64                `json:"serverTime"`
	RateLimits      []bnFuturesRateLimit `json:"rateLimits"`
	ExchangeFilters []interface{}        `json:"exchangeFilters"`
	Symbols         []bnFuturesSymbol    `json:"symbols"`
}

type bnFuturesRateLimit struct {
	RateLimitType string `json:"rateLimitType"`
	Interval      string `json:"interval"`
	IntervalNum   int64  `json:"intervalNum"`
	Limit         int64  `json:"limit"`
}

type bnFuturesSymbol struct {
	Symbol                string                   `json:"symbol"`
	Pair                  string                   `json:"pair"`
	ContractType          string                   `json:"contractType"`
	DeliveryDate          int64                    `json:"deliveryDate"`
	OnboardDate           int64                    `json:"onboardDate"`
	Status                string                   `json:"status"`
	MaintMarginPercent    string                   `json:"maintMarginPercent"`
	RequiredMarginPercent string                   `json:"requiredMarginPercent"`
	PricePrecision        int                      `json:"pricePrecision"`
	QuantityPrecision     int                      `json:"quantityPrecision"`
	BaseAssetPrecision    int                      `json:"baseAssetPrecision"`
	QuotePrecision        int                      `json:"quotePrecision"`
	UnderlyingType        string                   `json:"underlyingType"`
	UnderlyingSubType     []string                 `json:"underlyingSubType"`
	SettlePlan            int64                    `json:"settlePlan"`
	TriggerProtect        string                   `json:"triggerProtect"`
	OrderType             []string                 `json:"orderType"`
	TimeInForce           []string                 `json:"timeInForce"`
	Filters               []map[string]interface{} `json:"filters"`
	QuoteAsset            string                   `json:"quoteAsset"`
	MarginAsset           string                   `json:"marginAsset"`
	BaseAsset             string                   `json:"baseAsset"`
	LiquidationFee        string                   `json:"liquidationFee"`
	MarketTakeBound       string                   `json:"marketTakeBound"`
}

func (s *bnFuturesSymbol) LotSizeFilter() *bnLotSizeFilter {
	for _, filter := range s.Filters {
		if filter["filterType"].(string) == "LOT_SIZE" {
			f := &bnLotSizeFilter{}
			if i, ok := filter["maxQty"]; ok {
				f.MaxQuantity = i.(string)
			}
			if i, ok := filter["minQty"]; ok {
				f.MinQuantity = i.(string)
			}
			if i, ok := filter["stepSize"]; ok {
				f.StepSize = i.(string)
			}
			return f
		}
	}
	return nil
}

func (s *bnFuturesSymbol) PriceFilter() *bnPriceFilter {
	for _, filter := range s.Filters {
		if filter["filterType"].(string) == "PRICE_FILTER" {
			f := &bnPriceFilter{}
			if i, ok := filter["maxPrice"]; ok {
				f.MaxPrice = i.(string)
			}
			if i, ok := filter["minPrice"]; ok {
				f.MinPrice = i.(string)
			}
			if i, ok := filter["tickSize"]; ok {
				f.TickSize = i.(string)
			}
			return f
		}
	}
	return nil
}

type bnSpotExchangeInfo struct {
	Timezone        string            `json:"timezone"`
	ServerTime      int64             `json:"serverTime"`
	RateLimits      []bnSpotRateLimit `json:"rateLimits"`
	ExchangeFilters []interface{}     `json:"exchangeFilters"`
	Symbols         []bnSpotSymbol    `json:"symbols"`
}

// bnLotSizeFilter define lot size filter of symbol
type bnLotSizeFilter struct {
	MaxQuantity string `json:"maxQty"`
	MinQuantity string `json:"minQty"`
	StepSize    string `json:"stepSize"`
}

// bnPriceFilter define price filter of symbol
type bnPriceFilter struct {
	MaxPrice string `json:"maxPrice"`
	MinPrice string `json:"minPrice"`
	TickSize string `json:"tickSize"`
}

type bnSpotRateLimit struct {
	RateLimitType string `json:"rateLimitType"`
	Interval      string `json:"interval"`
	IntervalNum   int64  `json:"intervalNum"`
	Limit         int64  `json:"limit"`
}

type bnSpotSymbol struct {
	Symbol                     string                   `json:"symbol"`
	Status                     string                   `json:"status"`
	BaseAsset                  string                   `json:"baseAsset"`
	BaseAssetPrecision         int                      `json:"baseAssetPrecision"`
	QuoteAsset                 string                   `json:"quoteAsset"`
	QuotePrecision             int                      `json:"quotePrecision"`
	QuoteAssetPrecision        int                      `json:"quoteAssetPrecision"`
	BaseCommissionPrecision    int32                    `json:"baseCommissionPrecision"`
	QuoteCommissionPrecision   int32                    `json:"quoteCommissionPrecision"`
	Orderevent                 []string                 `json:"orderevent"`
	IcebergAllowed             bool                     `json:"icebergAllowed"`
	OcoAllowed                 bool                     `json:"ocoAllowed"`
	QuoteOrderQtyMarketAllowed bool                     `json:"quoteOrderQtyMarketAllowed"`
	IsSpotTradingAllowed       bool                     `json:"isSpotTradingAllowed"`
	IsMarginTradingAllowed     bool                     `json:"isMarginTradingAllowed"`
	Filters                    []map[string]interface{} `json:"filters"`
	Permissions                []string                 `json:"permissions"`
}

func (s *bnSpotSymbol) LotSizeFilter() *bnLotSizeFilter {
	for _, filter := range s.Filters {
		if filter["filterType"].(string) == "LOT_SIZE" {
			f := &bnLotSizeFilter{}
			if i, ok := filter["maxQty"]; ok {
				f.MaxQuantity = i.(string)
			}
			if i, ok := filter["minQty"]; ok {
				f.MinQuantity = i.(string)
			}
			if i, ok := filter["stepSize"]; ok {
				f.StepSize = i.(string)
			}
			return f
		}
	}
	return nil
}

func (s *bnSpotSymbol) PriceFilter() *bnPriceFilter {
	for _, filter := range s.Filters {
		if filter["filterType"].(string) == "PRICE_FILTER" {
			f := &bnPriceFilter{}
			if i, ok := filter["maxPrice"]; ok {
				f.MaxPrice = i.(string)
			}
			if i, ok := filter["minPrice"]; ok {
				f.MinPrice = i.(string)
			}
			if i, ok := filter["tickSize"]; ok {
				f.TickSize = i.(string)
			}
			return f
		}
	}
	return nil
}

type bnFuturesOrderResponse struct {
	Symbol            string `json:"symbol"`
	OrderID           int64  `json:"orderId"`
	ClientOrderID     string `json:"clientOrderId"`
	Price             string `json:"price"`
	OrigQuantity      string `json:"origQty"`
	ExecutedQuantity  string `json:"executedQty"`
	CumQuote          string `json:"cumQuote"`
	ReduceOnly        bool   `json:"reduceOnly"`
	Status            string `json:"status"`
	StopPrice         string `json:"stopPrice"`
	TimeInForce       string `json:"timeInForce"`
	Type              string `json:"type"`
	Side              string `json:"side"`
	UpdateTime        int64  `json:"updateTime"`
	WorkingType       string `json:"workingType"`
	ActivatePrice     string `json:"activatePrice"`
	PriceRate         string `json:"priceRate"`
	AvgPrice          string `json:"avgPrice"`
	PositionSide      string `json:"positionSide"`
	ClosePosition     bool   `json:"closePosition"`
	PriceProtect      bool   `json:"priceProtect"`
	RateLimitOrder10s string `json:"rateLimitOrder10s,omitempty"`
	RateLimitOrder1m  string `json:"rateLimitOrder1m,omitempty"`
}

type bnSpotCreateOrderResponse struct {
	Symbol                   string `json:"symbol"`
	OrderID                  int64  `json:"orderId"`
	ClientOrderID            string `json:"clientOrderId"`
	TransactTime             int64  `json:"transactTime"`
	Price                    string `json:"price"`
	OrigQuantity             string `json:"origQty"`
	ExecutedQuantity         string `json:"executedQty"`
	CummulativeQuoteQuantity string `json:"cummulativeQuoteQty"`
	IsIsolated               bool   `json:"isIsolated"` // for isolated margin
	Status                   string `json:"status"`
	TimeInForce              string `json:"timeInForce"`
	Type                     string `json:"type"`
	Side                     string `json:"side"`
	// for order response is set to FULL
	Fills                 []*bnSpotFill `json:"fills"`
	MarginBuyBorrowAmount string        `json:"marginBuyBorrowAmount"` // for margin
	MarginBuyBorrowAsset  string        `json:"marginBuyBorrowAsset"`
}

type bnSpotFill struct {
	TradeID         int64  `json:"tradeId"`
	Price           string `json:"price"`
	Quantity        string `json:"qty"`
	Commission      string `json:"commission"`
	CommissionAsset string `json:"commissionAsset"`
}
