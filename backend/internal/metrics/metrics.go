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

	// MQLag tracks the consumer lag (pending events) for the in-process bus (F009 skeleton).
	MQLag = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "echoline_mq_lag",
		Help: "Approximate MQ consumer lag (pending events not yet processed)",
	})

	// HotConversations tracks the number of conversations with active WS connections (E009 skeleton).
	HotConversations = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "echoline_hot_conversations",
		Help: "Number of conversations with at least one active WebSocket subscriber",
	})

	WSMessagesDropped = promauto.NewCounter(prometheus.CounterOpts{
		Name: "echoline_ws_messages_dropped_total",
		Help: "WebSocket messages dropped due to full send buffer",
	})
)

// Handler returns Prometheus scrape handler.
func Handler() http.Handler {
	return promhttp.Handler()
}

// ObserveMessageSend records send latency.
func ObserveMessageSend(start time.Time) {
	MessageSendLatency.Observe(time.Since(start).Seconds())
}
