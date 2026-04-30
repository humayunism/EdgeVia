# Edgevia Architecture

**Version:** 1.0.0  
**Source of truth:** `docs/edgevia-srs.pdf`  
**Status:** Revised architecture blueprint

## 1. Overview

Edgevia is a multi-tenant reverse proxy and traffic management platform built to protect customer applications during sudden traffic spikes. It sits between the public internet and a customer's origin server, enforcing rate limits, activating a virtual waiting room when demand exceeds capacity, and preserving service availability during high-pressure events such as flash sales, ticket launches, and viral traffic surges.

The platform is designed around three goals:

1. Protect origin infrastructure from overload.
2. Provide fair user access during peak demand.
3. Offer enterprise-style traffic controls at startup-friendly cost.

## 2. Architecture Summary

Edgevia uses a polyglot microservices architecture deployed to Kubernetes in production and Docker Compose for local development.

### Core architectural characteristics

- Go handles the latency-sensitive proxy path.
- NestJS owns business logic, tenant management, billing, and analytics APIs.
- Next.js provides the operator dashboard and customer-facing management UI.
- Redis supports rate limiting, queue state, caching, and pub/sub.
- PostgreSQL stores durable tenant, site, billing, and analytics data.
- gRPC is used for internal service-to-service communication.
- Socket.io pushes real-time operational metrics to the dashboard.

## 3. High-Level System Components

| Component | Primary Responsibility | Technology |
| --- | --- | --- |
| Proxy Engine | Reverse proxying, rate limiting, queueing, circuit breaking, SSL termination | Go |
| Management API | Auth, site management, analytics, billing, events, notifications | Node.js + NestJS |
| Dashboard | Real-time control plane and observability UI | Next.js 14 |
| Cache and Queue | Token buckets, waiting room queues, cached site config, pub/sub | Redis 7 |
| System of Record | Tenants, sites, events, certificates, usage, analytics | PostgreSQL + Prisma |
| Async Workers | Metering, notifications, aggregation, billing workflows | BullMQ |
| Observability | Metrics, dashboards, logs, traces | Prometheus, Grafana, Zap, Pino, OpenTelemetry |

## 4. Request Flow

### Normal request path

1. A client request reaches the public load balancer on port `80` or `443`.
2. Traffic is forwarded to the Go Proxy service.
3. The proxy resolves the tenant and site configuration for the requested domain.
4. Redis is checked for token bucket state and queue status.
5. If the request is within threshold, the proxy forwards it to the customer's origin server.
6. Metrics are emitted to Prometheus and reported upstream to the API layer.

### Overload path

1. The proxy detects that the token bucket is exhausted or queue depth exceeds the configured threshold.
2. The request is placed into a tenant-isolated FIFO queue in Redis.
3. The user receives a branded waiting room page.
4. Queue position updates are delivered via WebSocket or SSE.
5. When capacity becomes available, the request is released fairly back into the allowed path.

### Failure protection path

1. The proxy monitors origin error rates.
2. If the error rate exceeds 50% within a 10-second window, the circuit breaker opens.
3. Requests are blocked or queued while the origin stabilizes.
4. Recovery is tested in a half-open state with limited traffic.
5. Customers are notified through webhook and email channels.

## 5. Core Service Design

### 5.1 Proxy Engine

The Go proxy is the system’s performance-critical edge component. It is responsible for:

- TLS termination and automatic certificate provisioning
- Reverse proxying to customer origins
- Tenant-aware request routing
- Redis-backed token bucket rate limiting
- Virtual waiting room enforcement
- Circuit breaker evaluation
- Prometheus metrics emission

The design choice is intentional: Go’s concurrency model supports high connection counts and low request overhead, making it suitable for the hot path.

### 5.2 Management API

The NestJS API acts as the platform control plane. It owns:

- JWT-based authentication
- tenant and site lifecycle management
- dashboard data aggregation
- billing and Stripe integration
- event mode configuration
- notification delivery
- configuration sync for proxy consumers

