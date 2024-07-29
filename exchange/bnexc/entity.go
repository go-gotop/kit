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

type bnMarginCreateOrderResponse struct {
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

type bnSpotSearchOrderReponse struct {
	ClientOrderID     string `json:"clientOrderId"`
	OrderID           int64  `json:"orderId"`
	Status            string `json:"status"`
	Symbol            string `json:"symbol"`
	Volume            string `json:"origQty"` // 原始交易数量
	Price             string `json:"price"`
	OrigQuoteOrderQty string `json:"origQuoteOrderQty"` //原始交易金额
	FilledQuoteVolume string `json:"cummulativeQuoteQty"`
	FilledVolume      string `json:"executedQty"`
	Side              string `json:"side"`
	TimeInForce       string `json:"timeInForce"`
	OrderType         string `json:"type"`
	CreatedTime       int64  `json:"time"`
	UpdateTime        int64  `json:"updateTime"`
}

type bnFuturesSearchOrderResponse struct {
	ClientOrderID     string `json:"clientOrderId"`
	OrderID           int64  `json:"orderId"`
	Status            string `json:"status"`
	Symbol            string `json:"symbol"`
	AvgPrice          string `json:"avgPrice"`
	Volume            string `json:"origQty"`
	Price             string `json:"price"`
	FilledQuoteVolume string `json:"cumQuote"`
	FilledVolume      string `json:"executedQty"`
	Side              string `json:"side"`
	PositionSide      string `json:"positionSide"`
	TimeInForce       string `json:"timeInForce"`
	OrderType         string `json:"type"`
	CreatedTime       int64  `json:"time"`
	UpdateTime        int64  `json:"updateTime"`
}

type bnSpotTrades struct {
	Symbol          string `json:"symbol"`
	ID              int64  `json:"id"`
	OrderID         int64  `json:"orderId"`
	Price           string `json:"price"`
	Quantity        string `json:"qty"`
	Commission      string `json:"commission"`
	CommissionAsset string `json:"commissionAsset"`
	Time            int64  `json:"time"`
	IsBuyer         bool   `json:"isBuyer"`
	IsMaker         bool   `json:"isMaker"`
}

type bnFuturesTrades struct {
	Symbol          string `json:"symbol"`
	ID              int64  `json:"id"`
	OrderID         int64  `json:"orderId"`
	Price           string `json:"price"`
	Quantity        string `json:"qty"`
	Commission      string `json:"commission"`
	CommissionAsset string `json:"commissionAsset"`
	Time            int64  `json:"time"`
	IsBuyer         bool   `json:"buyer"`
	IsMaker         bool   `json:"maker"`
}

type bnPremiumIndex struct {
	Symbol               string `json:"symbol"`
	MarkPrice            string `json:"markPrice"`
	IndexPrice           string `json:"indexPrice"`
	EstimatedSettlePrice string `json:"estimatedSettlePrice"`
	LastFundingRate      string `json:"lastFundingRate"`
	InterestRate         string `json:"interestRate"`
	NextFundingTime      int64  `json:"nextFundingTime"`
	Time                 int64  `json:"time"`
}

type bnMarginInterestRate struct {
	Asset                  string `json:"asset"`
	NextHourlyInterestRate string `json:"nextHourlyInterestRate"`
}

type bnMarginInventory struct {
	Assets     map[string]string `json:"assets"` // 假设所有资产值都是字符串类型
	UpdateTime int64             `json:"updateTime"`
}
