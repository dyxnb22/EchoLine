package worker

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/echoline/echoline/backend/internal/config"
	"github.com/echoline/echoline/backend/internal/db"
	"github.com/echoline/echoline/backend/internal/eventbus"
	"github.com/echoline/echoline/backend/internal/metrics"
	"github.com/echoline/echoline/backend/internal/migrate"
	"github.com/echoline/echoline/backend/internal/outbox"
	"github.com/echoline/echoline/backend/internal/push"
	"github.com/echoline/echoline/backend/internal/search"
	"github.com/echoline/echoline/backend/internal/webhook"
	workerpkg "github.com/echoline/echoline/backend/internal/worker"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := migrate.Up(ctx, cfg.DatabaseURL); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer pool.Close()

	outboxRepo := outbox.NewRepository(pool)
	searchRepo := search.NewRepository(pool)
	memBus := eventbus.NewMemoryPublisher(256)
	memPub := eventbus.NewMemoryBytesPublisher(memBus)

	var kafkaPub eventbus.BytesPublisher
	var kafkaConsumer *eventbus.KafkaConsumer
	if cfg.KafkaBrokers != "" {
		kp := eventbus.NewKafkaPublisher(cfg.KafkaBrokers, eventbus.TopicMessageCreated)
		kafkaPub = kp
		defer kp.Close()

		// F009: consumer for lag tracking.
		kafkaConsumer = eventbus.NewKafkaConsumer(cfg.KafkaBrokers, eventbus.TopicMessageCreated, "echoline-lag-probe")
		defer kafkaConsumer.Close()
	}

	drainer := outbox.NewPublisher(outboxRepo, kafkaPub, memPub, logger)
	go drainer.Run(ctx)

	msgHandler := workerpkg.NewMessageCreatedHandler(searchRepo, logger)
	pushRepo := push.NewRepository(pool)
	pushWorker := push.NewWorker(pushRepo, push.NewMockProvider(logger), logger)
	fanoutWorker := workerpkg.NewFanoutWorker(pool, pushWorker, logger)
	webhookRepo := webhook.NewRepository(pool)
	webhookDispatcher := webhook.NewDispatcher(cfg.WebhookURL)
	webhookRetry := webhook.NewRetryWorker(webhookRepo, webhookDispatcher)

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if n, err := webhookRetry.RunOnce(ctx); err != nil {
					logger.Warn("webhook retry", "error", err)
				} else if n > 0 {
					logger.Info("webhook retry delivered", "count", n)
				}
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if count, err := outboxRepo.CountPending(ctx); err == nil {
					metrics.OutboxPending.Set(float64(count))
				}
				// F009: update Kafka lag metric when broker is configured.
				if kafkaConsumer != nil {
					metrics.MQLag.Set(float64(kafkaConsumer.LagEstimate()))
				}
			}
		}
	}()

	go func() {
		for evt := range memBus.C() {
			if evt.Type != eventbus.TopicMessageCreated {
				continue
			}
			if err := msgHandler.Handle(ctx, evt.Payload); err != nil {
				logger.Error("message.created handler", "error", err)
			}
			if err := fanoutWorker.Handle(ctx, evt.Payload); err != nil {
				logger.Error("fanout worker", "error", err)
			}
			logger.Info("notification worker processed", "topic", evt.Type)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	logger.Info("worker started", "kafka", cfg.KafkaBrokers != "")
	<-stop
	cancel()
	time.Sleep(200 * time.Millisecond)
	logger.Info("worker stopped")
}
