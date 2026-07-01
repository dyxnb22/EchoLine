package server

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/echoline/echoline/backend/internal/admin"
	"github.com/echoline/echoline/backend/internal/audit"
	"github.com/echoline/echoline/backend/internal/auth"
	"github.com/echoline/echoline/backend/internal/block"
	"github.com/echoline/echoline/backend/internal/cache"
	"github.com/echoline/echoline/backend/internal/config"
	"github.com/echoline/echoline/backend/internal/conversation"
	"github.com/echoline/echoline/backend/internal/delivery"
	"github.com/echoline/echoline/backend/internal/device"
	"github.com/echoline/echoline/backend/internal/eventbus"
	"github.com/echoline/echoline/backend/internal/media"
	"github.com/echoline/echoline/backend/internal/message"
	"github.com/echoline/echoline/backend/internal/notification"
	"github.com/echoline/echoline/backend/internal/outbox"
	"github.com/echoline/echoline/backend/internal/pin"
	"github.com/echoline/echoline/backend/internal/presence"
	"github.com/echoline/echoline/backend/internal/rate_limit"
	"github.com/echoline/echoline/backend/internal/realtime"
	"github.com/echoline/echoline/backend/internal/redisx"
	"github.com/echoline/echoline/backend/internal/report"
	"github.com/echoline/echoline/backend/internal/search"
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
	outboxRepo := outbox.NewRepository(pool)
	attachmentRepo := media.NewRepository(pool)
	searchRepo := search.NewRepository(pool)
	cursorRepo := sync.NewCursorRepository(pool)
	msgRepo := message.NewRepository(pool, outboxRepo)
	msgSvc := message.NewService(msgRepo, convRepo, attachmentRepo, nil)
	deliveryRepo := delivery.NewRepository(pool)

	memBus := eventbus.NewMemoryPublisher(256)

	var presenceTracker realtime.PresenceTracker
	var limiter rate_limit.Limiter
	var listCache *cache.ConversationListCache
	if opts.Redis != nil {
		presenceTracker = presence.NewStore(opts.Redis, 0)
		limiter = rate_limit.NewRedisLimiter(opts.Redis)
		listCache = cache.NewConversationListCache(opts.Redis)
	}

	rt := realtime.NewServer(authSvc, msgSvc, convRepo, deliveryRepo, presenceTracker, logger)
	rt.SetHubMetrics()

	convHandler := conversation.NewHandler(convRepo)
	convHandler.SetListCache(listCache)

	var mediaHandler *media.Handler
	if cfg.S3Endpoint != "" {
		if mediaClient, err := media.NewClient(cfg); err == nil {
			mediaHandler = media.NewHandler(mediaClient, attachmentRepo)
		} else {
			logger.Warn("media client unavailable", "error", err)
		}
	}

	pinRepo := pin.NewRepository(pool)
	blockRepo := block.NewRepository(pool)
	reportRepo := report.NewRepository(pool)
	notifRepo := notification.NewRepository(pool)

	return &Server{
		cfg:          cfg,
		pool:         pool,
		logger:       logger,
		auth:         authSvc,
		conv:         convHandler,
		msg:          message.NewHandler(msgSvc, convRepo, attachmentRepo, auditRepo),
		sync:         sync.NewHandler(convRepo, msgSvc, cursorRepo),
		search:       search.NewHandler(searchRepo),
		delivery:     delivery.NewHandler(deliveryRepo, convRepo),
		realtime:     rt,
		limiter:      limiter,
		memBus:       memBus,
		outboxRepo:   outboxRepo,
		media:        mediaHandler,
		pin:          pin.NewHandler(pinRepo, convRepo),
		block:        block.NewHandler(blockRepo),
		report:       report.NewHandler(reportRepo, convRepo),
		notification: notification.NewHandler(notifRepo),
		adminHandler: admin.NewHandler(pool, authSvc),
		dlqHandler:   outbox.NewDLQHandler(pool),
		userRepo:     userRepo,
		profileRepo:  user.NewProfileRepository(pool),
		deviceH:      device.NewHandler(pool),
		mute:         conversation.NewMuteHandler(pool, convRepo),
	}
}

// Server is the HTTP API server.
type Server struct {
	cfg          config.Config
	pool         *pgxpool.Pool
	logger       *slog.Logger
	httpServer   *http.Server
	auth         *auth.Service
	conv         *conversation.Handler
	msg          *message.Handler
	sync         *sync.Handler
	search       *search.Handler
	delivery     *delivery.Handler
	realtime     *realtime.Server
	limiter      rate_limit.Limiter
	memBus       *eventbus.MemoryPublisher
	outboxRepo   *outbox.Repository
	media        *media.Handler
	pin          *pin.Handler
	block        *block.Handler
	report       *report.Handler
	notification *notification.Handler
	adminHandler *admin.Handler
	dlqHandler   *outbox.DLQHandler
	userRepo     *user.Repository
	profileRepo  *user.ProfileRepository
	deviceH      *device.Handler
	mute         *conversation.MuteHandler
}

// OutboxRepo exposes outbox repository for workers.
func (s *Server) OutboxRepo() *outbox.Repository {
	return s.outboxRepo
}

// MemoryBus exposes the in-process event bus for workers/tests.
func (s *Server) MemoryBus() *eventbus.MemoryPublisher {
	return s.memBus
}

// applyRateLimits wraps handlers with Redis rate limiting when configured.
func (s *Server) applyRateLimits(mux *http.ServeMux) {
	loginMW := rate_limit.Middleware(s.limiter, "login", 20, time.Minute, rate_limit.IPKey)
	convSendMW := rate_limit.AuthConversationMiddleware(s.limiter, "conv_send", 60, time.Minute)

	mux.Handle("POST /api/auth/login", loginMW(http.HandlerFunc(s.auth.HandleLogin)))
	mux.Handle(
		"POST /api/conversations/{id}/messages",
		auth.RequireAuth(s.auth, convSendMW(http.HandlerFunc(s.msg.HandleSend))),
	)
}
