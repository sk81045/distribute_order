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

func (rs *RedisStore) SetList(key string, val string) (ok bool) {
	dd := global.H_REDIS.LPush(context.Background(), key, val).Val()
	if dd > 0 {
		return true
	} else {
		return false
	}
}

func (rs *RedisStore) LRange(key string, start int64, end int64) (val []string) {
	return global.H_REDIS.LRange(context.Background(), key, start, end).Val()

	// for _, i := range val {
	// 	fmt.Println("GetList keys:", i)
	// }

}

func (rs *RedisStore) BRPopLPush(key1 string, key2 string, timeout time.Duration) (val string) {
	return global.H_REDIS.BRPopLPush(context.Background(), key1, key2, timeout).Val()
}

func (rs *RedisStore) LLen(key string) (le int64) {
	return global.H_REDIS.LLen(context.Background(), key).Val()
}

func (rs *RedisStore) LRpop(key string) (val string) {
	return global.H_REDIS.RPop(context.Background(), key).Val()

	// for _, i := range val {
	// 	fmt.Println("rpop-->:", i)
	// }

	// fmt.Println("LPop-->", val)
}

func (rs *RedisStore) ListLPop(key string) {
	val := global.H_REDIS.LRange(context.Background(), key, 0, -1).Val()
	for _, i := range val {
		err := global.H_REDIS.LRem(context.Background(), key, 1, i).Val()
		fmt.Println("LPush:", err)
	}

	// fmt.Println("GetList val:", val)
}
