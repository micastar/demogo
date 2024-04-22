package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/micastar/shorten-url/routes"
)

func setuprRoutes(r *gin.Engine) {
	r.GET("/:url", routes.ResolveURL)
	r.POST("/api/v1", routes.ShortenURL)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println(err)
	}

	r := gin.New()

	r.Use(gin.Logger())

	setuprRoutes(r)

	log.Fatal(r.Run(os.Getenv("APP_ADDR")))
}
