# FairFlow

**Software Requirements Specification & System Architecture Blueprint**

**Version:** 1.0.0
**Release Date:** April 2026
**Author:** Humayun Kabir
**Status:** Active Development

---

# 1. Project Overview

**FairFlow** is a high-performance, multi-tenant **reverse proxy and traffic management platform** built for:

* E-commerce businesses
* SaaS startups
* Any web service experiencing sudden traffic spikes

Examples include:

* Flash sales
* Product drops
* Ticket releases
* Viral traffic events

FairFlow sits **between the internet and the customer's origin server**.

When traffic exceeds configured thresholds, FairFlow activates a **Virtual Waiting Room**, placing users in a **fair queue instead of allowing the server to crash**.

The customer's service stays online and revenue is protected.

---

# 1.1 The Problem

Small and mid-size businesses face several issues:

* Cloudflare Enterprise is extremely expensive
* Shared hosting and budget VPS servers crash during traffic spikes
* Flash sales often cause server failures
* Existing solutions are enterprise-only or overly complex

Result: **direct revenue loss.**

---

# 1.2 The Solution

FairFlow provides **enterprise-grade traffic protection at startup pricing**.

Key capabilities:

* One CNAME configuration
* Automatic SSL provisioning
* Real-time traffic dashboard
* Fair customer queue during overload

---

# 1.3 Target Market

| Segment                          | Example                         | Pain Point                     |
| -------------------------------- | ------------------------------- | ------------------------------ |
| E-commerce (Shopify/WooCommerce) | Fashion drops, sneaker releases | Server crash = lost sales      |
| Ticketing Platforms              | Concert / event ticket sales    | Bot traffic, fair queue needed |
| SaaS launching on Product Hunt   | Launch day traffic spike        | 10x traffic overnight          |
| Custom VPS businesses            | Self-hosted services            | No CDN or protection layer     |

---

# 2. Business Model & Pricing

## 2.1 Pricing Strategy

FairFlow uses a **hybrid SaaS model**:

* Monthly subscription
* Event-based add-ons
* Usage-based overage

| Plan         | Monthly Price | Included                                        | Overage          |
| ------------ | ------------- | ----------------------------------------------- | ---------------- |
| Starter      | $29/mo        | 3 sites, 1M requests/month                      | $0.002 / 1K req  |
| Growth       | $79/mo        | 10 sites, 5M requests, branding, analytics      | $0.0015 / 1K req |
| Business     | $199/mo       | Unlimited sites, 20M requests, priority support | $0.001 / 1K req  |
| Event Add-on | $49/event     | Dedicated event protection                      | Included         |

---

## 2.2 Revenue Streams

Primary revenue channels:

* Monthly SaaS subscriptions (MRR)
* Event-based add-ons
* Overage billing for extra requests
* Enterprise contracts (future roadmap)

---

# 3. System Architecture

## 3.1 Architecture Overview

FairFlow uses a **polyglot microservices architecture** deployed on **Kubernetes**.

External traffic flows through the **Go Proxy Engine**.

Internal services communicate via **gRPC**.

| Component         | Technology       |
| ----------------- | ---------------- |
| Proxy Engine      | Go               |
| Management API    | Node.js + NestJS |
| Dashboard         | Next.js          |
| Communication     | gRPC             |
| Real-time updates | Socket.io        |

---

## 3.2 Technology Stack

| Layer                  | Technology              | Justification                       |
| ---------------------- | ----------------------- | ----------------------------------- |
| Proxy Engine           | Go (Golang)             | Handles 50K+ concurrent connections |
| Management API         | Node.js + NestJS        | Scalable modular architecture       |
| Dashboard              | Next.js 14              | SSR + real-time UI                  |
| Database               | PostgreSQL + Prisma     | ACID compliance                     |
| Cache                  | Redis 7                 | Rate limiting + queue               |
| Job Queue              | BullMQ                  | Async processing                    |
| Internal Communication | gRPC + Protocol Buffers | High performance                    |
| Real-time UI           | Socket.io               | Live dashboard metrics              |
| SSL Automation         | Let's Encrypt           | Automatic certificate provisioning  |
| Metrics                | Prometheus + Grafana    | Monitoring and alerts               |
| Logging                | Zap (Go) + Pino (Node)  | Structured logging                  |
| Tracing                | OpenTelemetry           | Distributed tracing                 |
| Containers             | Docker + Kubernetes     | Orchestration                       |
| Object Storage         | AWS S3 / Cloudflare R2  | Static assets                       |
| Billing                | Stripe                  | Subscription + usage                |
| Email                  | Resend                  | Transactional notifications         |
| DNS                    | Cloudflare API          | Automated DNS management            |

