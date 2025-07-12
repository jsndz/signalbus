package metrics

import "github.com/prometheus/client_golang/prometheus"

var HttpRequestTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_total_requests",
		Help: "Total number of HTTP requests",

	},
	[]string{"endpoint","status","method"},
)


var HttpRequestDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name: "http_request_duration_seconds",
		Help: "Duration for http requests",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"endpoint", "method"},
)


var KafkaPublisherSuccess = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "kafka_publish_success_total",
		Help: "Total number of successful publishes",
	},
	[]string{"topic"},
)

var KafkaPublisherFailure = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "kafka_publish_failure_total",
		Help: "Total number of Failed publishes",
	},
	[]string{"topic"},

)

var KafkaSubscriberSuccess = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "kafka_subscriber_success_total",
		Help: "Total number of successful subscribes",
	},
	[]string{"topic"},

)

var KafkaSubscriberFailure = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "kafka_subscriber_failure_total",
		Help: "Total number of Failed subscribes",
	},
	[]string{"topic"},

)

var ExternalAPISuccess = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "external_api_success_total",
		Help: "Successful external API calls",
	},
	[]string{"provider", "service"},
)

var ExternalAPIFailure = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "external_api_failure_total",
		Help: "Failed external API calls",
	},
	[]string{"provider", "service"},
)

var NotificationSendDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "notification_send_duration_seconds",
		Help:    "Time taken to send notifications via external providers",
		Buckets: prometheus.DefBuckets, 
	},
	[]string{"provider", "service"},
)


func InitAPIMetrics(){
	prometheus.MustRegister(HttpRequestDuration)
	prometheus.MustRegister(HttpRequestTotal)
	prometheus.MustRegister(KafkaPublisherFailure)
	prometheus.MustRegister(KafkaPublisherSuccess)
}

func InitEmailMetrics(){
	prometheus.MustRegister(HttpRequestDuration)
	prometheus.MustRegister(HttpRequestTotal)
	prometheus.MustRegister(KafkaSubscriberSuccess)
	prometheus.MustRegister(KafkaSubscriberFailure)
	prometheus.MustRegister(ExternalAPIFailure)
	prometheus.MustRegister(ExternalAPISuccess)
	prometheus.MustRegister(NotificationSendDuration)
}

func InitSMSMetrics(){
	prometheus.MustRegister(HttpRequestDuration)
	prometheus.MustRegister(HttpRequestTotal)
	prometheus.MustRegister(KafkaSubscriberSuccess)
	prometheus.MustRegister(KafkaSubscriberFailure)
	prometheus.MustRegister(ExternalAPIFailure)
	prometheus.MustRegister(ExternalAPISuccess)
	prometheus.MustRegister(NotificationSendDuration)
}