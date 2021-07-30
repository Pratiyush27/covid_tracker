package cache

import (
	"encoding/json"
	"context"
	"github.com/go-redis/redis/v8"
	"time"
)

var ctx = context.Background()



type CacheItf interface {
	Set(key string, data interface{}, expiration time.Duration) error
	Get(key string) ([]byte, error)
	FlushDB() error
}


func InitRedisCache() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		Password: "",
		DB: 0,
	})

	return client
}

func Set(client *redis.Client,key string, data interface{}, expiration time.Duration) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return client.Set(ctx,key, b, expiration).Err()
}

func Get(client *redis.Client,key string) ([]byte, error) {
	result, err := client.Get(ctx,key).Bytes()
	if err == redis.Nil {
		return nil,nil
	}

	return result, err
}
func FlushDB(client *redis.Client) error {
	err := client.FlushDB(ctx)
	if err != nil {
		return err.Err()
	}

	return err.Err()
}