This service is optimized for maintainability, modularity, and typed business workflows rather than raw edge-path performance.

### 5.3 Dashboard

The Next.js dashboard is the operational nerve center for customers and platform operators. It provides:

- live visitors, queue depth, RPS, latency, and health views
- multi-site switching
- inline RPS updates
- waiting room customization
- event mode controls
- billing and usage visibility
- alert and webhook configuration

## 6. Technology Stack

| Layer | Technology | Why it is used |
| --- | --- | --- |
| Proxy Engine | Go | High concurrency, low latency, efficient network I/O |
| API Layer | Node.js + NestJS | Modular business logic, typed architecture, strong ecosystem |
| Frontend | Next.js 14 App Router | SSR for first load, responsive operator UI, real-time patterns |
| Database | PostgreSQL + Prisma | Durable relational storage with strong consistency |
| Cache / Queue | Redis 7 | Fast token bucket operations, queue state, pub/sub |
| Job Queue | BullMQ | Background jobs for metering, email, and aggregation |
| Internal RPC | gRPC + Protocol Buffers | Efficient binary contracts between services |
| Real-Time Updates | Socket.io | Dashboard live updates |
| SSL Automation | Let's Encrypt + Go `autocert` | Automatic certificate issuance and renewal |
| Metrics | Prometheus + Grafana | Operational monitoring and alerting |
| Logging | Zap and Pino | Structured low-overhead application logs |
| Tracing | OpenTelemetry | Distributed tracing across services |
| Object Storage | AWS S3 or Cloudflare R2 | Waiting room assets and exports |
| Billing | Stripe | Subscription and usage-based billing |
| Email | Resend | Transactional messaging |
| DNS Automation | Cloudflare API | Programmatic customer domain setup |

## 7. Feature Architecture

### 7.1 Zero-Configuration Onboarding

Customer onboarding is designed to be fast and low-friction:

1. The customer creates a CNAME record pointing their domain to Edgevia.
2. The proxy detects the incoming domain through SNI.
3. Edgevia provisions an SSL certificate automatically.
4. Origin configuration is stored in PostgreSQL and cached in Redis.
5. The site becomes traffic-ready, with the SRS targeting completion within roughly 60 seconds.

### 7.2 Rate Limiting

Rate limiting uses a tenant-isolated token bucket model stored in Redis.

Key pattern:

```text
tenant:{tenantId}:bucket:{domain}
```

Behavior:

- buckets refill at the configured requests-per-second rate
- bursts are allowed for short windows, such as `2x RPS for 5 seconds`
- when the bucket empties, requests are queued rather than dropped

### 7.3 Virtual Waiting Room

The waiting room is triggered when incoming traffic exceeds configured capacity. It is designed to preserve fairness and protect origin stability.

Capabilities include:

- tenant-branded queue page
- FIFO admission order
- real-time queue position updates
- bot and crawler bypass for verified search engines
- preserved client IP forwarding through `X-Forwarded-For`

### 7.4 Circuit Breaker

The circuit breaker protects unhealthy origins from cascading failure.

States:

```text
CLOSED -> OPEN -> HALF-OPEN
```

Open-state criteria from the SRS:

- origin error rate greater than `50%`
- measured over a `10 second` window

Recovery behavior:

- limited test traffic in half-open mode
- automatic resumption if the origin is healthy again
- webhook and email notification on open events

## 8. Multi-Tenant Isolation Model

Tenant isolation is enforced at several layers:

- Redis key namespacing with `tenant:{tenantId}:...`
- PostgreSQL Row-Level Security for persistent data boundaries
- Kubernetes `NetworkPolicy` isolation between service contexts
- audit logging for tenant configuration changes

This layered model prevents cross-tenant leakage across cache, queue, storage, and network boundaries.

## 9. Real-Time Dashboard Metrics

