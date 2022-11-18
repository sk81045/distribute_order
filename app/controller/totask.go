package controller

import (
	"Hwgen/global"
	"context"
	"fmt"
	"go.uber.org/zap"
	"time"
)

type RedisStore struct {
	Expiration time.Duration
	PreKey     string
	Context    context.Context
}

func (rs *RedisStore) Set(id string, value string) {
	fmt.Println("ddd")
	err := global.H_REDIS.Set(context.Background(), id, value, rs.Expiration).Err()
	if err != nil {
		global.H_LOG.Error("RedisStoreSetError!", zap.Error(err))
	}
}

func (rs *RedisStore) Get(key string, clear bool) string {
	val, err := global.H_REDIS.Get(context.Background(), key).Result()
	if err != nil {
		global.H_LOG.Error("RedisStoreGetError!", zap.Error(err))
		return "RedisStoreGetError"
	}
	if clear {
		err := global.H_REDIS.Del(context.Background(), key).Err()
		if err != nil {
			global.H_LOG.Error("RedisStoreClearError!", zap.Error(err))
			return ""
		}
	}
	return val
}

func (rs *RedisStore) Scan(key string, limit int64) {
	var cursor uint64
	keys, cursor, err := global.H_REDIS.Scan(context.Background(), cursor, "*", limit).Result()
	if err != nil {
		fmt.Println("scan keys failed err:", err)
	}
	fmt.Println("cursor:", cursor)
	fmt.Println("keys:", keys)
}

func (rs *RedisStore) SetTo(key string, val string) {
	keys := global.H_REDIS.SAdd(context.Background(), key, val)

	fmt.Println("SetTo keys:", keys)
}

func (rs *RedisStore) GetTo(key string) {
	keys := global.H_REDIS.SMembers(context.Background(), key)

	fmt.Println("GetList keys:", keys)
}

func (rs *RedisStore) SetList(key string, val string) {
	keys := global.H_REDIS.RPush(context.Background(), key, val)
	fmt.Println("SetList keys:", keys)
}

func (rs *RedisStore) GetList(key string) {
	val := global.H_REDIS.LRange(context.Background(), key, 0, -1).Val()

	for _, i := range val {
		fmt.Println("GetList keys:", i)
	}

	// fmt.Println("GetList val:", val)
}

func (rs *RedisStore) ListLPop(key string) {
	val := global.H_REDIS.LRange(context.Background(), key, 0, -1).Val()
	for _, i := range val {
		err := global.H_REDIS.LRem(context.Background(), key, 1, i).Val()
		fmt.Println("LPush:", err)
	}

	// fmt.Println("GetList val:", val)
}
