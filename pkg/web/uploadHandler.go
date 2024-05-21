package web

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/micastar/file-to-storage-and-share/config"
	"github.com/micastar/file-to-storage-and-share/pkg/db"
	"github.com/micastar/file-to-storage-and-share/pkg/metric"
	"github.com/micastar/file-to-storage-and-share/pkg/utils"
	"github.com/prometheus/client_golang/prometheus"
)

type uploadHandler struct {
	metric *metric.Metrics
}

func (uh *uploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		uploadFile(w, r, uh.metric)
	default:
		w.Header().Set("Allow", "POST")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

	}

}

func uploadFile(w http.ResponseWriter, r *http.Request, m *metric.Metrics) {

	now := time.Now().UTC()

	parseData(w, r)
	filename, uploadPath := retrieveData(w, r)
	fileID, fileMetadata := generateFileID(filename, uploadPath)

	go setFileWithRedis(fileID, fileMetadata, w)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := fmt.Sprintf(`{"message": "File uploaded successfully", "file_id": "%s"}`, fileID)
	w.Write([]byte(response))

	m.UpRequestCounter.With(prometheus.Labels{"type": "upload"}).Inc()
	m.UpRequestDuration.Observe(time.Since(now).Seconds())
}

func setFileWithRedis(fileID string, fileMetadata map[string]any, w http.ResponseWriter) {

	var ctx = context.Background()

	log.Println("setFileWithRedis: ", fileID, "\t", fileMetadata)

	var rdb = db.RedisClient(config.REDIS_DB_MAIN)
	defer rdb.Close()

	var incr = db.RedisClient(config.REDIS_DB_INCR)
	defer incr.Close()

	// Store file metadata in Redis with TTL
	err := rdb.HSet(ctx, fileID, fileMetadata).Err()
	if err != nil {
		log.Println("Redis Main DB: ", err)
		http.Error(w, "Failed save file metadata to Main", http.StatusInternalServerError)
		return
	}
	err = incr.HSet(ctx, fileID, fileMetadata).Err()
	if err != nil {
		log.Println("Redis Incr DB: ", err)
		http.Error(w, "Failed save file metadata to Incr", http.StatusInternalServerError)
		return
	}

	err = rdb.Expire(ctx, fileID, config.REDIS_TTL_MAIN).Err()
	if err != nil {
		log.Println("Main: ", err)
		http.Error(w, "Failed to set TTL for file metadata", http.StatusInternalServerError)
		return
	}
	err = incr.Expire(ctx, fileID, config.REDIS_TTL_INCR).Err()
	if err != nil {
		log.Println("Incr: ", err)
		http.Error(w, "Failed set TTL for file metadata to Incr", http.StatusInternalServerError)
		return
	}
}

func parseData(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	// https://freshman.tech/file-upload-golang/
	err := r.ParseMultipartForm(10 << 20) // Set max memory to 10MB
	if err != nil {
		log.Println("parseData: ", err)
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}
}

func retrieveData(w http.ResponseWriter, r *http.Request) (string, string) {

	file, handler, err := r.FormFile("file")
	if err != nil {
		log.Println("retrieveData: ", err)
		http.Error(w, "Unable to retrieve file from form data", http.StatusBadRequest)
		return "", ""
	}
	defer file.Close()
	filename := filepath.Base(handler.Filename)

	outFile, uploadPath := saveFile(w, filename)
	defer outFile.Close()

	utils.Copy2Dst(w, outFile, file)

	return filename, uploadPath
}

const uploadDir = "./assets"

func saveFile(w http.ResponseWriter, filename string) (*os.File, string) {

	os.MkdirAll(uploadDir, 0o755)
	filePath := filepath.Join(uploadDir, filename)

	// Create the file on disk
	destFile, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Unable to create file on disk", http.StatusInternalServerError)
		return nil, ""
	}

	return destFile, filePath
}

func generateFileID(filename string, uploadPath string) (string, map[string]any) {
	fileID := fmt.Sprintf("file:%d", time.Now().UnixNano())
	fileMetadata := map[string]any{
		"filename": filename,
		"path":     uploadPath,
	}
	return fileID, fileMetadata
}
