package metrics

import "github.com/prometheus/client_golang/prometheus"

var HttpRequestsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests received",
	},
	[]string{"endpoint", "status", "method"},
)

var HttpRequestDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Duration of HTTP requests in seconds",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"endpoint", "method"},
)

var HttpErrorsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_errors_total",
		Help: "Total number of failed HTTP requests (4xx/5xx)",
	},
	[]string{"endpoint", "status", "method"},
)

var HttpRateLimitRejectionsTotal = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "http_rate_limit_rejections_total",
		Help: "Total number of HTTP requests rejected due to rate limiting",
	},
)



var NotificationsAttemptedTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "notifications_attempted_total",
		Help: "Total number of notifications attempted",
	},
	[]string{"channel", "status", "provider"},
)

var NotificationSendDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "notification_send_duration_seconds",
		Help:    "Time taken to send notifications via external providers",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"provider", "channel"},
)

var NotificationRetriesTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "notification_retries_total",
		Help: "Total number of notification retries",
	},
	[]string{"reason", "channel"},
)

var NotificationDLQTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "notification_dlq_total",
		Help: "Total number of notifications sent to DLQ",
	},
	[]string{"reason", "channel"},
)


var KafkaPublishFailureTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "kafka_publish_failure_total",
		Help: "Total number of failed Kafka publishes",
	},
	[]string{"topic"},
)

var KafkaSubscriberFailureTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "kafka_subscriber_failure_total",
		Help: "Total number of failed Kafka subscribes",
	},
	[]string{"topic"},
)

var KafkaConsumerLag = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "kafka_consumer_lag",
		Help: "Lag of Kafka consumer group per topic/partition",
	},
	[]string{"group", "topic", "partition"},
)

var KafkaRebalancesTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "kafka_rebalances_total",
		Help: "Number of Kafka consumer group rebalances",
	},
	[]string{"group"},
)



var ExternalAPISuccessTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "external_api_success_total",
		Help: "Total number of successful external API calls",
	},
	[]string{"provider", "service"},
)

var ExternalAPIFailureTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "external_api_failure_total",
		Help: "Total number of failed external API calls",
	},
	[]string{"provider", "service"},
)

var ExternalAPIDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "external_api_duration_seconds",
		Help:    "Duration of external API calls in seconds",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"provider", "service"},
)

func InitAPIMetrics() {
	prometheus.MustRegister(HttpRequestsTotal)
	prometheus.MustRegister(HttpRequestDuration)
	prometheus.MustRegister(HttpErrorsTotal)
	prometheus.MustRegister(HttpRateLimitRejectionsTotal)
}

func InitWorkerMetrics() {
	prometheus.MustRegister(NotificationsAttemptedTotal)
	prometheus.MustRegister(NotificationSendDuration)
	prometheus.MustRegister(NotificationRetriesTotal)
	prometheus.MustRegister(NotificationDLQTotal)
	prometheus.MustRegister(ExternalAPISuccessTotal)
	prometheus.MustRegister(ExternalAPIFailureTotal)
	prometheus.MustRegister(ExternalAPIDuration)
}

func InitKafkaMetrics() {
	prometheus.MustRegister(KafkaPublishFailureTotal)
	prometheus.MustRegister(KafkaSubscriberFailureTotal)
	prometheus.MustRegister(KafkaConsumerLag)
	prometheus.MustRegister(KafkaRebalancesTotal)
}


