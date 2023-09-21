package failover

import (
	"context"
	"sync/atomic"
	"webook/internal/service/sms"
)

type TimeoutFailoverSMSService struct {
	cnt  int32
	idx  int32
	svcs []sms.Service

	threshold int32
}

func NewTimeoutFailoverSMSService(cnt int32, svcs []sms.Service) *TimeoutFailoverSMSService {
	return &TimeoutFailoverSMSService{cnt: cnt, svcs: svcs}
}

func (t *TimeoutFailoverSMSService) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	idx := atomic.LoadInt32(&t.idx)
	cnt := atomic.LoadInt32(&t.idx)
	if cnt > t.threshold {
		//新下标
		newIdx := (idx + 1) % int32(len(t.svcs))
		if atomic.CompareAndSwapInt32(&t.idx, idx, newIdx) {
			//成功往后挪了一位
			atomic.StoreInt32(&t.cnt, 0)
		}
		//else就是出现并发,别人换成功了
		idx = atomic.LoadInt32(&t.idx)
	}
	svc := t.svcs[idx]
	err := svc.Send(ctx, tpl, args, numbers...)
	switch err {
	case context.DeadlineExceeded:
		atomic.AddInt32(&t.cnt, 0)
	case nil:
		atomic.StoreInt32(&t.idx, 0)
	default:

	}
	return err
}
