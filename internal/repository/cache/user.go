package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
	"webook/internal/domain"
)

var (
	ErrKeyNotExist = redis.Nil
)

type UserCache struct {
	client     redis.Cmdable
	expiration time.Duration
}

// A 用到了 B, B 一定是接口
// A 用到了 B, B 一定是 A 的字段
// A 用到了 B, A	绝对不初始化 B, 而是外面注入

func NewUserCache(client redis.Cmdable) *UserCache {
	return &UserCache{
		client:     client,
		expiration: time.Minute * 15,
	}
}

//若error为nil 则默认缓存中有数据
//若没有数据 则返回一个特定的error

func (cache *UserCache) Get(ctx context.Context, id int) (domain.User, error) {
	key := cache.key(id)
	bytes, err := cache.client.Get(ctx, key).Bytes()
	if err != nil {
		return domain.User{}, err
	}
	var u domain.User
	err = json.Unmarshal(bytes, &u)
	return u, err
}

func (cache *UserCache) Set(ctx context.Context, u domain.User) error {
	val, err := json.Marshal(u)
	if err != nil {
		return err
	}
	key := cache.key(u.Id)
	return cache.client.Set(ctx, key, val, cache.expiration).Err()
}

func (cache *UserCache) key(id int) string {
	return fmt.Sprintf("user:info:%s", id)
}
