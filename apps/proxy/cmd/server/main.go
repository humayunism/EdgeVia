package main

import (
	"log"
	"net/http"
	"os"

	"github.com/edgevia/proxy/internal/logger"
	"github.com/edgevia/proxy/internal/metrics"
	"github.com/edgevia/proxy/internal/proxy"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	logger.Init()
	metrics.Init()

	handler := proxy.NewHandler()
	metricsPort := getenv("METRICS_PORT", "9090")
	httpAddr := getenv("HTTP_ADDR", ":8080")
	httpsAddr := getenv("HTTPS_ADDR", ":8443")

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		log.Printf("metrics listening on :%s", metricsPort)
		if err := http.ListenAndServe(":"+metricsPort, mux); err != nil {
			log.Fatal(err)
		}
	}()

	log.Printf("Edgevia Proxy starting on %s (HTTP) and %s (HTTPS fallback)", httpAddr, httpsAddr)
	go func() {
		if err := http.ListenAndServe(httpAddr, handler); err != nil {
			log.Fatal(err)
		}
	}()

	// HTTPS with Let's Encrypt autocert
	if err := proxy.ListenTLS(httpsAddr, handler); err != nil {
		log.Fatal(err)
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
