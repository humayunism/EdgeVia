package proxy

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"sync"
	"time"

	"github.com/edgevia/proxy/internal/metrics"
	"github.com/edgevia/proxy/internal/queue"
	"github.com/edgevia/proxy/internal/ratelimit"
	"github.com/edgevia/proxy/internal/tenant"
)

// Handler is the core reverse proxy handler
type Handler struct {
	resolver    *tenant.Resolver
	waitingRoom *queue.WaitingRoom
	mu          sync.Mutex
	limiters    map[string]*ratelimit.TokenBucket
	proxies     map[string]*httputil.ReverseProxy
	pageTmpl    *template.Template
}

func NewHandler() *Handler {
	return &Handler{
		resolver:    tenant.NewResolver(),
		waitingRoom: queue.NewWaitingRoom(),
		limiters:    make(map[string]*ratelimit.TokenBucket),
		proxies:     make(map[string]*httputil.ReverseProxy),
		pageTmpl:    template.Must(template.New("waiting-room").Parse(waitingRoomHTML)),
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	host := normalizeHost(r.Host)

	cfg, err := h.resolver.Resolve(host)
	if err != nil {
		http.Error(w, "unable to resolve site configuration", http.StatusBadGateway)
		return
	}
	if !cfg.Active {
		http.Error(w, "site is inactive", http.StatusServiceUnavailable)
		return
	}

	limiter := h.limiterFor(cfg)
	allowed, position, waitTime := limiter.Allow(r.Context(), cfg.TenantID, cfg.Domain)
	if !allowed {
		position, err := h.enqueueVisitor(r.Context(), cfg, r)
		if err != nil {
			http.Error(w, "unable to enqueue visitor", http.StatusServiceUnavailable)
			return
		}
		h.renderWaitingRoom(w, r, cfg, int(position), waitTime)
		metrics.QueueDepth.WithLabelValues(cfg.TenantID, cfg.Domain).Set(float64(h.waitingRoom.Depth(cfg.TenantID, cfg.Domain)))
		metrics.RequestsTotal.WithLabelValues(cfg.TenantID, cfg.Domain, "queued").Inc()
		return
	}

	proxy, err := h.proxyFor(cfg.OriginURL)
	if err != nil {
		http.Error(w, "invalid origin configuration", http.StatusBadGateway)
		return
	}

	proxy.ServeHTTP(w, h.withForwardHeaders(r))
	metrics.RequestsTotal.WithLabelValues(cfg.TenantID, cfg.Domain, "proxied").Inc()
	metrics.ProxyLatency.WithLabelValues(cfg.TenantID, cfg.Domain).Observe(time.Since(start).Seconds())
	_ = position
}

func ListenTLS(addr string, handler http.Handler) error {
	// TODO: integrate golang.org/x/crypto/acme/autocert
	// For now, plain HTTP fallback
	return http.ListenAndServe(addr, handler)
}

func (h *Handler) proxyFor(origin string) (*httputil.ReverseProxy, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if proxy, ok := h.proxies[origin]; ok {
		return proxy, nil
	}

	parsed, err := tenant.ValidateOrigin(origin)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(parsed)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("proxy upstream error: %v", err)
		http.Error(w, "origin unavailable", http.StatusBadGateway)
	}

	h.proxies[origin] = proxy
	return proxy, nil
}

func (h *Handler) withForwardHeaders(r *http.Request) *http.Request {
	req := r.Clone(r.Context())
	if clientIP := clientIPFromRemoteAddr(r.RemoteAddr); clientIP != "" {
		appendForwardedFor(req.Header, clientIP)
		req.Header.Set("X-Real-IP", clientIP)
	}
	req.Header.Set("X-Forwarded-Host", r.Host)
	req.Header.Set("X-Forwarded-Proto", forwardedProto(r))
	return req
}

func (h *Handler) enqueueVisitor(ctx context.Context, cfg *tenant.Config, r *http.Request) (int64, error) {
	visitorID := clientIPFromRemoteAddr(r.RemoteAddr)
	if visitorID == "" {
		visitorID = r.RemoteAddr
	}
	return h.waitingRoom.Enqueue(ctx, cfg.TenantID, cfg.Domain, visitorID)
}

func (h *Handler) limiterFor(cfg *tenant.Config) *ratelimit.TokenBucket {
	key := cfg.TenantID + ":" + cfg.Domain

	h.mu.Lock()
	defer h.mu.Unlock()

	if limiter, ok := h.limiters[key]; ok &&
		limiter.RPS == cfg.RPSLimit &&
		limiter.Burst == cfg.BurstSize {
		return limiter
	}

	limiter := ratelimit.NewTokenBucket(cfg.RPSLimit, cfg.BurstSize)
	h.limiters[key] = limiter
	return limiter
}

func (h *Handler) renderWaitingRoom(w http.ResponseWriter, r *http.Request, cfg *tenant.Config, position int, wait time.Duration) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Retry-After", fmt.Sprintf("%.0f", wait.Seconds()))
	w.WriteHeader(http.StatusTooManyRequests)

	_ = h.pageTmpl.Execute(w, struct {
		Domain   string
		Position int
		Wait     string
	}{
		Domain:   cfg.Domain,
		Position: position,
		Wait:     wait.Round(time.Second).String(),
	})
}

func normalizeHost(host string) string {
	if parsedHost, _, err := net.SplitHostPort(host); err == nil {
		return strings.ToLower(parsedHost)
	}
	return strings.ToLower(host)
}

func clientIPFromRemoteAddr(remoteAddr string) string {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err == nil {
		return host
	}
	return remoteAddr
}

func appendForwardedFor(header http.Header, clientIP string) {
	existing := strings.TrimSpace(header.Get("X-Forwarded-For"))
	if existing == "" {
		header.Set("X-Forwarded-For", clientIP)
		return
	}
	header.Set("X-Forwarded-For", existing+", "+clientIP)
}

func forwardedProto(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	return "http"
}

const waitingRoomHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Waiting Room</title>
  <style>
    body {
      font-family: Arial, sans-serif;
      background: #0f172a;
      color: #e2e8f0;
      margin: 0;
      display: grid;
      place-items: center;
      min-height: 100vh;
      padding: 24px;
    }
    main {
      max-width: 480px;
      background: rgba(15, 23, 42, 0.88);
      border: 1px solid #334155;
      border-radius: 16px;
      padding: 32px;
      box-shadow: 0 20px 45px rgba(15, 23, 42, 0.45);
    }
    h1 { margin-top: 0; }
    p { color: #cbd5e1; line-height: 1.6; }
    .pill {
      display: inline-block;
      padding: 10px 14px;
      border-radius: 999px;
      background: #1d4ed8;
      color: white;
      font-weight: 600;
      margin-top: 8px;
    }
  </style>
</head>
<body>
  <main>
    <h1>Traffic is high right now</h1>
    <p><strong>{{.Domain}}</strong> is temporarily pacing visitors to protect the origin.</p>
    <p class="pill">Queue position: {{.Position}}</p>
    <p>Please retry in about {{.Wait}}.</p>
  </main>
</body>
</html>`
