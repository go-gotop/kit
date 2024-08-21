package tencent

type Option func(*options)

type options struct {
	SecretID    string
	SecretKey   string
	AppID       string
	UserSession string // 用户自定义 session，server 会原样返回
	TimeOut     int    // 超时
}

func WithSecretID(secretID string) Option {
	return func(o *options) {
		o.SecretID = secretID
	}
}

func WithSecretKey(secretKey string) Option {
	return func(o *options) {
		o.SecretKey = secretKey
	}
}

func WithAppID(appID string) Option {
	return func(o *options) {
		o.AppID = appID
	}
}

func WithUserSession(userSession string) Option {
	return func(o *options) {
		o.UserSession = userSession
	}
}

func WithTimeOut(timeOut int) Option {
	return func(o *options) {
		o.TimeOut = timeOut
	}
}
