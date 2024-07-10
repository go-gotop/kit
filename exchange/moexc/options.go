package moexc

type Option func(*options)

type options struct {
	mockExchangEndpoint string
}

func WithWsEndpoint(mockExchangEndpoint string) Option {
	return func(o *options) {
		o.mockExchangEndpoint = mockExchangEndpoint
	}
}
