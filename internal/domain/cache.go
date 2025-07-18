package domain

import "time"

type CacheStore interface {
	Set(key string, value interface{}, expiration time.Duration) error
	Get(key string) (string, error)
	Del(key string) error
	Close() error
}
