package web

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/micastar/file-to-storage-and-share/config"
)

// var webServer *http.Server
func Server(webServer *http.Server) {

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
func Shutdown(webServer *http.Server) {
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
