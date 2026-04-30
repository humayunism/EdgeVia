<div align="center">

<br />

```text
███████╗██████╗  ██████╗ ███████╗██╗   ██╗██╗ █████╗
██╔════╝██╔══██╗██╔════╝ ██╔════╝██║   ██║██║██╔══██╗
█████╗  ██║  ██║██║  ███╗█████╗  ██║   ██║██║███████║
██╔══╝  ██║  ██║██║   ██║██╔══╝  ╚██╗ ██╔╝██║██╔══██║
███████╗██████╔╝╚██████╔╝███████╗ ╚████╔╝ ██║██║  ██║
╚══════╝╚═════╝  ╚═════╝ ╚══════╝  ╚═══╝  ╚═╝╚═╝  ╚═╝
```

**Traffic protection and virtual waiting rooms for modern products.**

Reverse proxy · Fair queueing · Auto SSL · Real-time dashboard

<br />

[![Go](https://img.shields.io/badge/Proxy-Go_1.22-00ADD8?logo=go)](https://go.dev)
[![NestJS](https://img.shields.io/badge/API-NestJS_10-E0234E?logo=nestjs)](https://nestjs.com)
[![Next.js](https://img.shields.io/badge/Dashboard-Next.js_14-black?logo=next.js)](https://nextjs.org)
[![PostgreSQL](https://img.shields.io/badge/DB-PostgreSQL-336791?logo=postgresql)](https://postgresql.org)
[![Redis](https://img.shields.io/badge/Cache-Redis_7-DC382D?logo=redis)](https://redis.io)
[![Kubernetes](https://img.shields.io/badge/Deploy-Kubernetes-326CE5?logo=kubernetes)](https://kubernetes.io)

[Architecture](docs/architecture.md) · [SRS PDF](docs/edgevia-srs.pdf)

</div>

## What Edgevia Is

Edgevia is a high-performance, multi-tenant reverse proxy and traffic management platform built for products that face sudden traffic spikes: flash sales, ticket launches, product drops, viral events, and launch-day surges.

It sits between the public internet and a customer origin server. When traffic stays within the configured rate, requests pass through normally. When demand exceeds capacity, Edgevia shifts users into a branded virtual waiting room instead of letting the origin collapse.

## Why It Exists

A traffic spike should not force a business to choose between downtime and enterprise-only infrastructure pricing. Edgevia is designed to give smaller teams the same core protection patterns used by larger platforms:

- reverse proxying at the edge
- per-tenant rate limiting
- fair FIFO queueing
- automatic SSL
- origin protection with circuit breaking
- live operational visibility through a dashboard

## Core Capabilities

### Shield Layer

- `SNI`-based domain resolution
- reverse proxying to customer origins
- token-bucket rate limiting per tenant and per domain
- circuit-breaker protection for unhealthy origins
- automatic certificate provisioning with Let's Encrypt

### Virtual Waiting Room

- fair `FIFO` queueing
- queue activation when traffic exceeds configured thresholds
- branded waiting-room experiences
- real-time position updates
- crawler bypass for verified Google and Bing bots

### Control Plane

- tenant and site management
- live metrics and health views
- inline rate-limit configuration
- event mode workflows
- billing and usage visibility
- audit logging and multi-tenant isolation

## Architecture At A Glance

Edgevia follows the architecture defined in the SRS and detailed in [docs/architecture.md](docs/architecture.md):

- `Go` handles the latency-sensitive proxy path
- `NestJS` owns authentication, site configuration, billing, analytics, and orchestration
- `Next.js 14` provides the dashboard experience
- `Redis` backs token buckets, waiting-room queues, pub/sub, and fast state
- `PostgreSQL` stores tenants, sites, events, analytics, SSL metadata, audit logs, and billing usage
- `gRPC` is used for internal service-to-service contracts
- `Socket.io` pushes live dashboard updates
- `Kubernetes` is the production deployment target

```text
Internet -> Load Balancer -> Go Proxy -> Origin
                              |      |
                              |      -> Waiting Room
                              |
                              -> Redis
                              -> NestJS API -> PostgreSQL
                                             -> BullMQ Workers -> Stripe

Browser -> Next.js Dashboard -> NestJS API -> Socket.io live updates
```

## Tech Stack

| Layer | Technology | Role |
|---|---|---|
| Proxy | Go 1.22 | Reverse proxy, rate limiting, queueing, metrics |
| API | Node.js + NestJS | Auth, config, analytics, billing, orchestration |
| Dashboard | Next.js 14 | UI and real-time control plane |
| Database | PostgreSQL + Prisma | Durable relational storage |
| Fast State | Redis 7 | Token buckets, queue state, pub/sub, streams |
| Jobs | BullMQ | Metering, alerts, async workflows |
| RPC | gRPC + Protobuf | Internal contracts between services |
| Observability | Prometheus + Grafana | Metrics, dashboards, alerting |
| Billing | Stripe | Plans, metering, invoices, webhooks |
| Email | Resend | Transactional notifications |
| Deploy | Docker + Kubernetes | Packaging and production orchestration |

## Repository Layout

```text
EdgeVia/
├── apps/
│   ├── proxy/       # Go proxy service
│   ├── api/         # NestJS management API
│   └── dashboard/   # Next.js dashboard
├── packages/
│   ├── proto/       # Shared protobuf definitions
│   └── types/       # Shared TypeScript types
├── infra/
│   ├── docker-compose.yml
│   └── k8s/
├── docs/
│   ├── architecture.md
│   └── edgevia-srs.pdf
├── scripts/
├── pnpm-workspace.yaml
└── turbo.json
```

## Current State Of This Repository

This repo already contains the three primary application workspaces and the main architectural building blocks:

- Go proxy entrypoint and internal modules for proxying, rate limiting, queueing, SSL, tenant resolution, metrics, and circuit breaking
- NestJS API workspace with Prisma schema and package dependencies
- Next.js dashboard workspace with Socket.io client and Zustand state
- Kubernetes manifests and Docker Compose infrastructure

This is still an active implementation repository, so some SRS-defined features are already scaffolded while others are still being completed. In a few places, the current codebase uses in-memory or placeholder implementations where the SRS targets Redis-backed or production-grade behavior.

## Local Development

### Prerequisites

| Tool | Version |
|---|---|
| Go | `1.22+` |
| Node.js | `20+` |
| pnpm | `9+` |
| Docker | recent |
| Docker Compose | recent |

### 1. Install workspace dependencies

```bash
pnpm install
```

### 2. Start local infrastructure

```bash
docker compose -f infra/docker-compose.yml up -d
```

This starts:

- PostgreSQL on `localhost:5432`
- Redis on `localhost:6379`
- Prometheus on `localhost:9090`
- Grafana on `localhost:3001`

### 3. Configure environment variables

The repository includes a root [`.env.example`](.env.example). Copy the values you need into the appropriate app-level environment files used by your local setup.

At minimum, the project expects values for:

- `DATABASE_URL`
- `REDIS_URL`
- `JWT_SECRET`
- `JWT_REFRESH_SECRET`
- `NEXT_PUBLIC_API_URL`

### 4. Generate Prisma client and run migrations

```bash
pnpm -C apps/api db:generate
pnpm -C apps/api db:migrate
```

### 5. Run the services

Open separate terminals and run:

```bash
pnpm -C apps/api dev
```

```bash
pnpm -C apps/dashboard dev
```

```bash
cd apps/proxy
go run cmd/server/main.go
```

Expected local ports:

- dashboard: `http://localhost:3000`
- API: `http://localhost:4000` if configured that way in your NestJS env
- proxy HTTP: `http://localhost:8080`
- proxy HTTPS fallback: `https://localhost:8443`
- Prometheus: `http://localhost:9090`
- Grafana: `http://localhost:3001`

## Running Individual Parts

### Proxy Only

```bash
cd apps/proxy
go run cmd/server/main.go
```

### API Only

```bash
pnpm -C apps/api dev
```

### Dashboard Only

```bash
pnpm -C apps/dashboard dev
```

## Configuration Notes

The SRS defines several important runtime concepts that show up across the codebase:

- token bucket key pattern: `tenant:{tenantId}:bucket:{domain}`
- waiting room queue key pattern: `tenant:{tenantId}:queue:{domain}`
- cached site config key pattern: `tenant:{tenantId}:config:{domain}`
- metrics stream key pattern: `tenant:{tenantId}:metrics:{domain}`
- circuit-breaker state key: `breaker:{domain}`

The architecture also assumes:

- PostgreSQL row-level tenant isolation
- Redis key namespacing by tenant
- gRPC communication between proxy and API
- Socket.io for dashboard live updates
- Kubernetes as the production target

## Planned Product Surface From The SRS

Edgevia’s SRS defines the following platform surface:

- zero-config onboarding through `CNAME` + SNI detection
- SSL auto-provisioning within roughly `60 seconds`
- branded waiting-room editor
- queue analytics and protected revenue views
- event mode for scheduled high-traffic periods
- Stripe subscriptions and hourly usage metering
- overage warnings at `80%` and `95%`
- audit logs for tenant configuration changes

## Documentation

- [Architecture](docs/architecture.md)
- [SRS PDF](docs/edgevia-srs.pdf)

## Development Notes

- The SRS is the source of truth for the intended platform behavior.
- The architecture document is the clean written summary of that SRS.
- The repo currently reflects a mix of implemented features and forward-looking structure from the blueprint.

## License

This repository does not currently include a license file. Add one before public distribution if you want the project to be open-source.
