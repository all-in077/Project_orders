package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// visitor tracks request statistics for a single client within the current window.
type visitor struct {
	count    int
	lastSeen time.Time
}

// RateLimiter implements a per-IP sliding window rate limiter.
// visitors maps each client IP to its request statistics.
// limit is the maximum number of requests allowed per window.
// window is the duration of the sliding time window (e.g. 1 minute).
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a RateLimiter with the given request limit and time window.
// Starts a background goroutine to periodically remove stale visitor entries.
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		limit:    limit,
		window:   window,
	}
	go rl.cleanup()
	return rl
}

// cleanup removes visitor entries that have been inactive for longer than the window.
// Runs every minute in a background goroutine.
// TODO: replace time.Tick with time.NewTicker and add stop channel for graceful shutdown.
func (rl *RateLimiter) cleanup() {
	for range time.Tick(time.Minute) {
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > rl.window {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Middleware returns a Gin handler that enforces the rate limit per client IP.
// Uses a sliding window: the window resets from the client's last seen request.
// Responds with 429 Too Many Requests when the limit is exceeded.
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		rl.mu.Lock()
		v, exists := rl.visitors[ip]
		if !exists || time.Since(v.lastSeen) > rl.window {
			rl.visitors[ip] = &visitor{count: 1, lastSeen: time.Now()}
			rl.mu.Unlock()
			c.Next()
			return
		}
		v.count++
		v.lastSeen = time.Now()
		count := v.count
		rl.mu.Unlock()

		if count > rl.limit {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		c.Next()
	}
}
