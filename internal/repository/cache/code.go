package cache

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
)

// error
var (
	ErrCodeSendTooMany   = errors.New("发送太频繁")
	ErrCodeVerifyTooMany = errors.New("验证次数太多")
	ErrUnknownForCode    = errors.New("未知错误")
)

// 该命令在编译时会将set_code的代码放进来这个luaSetCode变量里
//
//go:embed lua/set_code.lua
var luaSetCode string

//go:embed lua/verify_code.lua
var luaVerifyCode string

type CodeCache struct {
	client redis.Cmdable
}

func NewCodeCache(client redis.Cmdable) *CodeCache {
	return &CodeCache{
		client: client,
	}
}

// Set 存入Redis
func (c *CodeCache) Set(ctx context.Context, biz, phone, code string) error {
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
func (c *CodeCache) Verify(ctx context.Context, biz, phone, code string) (bool, error) {
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

func (c *CodeCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}
