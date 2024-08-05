package dfmock

import "github.com/shopspring/decimal"

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

type klineEvent struct {
	Symbol                   string          `json:"Symbol"`
	OpenTime                 int64           `json:"OpenTime"`
	Open                     decimal.Decimal `json:"Open"`
	High                     decimal.Decimal `json:"High"`
	Low                      decimal.Decimal `json:"Low"`
	Close                    decimal.Decimal `json:"Close"`
	Volume                   decimal.Decimal `json:"Volume"`
	CloseTime                int64           `json:"CloseTime"`
	QuoteAssetVolume         decimal.Decimal `json:"QuoteAssetVolume"`
	NumberOfTrades           int64           `json:"NumberOfTrades"`
	TakerBuyBaseAssetVolume  decimal.Decimal `json:"TakerBuyBaseAssetVolume"`
	TakerBuyQuoteAssetVolume decimal.Decimal `json:"TakerBuyQuoteAssetVolume"`
}
