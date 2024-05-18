package db

import (
	"context"
	"log"

	"github.com/micastar/file-to-storage-and-share/config"
	"github.com/redis/go-redis/v9"
)

func RedisClient(dbNO int) *redis.Client {

	rdb := redis.NewClient(&redis.Options{
		Addr:     config.REDIS_HOST,
		Password: config.REDIS_PWD,
		DB:       dbNO,
	})

	_, err := rdb.Ping(context.TODO()).Result()
	if err != nil {
		log.Fatal("Error Connection to Redis: ", err)
	}
	return rdb
}
