package redis

import (
	"context"
	"restapi/internal/app/config"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(config *config.RedisConfig) *RedisClient {
	cl := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       0,
	})

	client := &RedisClient{
		client: cl,
	}

	return client
}

func (client *RedisClient) SetValue(key string, value string) error {
	ctx := context.Background()

	err := client.client.Set(ctx, key, value, time.Hour).Err()
	if err != nil {
		return err
	}

	return nil
}

func (client *RedisClient) GetValue(key string) (string, error) {
	ctx := context.Background()

	val, err := client.client.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}

	return val, nil
}
