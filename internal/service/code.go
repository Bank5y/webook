package service

import (
	"context"
	"fmt"
	"math/rand"
	"webook/internal/repository"
	"webook/internal/service/sms"
)

// TODO
const codeTplId = ""

var (
	ErrCodeSendTooMany = repository.ErrCodeSendTooMany
)

type CodeService interface {
	Send(ctx context.Context, biz string, phone string) error
	Verify(ctx context.Context, biz string, phone string, inputCode string) (bool, error)
}

type CodeDevService struct {
	smsSvc sms.Service
	repo   repository.CodeRepository
}

func NewCodeDevService(smsSvc sms.Service, repo repository.CodeRepository) *CodeDevService {
	return &CodeDevService{smsSvc: smsSvc, repo: repo}
}

// Send 发送验证码
func (svc *CodeDevService) Send(ctx context.Context, biz string, phone string) error {

	//两个步骤
	//1.生成一个验证码
	code := svc.generateCode()
	//2.加入redis
	err := svc.repo.Store(ctx, biz, phone, code)
	if err != nil {
		return err
	}
	//发送
	err = svc.smsSvc.Send(ctx, codeTplId, []string{code}, phone)
	if err != nil {
		//若出错,不可以删除redis中的验证码
		//因为err可能是超时的err,无法知晓是否发送成功
		//若需要做重试功能,应该再传入一个支持重试功能的smsSvc
		return err
	}
	return err
}

func (svc *CodeDevService) Verify(ctx context.Context, biz string, phone string, inputCode string) (bool, error) {
	return svc.repo.Verify(ctx, biz, phone, inputCode)
}

func (svc *CodeDevService) generateCode() string {
	num := rand.Intn(1000000)
	return fmt.Sprintf("%06d", num)
}
