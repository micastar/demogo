package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

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
	devices  prometheus.Gauge
	info     *prometheus.GaugeVec
	upgrades *prometheus.CounterVec
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
		upgrades: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "myapp",
			Name:      "device_upgrade_total",
			Help:      "Number of upgraded devices",
		}, []string{"type"}),
	}
	reg.MustRegister(m.devices, m.info, m.upgrades)

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

	dMux := http.ServeMux{}
	rdh := registerDevicesHandler{metrics: m}
	mdh := managerDevicesHandler{metrics: m}
	dMux.Handle("/devices", rdh)
	dMux.Handle("/devices/", mdh)
	// dMux.HandleFunc("/devices", rdh.ServeHTTP)

	pMux := http.ServeMux{}
	promHandler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{EnableOpenMetrics: true})
	pMux.Handle("/metrics", promHandler)

	go func() {
		log.Fatal(http.ListenAndServe(":8833", &dMux))
	}()

	go func() {
		log.Fatal(http.ListenAndServe(":8844", &pMux))
	}()

	select {}
}

type registerDevicesHandler struct {
	metrics *metrics
}

func (rdh registerDevicesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getDevice(w, r)
	case "POST":
		createDevice(w, r, rdh.metrics)
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getDevice(w http.ResponseWriter, _ *http.Request) {
	b, err := json.Marshal(dvs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func createDevice(w http.ResponseWriter, r *http.Request, m *metrics) {
	var dv Device

	err := json.NewDecoder(r.Body).Decode(&dv)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// m.devices.Inc()

	dvs = append(dvs, dv)

	// for performance
	m.devices.Set(float64(len(dvs)))

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Device Created"))
}

func upgradeDevice(w http.ResponseWriter, r *http.Request, m *metrics) {
	path := strings.TrimPrefix(r.URL.Path, "/devices/")

	id, err := strconv.Atoi(path)
	if err != nil || id < 1 {
		http.NotFound(w, r)
	}

	var dv Device

	err = json.NewDecoder(r.Body).Decode(&dv)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for i := range dvs {
		if dvs[i].ID == id {
			dvs[i].Firmware = dv.Firmware
		}
	}

	sleep(1000)

	m.upgrades.With(prometheus.Labels{"type": "router"}).Inc()
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Upgrading...."))

}

type managerDevicesHandler struct {
	metrics *metrics
}

func (mdh managerDevicesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "PUT":
		upgradeDevice(w, r, mdh.metrics)
	default:
		w.Header().Set("Allow", "PUT")
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func sleep(ms int) {
	rand.NewSource(time.Now().UnixNano())
	now := time.Now()

	n := rand.Intn(ms + now.Second())
	time.Sleep(time.Duration(n) * time.Millisecond)
}
