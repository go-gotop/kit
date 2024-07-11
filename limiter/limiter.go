package limiter

type LimitType string

const (
	CreateOcoOrderLimit LimitType = "CREATE_OCO_ORDER_LIMIT" // 创建现货oco订单
	CreateOrderLimit    LimitType = "CREATE_ORDER"           // 创建订单
	UpdateOrderLimit    LimitType = "UPDATE_ORDER"           // 更新订单
	CancelOrderLimit    LimitType = "CANCEL_ORDER"           // 取消订单
	SearchOrderLimit    LimitType = "SEARCH_ORDER"           // 查询订单
	NormalRequestLimit  LimitType = "NORMAL_REQUEST"         // 普通请求
)

type LimiterReq struct {
	AccountId   string //  交易账户用户ID
	LimiterType LimitType
}

//go:generate mockgen -destination=../limiter/mocks/limiter.go -package=mklimiter . Limiter
type Limiter interface {
	WsAllow() bool
	SpotAllow(t *LimiterReq) bool
	FutureAllow(t *LimiterReq) bool
}
