package main

import (
	"encoding/json"
	"net"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Device struct {
	ID       int    `json:"id"`
	Mac      string `json:"mac"`
	Firmware string `json:"firmware"`
}

var dvs []Device

type metrics struct {
	devices prometheus.Gauge
}

func NewMetrics(reg prometheus.Registerer) *metrics {
	m := &metrics{
		devices: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "myapp",
			Name:      "connected_devices",
			Help:      "Unmber of currently connected devices",
		}),
	}
	reg.MustRegister(m.devices)

	return m
}

func init() {
	dvs = []Device{
		{1, "beaa-d46d6df4bd89", "2,1,6"},
		{2, "b1e5-d46d6df4bd89", "2,1,6"},
	}
}

func main() {

	reg := prometheus.NewRegistry()
	reg.MustRegister(collectors.NewGoCollector())
	m := NewMetrics(reg)

	m.devices.Set(float64(len(dvs)))

	promHandler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{EnableOpenMetrics: true, Registry: reg})

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
