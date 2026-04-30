# Edgevia Architecture

This document restates the architecture defined in `docs/edgevia-srs.pdf` and is aligned to the SRS titled `EDGEVIA Software Requirements Specification & System Architecture Blueprint`, version `1.0.0`, dated `April 2026`.

## 1. Architecture Summary

Edgevia is a high-performance, multi-tenant reverse proxy and traffic management platform designed to protect customer origin servers during sudden traffic spikes such as flash sales, product drops, ticket releases, and viral events.

The system follows a polyglot microservices architecture:

- `Go` powers the performance-critical proxy layer.
- `Node.js + NestJS` handles business logic, configuration, tenant management, billing, and integrations.
- `Next.js 14` provides the dashboard and operational control plane.
- `Redis` backs rate limiting, queue state, and real-time coordination.
- `PostgreSQL` stores durable tenant, site, billing, analytics, and certificate data.

Production deployment targets `Kubernetes` from day one. `Docker Compose` is used for local development only.

## 2. High-Level System View

External traffic enters Edgevia only through the Go Proxy Engine. Internal services communicate over `gRPC`, while dashboard updates are pushed in real time via `Socket.io`.

```text
                         Internet Traffic
                                |
                         Load Balancer
                           (80 / 443)
                                |
                                v
                     +----------------------+
                     |   Go Proxy Engine    |
                     |----------------------|
                     | SNI domain routing   |
                     | Token bucket limiting|
                     | Waiting room queue   |
                     | Circuit breaker      |
                     | SSL autocert         |
                     | Prometheus metrics   |
                     +----+-----------+-----+
                          |           |
                 Redis checks         | gRPC
                          |           v
                          |   +----------------------+
                          |   |   NestJS API         |
                          |   |----------------------|
                          |   | Auth                 |
                          |   | Site management      |
                          |   | Billing              |
                          |   | Analytics            |
                          |   | Config sync          |
                          |   +----+-----------+-----+
                          |        |           |
                          |        |           +------------------+
                          |        |                              |
                          v        v                              v
                   +-------------+   +------------------+   +------------------+
                   |   Redis 7   |   |   PostgreSQL     |   |  BullMQ Workers  |
                   |-------------|   |------------------|   |------------------|
                   | Buckets     |   | Tenants          |   | Usage metering   |
                   | Queues      |   | Sites            |   | Billing events   |
                   | Config cache|   | Events           |   | Emails           |
                   | Pub/Sub     |   | Analytics        |   +------------------+
                   | Streams     |   | SSL certs        |
                   +-------------+   | Audit logs       |
                                     | Billing usage    |
                                     +------------------+

 Browser <-------------------- Socket.io --------------------> Next.js Dashboard
```

## 3. Core Architectural Pattern

### 3.1 Polyglot Microservices

Edgevia separates concerns by execution profile:

- `Go` is used for low-latency request handling and concurrency-heavy proxy work.
- `NestJS` is used for modular business workflows such as authentication, billing, configuration, and analytics.
- `Next.js 14` is used for the operator-facing dashboard with SSR and real-time client updates.

### 3.2 Internal Communication Model

- `gRPC + Protocol Buffers` connect the Go proxy and NestJS API for configuration lookup, configuration update propagation, metrics reporting, and circuit-state retrieval.
- `Socket.io` delivers live metrics and state updates from the API to the dashboard.
- `Redis pub/sub` and `Redis streams` support internal real-time coordination.

## 4. Primary Services

### 4.1 Go Proxy Engine

The Go Proxy Engine is the only service exposed to public request traffic. It is responsible for:

- SNI-based domain detection and request routing
- reverse proxying to customer origin servers
- per-tenant, per-domain token-bucket rate limiting
- activation of the Virtual Waiting Room when thresholds are exceeded
- FIFO queue admission behavior
- crawler bypass for verified Google and Bing bots
- circuit-breaker protection for overloaded or failing origins
- automatic SSL provisioning and renewal with Let's Encrypt via `autocert`
- Prometheus metric emission

Key behaviors defined by the SRS:

- customer onboarding begins with a `CNAME` pointed to the Edgevia proxy
- the proxy detects the domain through `SNI`
- origin configuration is stored in PostgreSQL and cached in Redis
- SSL certificates are expected to auto-provision within `60 seconds`

### 4.2 NestJS Management API

The NestJS API is the control plane for Edgevia. It is responsible for:

