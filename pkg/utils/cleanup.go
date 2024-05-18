package utils

import (
	"context"
	"log"
	"os"

	"github.com/micastar/file-to-storage-and-share/config"
	"github.com/micastar/file-to-storage-and-share/pkg/db"
	"github.com/redis/go-redis/v9"
)

func CleanupExpiredFiles(inExpired string) {
	var incr = db.RedisClient(config.REDIS_DB_INCR)
	defer incr.Close()

	var ctx = context.Background()

	// Fetch file path from Redis
	fileMetadata, err := incr.HGetAll(ctx, inExpired).Result()
	if err == redis.Nil || len(fileMetadata) == 0 {
		log.Println("Failed to metadata is Empty:", inExpired, ":", err)
	} else if err != nil {
		log.Println("Failed to fetch metadata for inExpired:", inExpired, ":", err)
	}

	// Delete the file from storage
	filePath := fileMetadata["path"]
	err = os.Remove(filePath)
	if err != nil {
		log.Println("Failed to delete file:", filePath, ":", err)
	}
}
