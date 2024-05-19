package action

import (
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"

	"github.com/micastar/file-to-storage-and-share/config"
	"github.com/urfave/cli/v2"
)

func DownloadFile(c *cli.Context) error {
	fileID := c.String("id")
	outputPath := c.String("output")

	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:%s/download/", config.CHI_ADDR, config.CHI_PORT)+fileID, nil)
	if err != nil {
		// return fmt.Errorf("error creating HTTP request: %v", err)
		log.Printf("error creating HTTP request: %v", err)
		return nil
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("error making HTTP request: %v", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("failed to download file, status code: %v", resp.StatusCode)
		return nil
	}

	var out *os.File

	if outputPath == "." || outputPath == "" {

		fileName := resp.Header.Get("Content-Disposition")

		_, params, err := mime.ParseMediaType(fileName)
		if err != nil {
			log.Println(err)
		}
		fileName = params["filename"]
		out, _ = os.Create(fileName)
	} else {
		out, _ = os.Create(outputPath)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Fatal("Error coudn not write data\t", err)
	}

	return nil
}
