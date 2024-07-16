package dffile

import (
	"github.com/go-kratos/kratos/v2/log"
)

type Option func(*options)

type options struct {
	path   string
	logger *log.Helper
}

func WithLogger(logger *log.Helper) Option {
	return func(o *options) {
		o.logger = logger
	}
}

func WithPath(path string) Option {
	return func(o *options) {
		o.path = path
	}
}
