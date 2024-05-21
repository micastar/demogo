package action

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/micastar/file-to-storage-and-share/config"
	"github.com/urfave/cli/v2"
)

// req.Header.Set("Content-Type", writer.FormDataContentType())
var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

func UploadFile(c *cli.Context) error {
	fPath := c.String("file")

	currentDir, _ := os.Getwd()

	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)
	f, err := os.Open(filepath.Join(currentDir, fPath))
	if err != nil {
		log.Printf("error opening file: %v", err)
		return nil
	}

	go func() {
		defer f.Close()

		part, err := writer.CreateFormFile("file", filepath.Base(f.Name()))
		if err != nil {
			log.Printf("error creating form file: %v", err)
		}
		var buf = make([]byte, 1024)
		cnt, _ := io.CopyBuffer(part, f, buf)
		log.Printf("copy %d bytes from file %s in total\n", cnt, f.Name())
		writer.Close() //write the tail boundry
		pw.Close()
	}()

	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s:%s/upload", config.CHI_ADDR, config.CHI_PORT), pr)
	if err != nil {
		log.Printf("error creating HTTP request: %v", err)
		return nil
	}
	req.Close = true
	req.Header.Add("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("error making HTTP request: %v", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("failed to upload file, status code: %v\t%v", resp.StatusCode, resp)
		return nil
	}

	var buf = make([]byte, 1024)
	i, _ := resp.Body.Read(buf)
	log.Println(string(buf[:i]))
	return nil
}
