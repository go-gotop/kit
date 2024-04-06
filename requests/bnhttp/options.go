package bnhttp

import (
	"net/http"
)

type Option func(o *options)

type options struct {
	apiKey     string
	secretKey  string
	baseURL   string
	proxyUrl   string
	timeOffset int64
	httpClient *http.Client
}

func BaseUrl(b string) Option {
	return func(o *options) { o.baseURL = b }
}

func ProxyURL(p string) Option {
	return func(o *options) { o.proxyUrl = p }
}

func HttpClient(h *http.Client) Option {
	return func(o *options) { o.httpClient = h }
}

func APIKey(k string) Option {
	return func(o *options) { o.apiKey = k }
}

func SecretKey(s string) Option {
	return func(o *options) { o.secretKey = s }
}

func TimeOffset(t int64) Option {
	return func(o *options) { o.timeOffset = t }
}
