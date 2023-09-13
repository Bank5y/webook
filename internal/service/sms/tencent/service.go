package tencent

import (
	"context"
	"fmt"
	"github.com/ecodeclub/ekit"
	"github.com/ecodeclub/ekit/slice"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
)

type Service struct {
	appId    *string
	signName *string
	client   *sms.Client
}

func NewService(appId string, signName string, client *sms.Client) *Service {
	return &Service{appId: &appId, signName: &signName, client: client}
}
func (svc *Service) Send(ctx context.Context, tplId string, args []string, number ...string) error {
	req := sms.NewSendSmsRequest()
	req.SmsSdkAppId = svc.appId
	req.SignName = svc.signName
	req.TemplateId = ekit.ToPtr[string](tplId)
	req.PhoneNumberSet = svc.toStringPtrSlice(number)
	req.TemplateParamSet = svc.toStringPtrSlice(args)
	resp, err := svc.client.SendSms(req)
	if err != nil {
		return err
	}
	for _, status := range resp.Response.SendStatusSet {
		if status.Code == nil || *(status.Code) != "Ok" {
			return fmt.Errorf("发送短信失败%s,%s", *status.Code, *status.Message)
		}
	}
	return nil
}

func (svc *Service) toStringPtrSlice(src []string) []*string {
	return slice.Map[string, *string](src, func(idx int, src string) *string {
		return &src
	})
}
