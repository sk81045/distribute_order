package core

import (
	"context"

	"Hwgen/global"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

func Redis() *redis.Client {
	redisCfg := global.H_CONFIG.Redis
	client := redis.NewClient(&redis.Options{
		Addr:     redisCfg.Addr,
		Password: redisCfg.Password, // no password set
		DB:       redisCfg.DB,       // use default DB
	})
	pong, err := client.Ping(context.Background()).Result()
	if err != nil {
		global.H_LOG.Error("redis connect ping failed, err:", zap.Error(err))
	}

	global.H_LOG.Info("redis connect ping response:", zap.String("pong", pong))

	return client
}
