package rate_limit

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/echoline/echoline/backend/internal/apierror"
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
