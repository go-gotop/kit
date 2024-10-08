package streamokx

import (
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

type Option func(*options)

type options struct {
	logger          *log.Helper
	maxConnDuration time.Duration // 最大连接持续时间
	connectCount    int
}

func WithLogger(logger *log.Helper) Option {
	return func(o *options) {
		o.logger = logger
	}
}

func WithMaxConnDuration(maxConnDuration time.Duration) Option {
	return func(o *options) {
		o.maxConnDuration = maxConnDuration
	}
}

func WithConnectCount(connectCount int) Option {
	return func(o *options) {
		o.connectCount = connectCount
	}
}
