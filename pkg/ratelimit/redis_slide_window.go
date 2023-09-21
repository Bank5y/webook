package ratelimit

import (
	"context"
	_ "embed"
	"github.com/redis/go-redis/v9"
	"time"
)

//go:embed slide_window.lua
var slideWindow string

type RedisSlidingWindowLimiter struct {
	cmd redis.Cmdable

	//窗口大小
	interval time.Duration
	//阈值
	rate int
}

func NewRedisSlidingWindowLimiter(cmd redis.Cmdable, interval time.Duration, rate int) *RedisSlidingWindowLimiter {
	return &RedisSlidingWindowLimiter{cmd: cmd, interval: interval, rate: rate}
}

func (r *RedisSlidingWindowLimiter) Limit(ctx context.Context, key string) (bool, error) {
	return r.cmd.Eval(ctx, slideWindow, []string{key},
		r.interval, r.rate, time.Now().UnixMilli()).Bool()
}
