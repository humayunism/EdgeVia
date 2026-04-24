package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	RequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "fairflow_requests_total",
			Help: "Total requests handled by the proxy",
		},
		[]string{"tenant_id", "domain", "status"},
	)

	QueueDepth = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "fairflow_queue_depth",
			Help: "Current virtual waiting room queue depth",
		},
		[]string{"tenant_id", "domain"},
	)

	ProxyLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "fairflow_proxy_latency_seconds",
			Help:    "Proxy request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"tenant_id", "domain"},
	)
)

func Init() {
	prometheus.MustRegister(RequestsTotal, QueueDepth, ProxyLatency)
}
