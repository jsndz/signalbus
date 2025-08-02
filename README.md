# Signalbus

**Signalbus** is a cloud-native, event-driven notification hub built with Go, Kafka, Docker, and cloud APIs (Mailgun, Twilio). It handles real-time email and SMS notifications through scalable microservices. Based on various events the data is put kafka through which the microservices get the data. Here microservices send emails and SMS. Its a simulation and a learning project.

---

![](public/1.png)
![](public/2.png)
![](public/3.png)
![](public/4.png)

---

## Features

- Microservices built in Go
- Email notifications via Mailgun
- SMS notifications via Twilio
- Kafka-based event distribution
- Dockerized services with Docker Compose
- Observability with Prometheus + Grafana
- CI/CD and deployment on Railway
- Fully testable via REST endpoints

---

## Architecture

```

Client Services (e.g., Order, Auth)
|
v
[ Notification API Service ]
|
v
Kafka Topics
|
v
[Email Worker] \[SMS Worker] \[Future Workers...]
|
v
Mailgun & Twilio

```

---

## Services

| Service          | Description                           |
| ---------------- | ------------------------------------- |
| Notification API | Accepts events and publishes to Kafka |
| Email Worker     | Listens to Kafka and sends emails     |
| SMS Worker       | Listens to Kafka and sends SMS        |
| Kafka/Zookeeper  | Message broker infrastructure         |
| Prometheus       | Metrics collector                     |
| Grafana          | Dashboard visualization               |

---

## Tech Stack

- **Go**
- **Kafka (segmentio/kafka-go)**
- **Docker / Docker Compose**
- **Mailgun API**
- **Twilio API**
- **PostgreSQL (optional for logging)**
- **Prometheus + Grafana**
- **Railway / GitHub Actions for CI/CD**

---

## Setup Instructions

### 1. Clone the Repo

```bash
git clone https://github.com/jsndz/signalbus.git
cd signalbus
```

### 2. Environment Variables

Rename `.env.example`to `.env` and replace the variables

### 3. Run Locally with Docker

```bash
docker-compose up --build
```

### 4. Trigger an Event

Test with Postman or curl:

```bash
curl -X POST http://localhost:8080/notify/signup \
     -H "Content-Type: application/json" \
     -d '{"email": "user@example.com", "phone": "+11234567890"}'
```

---

## Docker Setups

How to Run Kafka separately:

- Run the docker image of `Kafka` and `zoo-keeper` (image details in docker-compose.yml)

Run docker compose(recommended):

- cd notification_hub
- docker compose build
- docker compose up -d

See logs:

- docker compose logs -f api
- docker compose logs -f email
- docker compose logs -f sms

Running API and Worker Separately (Manual Build):

- docker build -f deployments/Dockerfile.api -t notification_hub:latest .
- docker run -p 8080:8080 notification_hub:latest

- docker build -f deployments/Dockerfile.email -t email:latest .
- docker run -p 3001:3001 email:latest

- docker build -f deployments/Dockerfile.sms -t sms:latest .
- docker run -p 3000:3000 sms:latest

Stop everything:

- docker compose down

Example curl request:

```sh
curl -X POST http://localhost:8080/api/notify/signup \
     -H "Content-Type: application/json" \
     -d '{"username":"jsn", "password":"qwerty","email": "user@example.com", "phone": "+11234567890" }'

```

Response Example:

```json
{
  "message": "Signup event sent to Kafka"
}
```

Logs:

```md
email-1 | 2025/07/10 09:54:26 jsn qwerty
sms-1 | 2025/07/11 14:52:38 SMS sent to +91XXXXXXXXXX! Message SID: SMfXXXXXXXXXXXXXXXXXXXXXXXXXXX
email-1 | 202
```

Topics Learned:

- Docker compose
- Twilio and Sendgrid Integration
- Kafka Producer and Consumer
- Kafka Retry Logics
- makefile
- Grafana + prometheus
- logging with zap

---

## Observability

- Visit Prometheus: `http://localhost:9090`
- Visit Grafana: `http://localhost:3002` (default login: `admin/admin`)

---

## Deployment

