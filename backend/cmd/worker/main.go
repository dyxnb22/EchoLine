package main

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
	"github.com/echoline/echoline/backend/internal/migrate"
	"github.com/echoline/echoline/backend/internal/outbox"
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
	memBus := eventbus.NewMemoryPublisher(256)
	memPub := eventbus.NewMemoryBytesPublisher(memBus)

	var kafkaPub eventbus.BytesPublisher
	if cfg.KafkaBrokers != "" {
		kp := eventbus.NewKafkaPublisher(cfg.KafkaBrokers, eventbus.TopicMessageCreated)
		kafkaPub = kp
		defer kp.Close()
	}

	drainer := outbox.NewPublisher(outboxRepo, kafkaPub, memPub, logger)
	go drainer.Run(ctx)

	go func() {
		for evt := range memBus.C() {
			logger.Info("memory event consumed", "type", evt.Type, "bytes", len(evt.Payload))
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
