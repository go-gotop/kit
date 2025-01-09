package dfbinance

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

type binanceFuntureMarkPriceEvent struct {
	Event                string `json:"e"`
	Time                 int64  `json:"E"`
	Symbol               string `json:"s"`
	MarkPrice            string `json:"p"`
	IndexPrice           string `json:"i"` // 现货指数价格
	EstimatedSettlePrice string `json:"P"` // 预估结算价
	LastFundingRate      string `json:"r"` // 资金费率
	NextFundingTime      int64  `json:"T"` // 下个资金时间
}

type binanceFuturesMarkPriceStream struct {
	Stream string                          `json:"stream"`
	Data   []*binanceFuntureMarkPriceEvent `json:"data"`
}

type binanceFuturesMarkPriceSingleStream struct {
	Stream string                        `json:"stream"`
	Data   *binanceFuntureMarkPriceEvent `json:"data"`
}

type binanceKlineEvent struct {
	Event     string `json:"e"`
	Time      int64  `json:"E"`
	Symbol    string `json:"s"`
	KlineData struct {
		StartTime    int64  `json:"t"`
		EndTime      int64  `json:"T"`
		Symbol       string `json:"s"`
		Interval     string `json:"i"`
		FirstTradeID int64  `json:"f"`
		LastTradeID  int64  `json:"L"`
		OpenPrice    string `json:"o"`
		ClosePrice   string `json:"c"`
		HighPrice    string `json:"h"`
		LowPrice     string `json:"l"`
		Volume       string `json:"v"`
		TradeNum     int64  `json:"n"`
		IsClosed     bool   `json:"x"`
		QuoteVolume  string `json:"q"`
		TakerVolume  string `json:"V"`
		TakerQuote   string `json:"Q"`
		Ignore       string `json:"B"`
	} `json:"k"`
}