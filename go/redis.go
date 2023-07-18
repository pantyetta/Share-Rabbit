package main

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

var rdb *redis.Client

func init() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})
}

func add(key string, value string) error {
	err := rdb.Set(ctx, key, value, 20*time.Second).Err()
	if err != nil {
		return fmt.Errorf("add failed: %w", err)
	}
	return nil
}

func get(key string) (string, error) {
	val, err := rdb.Get(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("get failed: %w", err)
	}
	return val, nil
}
