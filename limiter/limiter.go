package limiter

type LimitType string

const (
	CreateOcoOrderLimit LimitType = "create_oco_order" // 创建现货oco订单
	CreateOrderLimit    LimitType = "create_order"     // 创建订单
	UpdateOrderLimit    LimitType = "update_order"     // 更新订单
	CancelOrderLimit    LimitType = "cancel_order"     // 取消订单
	SearchOrderLimit    LimitType = "search_order"     // 查询订单
	NormalRequestLimit  LimitType = "normal_request"   // 普通请求
)

type Limiter interface {
	WsAllow() bool
	SpotAllow(t LimitType) bool
	FutureAllow(t LimitType) bool
}
