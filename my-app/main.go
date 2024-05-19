package main

import (
	"encoding/json"
	"net"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Device struct {
	ID       int    `json:"id"`
	Mac      string `json:"mac"`
	Firmware string `json:"firmware"`
}

var dvs []Device
var version string

type metrics struct {
	devices prometheus.Gauge
	info    *prometheus.GaugeVec
}

func NewMetrics(reg prometheus.Registerer) *metrics {
	m := &metrics{
		devices: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "myapp",
			Name:      "connected_devices",
			Help:      "Unmber of currently connected devices",
		}),
		info: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "myapp",
			Name:      "info",
			Help:      "Information about the My App environment",
		}, []string{"version"}),
	}
	reg.MustRegister(m.devices, m.info)

	return m
}

func init() {
	version = "2.10.5"

	dvs = []Device{
		{1, "beaa-d46d6df4bd89", "2,1,6"},
		{2, "b1e5-d46d6df4bd89", "2,1,6"},
	}
}

func main() {

	reg := prometheus.NewRegistry()
	m := NewMetrics(reg)

	m.devices.Set(float64(len(dvs)))
	m.info.With(prometheus.Labels{"version": version}).Set(1)

	promHandler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{EnableOpenMetrics: true})

	http.Handle("/metrics", promHandler)
	http.HandleFunc("/devices", getDevice)
	webServer := &http.Server{}
	server, _ := net.Listen("tcp", ":8833")
	go func() {
		webServer.Serve(server)
	}()

	select {}
}

func getDevice(w http.ResponseWriter, r *http.Request) {
	b, err := json.Marshal(dvs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}
