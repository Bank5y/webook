package syntoasy

import (
	"context"
	"log"
	"sync/atomic"
	"time"
	"webook/internal/service/sms"
	"webook/pkg/ratelimit"
)

const (
	defaultRetryTime = time.Second * 10
	defaultRetryCnt  = 5
	defaultRate      = 5
)

type configs struct {
	retryTime time.Duration
	retryCnt  int
	rate      int
	respTime  int64
}

type Config interface {
	apply(*configs)
}

type configFunc func(*configs)

func (f configFunc) apply(config *configs) {
	f(config)
}

func WithRetryTime(time time.Duration) Config {
	return configFunc(func(c *configs) {
		c.retryTime = time
	})
}

func WithRetryCnt(cnt int) Config {
	return configFunc(func(c *configs) {
		c.retryCnt = cnt
	})
}
func WithRate(rate int) Config {
	return configFunc(func(c *configs) {
		c.rate = rate
	})
}

type TooManyErrFailoverService struct {
	svcs    []sms.Service
	config  configs
	limiter ratelimit.Limiter
	idx     int64
}

func NewTooManyErrFailoverService(svcs []sms.Service, limiter ratelimit.Limiter, confs ...Config) *TooManyErrFailoverService {
	configs := configs{
		retryTime: defaultRetryTime,
		retryCnt:  defaultRetryCnt,
		rate:      defaultRate,
	}
	for _, conf := range confs {
		conf.apply(&configs)
	}
	return &TooManyErrFailoverService{
		svcs:    svcs,
		limiter: limiter,
		config:  configs,
	}
}

func (t *TooManyErrFailoverService) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	idx := atomic.LoadInt64(&t.idx)
	limit, err := t.limiter.Limit(ctx, "cnt:send")
	if err != nil {
		log.Println(err)
		return err
	}
	if limit {
		newIdx := (idx + 1) % int64(len(t.svcs))
		if atomic.CompareAndSwapInt64(&t.idx, idx, newIdx) {
			t.config.respTime = 0
		}
		idx = atomic.LoadInt64(&t.idx)
		return err
	}
	svc := t.svcs[idx]

	startTime := time.Now().UnixNano()
	err = svc.Send(ctx, tpl, args, numbers...)
	endTime := time.Now().UnixNano()
	if err != nil {
		newIdx := (idx + 1) % int64(len(t.svcs))
		atomic.CompareAndSwapInt64(&t.idx, idx, newIdx)
		return err
	}

	if t.config.respTime == 0 {
		val := endTime - startTime
		atomic.StoreInt64(&t.config.respTime, val)
		return err
	}
	if endTime-startTime > 1000 || (endTime-startTime)/t.config.respTime > int64(t.config.rate) {
		atomic.StoreInt64(&t.config.respTime, 0)
		newIdx := (idx + 1) % int64(len(t.svcs))
		atomic.CompareAndSwapInt64(&t.idx, idx, newIdx)
	}
	return err
}
