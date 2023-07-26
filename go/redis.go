package main

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

var rdb *redis.Client

func InitRedis() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})
}

func Add(key string, value string) error {
	err := rdb.Set(ctx, key, value, 20*time.Second).Err()
	if err != nil {
		return fmt.Errorf("add failed: %w", err)
	}
	return nil
}

func Get(key string) (string, error) {
	val, err := rdb.Get(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("get failed: %w", err)
	}
	return val, nil
}

func Keys(pattern string) ([]string, error) {
	val, err := rdb.Keys(ctx, pattern).Result()

	if err != nil {
		return nil, fmt.Errorf("Keys failed: %w", err)
	}
	return val, nil
}
