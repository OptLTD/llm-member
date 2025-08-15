package config

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisConfig Redis配置
type RedisConfig struct {
	Host string // 主机地址
	Port string // 端口
	Pass string // 密码
	DB   string // 数据库编号
}

// GetRedisClient 创建Redis客户端连接
func GetRedisClient(cfg *RedisConfig) (*redis.Client, error) {
	db, _ := strconv.Atoi(cfg.DB)
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	rdb := redis.NewClient(&redis.Options{
		Addr: addr, DB: db, Password: cfg.Pass,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %v", err)
	}

	return rdb, nil
}
