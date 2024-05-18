package web

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/go-chi/render"
	"github.com/micastar/file-to-storage-and-share/config"
	"github.com/micastar/file-to-storage-and-share/pkg/db"
	"github.com/redis/go-redis/v9"
)

var webServer *http.Server

func Server() {

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RedirectSlashes)
	r.Use(middleware.StripSlashes)
	r.Use(httprate.LimitByIP(8, 5*time.Minute))

	r.Post("/upload", uploadFile)
	r.Get("/download/{fileID}", downloadFile)

	server, err := net.Listen("tcp", fmt.Sprintf("%s:%s", config.CHI_ADDR, config.CHI_PORT))
	if err != nil {
		log.Println("[Web] Failed to start the http server: ", err)
	}
	log.Println("[Web] HTTP server is listening on ", config.CHI_ADDR+":"+config.CHI_PORT)

	// Start the http server
	go func() {
		webServer = &http.Server{Handler: r}
		if err := webServer.Serve(server); !errors.Is(err, http.ErrServerClosed) {
			log.Println("[Web] HTTP server error: ", err)
		}
		log.Println("[Web] HTTP server is stopped.")
	}()
}

// Shutdown the http server
func Shutdown() {
	if webServer == nil {
		log.Println("HTTP server is not running, skip to shutdown")
		return
	}
	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), config.CHI_DefaultTimeOut)
	defer shutdownRelease()

	if err := webServer.Shutdown(shutdownCtx); err != nil {
		log.Println("Failed to shutdown the http server: ", err)
	}
	log.Println("HTTP server is shutdown")
	webServer = nil
}

func uploadFile(w http.ResponseWriter, r *http.Request) {

	var ctx = context.Background()

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // Set max memory to 10MB
	if err != nil {
		log.Println(err)
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}
	// Retrieve the file from form data
	file, handler, err := r.FormFile("file")
	if err != nil {
		log.Println(err)
		http.Error(w, "Unable to retrieve file from form data", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Save file to server
	filename := filepath.Base(handler.Filename)
	uploadPath := filepath.Join("./assets", filename)
	outFile, err := os.Create(uploadPath)
	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to create file", http.StatusInternalServerError)
		return
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, file)
	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to copy file data", http.StatusInternalServerError)
		return
	}

	// Generate a unique file ID
	fileID := fmt.Sprintf("file:%d", time.Now().UnixNano())
	fileMetadata := map[string]interface{}{
		"filename": filename,
		"path":     uploadPath,
	}

	log.Println(fileID, "\t", fileMetadata)

	var rdb = db.RedisClient(config.REDIS_DB_MAIN)
	defer rdb.Close()

	var incr = db.RedisClient(config.REDIS_DB_INCR)
	defer incr.Close()

	// Store file metadata in Redis with TTL
	err = rdb.HSet(ctx, fileID, fileMetadata).Err()
	if err != nil {
		log.Println(err)
		http.Error(w, "Failed save file metadata to Main", http.StatusInternalServerError)
		return
	}
	err = incr.HSet(ctx, fileID, fileMetadata).Err()
	if err != nil {
		log.Println(err)
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

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := fmt.Sprintf(`{"message": "File uploaded successfully", "file_id": "%s"}`, fileID)
	w.Write([]byte(response))

}

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
	log.Println(fileMetadata)
	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to retrieve file metadata", http.StatusInternalServerError)
		return
	}

	// Open the file
	filePath := fileMetadata["path"]
	file, err := os.Open(filePath)
	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to open file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Set the Content-Disposition header to force download
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileMetadata["filename"]))

	// Copy file data to response writer
	_, err = io.Copy(w, file)
	if err != nil {
		log.Println(err)
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
