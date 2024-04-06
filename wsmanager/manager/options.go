package manager

import (
	"time"

	"github.com/go-gotop/kit/limiter"
)

type ConnConfig func(*connConfig)

type connConfig struct {
	maxConn         int             // 最大连接数
	maxConnDuration time.Duration   // 最大连接持续时间
	connLimiter     limiter.Limiter // 连接限流器
	isCheckReConn   bool            // 是否检查重连
}

func WithMaxConn(maxConn int) ConnConfig {
	return func(c *connConfig) {
		c.maxConn = maxConn
	}
}

func WithMaxConnDuration(maxConnDuration time.Duration) ConnConfig {
	return func(c *connConfig) {
		c.maxConnDuration = maxConnDuration
	}
}

func WithConnLimiter(connLimiter limiter.Limiter) ConnConfig {
	return func(c *connConfig) {
		c.connLimiter = connLimiter
	}
}

func WithCheckReConn(isCheckReConn bool) ConnConfig {
	return func(c *connConfig) {
		c.isCheckReConn = isCheckReConn
	}
}