| Metric | Presentation | Source | Update Rate |
| --- | --- | --- | --- |
| Live visitors | Animated counter | Redis pub/sub | 1s |
| Queue depth | Counter and chart | Redis queue length | 500ms |
| System health | Pulsing status icon | Circuit breaker state | 2s |
| Protected revenue | Currency counter | PostgreSQL + Stripe | 5s |
| Current RPS | Gauge / speedometer | Prometheus | 1s |
| Origin response time | Line chart | Go proxy latency | 1s |

## 10. Billing Architecture

Billing is subscription-based and integrated with Stripe.

Supported billing workflows:

- subscription lifecycle management
- hourly usage metering via BullMQ workers
- overage alerts at `80%` and `95%` of plan limits
- self-service plan upgrades and downgrades
- downloadable invoice history
- event add-on purchase flow

## 11. Repository Structure

The SRS describes a monorepo organized around three applications plus shared packages. The current repository aligns with that structure:

```text
EdgeVia/
├── apps/
│   ├── proxy/        # Go reverse proxy and edge runtime
│   ├── api/          # NestJS management API
│   └── dashboard/    # Next.js dashboard
├── packages/
│   ├── proto/        # Shared protobuf definitions
│   └── types/        # Shared application types
├── infra/            # Kubernetes and infrastructure assets
├── docs/             # Architecture, API, onboarding, SRS
├── scripts/          # Setup and seed automation
├── pnpm-workspace.yaml
└── turbo.json
```

### Intended module boundaries

- `apps/proxy/internal/proxy`: reverse proxy transport and handlers
- `apps/proxy/internal/ratelimit`: Redis token bucket logic
- `apps/proxy/internal/queue`: waiting room queue and position handling
- `apps/proxy/internal/circuitbreaker`: origin health protection
- `apps/proxy/internal/ssl`: certificate automation
- `apps/api/src/modules`: business modules such as auth, sites, analytics, billing, tenants, notifications, and events
- `apps/dashboard`: operator UI, charts, site switching, settings, and waiting room editor

## 12. Data Model

### Primary relational tables

| Table | Key Fields | Purpose |
| --- | --- | --- |
| `tenants` | `id`, `name`, `plan`, `stripe_customer_id`, `created_at` | Customer account records |
| `sites` | `id`, `tenant_id`, `domain`, `origin_url`, `rps_limit`, `active` | Protected domains and origin config |
| `events` | `id`, `site_id`, `name`, `starts_at`, `ends_at`, `active` | Scheduled event mode windows |
| `analytics_hourly` | `id`, `site_id`, `hour`, `req_count`, `queued_count`, `blocked_count` | Aggregated traffic analytics |
| `ssl_certs` | `id`, `site_id`, `domain`, `cert_pem`, `expires_at`, `renewed_at` | Certificate cache metadata |
| `audit_logs` | `id`, `tenant_id`, `action`, `payload`, `created_at` | Configuration audit trail |
| `billing_usage` | `id`, `tenant_id`, `period_start`, `period_end`, `req_count`, `billed` | Usage metering for billing |

### Redis key patterns

| Key Pattern | Type | Purpose |
| --- | --- | --- |
| `tenant:{id}:bucket:{domain}` | Hash | Token bucket state |
| `tenant:{id}:queue:{domain}` | List | FIFO waiting room queue |
| `tenant:{id}:config:{domain}` | Hash | Cached site configuration |
| `tenant:{id}:metrics:{domain}` | Stream | Real-time metric stream |
| `breaker:{domain}` | String | Circuit breaker state |

## 13. Service Contracts

### REST API

| Method | Endpoint | Description | Auth |
| --- | --- | --- | --- |
| `POST` | `/auth/register` | Create a tenant account | None |
| `POST` | `/auth/login` | Issue JWT access token | None |
| `POST` | `/auth/refresh` | Refresh access token | Refresh token |
| `GET` | `/sites` | List tenant sites | JWT |
| `POST` | `/sites` | Add a protected site | JWT |
| `PATCH` | `/sites/:id/config` | Update origin URL and RPS limit | JWT |
| `DELETE` | `/sites/:id` | Remove a site from protection | JWT |
| `GET` | `/analytics/:siteId` | Retrieve traffic analytics | JWT |
| `POST` | `/events` | Create an event mode session | JWT |
| `GET` | `/billing/usage` | Compare current usage with plan limits | JWT |
| `POST` | `/billing/upgrade` | Upgrade the subscription plan | JWT |

