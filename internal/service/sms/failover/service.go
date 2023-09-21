package failover

import (
	"context"
	"errors"
	"log"
	"sync/atomic"
	"webook/internal/service/sms"
)

type SMSFailoverService struct {
	services []sms.Service
	idx      uint64
}

func NewSMSFailoverService(services []sms.Service) *SMSFailoverService {
	return &SMSFailoverService{services: services}
}

// Send 服务轮询
func (s SMSFailoverService) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	for _, svc := range s.services {
		err := svc.Send(ctx, tpl, args, numbers...)
		if err != nil {
			return nil
		}
		//正常流程 输出日志
		//做好监控
		log.Println(err)
	}
	return errors.New("全部服务都失败了")
}

// SendV1 动态指针轮询
func (s SMSFailoverService) SendV1(ctx context.Context, tpl string, args []string, numbers ...string) error {
	idx := atomic.AddUint64(&s.idx, 1)
	length := uint64(len(s.services))
	for i := idx; i < idx+length; i++ {
		svc := s.services[int(i%length)]
		err := svc.Send(ctx, tpl, args, numbers...)
		switch err {
		case nil:
			return nil
		case context.DeadlineExceeded, context.Canceled:
			return err
		default:
			//输出日志
		}
	}
	return errors.New("全部服务都失败了")
}