- tenant registration and authentication
- JWT access token issuance and refresh handling
- protected site CRUD operations
- configuration changes such as RPS limit and origin URL updates
- analytics retrieval
- event mode activation
- billing and subscription workflows
- Stripe webhook processing
- usage metering coordination
- real-time dashboard event publishing
- audit logging

The API persists long-lived data in PostgreSQL and coordinates fast-changing operational data through Redis.

### 4.3 Next.js Dashboard

The dashboard is the single-page operational nerve center for customers. It provides:

- live visitor counts
- queue depth visualization
- system health display
- protected revenue counters
- current RPS gauges
- origin response-time charts
- site switching across protected domains
- inline RPS configuration updates without page reload
- waiting-room editing with live preview
- traffic analytics for `7 / 30 / 90` day windows
- event mode controls
- billing and plan management
- alert settings including webhook URL, email, and Slack integration

### 4.4 Redis

Redis is used for low-latency shared state:

- token bucket state for rate limiting
- FIFO waiting-room queue state
- cached site configuration
- request metrics streams
- pub/sub for live dashboard metrics
- circuit-breaker state
- BullMQ backing data

### 4.5 PostgreSQL

PostgreSQL is the durable system of record for:

- tenants
- protected sites
- scheduled event-mode records
- hourly analytics aggregates
- SSL certificate metadata and certificate cache
- audit logs
- billing usage records

## 5. Technology Stack

| Layer | Technology | Purpose |
|---|---|---|
| Proxy Engine | Go (Golang) | Handles 50K+ concurrent connections with sub-ms latency |
| Management API | Node.js + NestJS | Typed, modular business logic for auth, config, analytics, and billing |
| Frontend Dashboard | Next.js 14 (App Router) | SSR first load with real-time operator experience |
| Primary Database | PostgreSQL + Prisma ORM | Durable ACID storage for tenants, billing, and site state |
| Cache + Queue | Redis 7 | Token buckets, queue state, config cache, pub/sub, streams |
| Job Queue | BullMQ | Background jobs for analytics, email, and billing events |
| Internal Comms | gRPC + Protocol Buffers | Type-safe service-to-service communication |
| Real-Time UI | Socket.io | Dashboard live updates |
| SSL Automation | Let's Encrypt + `autocert` | Automatic certificate provisioning per domain |
| Metrics | Prometheus + Grafana | Monitoring, visibility, and alerting |
| Logging | Zap (Go) + Pino (Node.js) | Structured low-overhead JSON logging |
| Tracing | OpenTelemetry | Distributed tracing across proxy and API |
| Containers | Docker + Kubernetes | Packaging and production orchestration |
| Object Storage | AWS S3 or Cloudflare R2 | Waiting-room assets, logos, analytics exports |
| Billing | Stripe | Subscriptions, metering, invoices, webhooks |
| Email | Resend | Transactional alerts and onboarding email |
| DNS | Cloudflare API | Programmatic CNAME and DNS-related setup support |

## 6. Request Lifecycle

### 6.1 Protected Request Flow

1. A browser sends HTTPS traffic to Edgevia through the public load balancer.
2. The Go proxy resolves the requested domain using `SNI`.
3. The proxy loads site configuration from Redis or retrieves it from the API via `gRPC`.
4. The proxy checks the tenant/domain token bucket in Redis.
5. If capacity exists, the request is forwarded to the origin server.
6. If capacity is exhausted, the request is added to the waiting-room queue instead of being dropped.
7. If the queue threshold is exceeded, the user is served the branded waiting-room page.
8. Queue position is updated in real time via `WebSocket or SSE` as defined by the SRS.
9. Verified search-engine crawlers can bypass the queue.
10. Circuit-breaker logic can stop forwarding traffic to an unhealthy origin and hold or shed traffic safely.

### 6.2 Dashboard Flow

1. The customer uses the dashboard in the browser.
2. The dashboard communicates with the NestJS API for CRUD and reporting operations.
3. Real-time metrics are pushed from the API to the dashboard through `Socket.io`.
4. The API reads durable state from PostgreSQL and fast-changing state from Redis.

### 6.3 Billing Flow

1. Request usage is accumulated operationally.
2. BullMQ workers report usage to Stripe every hour.
3. The API exposes billing usage and plan-management endpoints to the dashboard.
4. Alerts are raised at `80%` and `95%` of plan limits.

## 7. Core Protection Mechanisms

### 7.1 Zero-Configuration Onboarding

- customer points a `CNAME` record to Edgevia
- proxy detects the new domain automatically via `SNI`
- certificate issuance is handled automatically
- origin URL is stored in PostgreSQL and cached in Redis

