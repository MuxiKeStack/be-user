package ioc

import (
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

func InitRedisCmd(client *redis.Client) redis.Cmdable {
	return client
}

func InitRedisClient() *redis.Client {
	type Config struct {
		Addr     string `yaml:"addr"`
		Password string `yaml:"password"`
	}
	var cfg Config
	err := viper.UnmarshalKey("redis", &cfg)
	if err != nil {
		panic(err)
	}
	return redis.NewClient(&redis.Options{Addr: cfg.Addr, Password: cfg.Password})
}
