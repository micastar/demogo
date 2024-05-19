package web

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/micastar/file-to-storage-and-share/config"
	"github.com/micastar/file-to-storage-and-share/pkg/db"
)

type uploadHandler struct{}

func newUploadHandler() uploadHandler {
	return uploadHandler{}
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	var u = newUploadHandler()

	u.parseData(w, r)
	filename, uploadPath := u.retrieveData(w, r)
	fileID, fileMetadata := u.generateFileID(filename, uploadPath)

	go u.setFileWithRedis(fileID, fileMetadata, w)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := fmt.Sprintf(`{"message": "File uploaded successfully", "file_id": "%s"}`, fileID)
	w.Write([]byte(response))

}

func (u uploadHandler) setFileWithRedis(fileID string, fileMetadata map[string]any, w http.ResponseWriter) {

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

func (u uploadHandler) parseData(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	// https://freshman.tech/file-upload-golang/
	err := r.ParseMultipartForm(10 << 20) // Set max memory to 10MB
	if err != nil {
		log.Println("parseData: ", err)
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}
}

func (u uploadHandler) retrieveData(w http.ResponseWriter, r *http.Request) (string, string) {

	file, handler, err := r.FormFile("file")
	if err != nil {
		log.Println("retrieveData: ", err)
		http.Error(w, "Unable to retrieve file from form data", http.StatusBadRequest)
		return "", ""
	}
	defer file.Close()
	filename := filepath.Base(handler.Filename)

	outFile, uploadPath := u.saveFile(w, filename)
	defer outFile.Close()

	u.copy2Dst(w, outFile, file)

	return filename, uploadPath
}

func (u uploadHandler) saveFile(w http.ResponseWriter, filename string) (*os.File, string) {
	uploadPath := filepath.Join("./assets", filename)
	outFile, err := os.Create(strings.TrimSpace(uploadPath))
	if err != nil {
		log.Println("saveFile: ", err)
		http.Error(w, "Failed to create file", http.StatusInternalServerError)
		return nil, ""
	}

	return outFile, uploadPath
}

func (u uploadHandler) copy2Dst(w http.ResponseWriter, outFile *os.File, file multipart.File) {
	_, err := io.Copy(outFile, file)
	if err != nil {
		log.Println("copy2Dst: ", err)
		http.Error(w, "Failed to copy file data", http.StatusInternalServerError)
		return
	}
}

func (u uploadHandler) generateFileID(filename string, uploadPath string) (string, map[string]any) {
	fileID := fmt.Sprintf("file:%d", time.Now().UnixNano())
	fileMetadata := map[string]any{
		"filename": filename,
		"path":     uploadPath,
	}
	return fileID, fileMetadata
}
