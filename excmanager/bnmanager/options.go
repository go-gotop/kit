package bnmanager

import (
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

type Option func(*options)

type options struct {
	logger               *log.Helper
	maxConn              int           // 最大连接数
	maxConnDuration      time.Duration // 最大连接持续时间
	listenKeyExpire      time.Duration // listenkey 过期时间
	checkListenKeyPeriod time.Duration // 检查 listenkey 的周期
}

func WithLogger(logger *log.Helper) Option {
	return func(o *options) {
		o.logger = logger
	}
}

func WithMaxConn(maxConn int) Option {
	return func(o *options) {
		o.maxConn = maxConn
	}
}

func WithMaxConnDuration(maxConnDuration time.Duration) Option {
	return func(o *options) {
		o.maxConnDuration = maxConnDuration
	}
}

func WithListenKeyExpire(listenKeyExpire time.Duration) Option {
	return func(o *options) {
		o.listenKeyExpire = listenKeyExpire
	}
}

func WithCheckListenKeyPeriod(checkListenKeyPeriod time.Duration) Option {
	return func(o *options) {
		o.checkListenKeyPeriod = checkListenKeyPeriod
	}
}
