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



TODO:

- [] http_rate_limit_rejections_total (only if rate limiting applies).