# Signalbus — Project Overview and Skills Summary

## Executive Summary
Signalbus is a cloud‑native, event‑driven notification hub built in Go. It publishes user and system events to Kafka and fan‑outs to channel workers (Email/SMS) that deliver through Mailgun and Twilio. The system is containerized with Docker Compose, observable with Prometheus + Grafana, and deployable to cloud (Railway). It demonstrates production‑grade patterns: idempotency, retries with DLQ, rate limiting, multi‑tenancy, structured logging, metrics, and a growing persistence/query surface.

## Key Achievements (from README and todos)
- Implemented Notification API with routes like `/notify/signup`, publishing to Kafka topics.
- Built `email_worker` and `sms_worker` services consuming Kafka and delivering via Mailgun and Twilio.
- End‑to‑end flows verified: API → Kafka → Email/SMS workers, with real provider calls.
- Containerized all services and infrastructure via Docker Compose; individual Dockerfiles per service.
- Observability stack: Prometheus scraping all services and Grafana dashboards for API, channel success/failure, and Kafka lag.
- Cloud deployment to Railway with CI/CD via GitHub Actions (build, test, deploy pipeline).
- Implemented:
  - Short‑term and long‑term rate limiting.
  - Redis integration for caching/idempotency support.
  - Persistence layer and repositories for notifications, attempts, tenants, templates, API keys.
  - Idempotency strategies (payload keys + transport‑level insights; docs clarify producer idempotency vs semantic idempotency).
  - Dead Letter Queue (DLQ) and retry patterns.
- Multi‑tenancy groundwork: tenant column propagation, scoped queries, quotas.
- Template engine for email/SMS with localization support.
- OpenTelemetry tracing foundations and standardized metrics naming/labels; Grafana dashboards provided.

## Skills and Expertise Demonstrated
- Backend Engineering (Go)
  - HTTP servers, routing, middlewares (auth, rate limiting, idempotency, metrics).
  - Strong typing and package design under `pkg/` (gomailer, gosms, kafka, templates, repositories, models).
  - Robust error handling, structured logging with zap, configuration management.
- Distributed Systems & Messaging
  - Kafka producers/consumers, topics, consumer groups, rebalance awareness.
  - Delivery semantics, idempotency keys, DLQ/redrive design.
  - Event‑driven architecture and fan‑out patterns.
- Cloud & DevOps
  - Docker/Docker Compose, multi‑stage Dockerfiles, container orchestration locally.
  - CI/CD with GitHub Actions; environment configuration and secret handling.
  - Railway deployment workflows; production‑style configs.
- Observability & Reliability
  - Prometheus metrics (RED/USE methodology, low‑cardinality label discipline).
  - Grafana dashboards for API throughput/latency, provider success rates, Kafka consumer lag.
  - OpenTelemetry tracing design across services; span modeling.
  - Logging with correlation IDs and structured fields.
- Data & Persistence
  - PostgreSQL schema design for notifications, delivery attempts, templates, tenants, and API keys.
  - Redis for caching/idempotency; repository pattern for data access.
  - Indexing strategies and pagination for query surfaces.
- Email & SMS Domain Knowledge
  - Email deliverability (SPF, DKIM, DMARC), SMTP vs API providers; TLS/STARTTLS.
  - SMS normalization/validation to E.164, delivery receipts (DLR) via webhooks.
  - Provider integrations: Mailgun and Twilio drivers and operational behavior.
- Security & Multi‑tenancy
  - API keys and scope‑based access.
  - Per‑tenant quotas and rate limiting; tenant scoping across reads/writes.
  - Idempotency keys and deduplication strategies.

## Tech Stack (with rationale)
- Language: Go (performance, concurrency, strong ecosystem for microservices).
- Messaging: Kafka via `segmentio/kafka-go` (event distribution, consumer groups, scalability).
- Containers: Docker + Docker Compose (local reproducibility, multi‑service orchestration).
- Email: Mailgun API; SMTP driver in `pkg/gomailer` (extensible provider abstraction).
- SMS: Twilio API via `pkg/gosms` (number normalization, validation, DLR groundwork).
- Datastores: PostgreSQL (consistency, relational queries), Redis (cache/idempotency store).
- Observability: Prometheus (metrics), Grafana (dashboards), OpenTelemetry (tracing), zap (logs).
- Deployment: Railway (PaaS), GitHub Actions (CI/CD pipeline).

## Architecture (high level)
- Notification API (`cmd/notification_api`)
  - Endpoints: publish topic events, policy‑driven fan‑out (`/notify`, `/publish/:topic`).
  - Middlewares: API key auth, rate limiting (short/long term), idempotency, metrics.
  - Produces events to Kafka with keys to preserve per‑user ordering.
- Workers
  - Email Worker (`cmd/email_worker`): consumes `notifications.email`/`user_signup`, renders templates, sends via gomailer, retries with exponential backoff + jitter, DLQ on permanent failure.
  - SMS Worker (`cmd/sms_worker`): consumes `notifications.sms`, normalizes to E.164, sends via gosms, per‑tenant throttling, retries + DLQ; DLR webhook planned in API.
- Persistence & Repos (`pkg/repositories`, `pkg/models`)
  - Entities for notifications, delivery attempts, API keys, tenants, templates; search and pagination endpoints.
- Templates (`pkg/templates`)
  - Channel‑aware rendering, localization, policy engine to map topic → channels honoring preferences.
- Observability
  - Metrics across API/workers (standardized names), Grafana dashboards, tracing spans across API → Kafka → worker → provider.

## Reliability Patterns Implemented
- Idempotency: semantic keys in payload; understanding of producer‑level idempotency trade‑offs with `segmentio/kafka-go`.
- Retries & DLQ: exponential backoff with jitter, DLQ for permanent failures; future redrive endpoint.
- Rate Limiting: short‑term middleware and long‑term quotas; per‑tenant throttling for SMS.
- Kafka Consumer Health: awareness of lag, rebalances; planned metrics `kafka_consumer_lag`, `kafka_rebalances_total`.

## Metrics Strategy (RED/USE highlights)
- API: `http_requests_total`, `http_request_duration_seconds`, `http_errors_total`, `http_rate_limit_rejections_total`.
- Workers: `notifications_attempted_total`, `notification_send_duration_seconds`, `notification_retries_total`, `notification_dlq_total` with low‑cardinality labels: route, code, provider, channel, tenant, status.
- Kafka: `kafka_consumer_lag`, `kafka_rebalances_total`, `kafka_commit_latency_seconds`.

## Roadmap (selected next steps)
- Add DLR webhook to process SMS delivery receipts and update statuses.
- Finish migrations and redrive API; add openapi/swagger UI exposure.
- Expand tracing and add alerting rules for failure rates and consumer lag.
- Extend providers (e.g., SendGrid, Vonage) and push channels (FCM/APNs/Web Push).
- Packaging for production: distroless images, non‑root, Helm chart, examples.

## How to Talk About This Project (skills angle)
- Designed and shipped an event‑driven notifications platform in Go with Kafka, integrating Mailgun and Twilio, with Dockerized microservices and cloud CI/CD.
- Implemented reliability patterns (idempotency, retries, DLQ), multi‑tenant rate limiting, and standardized observability with Prometheus/Grafana and OpenTelemetry.
- Built reusable libraries (`pkg/gomailer`, `pkg/gosms`) and a clean repository/template architecture enabling future channels and providers.
