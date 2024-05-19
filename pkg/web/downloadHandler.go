package web

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/micastar/file-to-storage-and-share/config"
	"github.com/micastar/file-to-storage-and-share/pkg/db"
	"github.com/redis/go-redis/v9"
)

func downloadFile(w http.ResponseWriter, r *http.Request) {
	var ctx = context.Background()
	var rdb = db.RedisClient(config.REDIS_DB_MAIN)
	defer rdb.Close()

	// Extract file ID from URL query parameter
	fileID := chi.URLParam(r, "fileID")
	if fileID == "" {
		render.Respond(w, r, "Type a key")
		return
	}

	// Retrieve file metadata from Redis based on file ID
	fileMetadata, err := getFileMetadataFromRedis(fileID, ctx, rdb)
	log.Println("fileMetadata: ", fileMetadata)
	if err != nil {
		log.Println("fileMetadata err: ", err)
		http.Error(w, "Failed to retrieve file metadata", http.StatusInternalServerError)
		return
	}

	// Open the file
	filePath := fileMetadata["path"]
	file, err := os.Open(filePath)
	if err != nil {
		log.Println("file error: ", err)
		http.Error(w, "Failed to open file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Set the Content-Disposition header to only download
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileMetadata["filename"]))

	// Copy file data to response writer
	_, err = io.Copy(w, file)

	if err != nil {
		log.Println("copy error: ", err)
		http.Error(w, "Failed to write file data to response", http.StatusInternalServerError)
		return
	}

}
func getFileMetadataFromRedis(fileID string, ctx context.Context, rdb *redis.Client) (map[string]string, error) {
	// Retrieve file metadata from Redis based on file ID
	fileMetadata, err := rdb.HGetAll(ctx, fileID).Result()
	if err != nil {
		return nil, err
	}

	return fileMetadata, nil
}
