package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
)

// RateLimiter stores rate limiters for different IPs
type RateLimiter struct {
	ips         map[string]*rate.Limiter
	mu          *sync.RWMutex
	rate        rate.Limit
	burst       int
	ttl         time.Duration
	lastCleanup time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(r rate.Limit, b int, ttl time.Duration) *RateLimiter {
	return &RateLimiter{
		ips:         make(map[string]*rate.Limiter),
		mu:          &sync.RWMutex{},
		rate:        r,
		burst:       b,
		ttl:         ttl,
		lastCleanup: time.Now(),
	}
}

// getLimiter returns the rate limiter for the given IP
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.ips[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.ips[ip] = limiter
	}

	// Cleanup old entries if needed
	if time.Since(rl.lastCleanup) > rl.ttl {
		rl.cleanup()
	}

	return limiter
}

// cleanup removes old rate limiters
func (rl *RateLimiter) cleanup() {
	rl.lastCleanup = time.Now()
	for ip := range rl.ips {
		delete(rl.ips, ip)
	}
}

// RateLimit creates a rate limiting middleware
func RateLimit(requests int, per time.Duration) echo.MiddlewareFunc {
	limiter := NewRateLimiter(rate.Limit(float64(requests)/per.Seconds()), requests, per)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()
			if !limiter.getLimiter(ip).Allow() {
				return echo.NewHTTPError(http.StatusTooManyRequests, "Rate limit exceeded")
			}
			return next(c)
		}
	}
}
