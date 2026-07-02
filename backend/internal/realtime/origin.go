package realtime

import (
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

var (
	wsAllowedOrigins     []string
	wsAllowedOriginsOnce sync.Once
)

func loadWSAllowedOrigins() {
	raw := strings.TrimSpace(os.Getenv("WS_ALLOWED_ORIGINS"))
	if raw == "" {
		wsAllowedOrigins = []string{
			"http://localhost:5173",
			"http://127.0.0.1:5173",
			"http://localhost:3000",
			"http://127.0.0.1:3000",
		}
		return
	}
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			wsAllowedOrigins = append(wsAllowedOrigins, part)
		}
	}
}

func checkWSOrigin(r *http.Request) bool {
	wsAllowedOriginsOnce.Do(loadWSAllowedOrigins)

	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return true
	}

	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	normalized := strings.TrimRight(u.Scheme+"://"+u.Host, "/")
	for _, allowed := range wsAllowedOrigins {
		if normalized == strings.TrimRight(allowed, "/") {
			return true
		}
	}
	return false
}
