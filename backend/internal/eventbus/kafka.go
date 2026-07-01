package eventbus

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

// KafkaPublisher publishes events to Kafka/Redpanda.
type KafkaPublisher struct {
	writers map[string]*kafka.Writer
}

// NewKafkaPublisher creates writers for the given topics.
func NewKafkaPublisher(brokers string, topics ...string) *KafkaPublisher {
	addrs := strings.Split(brokers, ",")
	p := &KafkaPublisher{writers: make(map[string]*kafka.Writer, len(topics))}
	for _, topic := range topics {
		p.writers[topic] = &kafka.Writer{
			Addr:         kafka.TCP(addrs...),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireOne,
			BatchTimeout: 10 * time.Millisecond,
		}
	}
	return p
}

// Publish writes an event to a topic.
func (p *KafkaPublisher) Publish(ctx context.Context, topic string, payload []byte) error {
	w, ok := p.writers[topic]
	if !ok {
		return fmt.Errorf("unknown topic: %s", topic)
	}
	return w.WriteMessages(ctx, kafka.Message{
		Key:   nil,
		Value: payload,
	})
}

// Close closes all writers.
func (p *KafkaPublisher) Close() error {
	var first error
	for _, w := range p.writers {
		if err := w.Close(); err != nil && first == nil {
			first = err
		}
	}
	return first
}

// KafkaConsumer reads events from a topic.
type KafkaConsumer struct {
	reader *kafka.Reader
}

// NewKafkaConsumer creates a consumer for a topic/group.
func NewKafkaConsumer(brokers, topic, groupID string) *KafkaConsumer {
	addrs := strings.Split(brokers, ",")
	return &KafkaConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: addrs,
			Topic:   topic,
			GroupID: groupID,
		}),
	}
}

// Read blocks until the next message or context cancel.
func (c *KafkaConsumer) Read(ctx context.Context) (kafka.Message, error) {
	return c.reader.ReadMessage(ctx)
}

// Close closes the consumer.
func (c *KafkaConsumer) Close() error {
	return c.reader.Close()
}

// LagEstimate returns the approximate consumer lag from reader stats.
// This reflects messages not yet committed by this reader instance.
func (c *KafkaConsumer) LagEstimate() int64 {
	return c.reader.Stats().Lag
}
