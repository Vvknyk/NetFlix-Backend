package config

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func ConnectRedis() {
	redisURL := os.Getenv("REDIS_URL")

	RedisClient = redis.NewClient(&redis.Options{
		Addr: redisURL,
	})

	ctx := context.Background()

	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatal("Redis connection failed:", err)
	}

	fmt.Println("Redis connected successfully")
}