### gRPC contracts

| RPC Method | Request | Response | Direction |
| --- | --- | --- | --- |
| `GetSiteConfig` | `domain: string` | Site config including origin, RPS, and tenant ID | Go -> Node |
| `UpdateSiteConfig` | `SiteConfig` | `Ack` | Node -> Go |
| `ReportMetrics` | `MetricsBatch` | `Ack` | Go -> Node |
| `GetCircuitState` | `domain: string` | `CircuitState` | Go -> Node |

## 14. Kubernetes Deployment Model

### Planned service sizing from the SRS

| Service | Replicas | Resource Notes | Scale Policy |
| --- | --- | --- | --- |
| Proxy | `3 min -> 20 max` | `256MB RAM`, `250m CPU` each | Scale when CPU > 60% |
| API | `2 min -> 8 max` | `512MB RAM`, `500m CPU` each | Scale when CPU > 70% |
| Dashboard | `2 min -> 4 max` | `256MB RAM`, `250m CPU` each | Scale when CPU > 80% |
| Redis | `1` StatefulSet | `1GB RAM` guaranteed | Manual scaling |
| PostgreSQL | `1` StatefulSet | `2GB RAM`, `1 CPU` | Manual scaling |

### Production traffic topology

```text
Internet
  -> LoadBalancer :80/:443
  -> Go Proxy
  -> Redis               (rate limits, queue state)
  -> Origin Server       (if allowed)
  -> Waiting Room HTML   (if queued)
  -> NestJS API via gRPC (config sync, metrics reporting)

Browser
  -> NestJS API via Socket.io (real-time dashboard updates)
```

## 15. Security Considerations

| Threat | Mitigation |
| --- | --- |
| Tenant data leakage | PostgreSQL RLS and Redis key namespacing by tenant |
| JWT token theft | Short-lived access tokens and secure `httpOnly` refresh cookies |
| DDoS against Edgevia itself | Upstream protection such as Cloudflare in front of platform infrastructure |
| SSL certificate exposure | Encrypted certificate storage and no certificate logging |
| Stripe webhook spoofing | Signature verification using the Stripe webhook header |
| Bot traffic in queue | User-Agent screening, behavioral analysis, and optional CAPTCHA |
| gRPC interception | mTLS between proxy and API services inside Kubernetes |

## 16. Delivery Roadmap

| Phase | Weeks | Deliverables | Priority |
| --- | --- | --- | --- |
| Phase 1: Proxy Core | 1-3 | Go proxy, Redis token bucket, SSL automation, tenant config resolver, Prometheus metrics | Critical |
| Phase 2: API Service | 4-5 | NestJS API, PostgreSQL schema, Prisma, JWT auth, gRPC client, site CRUD | Critical |
| Phase 3: Waiting Room | 6-7 | Queue logic, SSE updates, circuit breaker, crawler bypass, waiting room templates | High |
| Phase 4: Dashboard | 8-9 | Next.js dashboard, Socket.io updates, Zustand store, charts, site switcher, inline RPS config | High |
| Phase 5: Billing | 10 | Stripe subscriptions, metering, event add-ons, invoice dashboard | High |
| Phase 6: Kubernetes | 11 | K8s manifests, HPA, Grafana dashboards, staging and production deployment | High |
| Phase 7: Polish | 12 | Onboarding, notifications, waiting room editor, documentation, beta launch | Medium |

## 17. Notes

- This revision standardizes the platform name as **Edgevia** throughout the document.
- A stray reference to **FairFlow** appears in the SRS text; this document treats it as a naming carryover rather than a separate system.
- This file is intentionally architecture-focused and leaves pricing, market positioning, and business narrative to the SRS PDF.
