package main

import (
	"net"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	http.Handle("/metrics", promhttp.Handler())

	webServer := &http.Server{}
	server, _ := net.Listen("tcp", ":8833")
	go func() {
		webServer.Serve(server)
	}()

	select {}
}
