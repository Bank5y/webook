package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"sync"
	"time"
)

// error
var (
	ErrCodeSendTooMany   = errors.New("发送太频繁")
	ErrCodeVerifyTooMany = errors.New("验证次数太多")
	ErrUnknownForCode    = errors.New("未知错误")
)
var (
	lock = sync.Mutex{}
)

// 该命令在编译时会将set_code的代码放进来这个luaSetCode变量里
//
//go:embed lua/set_code.lua
var luaSetCode string

//go:embed lua/verify_code.lua
var luaVerifyCode string

type CodeCache interface {
	Set(ctx context.Context, biz, phone, code string) error
	Verify(ctx context.Context, biz, phone, code string) (bool, error)
}

type LocalCode struct {
	Code     string
	TryCount int
	Ctime    int64
}

type CodeLocalCache struct {
	CacheMap map[string]LocalCode
}

// NewCodeLocalCache 本地缓存
func NewCodeLocalCache() CodeCache {
	return &CodeLocalCache{
		CacheMap: make(map[string]LocalCode, 10000),
	}
}

func (c *CodeLocalCache) Set(ctx context.Context, biz, phone, code string) error {
	key := biz + ":" + phone
	lock.Lock()
	if localCode, ok := c.CacheMap[key]; ok {
		now := time.Now().UnixMilli()
		if (now - localCode.Ctime) < 60*1000 {
			lock.Unlock()
			return ErrCodeSendTooMany
		}
		var result LocalCode
		result = c.CacheMap[key]
		result.Ctime = now
		lock.Unlock()
		return nil
	}
	now := time.Now().UnixMilli()
	cache := LocalCode{
		Code:     code,
		TryCount: 3,
		Ctime:    now,
	}
	c.CacheMap[key] = cache
	lock.Unlock()
	return nil
}

func (c *CodeLocalCache) Verify(ctx context.Context, biz, phone, code string) (bool, error) {
	key := biz + ":" + phone
	lock.Lock()
	if localCode, ok := c.CacheMap[key]; ok {
		now := time.Now().UnixMilli()
		if localCode.TryCount <= 0 || now-localCode.Ctime > 10*60*6000 {
			delete(c.CacheMap, key)
			lock.Unlock()
			return false, nil
		}
		delete(c.CacheMap, key)
		lock.Unlock()
		return true, nil
	}
	lock.Unlock()
	return false, nil
}

type CodeRedisCache struct {
	client redis.Cmdable
}

func NewCodeRedisCache(client redis.Cmdable) CodeCache {
	return &CodeRedisCache{
		client: client,
	}
}

// Set 存入Redis
func (c *CodeRedisCache) Set(ctx context.Context, biz, phone, code string) error {
	val, err := c.client.Eval(ctx, luaSetCode, []string{c.key(biz, phone)}, code).Int()
	if err != nil {
		return err
	}
	switch val {
	case 0:
		//没有问题
		return nil
	case -1:
		//发送太频繁
		return ErrCodeSendTooMany
	default:
		return errors.New("系统错误")
		//系统错误
	}
}

// Verify 验证
func (c *CodeRedisCache) Verify(ctx context.Context, biz, phone, code string) (bool, error) {
	println(biz)
	val, err := c.client.Eval(ctx, luaVerifyCode, []string{c.key(biz, phone)}, code).Int()
	if err != nil {
		return false, err
	}
	switch val {
	case 0:
		return true, nil
	case -1:
		return false, ErrCodeVerifyTooMany
	case -2:
		return false, nil
	default:
		return false, ErrUnknownForCode
	}
}

func (c *CodeRedisCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}
