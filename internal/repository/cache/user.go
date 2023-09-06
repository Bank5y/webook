package cache

import (
	"context"
	"github.com/redis/go-redis/v9"
	"webook/internal/domain"
)

type UserCache struct {
	client *redis.Cmdable
}

// A 用到了 B, B 一定是接口
// A 用到了 B, B 一定是 A 的字段
// A 用到了 B, A	绝对不初始化 B, 而是外面注入

func NewUserCache(client *redis.Cmdable) *UserCache {
	return &UserCache{client: client}
}

func (u *UserCache) GetUser(ctx context.Context, email string) (domain.User, error) {

}
