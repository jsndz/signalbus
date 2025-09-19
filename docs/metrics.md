Cardinality means the number of distinct (unique) values something can take.

Examples:

In a set {1, 2, 3, 3, 2}, the cardinality is 3 (because the unique values are {1, 2, 3}).

In metrics high cardinality is problem for because lets say the label "user_id" which has very high cardinality cause everything is unique.

If you accidentally add something like user_id as a label, and you have 1M users then Suddenly Prometheus has 1M+ time series to store and query.

So if you have:

2 methods (GET, POST)
10 routes (/login, /signup, …)
5 status codes (200, 400, 401, 500, 503)
That’s 2 × 10 × 5 = 100 unique time series for just this one metric.

So ,
Rule of thumb:
Low cardinality labels = good (route, code, provider, channel).
High cardinality labels = dangerous (user_id, email, notification_id).

Aim for O(10–100) unique values per label; O(1000+) is risky unless you know memory sizing.


RED and USE:
RED (for services, like your API or worker)

Think: “How is my service behaving from the user’s perspective?”

R = Rate → how many requests per second are coming in.

Example: http_requests_total{route="/notify"}

Tells you: is traffic normal, too low, or spiking?

E = Errors → how many of those requests failed.

Example: notification_send_errors_total{channel="email", provider="sendgrid"}

Tells you: are users seeing errors more than usual?

D = Duration → how long requests take (latency).

Example: http_request_duration_seconds{route="/notify"}

Tells you: are users waiting too long for a response?

 Together, RED = user experience view. If rate drops, errors spike, or duration is high → users are unhappy.

 USE (for resources, like Kafka, DB, CPU, memory)

Think: “Are my servers/infrastructure healthy?”

U = Utilization → how busy is the resource.

Example: CPU at 80%, DB connections 90% full.

If utilization is always near 100% → bottleneck risk.

S = Saturation → how much extra work is piling up.

Example: Kafka consumer lag = 10,000 messages waiting.

Example: DB connection pool has 50 clients waiting.

Tells you: the system is overloaded, not just busy.

E = Errors → any failures at the infra level.

Example: DB timeouts, Kafka rebalances, network errors.

Tells you: infra itself is struggling.

 Together, USE = system health view. Even if RED looks good, USE can show hidden problems (like lag building up).



## **1. Why are metrics used?**

Metrics are a way of **quantifying system behavior** so that humans and automated systems can observe, alert, and act.
They answer **what’s happening in your system** — both when things are working and when they’re not.

Think of metrics as **structured, numeric observations**. For example:

* Instead of “API feels slow,” a metric says “P95 latency = 1.2s.”
* Instead of “Some notifications didn’t go through,” a metric says “5% of SMS sends failed with error=provider\_down.”

Metrics help you:

* **Detect problems** (alerting when something is wrong).
* **Understand performance** (latency, throughput, error rate).
* **Plan capacity** (how many requests, workers, or resources you need).
* **Debug incidents** (which part of the system failed, where to look).

So yes — your phrasing is right: metrics help us know what’s correct and what’s not, but also *how well* the system is behaving, not just whether it’s broken.

---

## **2. On what basis should I create metrics?**

There are a few widely adopted frameworks to guide metric design:

### **RED Method** (popular for user-facing systems, APIs, microservices)

* **R**ate → How many requests/operations are happening?
* **E**rrors → How many of them are failing?
* **D**uration → How long do they take?

This ensures you always capture the three critical signals of request/response systems.

### **USE Method** (for infrastructure/resources: CPUs, workers, queues, Kafka)

* **U**tilization → How busy is the resource? (e.g., CPU 70%)
* **S**aturation → How “full” is it? (e.g., queue length, consumer lag)
* **E**rrors → Failures in the resource (e.g., worker retries, disk errors).

Rule of thumb:

* Use **RED** for APIs, user requests, background jobs.
* Use **USE** for infrastructure, queues, workers, Kafka.

---

## **3. Is there a formula for designing metrics?**

Not an exact formula, but a **process**:

1. **Identify the component/layer.** (API, worker, Kafka, DB, etc.)
2. **Ask: what does success mean here? What does failure look like?**

   * API → success = fast 200 OK responses, failure = errors or slowness.
   * Worker → success = notification delivered, failure = retries, DLQ.
   * Kafka → success = consumers keeping up, failure = lag, rebalance storms.
3. **Choose RED or USE framework.**

   * API → RED.
   * Worker → hybrid (RED for send attempts, USE for worker resources).
   * Kafka → USE.
4. **Define metric name, type, labels.**

   * Be consistent in naming (e.g., `http_requests_total`, `worker_retries_total`).
   * Labels let you slice/dice: tenant, provider, status.
   * Pick type: counter, gauge, histogram.
5. **Ensure actionability.**

   * Every metric should answer: *“What decision does this help me make?”*
   * If you can’t act on it, maybe it’s not worth measuring.

---

## **4. How do industry-grade systems design metrics?**

Industry practices generally follow these steps:

1. **Start with SLOs/SLIs** (Service Level Objectives/Indicators):

   * Example: “99.9% of notifications delivered within 5 seconds.”
   * Metrics are designed to measure whether you’re meeting that.

2. **Standardize metric naming and labels.**

   * Teams agree on a schema (`_total`, `_duration_seconds`, `_errors_total`).
   * Consistent labels (`tenant`, `status`, `provider`) make dashboards reusable.

3. **Follow RED/USE + custom business metrics.**

   * RED/USE covers system health.
   * Business-specific metrics (e.g., “notifications sent per tenant per day”) cover domain needs.

4. **Minimal but sufficient.**

   * Too few metrics → blind spots.
   * Too many metrics → hard to monitor.
   * The balance is: *only keep metrics that are useful for alerting, debugging, or capacity planning.*

5. **Integrate with monitoring stack.**

   * Collect with Prometheus/OpenTelemetry.
   * Visualize in Grafana.
   * Alert in PagerDuty/Slack.

---

## **5. Your Example Layers (with RED/USE lens)**

### **API Layer (`notification_api`) → RED**

* `http_requests_total` (Counter) — labels: method, route, status.
* `http_request_duration_seconds` (Histogram) — labels: route, method.
* `http_errors_total` (Counter) — labels: status (4xx, 5xx).
* `http_rate_limit_rejections_total` (Counter).

### **Worker Layer (`sms_worker`, `email_worker`) → RED + USE**

* `notifications_attempted_total` (Counter) — labels: channel, provider, tenant, status.
* `notification_send_duration_seconds` (Histogram) — labels: channel, provider.
* `notification_retries_total` (Counter) — labels: reason, channel.
* `notification_dlq_total` (Counter) — labels: channel, reason.

### **Kafka Layer → USE**

* `kafka_consumer_lag` (Gauge) — labels: group, topic, partition.
* `kafka_rebalances_total` (Counter) — labels: group.
* `kafka_commit_latency_seconds` (Histogram).

---

