package ratelimit

import "context"

type Limiter interface {
	// Limit 触发限流器 key:限流对象
	//bool 代表是否限流
	//err 限流器本身有没有错误
	Limit(ctx context.Context, key string) (bool, error)
}
