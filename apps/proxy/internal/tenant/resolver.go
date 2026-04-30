package tenant

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// Config holds the per-tenant, per-site configuration.
// Cached in Redis, source of truth in PostgreSQL.
type Config struct {
	TenantID  string
	Domain    string
	OriginURL string
	RPSLimit  int
	BurstSize int
	Active    bool
}

// Resolver resolves tenant configuration from a domain name.
type Resolver struct {
	// redisClient redis.Client
	// grpcClient  apiclient.Client
	defaultOrigin string
	defaultRPS    int
	defaultBurst  int
	overrides     map[string]Config
}

func NewResolver() *Resolver {
	return &Resolver{
		defaultOrigin: getenv("EDGEVIA_DEFAULT_ORIGIN", "http://localhost:3001"),
		defaultRPS:    getenvInt("EDGEVIA_DEFAULT_RPS", 100),
		defaultBurst:  getenvInt("EDGEVIA_DEFAULT_BURST", 200),
		overrides:     parseOverrides(os.Getenv("EDGEVIA_SITE_CONFIGS")),
	}
}

// Resolve looks up config for a domain.
// Cache-first: Redis → gRPC → PostgreSQL
func (r *Resolver) Resolve(domain string) (*Config, error) {
	host := normalizeDomain(domain)
	if host == "" {
		return nil, fmt.Errorf("missing domain")
	}

	if cfg, ok := r.overrides[host]; ok {
		copy := cfg
		copy.Domain = host
		if copy.RPSLimit <= 0 {
			copy.RPSLimit = r.defaultRPS
		}
		if copy.BurstSize <= 0 {
			copy.BurstSize = r.defaultBurst
		}
		if copy.OriginURL == "" {
			copy.OriginURL = r.defaultOrigin
		}
		return &copy, nil
	}

	// TODO: HGETALL tenant:*:config:{domain}
	// Fallback: gRPC call to NestJS API
	return &Config{
		TenantID:  "demo",
		Domain:    host,
		OriginURL: r.defaultOrigin,
		RPSLimit:  r.defaultRPS,
		BurstSize: r.defaultBurst,
		Active:    true,
	}, nil
}

func parseOverrides(raw string) map[string]Config {
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	configs := make(map[string]Config)
	for _, entry := range strings.Split(raw, ";") {
		parts := strings.Split(entry, "=")
		if len(parts) != 2 {
			continue
		}

		host := normalizeDomain(parts[0])
		if host == "" {
			continue
		}

		fields := strings.Split(parts[1], ",")
		cfg := Config{
			TenantID:  "demo",
			Domain:    host,
			OriginURL: strings.TrimSpace(fields[0]),
			Active:    true,
		}

		if len(fields) > 1 {
			cfg.RPSLimit = atoiDefault(fields[1], 0)
		}
		if len(fields) > 2 {
			cfg.BurstSize = atoiDefault(fields[2], 0)
		}

		configs[host] = cfg
	}

	return configs
}

func normalizeDomain(domain string) string {
	host := strings.TrimSpace(domain)
	if host == "" {
		return ""
	}

	if parsedHost, _, err := net.SplitHostPort(host); err == nil {
		host = parsedHost
	}

	return strings.ToLower(host)
}

func getenv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	return atoiDefault(os.Getenv(key), fallback)
}

func atoiDefault(raw string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func ValidateOrigin(raw string) (*url.URL, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("origin must include scheme and host")
	}
	return parsed, nil
}
