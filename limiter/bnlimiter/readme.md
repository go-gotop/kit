## 现货/杠杠

每分钟不超过 6000 权重请求（request_weight
每 5 分钟不超过 61000 次请求（raw_requests
每 10s 不超过 100 次下单请求（orders
每天不超过 200000 次下单请求（orders

## 合约

每分钟不超过 2400 权重请求（request_weight
每 10s 不超过 300 次下单请求（orders
每分钟不超过 1200 次下单请求（orders

## 账户信息（资产余额

- 现货：
  账户信息接口 （weight：20）
- 合约：
  账户余额接口（weight：5
  管理 listenkey（weight：1，生成、延长、删除分别 weight 是 1，一个 listenkey 有效期 60 分钟
  使用 ws+listenkey 监听账户余额变动

## 订单操作

- 现货
  下单（weight：1）、oco 下单（weight：2）
  撤销订单（weight：1）
  查询订单（weight：1）、ws 订单状态推送
- 合约：
  下单（weight：0）
  撤销订单（weight：1）
  查询订单（weight：1）、ws 订单状态推送




  我主要想测试速度率，也就是单位时间内的调用次数，单位时间内的权重限制，单位时间内的下单次数限制
