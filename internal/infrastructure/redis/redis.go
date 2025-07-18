package redis

import (
	"context"
	"fmt"
	"log"
	"time"
	"xyz-multifinance-api/config"
	"xyz-multifinance-api/internal/domain"

	"github.com/redis/go-redis/v9"
)

var Ctx = context.Background()

type RedisCacheStore struct {
	client *redis.Client
}

func NewRedisCacheStore(client *redis.Client) domain.CacheStore {
	return &RedisCacheStore{client: client}
}

func InitRedisClient(cfg *config.Config) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: "",
		DB:       0,
		PoolSize: 10,
	})

	pong, err := rdb.Ping(Ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}
	log.Printf("Successfully connected to Redis: %s", pong)

	return rdb, nil
}

func (r *RedisCacheStore) Set(key string, value interface{}, expiration time.Duration) error {
	err := r.client.Set(Ctx, key, value, expiration).Err()
	if err != nil {
		log.Printf("Redis Set error for key %s: %v", key, err)
		return fmt.Errorf("redis set failed: %w", err)
	}
	return nil
}

func (r *RedisCacheStore) Get(key string) (string, error) {
	val, err := r.client.Get(Ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("key %s not found in redis", key)
		}
		log.Printf("Redis Get error for key %s: %v", key, err)
		return "", fmt.Errorf("redis get failed: %w", err)
	}
	return val, nil
}

func (r *RedisCacheStore) Del(key string) error {
	err := r.client.Del(Ctx, key).Err()
	if err != nil {
		log.Printf("Redis Del error for key %s: %v", key, err)
		return fmt.Errorf("redis delete failed: %w", err)
	}
	return nil
}

func (r *RedisCacheStore) Close() error {
	return r.client.Close()
}
