<div align="center">

<br />

```
███████╗██████╗  ██████╗ ███████╗██╗   ██╗██╗ █████╗
██╔════╝██╔══██╗██╔════╝ ██╔════╝██║   ██║██║██╔══██╗
█████╗  ██║  ██║██║  ███╗█████╗  ██║   ██║██║███████║
██╔══╝  ██║  ██║██║   ██║██╔══╝  ╚██╗ ██╔╝██║██╔══██║
███████╗██████╔╝╚██████╔╝███████╗ ╚████╔╝ ██║██║  ██║
╚══════╝╚═════╝  ╚═════╝ ╚══════╝  ╚═══╝  ╚═╝╚═╝  ╚═╝
```

**Enterprise-grade traffic protection for the rest of us.**

Reverse proxy · Virtual waiting room · Auto SSL · Real-time dashboard

<br />

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Proxy-Go_1.22-00ADD8?logo=go)](https://go.dev)
[![Node.js](https://img.shields.io/badge/API-NestJS-E0234E?logo=nestjs)](https://nestjs.com)
[![Next.js](https://img.shields.io/badge/Dashboard-Next.js_14-black?logo=next.js)](https://nextjs.org)
[![PostgreSQL](https://img.shields.io/badge/DB-PostgreSQL-336791?logo=postgresql)](https://postgresql.org)
[![Redis](https://img.shields.io/badge/Cache-Redis_7-DC382D?logo=redis)](https://redis.io)
[![Kubernetes](https://img.shields.io/badge/Deploy-Kubernetes-326CE5?logo=kubernetes)](https://kubernetes.io)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

<br />

[**Live Demo**](https://edgevia.io) · [**Documentation**](https://docs.edgevia.io) · [**Report Bug**](https://github.com/yourusername/edgevia/issues) · [**Request Feature**](https://github.com/yourusername/edgevia/issues)

</div>

---

## The Problem

Your product drops at 10 AM. Traffic hits 50x your normal load in 30 seconds. Your server crashes. Customers get a 502. Revenue walks out the door.

**You don't need Cloudflare Enterprise ($2,500/mo) to survive a flash sale.**

---

## What is Edgevia?

Edgevia sits between the internet and your origin server. When traffic exceeds your configured threshold, it activates a **Virtual Waiting Room** — holding users in a fair, branded queue instead of letting your server crash.

One CNAME change. Automatic SSL. Your business keeps running.

```
Before Edgevia                     After Edgevia
─────────────────                  ─────────────────────────────────────
                                   
  10,000 users                       10,000 users
       │                                   │
       ▼                                   ▼
  Your Server  ──── 💥 CRASH         ┌─────────────────────┐
                                     │     Edgevia Proxy   │
                                     │   (Go, sub-ms)      │
                                     └────────┬────────────┘
                                              │
                              ┌───────────────┴───────────────┐
                              │                               │
                        ┌─────▼──────┐               ┌───────▼───────┐
                        │  Allowed   │               │ Virtual Queue │
                        │  (RPS OK)  │               │  (fair FIFO)  │
                        └─────┬──────┘               └───────┬───────┘
                              │                              │
                              ▼                              ▼
                        Your Server ✅              User sees position,
                                                   waits their turn 🎟️
```

---

## Features

### 🛡️ The Shield — Core Proxy Engine
- **Zero-config onboarding** — add one CNAME, everything else is automatic
- **Auto SSL** — Let's Encrypt certificates provisioned in under 60 seconds per domain
- **Token bucket rate limiting** — per-tenant, per-domain, configurable RPS with burst allowance
- **Circuit breaker** — auto-detects origin failure, stops the cascade before it starts

### 🎟️ Virtual Waiting Room
- Fair **FIFO queue** — no position jumping, no bots cutting the line
- Real-time position updates via **SSE** (no polling)
- **Fully branded** — custom HTML/CSS waiting room per site
- **SEO-safe** — verified crawlers (Google, Bing) bypass the queue automatically

### 📊 Real-Time Dashboard
- Live visitor count, queue depth, RPS gauge — updated every 500ms–1s
- Inline RPS config — change rate limits without a page reload
- **Protected revenue counter** — see what Edgevia saved you in real-time
- Traffic analytics: 7 / 30 / 90 day breakdowns, peak hour heatmaps
- **Event Mode** — dedicated protection for scheduled flash sales and drops

### 🏢 Multi-Tenant Architecture
- Full tenant isolation — separate Redis namespaces, PostgreSQL Row-Level Security
- Unlimited sites per account (Business plan)
- Audit log for every configuration change

### 💳 Billing Built In
- Stripe-powered subscriptions with usage metering
- Overage alerts at 80% and 95% of plan limit
- One-click Event Add-on activation
- Self-service plan upgrades and invoice history

---

## Architecture

Edgevia is a **polyglot microservices monorepo**. The performance-critical proxy is written in Go. Business logic lives in NestJS. The dashboard is Next.js 14. Services communicate via gRPC internally.

```
                        ┌─────────────────────────────────┐
                        │         Internet Traffic         │
                        └──────────────┬──────────────────┘
                                       │ HTTPS (80/443)
                        ┌──────────────▼──────────────────┐
                        │     Go Proxy Engine              │
                        │  · SNI-based domain routing      │
                        │  · Token bucket rate limiting    │
                        │  · Circuit breaker               │
                        │  · Virtual waiting room (SSE)    │
                        │  · Let's Encrypt autocert        │
                        │  · Prometheus metrics            │
                        └──────┬───────────────┬──────────┘
                               │ gRPC           │ Redis
                 ┌─────────────▼────┐    ┌──────▼───────────┐
                 │  NestJS API       │    │    Redis 7        │
                 │  · JWT auth       │    │  · Token buckets  │
                 │  · Site CRUD      │    │  · Queue state    │
                 │  · Stripe billing │    │  · Pub/Sub        │
                 │  · BullMQ workers │    │  · Metrics stream │
                 │  · Socket.io push │    └──────────────────┘
                 └──────┬───────────┘
                        │ Prisma ORM
                 ┌──────▼───────────┐       ┌──────────────────┐
                 │   PostgreSQL      │       │  Next.js 14      │
                 │  · Tenants        │       │  · App Router    │
                 │  · Sites          │◄──────│  · Socket.io     │
                 │  · Analytics      │  REST │  · Zustand store │
                 │  · Billing usage  │       │  · Recharts      │
                 │  · SSL certs      │       └──────────────────┘
                 └──────────────────┘
```

### Tech Stack

| Layer | Technology | Why |
|---|---|---|
| **Proxy Engine** | Go 1.22 | goroutines handle 50K+ concurrent connections at sub-ms latency |
| **Management API** | Node.js + NestJS | Modular, typed, decorator-driven — perfect for business logic |
| **Frontend** | Next.js 14 (App Router) | SSR for first load, real-time via Socket.io |
| **Database** | PostgreSQL + Prisma | ACID compliance for billing and multi-tenant data |
| **Cache + Queue** | Redis 7 | Token buckets, FIFO queues, pub/sub, BullMQ |
| **Internal Comms** | gRPC + Protobuf | Type-safe binary protocol between Go and Node.js |
| **SSL** | Let's Encrypt + autocert | Auto-provision and renew certs per customer domain |
| **Metrics** | Prometheus + Grafana | Operational visibility and alerting |
| **Billing** | Stripe | Subscriptions, usage metering, webhooks |
| **Deployment** | Kubernetes | Auto-scaling, HA, HPA per service |

---

## Pricing

| Plan | Price | Sites | Requests/mo | Features |
|---|---|---|---|---|
| **Starter** | $29/mo | 3 | 1M | Basic waiting room |
| **Growth** | $79/mo | 10 | 5M | Custom branding, analytics |
| **Business** | $199/mo | Unlimited | 20M | Priority support, API access |
| **Event Add-on** | $49/event | — | — | Dedicated event mode + real-time queue dashboard |

Overages billed per 1K requests beyond plan limit. No surprises.

---

## Getting Started

### Prerequisites

| Tool | Version |
|---|---|
| Go | 1.22+ |
| Node.js | 20+ |
| pnpm | 9+ |
| Docker + Docker Compose | latest |
| Redis | 7+ |
| PostgreSQL | 15+ |

### Local Development

```bash
# 1. Clone the repository
git clone https://github.com/yourusername/edgevia.git
cd edgevia

# 2. Install Node.js dependencies (all workspaces)
pnpm install

# 3. Start infrastructure (PostgreSQL + Redis)
docker-compose up -d postgres redis

# 4. Set up environment variables
cp apps/api/.env.example apps/api/.env
cp apps/dashboard/.env.example apps/dashboard/.env.local

# 5. Run database migrations
cd apps/api && pnpm prisma migrate dev

# 6. Start all services in parallel (Turborepo)
pnpm dev
```

Services will be available at:
- **Dashboard** → http://localhost:3000
- **API** → http://localhost:4000
- **Proxy** → http://localhost:8080
- **Prometheus** → http://localhost:9090

### Go Proxy Only

```bash
cd apps/proxy
go mod download
go run cmd/server/main.go
```

---

## Project Structure

``` 
edgevia/
├── apps/
│   ├── api/                        # NestJS management API
│   │   ├── prisma/
│   │   │   └── schema.prisma
│   │   ├── src/
│   │   │   ├── common/
│   │   │   │   ├── decorators/
│   │   │   │   ├── guards/
│   │   │   │   └── interceptors/
│   │   │   ├── grpc/
│   │   │   ├── modules/
│   │   │   │   ├── analytics/
│   │   │   │   ├── auth/
│   │   │   │   ├── billing/
│   │   │   │   ├── events/
│   │   │   │   ├── notifications/
│   │   │   │   ├── sites/
│   │   │   │   └── tenants/
│   │   │   └── main.ts
│   │   ├── Dockerfile
│   │   └── package.json
│   ├── dashboard/                  # Next.js frontend
│   │   ├── app/
│   │   │   ├── auth/
│   │   │   │   ├── login/
│   │   │   │   └── register/
│   │   │   ├── billing/
│   │   │   ├── dashboard/
│   │   │   ├── settings/
│   │   │   └── layout.tsx
│   │   ├── components/
│   │   │   ├── charts/
│   │   │   ├── dashboard/
│   │   │   ├── ui/
│   │   │   └── waiting-room/
│   │   ├── hooks/
│   │   │   └── useSocket.ts
│   │   ├── public/
│   │   ├── store/
│   │   │   └── useSiteStore.ts
│   │   ├── Dockerfile
│   │   └── package.json
│   └── proxy/                      # Go proxy service
│       ├── cmd/
│       │   └── server/
│       │       └── main.go
│       ├── internal/
│       │   ├── circuitbreaker/
│       │   │   └── breaker.go
│       │   ├── logger/
│       │   │   └── logger.go
│       │   ├── metrics/
│       │   │   └── prometheus.go
│       │   ├── proxy/
│       │   │   └── handler.go
│       │   ├── queue/
│       │   │   └── room.go
│       │   ├── ratelimit/
│       │   │   └── bucket.go
│       │   ├── ssl/
│       │   │   └── manager.go
│       │   └── tenant/
│       │       └── resolver.go
│       ├── proto/
│       ├── Dockerfile
│       └── go.mod
├── docs/
│   └── architecture.md
├── infra/
│   ├── docker-compose.yml
│   └── k8s/                        # Kubernetes manifests
│       ├── api/
│       ├── dashboard/
│       ├── monitoring/
│       ├── postgres/
│       ├── proxy/
│       │   ├── deployment.yaml
│       │   └── hpa.yaml
│       ├── redis/
│       └── namespace.yaml
├── packages/
│   ├── proto/
│   │   └── fairflow.proto
│   └── types/
│       └── api.types.ts
├── scripts/
│   └── setup.sh
├── .env.example
├── .gitignore
├── pnpm-workspace.yaml
├── turbo.json
└── README.md
```

---

## Roadmap

| Phase | Weeks | Deliverables | Status |
|---|---|---|---|
| **Phase 1** — Proxy Core | 1–3 | Go reverse proxy, Redis token bucket, Let's Encrypt SSL | 🔨 In Progress |
| **Phase 2** — API Service | 4–5 | NestJS API, PostgreSQL, Prisma, JWT auth, gRPC, site CRUD | 📋 Planned |
| **Phase 3** — Waiting Room | 6–7 | Virtual queue logic, SSE, circuit breaker, bot bypass | 📋 Planned |
| **Phase 4** — Dashboard | 8–9 | Next.js dashboard, Socket.io real-time, Recharts, Zustand | 📋 Planned |
| **Phase 5** — Billing | 10 | Stripe subscriptions, usage metering, event add-on flow | 📋 Planned |
| **Phase 6** — Kubernetes | 11 | K8s manifests, HPA, Grafana dashboards, staging deploy | 📋 Planned |
| **Phase 7** — Polish | 12 | Onboarding flow, email alerts, waiting room editor, beta | 📋 Planned |

---

## Environment Variables

### API (`apps/api/.env`)

```env
# Database
DATABASE_URL=postgresql://user:password@localhost:5432/edgevia

# Redis
REDIS_URL=redis://localhost:6379

# JWT
JWT_SECRET=your-secret-here
JWT_REFRESH_SECRET=your-refresh-secret-here

# Stripe
STRIPE_SECRET_KEY=sk_live_...
STRIPE_WEBHOOK_SECRET=whsec_...

# Resend (email)
RESEND_API_KEY=re_...

# Cloudflare (DNS automation)
CLOUDFLARE_API_TOKEN=...
CLOUDFLARE_ZONE_ID=...
```

### Dashboard (`apps/dashboard/.env.local`)

```env
NEXT_PUBLIC_API_URL=http://localhost:4000
NEXT_PUBLIC_SOCKET_URL=http://localhost:4000
```

### Proxy (`apps/proxy/.env`)

```env
REDIS_ADDR=localhost:6379
GRPC_API_ADDR=localhost:50051
METRICS_PORT=9090
```

---

## Security

| Threat | Mitigation |
|---|---|
| Tenant data leakage | PostgreSQL RLS + Redis key namespacing per `tenant_id` |
| JWT token theft | 15-minute access tokens + httpOnly refresh cookies |
| DDoS against Edgevia | Cloudflare in front of Edgevia infrastructure |
| SSL cert exposure | Certs stored encrypted in PostgreSQL, never logged |
| Stripe webhook spoofing | `stripe-signature` header verification on every webhook |
| Bot traffic in queue | User-Agent + behavioral analysis + CAPTCHA on waiting room |
| gRPC interception | mTLS between proxy and API inside Kubernetes |

---

## Contributing

Contributions are welcome. Please open an issue first to discuss what you'd like to change.

```bash
# Fork the repo, then:
git checkout -b feature/your-feature-name
git commit -m "feat: add your feature"
git push origin feature/your-feature-name
# Open a Pull Request
```

Please follow [Conventional Commits](https://www.conventionalcommits.org/) for commit messages.

---

## License

MIT © [Humayun Kabir](https://github.com/yourusername)

---

<div align="center">

Built with obsession by [Humayun Kabir](https://github.com/yourusername) · Dhaka, Bangladesh 🇧🇩

*If Edgevia saves your flash sale, consider giving it a ⭐*

</div>
