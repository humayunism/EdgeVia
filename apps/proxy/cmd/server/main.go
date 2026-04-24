package main

import (
	"log"
	"net/http"

	"github.com/fairflow/proxy/internal/logger"
	"github.com/fairflow/proxy/internal/metrics"
	"github.com/fairflow/proxy/internal/proxy"
)

func main() {
	logger.Init()
	metrics.Init()

	handler := proxy.NewHandler()

	log.Println("FairFlow Proxy starting on :8080 (HTTP) and :8443 (HTTPS)")
	go func() {
		if err := http.ListenAndServe(":8080", handler); err != nil {
			log.Fatal(err)
		}
	}()

	// HTTPS with Let's Encrypt autocert
	if err := proxy.ListenTLS(":8443", handler); err != nil {
		log.Fatal(err)
	}
}
