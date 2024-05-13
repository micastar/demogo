package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/redis/go-redis/v9"
)

var (
	redisConnMu sync.Mutex
)

func CreateClient(dbNo int) *redis.Client {
	redisConnMu.Lock()
	defer redisConnMu.Unlock()

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
