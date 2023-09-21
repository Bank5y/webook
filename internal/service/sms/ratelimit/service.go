package ratelimit

import (
	"context"
	"fmt"
	"webook/internal/service/sms"
	"webook/pkg/ratelimit"
)

type Service struct {
	svc     sms.Service
	limiter ratelimit.Limiter
}

func NewService(svc sms.Service, limiter ratelimit.Limiter) *Service {
	return &Service{svc: svc, limiter: limiter}
}

func (s *Service) Send(ctx context.Context, tpl string, args []string, number ...string) error {
	//加新特性
	limited, err := s.limiter.Limit(ctx, "sms:tencent")
	if err != nil {
		return fmt.Errorf("短信服务判断是否限流出现问题,%w", err)
	}
	if limited {
		return fmt.Errorf("触发了限流")
	}
	err = s.svc.Send(ctx, tpl, args, number...)
	//加一些代码 新特性
	return err
}
