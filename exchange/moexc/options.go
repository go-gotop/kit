package moexc

type Option func(*options)

type options struct {
	mockExchangEndpoint string
}

func WithMockExchangEndpoint(mockExchangEndpoint string) Option {
	return func(o *options) {
		o.mockExchangEndpoint = mockExchangEndpoint
	}
}
