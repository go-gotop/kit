package streamokx

type login struct {
	Op   string `json:"op"`
	Args []struct {
		ApiKey     string `json:"apiKey"`
		Passphrase string `json:"passphrase"`
		Timestamp  string `json:"timestamp"`
		Sign       string `json:"sign"`
	} `json:"args"`
}

type sub struct {
	Op   string `json:"op"`
	Args []struct {
		Channel  string `json:"channel"`
		InstType string `json:"instType"`
	} `json:"args"`
}

type okWsOrderUpdateEvent struct {
	Arg  okWsOrderUpdateArg    `json:"arg"`
	Data []okWsOrderUpdateData `json:"data"`
}

type okWsOrderUpdateArg struct {
	Channel    string `json:"channel"`
	InstType   string `json:"instType"`
	InstID     string `json:"instId"`
	InstFamily string `json:"instFamily"`
}

type okWsOrderUpdateData struct {
	InstType        string `json:"instType"`
	InstID          string `json:"instId"`
	OrderID         string `json:"ordId"`
	ClientOrderID   string `json:"clOrdId"`
	Px              string `json:"px"`              // 委托价格
	Sz              string `json:"sz"`              // 委托数量 币币/币币杠杆，以币为单位；交割/永续/期权 ，以张为单位
	FillNotionalUsd string `json:"fillNotionalUsd"` // 已成交金额
	OrderType       string `json:"ordType"`         // 订单类型
	Side            string `json:"side"`            // 订单方向
	PosSide         string `json:"posSide"`         // 持仓方向
	FillPx          string `json:"fillPx"`          // 最新成交价格
	FillSz          string `json:"fillSz"`          // 最新成交数量
	FillFee         string `json:"fillFees"`        // 最新一笔成交手续费金额 或 返佣金额，看正负数
	FillFeeCcy      string `json:"fillFeeCcy"`      // 最新一笔成交手续费币种 或 返佣币种
	AccFillSz       string `json:"accFillSz"`       // 累计成交数量
	AvgPx           string `json:"avgPx"`           // 平均成交价格
	State           string `json:"state"`           // 订单状态
	Lever           string `json:"lever"`           // 杠杆倍数
	FeeCcy          string `json:"feeCcy"`          // 手续费币种
	Fee             string `json:"fee"`             // 订单交易累计的手续费与返佣
	RebateCcy       string `json:"rebateCcy"`       // 返佣币种
	Rebate          string `json:"rebate"`          // 返佣累积金额
	UpdateTime      string `json:"uTime"`           // 更新时间
	CreateTime      string `json:"cTime"`           // 创建时间
	LastPx          string `json:"lastPx"`          // 最新成交价格
	ExecType        string `json:"execType"`        // 订单执行方向 T：taker M：maker
	Code            string `json:"code"`            // 错误码
	Msg             string `json:"msg"`             // 错误信息
}