- Services are deployed on [Railway](https://railway.app)
- GitHub Actions handle CI/CD on push to `main`
- Kafka managed via Railway Add-on or Upstash

---

## Grafana Dashboard Layout for Prometheus Metrics

### HTTP Metrics (API Gateway / Services)

**Panel Group Title:** `HTTP Metrics`

- **HTTP Request Rate**
  **Query:**

  ```promql
  rate(http_total_requests[1m])
  ```

  **Panel Type:** Time series (lines per `endpoint`)

- **Requests by Status**
  **Query:**

  ```promql
  sum by (status) (http_total_requests)
  ```

  **Panel Type:** Bar chart or Pie chart

- **Average Request Duration**
  **Query:**

  ```promql
  rate(http_request_duration_seconds_sum[1m]) / rate(http_request_duration_seconds_count[1m])
  ```

  **Panel Type:** Time series

- **P95 Request Duration**
  **Query:**

  ```promql
  histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le, endpoint))
  ```

  **Panel Type:** Time series

- **Latency Distribution (Heatmap)**
  **Query:**

  ```promql
  http_request_duration_seconds_bucket
  ```

  **Panel Type:** Heatmap

---

### Kafka Publisher & Subscriber Metrics

**Panel Group Title:** `Kafka Metrics`

- **Kafka Publish Success Rate**
  **Query:**

  ```promql
  rate(kafka_publish_success_total[1m])
  ```

  **Panel Type:** Time series (lines per `topic`)

- **Kafka Publish Failure Rate**
  **Query:**

  ```promql
  rate(kafka_publish_failure_total[1m])
  ```

  **Panel Type:** Time series or Stat (red)

- **Kafka Subscribe Success Rate**
  **Query:**

  ```promql
  rate(kafka_subscriber_success_total[1m])
  ```

  **Panel Type:** Time series

- **Kafka Subscribe Failure Rate**
  **Query:**

  ```promql
  rate(kafka_subscriber_failure_total[1m])
  ```

  **Panel Type:** Time series or Stat (red)

---

### External API Reliability

**Panel Group Title:** `External API`

- **External API Success Rate**
  **Query:**

  ```promql
  rate(external_api_success_total[1m])
  ```

  **Panel Type:** Bar chart (per `provider`, `service`)

- **External API Failure Rate**
  **Query:**

  ```promql
  rate(external_api_failure_total[1m])
  ```

  **Panel Type:** Bar chart or Time series

- **External API Failure Ratio**
  **Query:**

  ```promql
  sum(rate(external_api_failure_total[5m])) / (sum(rate(external_api_failure_total[5m])) + sum(rate(external_api_success_total[5m])))
  ```

  **Panel Type:** Stat (Red if >10%)

---

### Notification Performance

**Panel Group Title:** `Notifications`

- **Avg Notification Send Duration**
  **Query:**

  ```promql
  rate(notification_send_duration_seconds_sum[1m]) / rate(notification_send_duration_seconds_count[1m])
  ```

  **Panel Type:** Time series (per `provider`, `service`)

- **P95 Notification Duration**
  **Query:**

  ```promql
  histogram_quantile(0.95, sum(rate(notification_send_duration_seconds_bucket[5m])) by (le, provider, service))
  ```

  **Panel Type:** Time series

- **Send Duration Heatmap**
  **Query:**

  ```promql
  notification_send_duration_seconds_bucket
  ```

  **Panel Type:** Heatmap

---

## Health & Overview Summary

**Panel Group Title:** `System Summary`

- **Total HTTP Requests**
  **Query:**

  ```promql
  sum(http_total_requests)
  ```

  **Panel Type:** Stat

- **Kafka Failures (All)**
  **Query:**

  ```promql
  sum(rate(kafka_publish_failure_total[1m]) + rate(kafka_subscriber_failure_total[1m]))
  ```

  **Panel Type:** Stat

- **External API Error Rate (Total)**
  **Query:**

  ```promql
  sum(rate(external_api_failure_total[5m])) / (sum(rate(external_api_failure_total[5m])) + sum(rate(external_api_success_total[5m])))
  ```

  **Panel Type:** Stat

---

## Roadmap

- [x] Signup event handling
- [x] Email/SMS delivery
- [x] Prometheus + Grafana integration
- [x] Cloud deployment
- [ ] Add retry logic with Dead Letter Queues
- [ ] In-app notifications (web dashboard)
- [ ] Notification history with PostgreSQL
- [ ] OpenTelemetry tracing

---

## Contributing

Pull requests are welcome! Open an issue first for feedback or discussion.

---
