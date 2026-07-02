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
	if opts.Redis != nil {
		authSvc.SetRefreshStore(auth.NewRedisRefreshStore(opts.Redis))
	} else {
		authSvc.SetRefreshStore(auth.NewMemoryRefreshStore())
	}
	if cfg.PaymentSelfServe {
		logger.Warn("PAYMENT_SELF_SERVE is enabled — channel entitlements can be self-granted without payment verification")
	}
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
	var limiter rate_limit.Limiter = rate_limit.NewMemoryLimiter()
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
	rt.SetConvSendLimiter(limiter)
	rt.SetHubMetrics()
	rt.SetAttachments(attachmentRepo)

	convHandler := conversation.NewHandler(convRepo)
	convHandler.SetListCache(listCache)
	entitlementRepo := entitlement.NewRepository(pool)
	convHandler.SetEntitlementGate(entitlementRepo)

	var mediaHandler *media.Handler
	if cfg.S3Endpoint != "" {
		if mediaClient, err := media.NewClient(cfg); err == nil {
			mediaHandler = media.NewHandler(mediaClient, attachmentRepo, convRepo)
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
	searchHandler.SetMemberChecker(convRepo)
	searchHandler.SetOpenSearch(osClient)

	adminChecker := admin.NewCompositeAdminChecker(admin.NewStaticAdminChecker(cfg.AdminUserIDs), pool)
	deviceRepo := device.NewRepository(pool)
	var graphHandler *graph.Handler
	if cfg.GraphQLEnabled {
		graphHandler = graph.NewHandler(convRepo, cfg.GraphiQL)
		graphHandler.SetMessageSender(msgSvc)
		graphHandler.SetReactionAdder(graph.NewReactionService(reaction.NewRepository(pool), convRepo, msgRepo))
	}

	rt.SetDeviceTracker(deviceRepo)

	syncHandler := sync.NewHandler(convRepo, msgSvc, cursorRepo)
	syncHandler.SetAttachments(attachmentRepo)

	presenceHandler := presence.NewOnlineHandler(presenceChecker)
	presenceHandler.SetContactGate(convRepo)
	lastSeenHandler := presence.NewLastSeenHandler(lastSeenStore)
	lastSeenHandler.SetContactGate(convRepo)

	return &Server{
		cfg:            cfg,
		pool:           pool,
		logger:         logger,
		auth:           authSvc,
		conv:           convHandler,
		msg:            msgHandler,
		sync:           syncHandler,
		search:         searchHandler,
		delivery:       delivery.NewHandler(deliveryRepo, convRepo, msgRepo),
		realtime:       rt,
		limiter:        limiter,
		memBus:         memBus,
		outboxRepo:     outboxRepo,
		media:          mediaHandler,
		pin:            pin.NewHandler(pinRepo, convRepo, msgRepo),
		block:          block.NewHandler(blockRepo),
		report:         report.NewHandler(reportRepo, convRepo, msgRepo),
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
		presenceH:      presenceHandler,
		lastSeenH:      lastSeenHandler,
		export:         export.NewHandler(export.NewRepository(pool), convRepo),
		push:           push.NewHandler(push.NewRepository(pool)),
		payment:        func() *payment.Handler {
			ph := payment.NewHandler(payment.NewRepository(pool))
			ph.SetEntitlementGranter(entitlementRepo)
			ph.SetSelfServe(cfg.PaymentSelfServe)
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

// applyRateLimits wraps handlers with rate limiting (Redis when configured, in-memory fallback).
func (s *Server) applyRateLimits(mux *http.ServeMux) {
	loginMW := rate_limit.Middleware(s.limiter, "login", 20, time.Minute, rate_limit.IPKey)
	registerMW := rate_limit.Middleware(s.limiter, "register", 10, time.Minute, rate_limit.IPKey)
	refreshMW := rate_limit.Middleware(s.limiter, "refresh", 30, time.Minute, rate_limit.IPKey)
	convSendMW := rate_limit.AuthConversationMiddleware(s.limiter, "conv_send", 60, time.Minute)
	graphqlSendMW := rate_limit.Middleware(s.limiter, "graphql_send", 60, time.Minute, rate_limit.AuthUserKey)
	syncMW := rate_limit.Middleware(s.limiter, "sync", 120, time.Minute, rate_limit.AuthUserKey)
	ackMW := rate_limit.Middleware(s.limiter, "message_ack", 120, time.Minute, rate_limit.AuthUserKey)
	searchMW := rate_limit.Middleware(s.limiter, "search", 60, time.Minute, rate_limit.AuthUserKey)
	exportMW := rate_limit.Middleware(s.limiter, "export", 10, time.Minute, rate_limit.AuthUserKey)
	mediaMW := rate_limit.Middleware(s.limiter, "media_presign", 60, time.Minute, rate_limit.AuthUserKey)

	mux.Handle("POST /api/auth/register", registerMW(http.HandlerFunc(s.auth.HandleRegister)))
	mux.Handle("POST /api/auth/refresh", refreshMW(http.HandlerFunc(s.auth.HandleRefresh)))
	mux.Handle("POST /api/auth/login", loginMW(http.HandlerFunc(s.auth.HandleLogin)))
	mux.Handle(
		"POST /api/conversations/{id}/messages",
		auth.RequireAuth(s.auth, convSendMW(http.HandlerFunc(s.msg.HandleSend))),
	)
	if s.graph != nil {
		mux.Handle(
			"POST /graphql",
			auth.RequireAuth(s.auth, graphqlSendMW(http.HandlerFunc(s.graph.HandleGraphQL))),
		)
	}
	mux.Handle("POST /api/sync", auth.RequireAuth(s.auth, syncMW(http.HandlerFunc(s.sync.HandleSync))))
	mux.Handle("POST /api/messages/ack", auth.RequireAuth(s.auth, ackMW(http.HandlerFunc(s.delivery.HandleACK))))
	mux.Handle("GET /api/search/messages", auth.RequireAuth(s.auth, searchMW(http.HandlerFunc(s.search.HandleSearch))))
	mux.Handle("GET /api/conversations/{id}/export", auth.RequireAuth(s.auth, exportMW(http.HandlerFunc(s.export.HandleExport))))
	if s.media != nil {
		mux.Handle("POST /api/media/upload-url", auth.RequireAuth(s.auth, mediaMW(http.HandlerFunc(s.media.HandlePresignUpload))))
		mux.Handle("POST /api/media/download-url", auth.RequireAuth(s.auth, mediaMW(http.HandlerFunc(s.media.HandlePresignDownload))))
	}
	wsMW := rate_limit.Middleware(s.limiter, "ws_upgrade", 30, time.Minute, rate_limit.IPKey)
	mux.Handle("GET /ws", wsMW(http.HandlerFunc(s.realtime.HandleWS)))
}
