package redis

import (
	"CMS/config"
	"context"

	"github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client
var ctx = context.Background()

// Init 初始化 Redis 连接
func Init() {
	// 获取配置
	cfg := config.LoadedConfig

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// 测试连接
	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		panic("Redis 连接失败: " + err.Error())
	}
}
