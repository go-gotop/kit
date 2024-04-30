package limiter

const (
	WsConnectLimit         = "WS_CONNECT"
	SpotCreateOrderLimit   = "SPOT_CREATE_ORDER"
	SpotNormalRequestLimit = "SPOT_NORMAL_REQUEST"
	FutureCreateOrderLimit = "FUTURE_CREATE_ORDER"

	// redis key
	SpotWeight = "SPOT_WEIGHT"
	FutureWeight = "FUTURE_WEIGHT"
	SpotLastRestTime = "SPOT_LAST_RESET_TIME"
	FutureLastRestTime = "FUTURE_LAST_RESET_TIME"
)

type Option func(*Options)

type WeightType int64

type PeriodLimit struct {
	WsConnectPeriod         string
	WsConnectTimes          int64
	SpotCreateOrderPeriod   string
	SpotCreateOrderTimes    int64
	SpotNormalRequestPeriod string
	SpotNormalRequestTimes  int64
	FutureCreateOrderPeriod string
	FutureCreateOrderTimes  int64
}

type Options struct {
	// 请求次数限制
	PeriodLimitArray []PeriodLimit

	// 权重
	CreateSpotOrderWeights   WeightType
	CreateOcoOrderWeights    WeightType
	CreateFutureOrderWeights WeightType
	CancelSpotOrderWeights   WeightType
	CancelFutureOrderWeights WeightType
	SearchSpotOrderWeights   WeightType
	SearchFutureOrderWeights WeightType
	UpdateSpotOrderWeights   WeightType
	UpdateFutureOrderWeights WeightType
	OtherWeights             WeightType
}

func WithPeriodLimitArray(p []PeriodLimit) Option {
	return func(o *Options) {
		o.PeriodLimitArray = p
	}
}

func WithCreateSpotOrderWeights(w WeightType) Option {
	return func(o *Options) {
		o.CreateSpotOrderWeights = w
	}
}

func WithCreateOcoOrderWeights(w WeightType) Option {
	return func(o *Options) {
		o.CreateOcoOrderWeights = w
	}
}

func WithCreateFutureOrderWeights(w WeightType) Option {
	return func(o *Options) {
		o.CreateFutureOrderWeights = w
	}
}

func WithCancelSpotOrderWeights(w WeightType) Option {
	return func(o *Options) {
		o.CancelSpotOrderWeights = w
	}
}

func WithCancelFutureOrderWeights(w WeightType) Option {
	return func(o *Options) {
		o.CancelFutureOrderWeights = w
	}
}

func WithSearchSpotOrderWeights(w WeightType) Option {
	return func(o *Options) {
		o.SearchSpotOrderWeights = w
	}
}

func WithSearchFutureOrderWeights(w WeightType) Option {
	return func(o *Options) {
		o.SearchFutureOrderWeights = w
	}
}

func WithUpdateSpotOrderWeights(w WeightType) Option {
	return func(o *Options) {
		o.UpdateSpotOrderWeights = w
	}
}

func WithUpdateFutureOrderWeights(w WeightType) Option {
	return func(o *Options) {
		o.UpdateFutureOrderWeights = w
	}
}

func WithOtherWeights(w WeightType) Option {
	return func(o *Options) {
		o.OtherWeights = w
	}
}
