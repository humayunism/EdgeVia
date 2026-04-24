package tenant

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
}

// Resolve looks up config for a domain.
// Cache-first: Redis → gRPC → PostgreSQL
func (r *Resolver) Resolve(domain string) (*Config, error) {
	// TODO: HGETALL tenant:*:config:{domain}
	// Fallback: gRPC call to NestJS API
	return &Config{
		TenantID:  "demo",
		Domain:    domain,
		OriginURL: "http://localhost:3001",
		RPSLimit:  100,
		BurstSize: 200,
		Active:    true,
	}, nil
}
