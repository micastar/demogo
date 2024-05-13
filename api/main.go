package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/micastar/shorten-url/routes"
)

func setupRoutes(r *gin.Engine) {
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

	setupRoutes(r)

	log.Fatal(r.Run(os.Getenv("APP_PORT")))
}
