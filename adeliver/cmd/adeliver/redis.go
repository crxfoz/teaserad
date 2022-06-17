package main

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

func getRedisConns() ([]*redis.Client, error) {
	rdb1 := redis.NewClient(&redis.Options{
		Addr:     "redis-1:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	status := rdb1.Ping(context.Background())
	err := status.Err()
	if err != nil {
		return nil, fmt.Errorf("could not connect to redis-1: %w", err)
	}

	rdb2 := redis.NewClient(&redis.Options{
		Addr:     "redis-2:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	status = rdb2.Ping(context.Background())
	err = status.Err()
	if err != nil {
		return nil, fmt.Errorf("could not connect to redis-1: %w", err)
	}

	rdb3 := redis.NewClient(&redis.Options{
		Addr:     "redis-3:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	status = rdb3.Ping(context.Background())
	err = status.Err()
	if err != nil {
		return nil, fmt.Errorf("could not connect to redis-1: %w", err)
	}

	rdb4 := redis.NewClient(&redis.Options{
		Addr:     "redis-4:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	status = rdb4.Ping(context.Background())
	err = status.Err()
	if err != nil {
		return nil, fmt.Errorf("could not connect to redis-1: %w", err)
	}

	return []*redis.Client{rdb1, rdb2, rdb3, rdb4}, nil
}
