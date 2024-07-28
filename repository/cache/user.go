package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/MuxiKeStack/be-user/domain"
	"github.com/redis/go-redis/v9"
	"time"
)

var ErrKeyNotExists = redis.Nil

type UserCache interface {
	Get(ctx context.Context, uid int64) (domain.User, error)
	Set(ctx context.Context, user domain.User) error
	Del(ctx context.Context, uid int64) error
}

type RedisUserCache struct {
	cmd redis.Cmdable
}

func (cache *RedisUserCache) Get(ctx context.Context, uid int64) (domain.User, error) {
	key := cache.key(uid)
	val, err := cache.cmd.Get(ctx, key).Bytes()
	if err != nil {
		return domain.User{}, err
	}
	var u domain.User
	err = json.Unmarshal(val, &u)
	return u, err
}

func (cache *RedisUserCache) Set(ctx context.Context, user domain.User) error {
	key := cache.key(user.Id)
	val, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return cache.cmd.Set(ctx, key, val, time.Minute*30).Err()
}

func (cache *RedisUserCache) key(uid int64) string {
	return fmt.Sprintf("kstack:users:%d", uid)
}

func (cache *RedisUserCache) Del(ctx context.Context, uid int64) error {
	return cache.cmd.Del(ctx, cache.key(uid)).Err()
}

func NewRedisUserCache(cmd redis.Cmdable) UserCache {
	return &RedisUserCache{cmd: cmd}
}
