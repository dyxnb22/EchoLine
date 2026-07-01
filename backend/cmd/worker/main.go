package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/echoline/echoline/backend/internal/config"
	"github.com/echoline/echoline/backend/internal/eventbus"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	memBus := eventbus.NewMemoryPublisher(256)
	go consumeMemory(ctx, logger, memBus)

	if cfg.KafkaBrokers != "" {
		consumer := eventbus.NewKafkaConsumer(cfg.KafkaBrokers, eventbus.TopicMessageCreated, "echoline-worker")
		defer consumer.Close()
		go consumeKafka(ctx, logger, consumer)
		logger.Info("kafka consumer started", "brokers", cfg.KafkaBrokers)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	logger.Info("worker started")
	<-stop
	cancel()
	time.Sleep(100 * time.Millisecond)
	logger.Info("worker stopped")
}

func consumeMemory(ctx context.Context, logger *slog.Logger, bus *eventbus.MemoryPublisher) {
	for {
		select {
		case <-ctx.Done():
			return
		case evt, ok := <-bus.C():
			if !ok {
				return
			}
			handleEvent(logger, evt.Type, evt.Payload)
		}
	}
}

func consumeKafka(ctx context.Context, logger *slog.Logger, consumer *eventbus.KafkaConsumer) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := consumer.Read(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				logger.Error("kafka read", "error", err)
				time.Sleep(time.Second)
				continue
			}
			handleEvent(logger, msg.Topic, msg.Value)
		}
	}
}

func handleEvent(logger *slog.Logger, topic string, payload []byte) {
	switch topic {
	case eventbus.TopicMessageCreated:
		evt, err := eventbus.DecodeMessageCreated(payload)
		if err != nil {
			logger.Error("decode message.created", "error", err)
			return
		}
		logger.Info("message.created consumed",
			"message_id", evt.ID,
			"conversation_id", evt.ConversationID,
			"seq", evt.Seq,
			"preview", truncate(evt.Body, 32),
		)
	default:
		logger.Info("event consumed", "type", topic, "bytes", len(payload))
	}
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
