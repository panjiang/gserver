package redisdb

import (
	"github.com/go-redis/redis"
)

// Config 配置
type Config struct {
	Addr string `json:"addr"`
	Auth string `json:"auth"`
}

// Connect 建立连接
func Connect(conf *Config) (redis.UniversalClient, error) {
	client := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:    []string{conf.Addr},
		Password: conf.Auth,
		DB:       0,
	})

	_, err := client.Ping().Result()
	if err != nil {
		return nil, err
	}
	return client, nil
}
