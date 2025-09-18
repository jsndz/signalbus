# Why This Notification System Is Useful

This project is an **open-source, self-hostable notification system** designed to manage multi-channel communication (Email, SMS, Push, Chat) in a consistent, reliable, and observable way.

It is built to be flexible enough for **startups, enterprises, developer agencies, and open-source communities**.  
The system provides the building blocks to handle notifications at scale with reliability, observability, and security in mind.

---

## Example Use Cases

1. **Startups**  
   - Centralize notification logic in one service instead of duplicating it across backend services.  
   - Send onboarding emails, password resets, and order confirmations reliably.  
   - Integrate with Mailgun, Twilio, or SMTP quickly.  
   - Run entirely on Docker Compose or Kubernetes.

2. **Agencies / Consultancies**  
   - Manage multiple client apps from a single deployment.  
   - Keep each client’s data, templates, and API keys isolated.  
   - Avoid the operational overhead of running multiple servers.

3. **Enterprises**  
   - Isolate notifications for internal teams (HR, Sales, Support, Marketing).  
   - Track logs and delivery attempts per department.  
   - Enforce quotas and rate limits to prevent one team from overwhelming the system.

4. **Open-Source Communities & Platforms**  
   - Power hackathons, forums, or community apps.  
   - Each organization or event is a tenant with its own API keys and templates.  
   - One shared infrastructure can safely support many organizers.

---

## Key Features

### 1. Multi-Channel Support
- **Email**: SMTP or API providers (Mailgun, SendGrid, SES).  
- **SMS**: Twilio, Vonage, etc.  
- **Push**: Firebase Cloud Messaging (FCM), APNs, Web Push.  
- **Chat & Webhooks**: Slack, Discord, custom integrations.  

### 2. Kafka-Powered Event Delivery
- Publish/subscribe model for decoupled services.  
- Retry + Dead Letter Queue (DLQ) support.  
- Consumer scaling with Kafka consumer groups.  

### 3. Multi-Tenancy
- Separate tenants for apps, clients, or teams.  
- Each tenant has:
  - API keys
  - Templates
  - Logs and delivery attempts
  - Quotas and limits
- Prevents accidental data mixing.  
- Allows single deployment to serve multiple apps or clients.  

### 4. Template Management
- Store templates in Postgres.  
- Support for Go templates and MJML (HTML emails).  
- Localization support (per tenant and per channel).  
- User preferences: opt-outs, do-not-disturb windows.  

### 5. Reliability & Observability
- **Retries**: Exponential backoff with jitter.  
- **Dead Letter Queues**: Handle failed notifications gracefully.  
- **Metrics**: Exposed via Prometheus (`/metrics` endpoint).  
- **Dashboards**: Grafana dashboards for requests, lag, failures, latency.  
- **Tracing**: OpenTelemetry tracing across API, Kafka, and workers.  

### 6. Security & API Design
- REST API with OpenAPI/Swagger documentation.  
- API keys scoped per tenant, with rotation support.  
- HMAC signing for message integrity.  
- Idempotency keys to avoid duplicate sends.  
- Rate limiting (short-term burst and long-term quotas).  

### 7. Persistence & Querying
- Postgres schema for notifications, delivery attempts, tenants, API keys, templates.  
- APIs to search notification history, filter by tenant, channel, or status.  
- Redrive failed notifications from DLQ.  

### 8. Deployment & Ops
- Lightweight Go binaries with small Docker images.  
- `docker-compose` for local setup with Kafka, Postgres, Prometheus, Grafana, Mailpit.  
- Helm charts for Kubernetes deployments.  
- CI/CD pipelines with GitHub Actions (lint, test, build, publish, scan).  

### 9. Developer Experience
- Reusable Go libraries:  
  - `gomailer` for email providers  
  - `gosms` for SMS providers  
  - Future: `gopush` for push providers  
- Example clients in Go, Node.js, and TypeScript.  
- Easy local testing with Mailpit (fake SMTP inbox).  
- Contribution-friendly (Apache-2.0 license, contributing guide).  

---

## Multi-Tenancy in Action

Even in a self-hosted setup, multi-tenancy provides flexibility:

- **Startup Example**:  
  Acme Corp has two apps—Customer App and Admin Dashboard.  
  - Customer App → Tenant 1  
  - Admin Dashboard → Tenant 2  
  Notifications, templates, and logs are isolated.  

- **Agency Example**:  
  A consultancy manages 5 clients. Each client gets their own tenant.  
  Logs, quotas, and API keys are separated.  

- **Enterprise Example**:  
  HR, Marketing, and Support teams use the same deployment.  
  Each team’s data and templates remain private.  

- **Community Example**:  
  A hackathon platform hosts the system. Each hackathon organizer is a tenant.  
  They share infrastructure but stay isolated.

---

## Why This Matters

* **For small apps**: Use a single `default` tenant.
* **For growing teams**: Isolate environments or apps easily.
* **For agencies/enterprises**: Manage multiple clients or departments from one deployment.
* **For communities**: Shared infra, isolated tenants.

This flexibility ensures the system can start simple but scale into **multi-client SaaS-like setups** without a rewrite.


