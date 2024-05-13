package main

import (
	"html/template"
	"io"
	"log"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/micastar/shorten-url/helpers"
	"github.com/micastar/shorten-url/routes"
)

func setupRoutes(r *gin.Engine) {
	r.GET("/", routes.Homepage)
	r.GET("/:url", routes.ResolveURL)
	r.POST("/api", routes.ShortenURL)
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	r := gin.Default()

	t, err := loadTemplate()
	if err != nil {
		panic(err)
	}
	r.SetHTMLTemplate(t)

	setupRoutes(r)

	log.Fatal(r.Run(os.Getenv("APP_PORT")))
}

func loadTemplate() (*template.Template, error) {
	t := template.New("")
	for name, file := range helpers.Assets.Files {
		if file.IsDir() || !strings.HasSuffix(name, ".tmpl") {
			continue
		}
		h, err := io.ReadAll(file)
		if err != nil {
			return nil, err
		}
		t, err = t.New(name).Parse(string(h))
		if err != nil {
			return nil, err
		}
	}
	return t, nil
}
