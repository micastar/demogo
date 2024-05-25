package database

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

func CreateClient(dbNo int) *redis.Client {

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("DB_HOST"), os.Getenv("DB_PORT")),
		Password: "",
		DB:       dbNo,
	})

	_, err := rdb.Ping(context.TODO()).Result()
	if err != nil {
		log.Fatal("Error Connection to Redis: ", err)
	}

	return rdb
}