### 7.2 Token-Bucket Rate Limiting

Each tenant receives an isolated Redis-backed bucket:

- key pattern: `tenant:{tenantId}:bucket:{domain}`
- refill rate matches configured `RPS`
- when empty, traffic is queued rather than dropped
- burst allowance is configurable, for example `2x RPS for 5 seconds`

### 7.3 Virtual Waiting Room

- activates when queue depth exceeds threshold
- serves a branded customizable HTML/CSS waiting room
- maintains a fair `FIFO` queue with no position jumping
- exposes real-time position updates
- preserves `X-Forwarded-For` for origin-side IP logging

### 7.4 Circuit Breaker

The proxy protects origin servers from cascading failure using three states:

- `CLOSED`: normal forwarding
- `OPEN`: traffic blocked from origin
- `HALF_OPEN`: limited test traffic allowed

SRS-defined behavior:

- opens when origin error rate exceeds `50%` in a `10 second` window
- auto-recovery sends `10%` of traffic to probe origin health
- customers are notified through webhook or email when the circuit opens

## 8. Multi-Tenant Isolation Model

Tenant isolation is enforced across layers:

- all Redis keys are prefixed with `tenant:{tenantId}:`
- PostgreSQL enforces `Row-Level Security (RLS)` per tenant
- Kubernetes `NetworkPolicy` isolates tenant namespaces
- audit logs record tenant configuration changes

This architecture treats tenancy as a first-class boundary in both storage and runtime communication.

## 9. Data Architecture

### 9.1 Key PostgreSQL Tables

| Table | Key Fields | Purpose |
|---|---|---|
| `tenants` | `id`, `name`, `plan`, `stripe_customer_id`, `created_at` | One row per Edgevia customer |
| `sites` | `id`, `tenant_id`, `domain`, `origin_url`, `rps_limit`, `active` | Protected customer domains |
| `events` | `id`, `site_id`, `name`, `starts_at`, `ends_at`, `active` | Scheduled event-mode sessions |
| `analytics_hourly` | `id`, `site_id`, `hour`, `req_count`, `queued_count`, `blocked_count` | Aggregated traffic data |
| `ssl_certs` | `id`, `site_id`, `domain`, `cert_pem`, `expires_at`, `renewed_at` | Let's Encrypt certificate cache |
| `audit_logs` | `id`, `tenant_id`, `action`, `payload`, `created_at` | Configuration change history |
| `billing_usage` | `id`, `tenant_id`, `period_start`, `period_end`, `req_count`, `billed` | Stripe usage metering records |

### 9.2 Redis Key Patterns

| Key Pattern | Type | Purpose |
|---|---|---|
| `tenant:{id}:bucket:{domain}` | Hash | Token bucket state |
| `tenant:{id}:queue:{domain}` | List | Virtual waiting-room FIFO queue |
| `tenant:{id}:config:{domain}` | Hash | Cached site config including origin URL and RPS limit |
| `tenant:{id}:metrics:{domain}` | Stream | Real-time request metrics stream |
| `breaker:{domain}` | String | Circuit-breaker state: `CLOSED`, `OPEN`, `HALF_OPEN` |

## 10. Service Contracts

### 10.1 REST API Surface

| Method | Endpoint | Purpose | Auth |
|---|---|---|---|
| `POST` | `/auth/register` | Create new tenant account | None |
| `POST` | `/auth/login` | Issue JWT access token | None |
| `POST` | `/auth/refresh` | Refresh access token | Refresh token |
| `GET` | `/sites` | List tenant sites | JWT |
| `POST` | `/sites` | Add protected site | JWT |
| `PATCH` | `/sites/:id/config` | Update RPS limit and origin URL | JWT |
| `DELETE` | `/sites/:id` | Remove protected site | JWT |
| `GET` | `/analytics/:siteId` | Fetch traffic analytics | JWT |
| `POST` | `/events` | Create event mode session | JWT |
| `GET` | `/billing/usage` | Show current usage versus plan | JWT |
| `POST` | `/billing/upgrade` | Upgrade subscription plan | JWT |

### 10.2 gRPC Contract Between Go and Node.js

| RPC Method | Request | Response | Direction |
|---|---|---|---|
| `GetSiteConfig` | `domain: string` | `SiteConfig (rps, origin, tenant_id)` | Go -> Node |
| `UpdateSiteConfig` | `SiteConfig` | `Ack` | Node -> Go |
| `ReportMetrics` | `MetricsBatch` | `Ack` | Go -> Node |
| `GetCircuitState` | `domain: string` | `CircuitState` | Go -> Node |

