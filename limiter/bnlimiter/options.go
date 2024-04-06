package bnlimiter

type Option func(*options)

type WeightType int64

type PeriodLimit struct {
	wsConnectPeriod         string
	wsConnectTimes          int64
	spotCreateOrderPeriod   string
	spotCreateOrderTimes    int64
	spotNormalRequestPeriod string
	spotNormalRequestTimes  int64
	futureCreateOrderPeriod string
	futureCreateOrderTimes  int64
}

type options struct {
	// 请求次数限制
	periodLimitArray []PeriodLimit

	// 权重
	createSpotOrderWeights   WeightType
	createOcoOrderWeights    WeightType
	createFutureOrderWeights WeightType
	cancelSpotOrderWeights   WeightType
	cancelFutureOrderWeights WeightType
	searchSpotOrderWeights   WeightType
	searchFutureOrderWeights WeightType
	updateSpotOrderWeights   WeightType
	updateFutureOrderWeights WeightType
	otherWeights             WeightType
}

func WithCreateSpotOrderWeights(w WeightType) Option {
	return func(o *options) {
		o.createSpotOrderWeights = w
	}
}

func WithCreateOcoOrderWeights(w WeightType) Option {
	return func(o *options) {
		o.createOcoOrderWeights = w
	}
}

func WithCreateFutureOrderWeights(w WeightType) Option {
	return func(o *options) {
		o.createFutureOrderWeights = w
	}
}

func WithCancelSpotOrderWeights(w WeightType) Option {
	return func(o *options) {
		o.cancelSpotOrderWeights = w
	}
}

func WithCancelFutureOrderWeights(w WeightType) Option {
	return func(o *options) {
		o.cancelFutureOrderWeights = w
	}
}

func WithSearchSpotOrderWeights(w WeightType) Option {
	return func(o *options) {
		o.searchSpotOrderWeights = w
	}
}

func WithSearchFutureOrderWeights(w WeightType) Option {
	return func(o *options) {
		o.searchFutureOrderWeights = w
	}
}

func WithUpdateSpotOrderWeights(w WeightType) Option {
	return func(o *options) {
		o.updateSpotOrderWeights = w
	}
}

func WithUpdateFutureOrderWeights(w WeightType) Option {
	return func(o *options) {
		o.updateFutureOrderWeights = w
	}
}

func WithOtherWeights(w WeightType) Option {
	return func(o *options) {
		o.otherWeights = w
	}
}