---

# 4. Feature Specifications

---

# 4.1 Core Proxy Engine (The Shield)

## 4.1.1 Zero-Configuration Onboarding

Customer workflow:

1. Customer adds a **CNAME record**
2. Domain points to **FairFlow**
3. Proxy detects domain via **SNI**
4. SSL certificate is auto-provisioned
5. Origin server configuration is stored

Setup completes in **~60 seconds**.

---

## 4.1.2 Rate Limiting — Token Bucket Algorithm

Each tenant receives an isolated rate-limit bucket stored in Redis.

Key pattern:

```
tenant:{tenantId}:bucket:{domain}
```

Features:

* Configurable requests per second (RPS)
* Burst allowance
* Automatic refill

If the bucket becomes empty:

* Requests are **queued**, not dropped.

---

## 4.1.3 Virtual Waiting Room

Activated when traffic exceeds capacity.

Features:

* Branded waiting room page
* Real-time queue position updates
* FIFO fairness
* WebSocket / SSE updates

Additional behavior:

* Google/Bing bots bypass queue
* Client IP preserved via `X-Forwarded-For`

---

## 4.1.4 Circuit Breaker

Protects origin servers from cascading failure.

States:

```
CLOSED → OPEN → HALF-OPEN
```

Trigger condition:

* > 50% error rate within 10 seconds.

Recovery strategy:

* Gradually reintroduce traffic
* Send health probes

Notifications sent via:

* Email
* Webhooks

---

# 4.2 Multi-Tenant Architecture

Isolation strategy:

* Redis key prefix per tenant
* PostgreSQL Row-Level Security
* Kubernetes NetworkPolicies
* Tenant audit logging

---

# 4.3 Dashboard (Single-Page Nerve Center)

## 4.3.1 Live Metrics

| Metric            | Display          | Source              | Update Rate |
| ----------------- | ---------------- | ------------------- | ----------- |
| Live visitors     | Animated counter | Redis pub/sub       | 1s          |
| Queue depth       | Live chart       | Redis queue         | 500ms       |
| System health     | Status indicator | Circuit breaker     | 2s          |
| Protected revenue | Currency counter | PostgreSQL + Stripe | 5s          |
| RPS               | Gauge            | Prometheus          | 1s          |
| Origin latency    | Line chart       | Proxy metrics       | 1s          |

---

## 4.3.2 Dashboard Features

* Multi-site selector
* Live RPS configuration
* Waiting room visual editor
* Traffic analytics
* Event mode activation
* Alert configuration
* Webhook integrations

---

# 4.4 Billing & Subscription

Billing is powered by **Stripe**.

Features:

* Subscription management
* Usage metering
* Overage billing
* Plan upgrades/downgrades
* Invoice downloads
* Event add-on purchase flow

Alerts trigger at:

* **80% usage**
* **95% usage**

---

# 5. Project Folder Structure

Monorepo using:

* **pnpm workspaces**
* **Turborepo**

```
fairflow/

apps/
  proxy/          # Go reverse proxy
  api/            # NestJS backend
  dashboard/      # Next.js frontend

packages/
  proto/
  types/

infra/
  k8s/
  terraform/

docs/
  architecture.md
  api-reference.md
  onboarding.md

scripts/
  setup.sh
  seed.ts

docker-compose.yml
pnpm-workspace.yaml
turbo.json
README.md
```

---

# 6. Database Schema

## Key Tables

| Table            | Key Fields                                   | Purpose                    |
| ---------------- | -------------------------------------------- | -------------------------- |
| tenants          | id, name, plan, stripe_customer_id           | Customer accounts          |
| sites            | id, tenant_id, domain, origin_url, rps_limit | Protected domains          |
| events           | id, site_id, starts_at, ends_at              | Event mode scheduling      |
| analytics_hourly | site_id, hour, req_count                     | Aggregated traffic metrics |
| ssl_certs        | domain, cert_pem, expires_at                 | SSL certificate storage    |

---

# Documentation

Additional documentation lives in:

```
/docs
```

Files include:

* `architecture.md`
* `api-reference.md`
* `onboarding.md`

---

# Development Setup

Local development uses **Docker Compose**.

```
docker-compose up
```

Production deployments run on **Kubernetes**.

---

# Future Roadmap

Planned improvements:

* Enterprise traffic policies
* Advanced bot detection
* Multi-region edge deployment
* AI traffic anomaly detection
* Enterprise contracts

---

# License

Proprietary — FairFlow Platform.
