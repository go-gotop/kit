package sms

type SendMobileRequest struct {
	TemplateParam []string // 模版参数
	TemplateID    string   // 模版ID
	PhoneNumbers  []string // 手机号
}

type Sms interface {
	SendMobile(request *SendMobileRequest) error
}
