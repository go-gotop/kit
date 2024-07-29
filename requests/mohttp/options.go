package mohttp

import (
	"net/http"
)

type Option func(o *options)

type options struct {
	proxyUrl   string
	timeOffset int64
	httpClient *http.Client
}

func ProxyURL(p string) Option {
	return func(o *options) { o.proxyUrl = p }
}

func HttpClient(h *http.Client) Option {
	return func(o *options) { o.httpClient = h }
}

func TimeOffset(t int64) Option {
	return func(o *options) { o.timeOffset = t }
}
