package tencent

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-gotop/kit/sms"
	"github.com/tencentcloud/tencentcloud-sdk-go-intl-en/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go-intl-en/tencentcloud/common/profile"
	terros "github.com/tencentcloud/tencentcloud-sdk-go-intl-en/tencentcloud/common/errors"
	tsms "github.com/tencentcloud/tencentcloud-sdk-go-intl-en/tencentcloud/sms/v20210111"
)

var (
	ErrorMissingSecretID  = errors.New("missing secret id")
	ErrorMissingSecretKey = errors.New("missing secret key")
	ErrorMissingAppID     = errors.New("missing app id")
)

type tencent struct {
	client *tsms.Client
	opts   *options
}

func NewSmsTencent(opts ...Option) (sms.Sms, error) {
	o := &options{
		SecretID:    "",
		SecretKey:   "",
		AppID:       "",
		UserSession: "",
		TimeOut:     10,
	}

	for _, opt := range opts {
		opt(o)
	}

	if o.SecretID == "" {
		return nil, ErrorMissingSecretID
	}

	if o.SecretKey == "" {
		return nil, ErrorMissingSecretKey
	}

	if o.AppID == "" {
		return nil, ErrorMissingAppID
	}

	credential := common.NewCredential(
		o.SecretID,
		o.SecretKey,
	)

	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "sms.tencentcloudapi.com"
	cpf.HttpProfile.ReqMethod = "POST"

	client, _ := tsms.NewClient(credential, "ap-singapore", cpf)

	return &tencent{
		client: client,
		opts:   o,
	}, nil
}

func (t *tencent) SendMobile(request *sms.SendMobileRequest) error {
	req := tsms.NewSendSmsRequest()
	req.SmsSdkAppId = common.StringPtr(t.opts.AppID)
	req.SessionContext = common.StringPtr(t.opts.UserSession)
	req.TemplateId = common.StringPtr(request.TemplateID)
	req.TemplateParamSet = common.StringPtrs(request.TemplateParam)
	req.PhoneNumberSet = common.StringPtrs(request.PhoneNumbers)

	// 处理异常
	response, err := t.client.SendSms(req)
	if _, ok := err.(*terros.TencentCloudSDKError); ok {
		return err
	}
	// 非SDK异常，直接失败。实际代码中可以加入其他的处理。
	if err != nil {
		panic(err)
	}
	b, _ := json.Marshal(response.Response)
	// 打印返回的json字符串
	fmt.Printf("%s", b)

	return nil
}
