package server

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/echoline/echoline/backend/internal/audit"
	"github.com/echoline/echoline/backend/internal/auth"
	"github.com/echoline/echoline/backend/internal/config"
	"github.com/echoline/echoline/backend/internal/conversation"
	"github.com/echoline/echoline/backend/internal/delivery"
	"github.com/echoline/echoline/backend/internal/eventbus"
	"github.com/echoline/echoline/backend/internal/message"
	"github.com/echoline/echoline/backend/internal/presence"
	"github.com/echoline/echoline/backend/internal/rate_limit"
	"github.com/echoline/echoline/backend/internal/realtime"
	"github.com/echoline/echoline/backend/internal/redisx"
	"github.com/echoline/echoline/backend/internal/sync"
	"github.com/echoline/echoline/backend/internal/user"
)

// Options configures optional server dependencies.
type Options struct {
	Redis *redisx.Client
}

// NewWithOptions creates a server with optional Redis-backed services.
func NewWithOptions(cfg config.Config, pool *pgxpool.Pool, logger *slog.Logger, redis *redisx.Client) *Server {
	return newServer(cfg, pool, logger, Options{Redis: redis})
}

func newServer(cfg config.Config, pool *pgxpool.Pool, logger *slog.Logger, opts Options) *Server {
	userRepo := user.NewRepository(pool)
	authSvc := auth.NewService(userRepo, cfg.JWTSecret)
	auditRepo := audit.NewRepository(pool)
	authSvc.SetLoginAuditor(auditRepo)

	convRepo := conversation.NewRepository(pool)
	msgRepo := message.NewRepository(pool)
	msgSvc := message.NewService(msgRepo, convRepo, nil)
	deliveryRepo := delivery.NewRepository(pool)

	memBus := eventbus.NewMemoryPublisher(256)
	memPub := eventbus.NewMemoryBytesPublisher(memBus)
	var publisher eventbus.BytesPublisher = memPub
	if cfg.KafkaBrokers != "" {
		kafkaPub := eventbus.NewKafkaPublisher(cfg.KafkaBrokers, eventbus.TopicMessageCreated)
		publisher = eventbus.NewFanoutPublisher(kafkaPub, memPub)
	}
	msgSvc.SetPublisher(publisher)

	var presenceTracker realtime.PresenceTracker
	var limiter rate_limit.Limiter
	if opts.Redis != nil {
		presenceTracker = presence.NewStore(opts.Redis, 0)
		limiter = rate_limit.NewRedisLimiter(opts.Redis)
	}

	rt := realtime.NewServer(authSvc, msgSvc, convRepo, deliveryRepo, presenceTracker, logger)

	return &Server{
		cfg:      cfg,
		pool:     pool,
		logger:   logger,
		auth:     authSvc,
		conv:     conversation.NewHandler(convRepo),
		msg:      message.NewHandler(msgSvc, convRepo),
		sync:     sync.NewHandler(convRepo, msgSvc),
		delivery: delivery.NewHandler(deliveryRepo, convRepo),
		realtime: rt,
		limiter:  limiter,
		memBus:   memBus,
	}
}

// Server is the HTTP API server.
type Server struct {
	cfg        config.Config
	pool       *pgxpool.Pool
	logger     *slog.Logger
	httpServer *http.Server
	auth       *auth.Service
	conv       *conversation.Handler
	msg        *message.Handler
	sync       *sync.Handler
	delivery   *delivery.Handler
	realtime   *realtime.Server
	limiter    rate_limit.Limiter
	memBus     *eventbus.MemoryPublisher
}

// MemoryBus exposes the in-process event bus for workers/tests.
func (s *Server) MemoryBus() *eventbus.MemoryPublisher {
	return s.memBus
}

// applyRateLimits wraps handlers with Redis rate limiting when configured.
func (s *Server) applyRateLimits(mux *http.ServeMux) {
	loginMW := rate_limit.Middleware(s.limiter, "login", 20, time.Minute, rate_limit.IPKey)
	sendMW := rate_limit.Middleware(s.limiter, "send", 120, time.Minute, rate_limit.PathKey)

	mux.Handle("POST /api/auth/login", loginMW(http.HandlerFunc(s.auth.HandleLogin)))
	mux.Handle("POST /api/conversations/{id}/messages", sendMW(auth.RequireAuth(s.auth, http.HandlerFunc(s.msg.HandleSend))))
}
