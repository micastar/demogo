package action

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/micastar/file-to-storage-and-share/config"
	"github.com/urfave/cli/v2"
)

func UploadFile(c *cli.Context) error {
	fPath := c.String("file")

	currentDir, _ := os.Getwd()

	file, err := os.Open(filepath.Join(currentDir, fPath))
	if err != nil {
		// return fmt.Errorf("error opening file: %v", err)
		log.Printf("error opening file: %v", err)
		return nil
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(file.Name()))
	if err != nil {
		// return fmt.Errorf("error creating form file: %v", err)
		log.Printf("error creating form file: %v", err)
		return nil
	}
	_, err = io.Copy(part, file)
	if err != nil {
		log.Printf("error copying file data: %v", err)
		return nil
	}
	writer.Close()

	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s:%s/upload", config.CHI_ADDR, config.CHI_PORT), body)
	if err != nil {
		log.Printf("error creating HTTP request: %v", err)
		return nil
	}
	req.Close = true
	req.Header.Set("Content-Type", writer.FormDataContentType())

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
