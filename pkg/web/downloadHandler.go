package web

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/micastar/file-to-storage-and-share/config"
	"github.com/micastar/file-to-storage-and-share/pkg/db"
	"github.com/micastar/file-to-storage-and-share/pkg/metric"
	"github.com/micastar/file-to-storage-and-share/pkg/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
)

type downloadHandler struct {
	metric *metric.Metrics
}

func (dh *downloadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	downloadFile(w, r, dh.metric)
}

func downloadFile(w http.ResponseWriter, r *http.Request, m *metric.Metrics) {

	now := time.Now().UTC()

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
	file := openFile(filePath, w)
	defer file.Close()

	// Set the Content-Disposition header to only download
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileMetadata["filename"]))

	utils.Copy2Dst(w, w, file)

	m.DownloadRequestCounter.With(prometheus.Labels{"type": "download"}).Inc()
	m.DownloadRequestDuration.Observe(float64(time.Since(now).Seconds()))
}

func openFile(filePath string, w http.ResponseWriter) *os.File {
	file, err := os.Open(filePath)
	if err != nil {
		log.Println("file error: ", err)
		http.Error(w, "Failed to open file", http.StatusInternalServerError)
		return nil
	}
	return file
}

func getFileMetadataFromRedis(fileID string, ctx context.Context, rdb *redis.Client) (map[string]string, error) {
	// Retrieve file metadata from Redis based on file ID
	fileMetadata, err := rdb.HGetAll(ctx, fileID).Result()
	if err != nil {
		return nil, err
	}

	return fileMetadata, nil
}
