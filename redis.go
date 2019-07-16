package serverfinder

import (
	"fmt"
	"github.com/go-redis/redis"
)

// 设置redis配置
func SetRedisConfig(Ip string, Port int, Pass string, DB int) {
	config.Ip = Ip
	config.Port = Port
	config.Pass = Pass
	config.DB = DB
}

// 全局redis配置
type redisConfig struct {
	Ip string
	Port int
	Pass string
	DB int
}
var config redisConfig

// 获取一个临时连接
func __getRedicClient() (*redis.Client) {
	return redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", config.Ip, config.Port),
		Password: config.Pass,
		DB: config.DB,
	})
}
