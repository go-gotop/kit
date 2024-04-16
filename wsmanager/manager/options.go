package manager

import (
	"time"

	"github.com/go-gotop/kit/limiter"
	"github.com/go-kratos/kratos/v2/log"
)

type ConnConfig func(*connConfig)

type connConfig struct {
	logger 		    *log.Helper     // 日志记录器
	maxConn         int             // 最大连接数
	maxConnDuration time.Duration   // 最大连接持续时间
	connLimiter     limiter.Limiter // 连接限流器
	isCheckReConn   bool            // 是否检查重连
}

func WithLogger(logger log.Logger) ConnConfig {
	return func(c *connConfig) {
		c.logger = log.NewHelper(logger)
	}
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
