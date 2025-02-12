package database

import (
	"context"
	"fmt"
	"os"
	"storegestserver/utils"
	"strconv"

	"github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client

func InitRedis() error {
	Logger = utils.NewLogger()

	RedisClient = redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_HOST"),
		DB: func() int {
			db, err := strconv.Atoi(os.Getenv("REDIS_DB"))
			if err != nil {
				Logger.Error(fmt.Sprint("Invalid REDIS_DB value:", err))
				return 0
			}
			return db
		}(),
	})

	// Ping the Redis server to check if connection is alive
	ctx := context.Background()
	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		Logger.Error(fmt.Sprint("Failed to connect to Redis:", err))
		return err
	}

	Logger.Info(fmt.Sprintf("Connected with Redis: %s", os.Getenv("REDIS_HOST")))
	return nil
}

func CloseRedis() {
	if RedisClient != nil {
		err := RedisClient.Close()
		if err != nil {
			Logger.Error(fmt.Sprint("Error closing Redis connection:", err))
			return
		}
		Logger.Info("Redis connection closed")
	}
}
