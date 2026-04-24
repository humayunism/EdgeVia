package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

// Handler is the core reverse proxy handler
type Handler struct {
	// TODO: inject Redis client, tenant resolver, rate limiter
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Step 1: Resolve tenant config from domain (SNI / Host header)
	// Step 2: Check rate limit (token bucket)
	// Step 3: If allowed → forward to origin
	// Step 4: If rate limited → virtual waiting room

	// Placeholder: forward to hardcoded origin for now
	origin, _ := url.Parse("http://localhost:3001")
	proxy := httputil.NewSingleHostReverseProxy(origin)

	// Preserve real client IP
	r.Header.Set("X-Forwarded-For", r.RemoteAddr)
	r.Header.Set("X-Forwarded-Host", r.Host)
	r.Header.Set("X-Real-IP", r.RemoteAddr)

	proxy.ServeHTTP(w, r)
}

func ListenTLS(addr string, handler http.Handler) error {
	// TODO: integrate golang.org/x/crypto/acme/autocert
	// For now, plain HTTP fallback
	return http.ListenAndServe(":8081", handler)
}
