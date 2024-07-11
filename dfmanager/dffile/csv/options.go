package csv

type Options func(*options)

type options struct {
	start int64
	end   int64
}

func WithStart(start int64) Options {
	return func(o *options) {
		o.start = start
	}
}

func WithEnd(end int64) Options {
	return func(o *options) {
		o.end = end
	}
}