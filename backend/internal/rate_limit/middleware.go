package rate_limit

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/echoline/echoline/backend/internal/apierror"
	"github.com/echoline/echoline/backend/internal/auth"
	"github.com/echoline/echoline/backend/internal/redisx"
)

// Limiter checks request quotas.
type Limiter interface {
	Allow(ctx context.Context, key string, limit int64, window time.Duration) (bool, error)
}

// RedisLimiter uses Redis INCR for fixed-window limiting.
type RedisLimiter struct {
	redis *redisx.Client
}

// NewRedisLimiter creates a Redis-backed limiter.
func NewRedisLimiter(redis *redisx.Client) *RedisLimiter {
	return &RedisLimiter{redis: redis}
}

// Allow implements Limiter.
func (l *RedisLimiter) Allow(ctx context.Context, key string, limit int64, window time.Duration) (bool, error) {
	if l == nil || l.redis == nil {
		return true, nil
	}
	return l.redis.Allow(ctx, key, limit, window)
}

// Middleware rate-limits requests by key derived from the request.
func Middleware(limiter Limiter, keyPrefix string, limit int64, window time.Duration, keyFn func(*http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if limiter == nil {
				next.ServeHTTP(w, r)
				return
			}
			key := keyPrefix + ":" + keyFn(r)
			ok, err := limiter.Allow(r.Context(), key, limit, window)
			if err != nil {
				apierror.Write(w, r, http.StatusInternalServerError, "internal_error", "rate limiter error")
				return
			}
			if !ok {
				apierror.Write(w, r, http.StatusTooManyRequests, "rate_limited", "too many requests")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// IPKey returns client IP for rate limiting.
func IPKey(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	host := r.RemoteAddr
	if i := len(host) - 1; i >= 0 {
		for j := len(host) - 1; j >= 0; j-- {
			if host[j] == ':' {
				return host[:j]
			}
		}
	}
	return host
}

// PathKey combines path with IP.
func PathKey(r *http.Request) string {
	return fmt.Sprintf("%s:%s", r.URL.Path, IPKey(r))
}
func ConversationUserKey(r *http.Request) string {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		return PathKey(r)
	}
	path := r.URL.Path
	const prefix = "/api/conversations/"
	const suffix = "/messages"
	if len(path) > len(prefix)+len(suffix) && strings.HasSuffix(path, suffix) {
		raw := path[len(prefix) : len(path)-len(suffix)]
		return fmt.Sprintf("%s:%s", raw, claims.UserID)
	}
	return fmt.Sprintf("%s:%s", path, claims.UserID)
}

// AuthConversationMiddleware applies rate limits after auth context exists.
func AuthConversationMiddleware(limiter Limiter, keyPrefix string, limit int64, window time.Duration) func(http.Handler) http.Handler {
	return Middleware(limiter, keyPrefix, limit, window, ConversationUserKey)
}