## 11. Kubernetes Architecture

### 11.1 Service Sizing

| Service | Replicas | Resources | HPA Target |
|---|---|---|---|
| `proxy (Go)` | `3 min -> 20 max` | `256MB RAM`, `250m CPU` each | scale up at `CPU > 60%` |
| `api (NestJS)` | `2 min -> 8 max` | `512MB RAM`, `500m CPU` each | scale up at `CPU > 70%` |
| `dashboard (Next.js)` | `2 min -> 4 max` | `256MB RAM`, `250m CPU` each | scale up at `CPU > 80%` |
| `redis` | `1 StatefulSet` | `1GB RAM guaranteed` | manual scale only |
| `postgres` | `1 StatefulSet` | `2GB RAM`, `1 CPU` | manual scale only |

### 11.2 Traffic and Dependency Flow

```text
Internet
  -> Load Balancer (80/443)
  -> Go Proxy Service
  -> Redis for rate-limit and queue checks
  -> Origin Server if allowed
     or Waiting Room HTML if queued

Go Proxy
  -> NestJS API via gRPC for config sync and metrics reporting

Browser
  -> NestJS API via Socket.io for real-time dashboard updates

NestJS API
  -> PostgreSQL for persistent reads and writes
  -> Redis for queue management and pub/sub

BullMQ Workers
  -> Stripe API for hourly usage metering
```

## 12. Planned Repository Blueprint

The SRS defines a monorepo using `pnpm workspaces` and `Turborepo` with three applications and shared packages.

```text
EDGEVIA/
├── apps/
│   ├── proxy/
│   │   ├── cmd/server/main.go
│   │   ├── internal/proxy/
│   │   ├── internal/ratelimit/
│   │   ├── internal/queue/
│   │   ├── internal/circuitbreaker/
│   │   ├── internal/ssl/
│   │   ├── internal/tenant/
│   │   ├── internal/metrics/
│   │   ├── internal/logger/
│   │   ├── proto/config.proto
│   │   ├── go.mod
│   │   └── Dockerfile
│   ├── api/
│   │   ├── src/modules/auth/
│   │   ├── src/modules/sites/
│   │   ├── src/modules/analytics/
│   │   ├── src/modules/billing/
│   │   ├── src/modules/events/
│   │   ├── src/modules/socket/
│   │   ├── prisma/schema.prisma
│   │   ├── package.json
│   │   └── Dockerfile
│   └── dashboard/
│       ├── app/(auth)/
│       ├── app/(protected)/
│       ├── components/ui/
│       ├── components/charts/
│       ├── components/dashboard/
│       ├── components/waiting-room/
│       ├── hooks/
│       ├── store/
│       ├── package.json
│       └── Dockerfile
├── packages/
│   ├── proto/
│   └── types/
├── infra/
│   ├── k8s/
│   ├── docker-compose.yml
│   └── terraform/
├── docs/
├── scripts/
├── pnpm-workspace.yaml
├── turbo.json
└── README.md
```

## 13. Security Architecture

The SRS identifies the following security controls:

| Threat | Mitigation |
|---|---|
| Tenant data leakage | PostgreSQL RLS and Redis tenant namespacing |
| JWT token theft | `15 minute` access tokens and secure `httpOnly` refresh-token cookies |
| DDoS against Edgevia itself | Upstream DDoS protection, with Cloudflare in front of Edgevia infrastructure |
| SSL certificate exposure | Certificates stored encrypted in PostgreSQL and never logged |
| Stripe webhook spoofing | Stripe signature verification |
| Bot traffic in queue | User-agent and behavioral analysis plus CAPTCHA on the waiting room |
| gRPC interception | `mTLS` between proxy and API services in Kubernetes |

## 14. Operational Characteristics

The architecture is designed to deliver:

- sub-millisecond proxy-path latency
- support for `50K+` concurrent connections at the proxy tier
- graceful degradation through queuing instead of hard request drops
- origin protection through rate limiting and circuit breaking
- tenant-safe scaling through isolated keyspaces and database policies
- real-time observability through Prometheus, Grafana, Socket.io, and Redis streams

## 15. Architecture Intent

At its core, Edgevia is built around one principle: when customer traffic spikes, the platform should absorb pressure, preserve fairness, and keep the origin alive. The Go proxy handles the traffic edge, Redis manages fast state, PostgreSQL preserves durable truth, the NestJS API governs business workflows, and the Next.js dashboard gives operators live control over the system.
