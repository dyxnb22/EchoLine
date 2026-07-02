package server

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/echoline/echoline/backend/internal/admin"
	"github.com/echoline/echoline/backend/internal/ads"
	"github.com/echoline/echoline/backend/internal/audit"
	"github.com/echoline/echoline/backend/internal/auth"
	"github.com/echoline/echoline/backend/internal/block"
	"github.com/echoline/echoline/backend/internal/cache"
	"github.com/echoline/echoline/backend/internal/config"
	"github.com/echoline/echoline/backend/internal/conversation"
	"github.com/echoline/echoline/backend/internal/delivery"
	"github.com/echoline/echoline/backend/internal/device"
	"github.com/echoline/echoline/backend/internal/encryption"
	"github.com/echoline/echoline/backend/internal/entitlement"
	"github.com/echoline/echoline/backend/internal/eventbus"
	"github.com/echoline/echoline/backend/internal/export"
	"github.com/echoline/echoline/backend/internal/forward"
	"github.com/echoline/echoline/backend/internal/graph"
	"github.com/echoline/echoline/backend/internal/media"
	"github.com/echoline/echoline/backend/internal/message"
	"github.com/echoline/echoline/backend/internal/notification"
	"github.com/echoline/echoline/backend/internal/outbox"
	"github.com/echoline/echoline/backend/internal/payment"
	"github.com/echoline/echoline/backend/internal/pin"
	"github.com/echoline/echoline/backend/internal/presence"
	"github.com/echoline/echoline/backend/internal/push"
	"github.com/echoline/echoline/backend/internal/rate_limit"
	"github.com/echoline/echoline/backend/internal/reaction"
	"github.com/echoline/echoline/backend/internal/realtime"
	"github.com/echoline/echoline/backend/internal/recommendation"
	"github.com/echoline/echoline/backend/internal/redisx"
	"github.com/echoline/echoline/backend/internal/report"
	"github.com/echoline/echoline/backend/internal/search"
	"github.com/echoline/echoline/backend/internal/sync"
	"github.com/echoline/echoline/backend/internal/thread"
	"github.com/echoline/echoline/backend/internal/user"
	"github.com/echoline/echoline/backend/internal/webhook"
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

	msgHandler := message.NewHandler(msgSvc, convRepo, attachmentRepo, auditRepo)

	blockRepo := block.NewRepository(pool)
	msgSvc.SetBlockChecker(blockRepo)

	memBus := eventbus.NewMemoryPublisher(256)

	var presenceTracker realtime.PresenceTracker
	var limiter rate_limit.Limiter
	var listCache *cache.ConversationListCache
	var presenceChecker *presence.RedisOnlineChecker
	var lastSeenStore *presence.LastSeenStore
	if opts.Redis != nil {
		presenceTracker = presence.NewStore(opts.Redis, 0)
		limiter = rate_limit.NewRedisLimiter(opts.Redis)
		listCache = cache.NewConversationListCache(opts.Redis)
		presenceChecker = presence.NewRedisOnlineChecker(opts.Redis)
		lastSeenStore = presence.NewLastSeenStore(opts.Redis)
	}

	rt := realtime.NewServer(authSvc, msgSvc, convRepo, deliveryRepo, presenceTracker, logger)
	rt.SetHubMetrics()

	convHandler := conversation.NewHandler(convRepo)
	convHandler.SetListCache(listCache)
	entitlementRepo := entitlement.NewRepository(pool)
	convHandler.SetEntitlementGate(entitlementRepo)

	var mediaHandler *media.Handler
	if cfg.S3Endpoint != "" {
		if mediaClient, err := media.NewClient(cfg); err == nil {
			mediaHandler = media.NewHandler(mediaClient, attachmentRepo)
		} else {
			logger.Warn("media client unavailable", "error", err)
		}
	}

	pinRepo := pin.NewRepository(pool)
	reportRepo := report.NewRepository(pool)
	notifRepo := notification.NewRepository(pool)

	archiveRepo := conversation.NewArchiveRepository(pool)

	webhookDispatcher := webhook.NewDispatcher(cfg.WebhookURL)
	webhookRepo := webhook.NewRepository(pool)
	persistingWebhook := webhook.NewPersistingDispatcher(webhookDispatcher, webhookRepo)
	msgHandler.SetWebhookNotifier(message.FuncWebhookNotifier(persistingWebhook.DispatchMessageCreated))

	searchHandler := search.NewHandler(searchRepo)
	osClient := search.NewOpenSearchClient(cfg.OpenSearchURL)
	searchHandler.SetOpenSearch(osClient)

	adminChecker := admin.NewCompositeAdminChecker(admin.NewStaticAdminChecker(cfg.AdminUserIDs), pool)
	deviceRepo := device.NewRepository(pool)
	graphHandler := graph.NewHandler(convRepo, cfg.GraphiQL)
	graphHandler.SetMessageSender(msgSvc)
	graphHandler.SetReactionAdder(graph.NewReactionService(reaction.NewRepository(pool)))

	rt.SetDeviceTracker(deviceRepo)

	return &Server{
		cfg:            cfg,
		pool:           pool,
		logger:         logger,
		auth:           authSvc,
		conv:           convHandler,
		msg:            msgHandler,
		sync:           sync.NewHandler(convRepo, msgSvc, cursorRepo),
		search:         searchHandler,
		delivery:       delivery.NewHandler(deliveryRepo, convRepo),
		realtime:       rt,
		limiter:        limiter,
		memBus:         memBus,
		outboxRepo:     outboxRepo,
		media:          mediaHandler,
		pin:            pin.NewHandler(pinRepo, convRepo),
		block:          block.NewHandler(blockRepo),
		report:         report.NewHandler(reportRepo, convRepo),
		notification:   notification.NewHandler(notifRepo),
		adminHandler:   admin.NewHandler(pool, authSvc),
		dlqHandler:     outbox.NewDLQHandler(pool),
		dlqReplay:      outbox.NewDLQReplayHandler(outbox.NewDLQRepository(pool), outboxRepo),
		userRepo:       userRepo,
		profileRepo:    user.NewProfileRepository(pool),
		deviceH:        device.NewHandler(pool),
		mute:           conversation.NewMuteHandler(pool, convRepo),
		reaction:       reaction.NewHandler(reaction.NewRepository(pool), convRepo, msgRepo),
		thread:         thread.NewHandler(msgSvc, convRepo, thread.NewRepository(pool)),
		forward:        forward.NewHandler(msgSvc),
		presenceH:      presence.NewOnlineHandler(presenceChecker),
		lastSeenH:      presence.NewLastSeenHandler(lastSeenStore),
		export:         export.NewHandler(export.NewRepository(pool), convRepo),
		push:           push.NewHandler(push.NewRepository(pool)),
		payment:        func() *payment.Handler {
			ph := payment.NewHandler(payment.NewRepository(pool))
			ph.SetEntitlementGranter(entitlementRepo)
			ph.SetPaidChannelChecker(entitlementRepo)
			return ph
		}(),
		ads:            ads.NewHandler(ads.NewRepository(pool), conversation.NewOwnerChecker(convRepo), convRepo),
		recommendation: recommendation.NewHandler(recommendation.NewRepository(pool)),
		archive:        conversation.NewArchiveHandler(archiveRepo, convRepo),
		encryption:     encryption.NewHandler(encryption.NewRepository(pool)),
		entitlement:    entitlement.NewHandler(entitlementRepo, conversation.NewOwnerChecker(convRepo), adminChecker),
		webhook:        webhookDispatcher,
		webhookRepo:    webhookRepo,
		opensearch:     osClient,
		adminChecker:   adminChecker,
		graph:          graphHandler,
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
	dlqReplay    *outbox.DLQReplayHandler
	userRepo     *user.Repository
	profileRepo  *user.ProfileRepository
	deviceH      *device.Handler
	mute         *conversation.MuteHandler
	reaction     *reaction.Handler
	thread       *thread.Handler
	forward      *forward.Handler
	presenceH    *presence.OnlineHandler
	lastSeenH    *presence.LastSeenHandler
	export       *export.Handler
	push         *push.Handler
	payment      *payment.Handler
	ads          *ads.Handler
	recommendation *recommendation.Handler
	archive      *conversation.ArchiveHandler
	encryption   *encryption.Handler
	entitlement  *entitlement.Handler
	webhook      *webhook.Dispatcher
	webhookRepo  *webhook.Repository
	opensearch   *search.OpenSearchClient
	adminChecker admin.AdminChecker
	graph        *graph.Handler
}

// WebhookRepo exposes webhook delivery repository for workers.
func (s *Server) WebhookRepo() *webhook.Repository {
	return s.webhookRepo
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
	loginMW := rate_limit.Middleware(s.limiter, "login", 5, time.Minute, rate_limit.IPKey)
	registerMW := rate_limit.Middleware(s.limiter, "register", 3, time.Minute, rate_limit.IPKey)
	convSendMW := rate_limit.AuthConversationMiddleware(s.limiter, "conv_send", 60, time.Minute)

	mux.Handle("POST /api/auth/login", loginMW(http.HandlerFunc(s.auth.HandleLogin)))
	mux.Handle("POST /api/auth/register", registerMW(http.HandlerFunc(s.auth.HandleRegister)))
	mux.Handle(
		"POST /api/conversations/{id}/messages",
		auth.RequireAuth(s.auth, convSendMW(http.HandlerFunc(s.msg.HandleSend))),
	)
}
