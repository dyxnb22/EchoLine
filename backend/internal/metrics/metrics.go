package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	HTTPRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "echoline_http_requests_total",
		Help: "Total HTTP requests",
	}, []string{"method", "path", "status"})

	WSConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "echoline_ws_connections",
		Help: "Active WebSocket connections",
	})

	MessageSendLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "echoline_message_send_duration_seconds",
		Help:    "Message send handler latency",
		Buckets: prometheus.DefBuckets,
	})

	OutboxPending = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "echoline_outbox_pending",
		Help: "Pending outbox events (updated by worker)",
	})

	MQEventsConsumed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "echoline_mq_events_consumed_total",
		Help: "Events consumed from bus",
	}, []string{"topic"})
)

// Handler returns Prometheus scrape handler.
func Handler() http.Handler {
	return promhttp.Handler()
}

// ObserveMessageSend records send latency.
func ObserveMessageSend(start time.Time) {
	MessageSendLatency.Observe(time.Since(start).Seconds())
}
