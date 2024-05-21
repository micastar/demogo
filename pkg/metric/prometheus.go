package metric

import (
	"github.com/micastar/file-to-storage-and-share/config"
	"github.com/prometheus/client_golang/prometheus"
)

// https://github.com/megaease/easeprobe/blob/7bb241a928c73f6cf363af1533360bcf5d94b9da/metric/prometheus.go#L32
type MetricType interface {
	*prometheus.CounterVec | *prometheus.GaugeVec | *prometheus.HistogramVec | *prometheus.SummaryVec
}

type Metrics struct {
	UpRequestDuration       prometheus.Summary
	DownloadRequestDuration prometheus.Summary
	UpRequestCounter        *prometheus.CounterVec
	DownloadRequestCounter  *prometheus.CounterVec
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		UpRequestDuration: prometheus.NewSummary(prometheus.SummaryOpts{
			Namespace:  config.DefaultProg,
			Name:       "up_request_duration_seconds",
			Help:       "Up duration of the request",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		}),
		DownloadRequestDuration: prometheus.NewSummary(prometheus.SummaryOpts{
			Namespace:  config.DefaultProg,
			Name:       "download_request_duration_seconds",
			Help:       "Download duration of the request",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		}),
		UpRequestCounter: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: config.DefaultProg,
			Name:      "up_request_total",
			Help:      "Up number of the request",
		}, []string{"type"}),
		DownloadRequestCounter: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: config.DefaultProg,
			Name:      "download_request_total",
			Help:      "Download number of the request",
		}, []string{"type"}),
	}

	reg.MustRegister(m.UpRequestDuration, m.DownloadRequestDuration, m.UpRequestCounter, m.DownloadRequestCounter)

	return m
}
