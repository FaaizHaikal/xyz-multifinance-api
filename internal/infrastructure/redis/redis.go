package redis

import (
	"context"
	"fmt"
	"log"
	"time"
	"xyz-multifinance-api/config"

	"github.com/redis/go-redis/v9"
)

var Ctx = context.Background()

func InitRedis(cfg *config.Config) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: "",
		DB:       0,
		PoolSize: 10,
	})

	// Ping check
	pong, err := rdb.Ping(Ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}
	log.Printf("Successfully connected to Redis: %s", pong)

	return rdb, nil
}

func Set(key string, value interface{}, expiration time.Duration) error {
	err := RDB.Set(Ctx, key, value, expiration).Err()
	if err != nil {
		log.Printf("Redis Set error for key %s: %v", key, err)
		return fmt.Errorf("redis set failed: %w", err)
	}
	return nil
}

func Get(key string) (string, error) {
	val, err := RDB.Get(Ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("key %s not found in redis", key)
		}
		log.Printf("Redis Get error for key %s: %v", key, err)
		return "", fmt.Errorf("redis get failed: %w", err)
	}
	return val, nil
}

func Del(key string) error {
	err := RDB.Del(Ctx, key).Err()
	if err != nil {
		log.Printf("Redis Del error for key %s: %v", key, err)
		return fmt.Errorf("redis delete failed: %w", err)
	}
	return nil
}

var RDB *redis.Client
