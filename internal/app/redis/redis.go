package red

import (
	"context"
	"restapi/internal/app/config"

	"github.com/redis/go-redis/v9"
)

type RedClient struct {
	client *redis.Client
}

func NewRedClient(config *config.RedisConfig) *RedClient {
	cl := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       0,
	})

	client := &RedClient{
		client: cl,
	}

	return client
}

func (client *RedClient) SetValue(key string, value string) error {
	ctx := context.Background()

	err := client.client.Set(ctx, key, value, 0).Err()
	if err != nil {
		return err
	}

	return nil
}

func (client *RedClient) GetValue(key string) (string, error) {
	ctx := context.Background()

	val, err := client.client.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}

	return val, nil
}
