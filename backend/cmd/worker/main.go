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
	"github.com/echoline/echoline/backend/internal/eventbus"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	_, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	bus := eventbus.NewMemoryPublisher(256)
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for evt := range bus.C() {
			logger.Info("event consumed", "type", evt.Type)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	logger.Info("worker started")
	<-stop
	cancel()
	time.Sleep(100 * time.Millisecond)
	logger.Info("worker stopped")
}
