package utils

import (
	"io"
	"log"
	"net/http"
)

func Copy2Dst(w http.ResponseWriter, outFile io.Writer, inFile io.Reader) {
	_, err := io.Copy(outFile, inFile)
	if err != nil {
		log.Println("copy2Dst: ", err)
		http.Error(w, "Failed to copy file data", http.StatusInternalServerError)
		return
	}
}
