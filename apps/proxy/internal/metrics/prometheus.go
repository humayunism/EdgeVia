package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	RequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "edgevia_requests_total",
			Help: "Total requests handled by the Edgevia proxy",
		},
		[]string{"tenant_id", "domain", "status"},
	)

	QueueDepth = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "edgevia_queue_depth",
			Help: "Current Edgevia virtual waiting room queue depth",
		},
		[]string{"tenant_id", "domain"},
	)

	ProxyLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "edgevia_proxy_latency_seconds",
			Help:    "Edgevia proxy request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"tenant_id", "domain"},
	)
)

func Init() {
	prometheus.MustRegister(RequestsTotal, QueueDepth, ProxyLatency)
}
