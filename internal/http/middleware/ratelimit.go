package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter implements per-IP rate limiting using token bucket algorithm
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     rate.Limit // requests per second
	burst    int        // burst size
	cleanup  *time.Ticker
}

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewRateLimiter creates a new rate limiter
// rps: requests per second (e.g., 0.1 = 1 request per 10 seconds)
// burst: maximum burst size
func NewRateLimiter(rps float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate.Limit(rps),
		burst:    burst,
		cleanup:  time.NewTicker(5 * time.Minute), // Clean up old visitors every 5 minutes
	}

	// Start cleanup goroutine
	go rl.cleanupVisitors()

	return rl
}

// cleanupVisitors removes old visitors that haven't been seen in 10 minutes
func (rl *RateLimiter) cleanupVisitors() {
	for range rl.cleanup.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, v := range rl.visitors {
			if now.Sub(v.lastSeen) > 10*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// getVisitor returns the rate limiter for a given IP, creating one if needed
func (rl *RateLimiter) getVisitor(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(rl.rate, rl.burst)
		rl.visitors[ip] = &visitor{
			limiter:  limiter,
			lastSeen: time.Now(),
		}
		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

// Limit wraps an HTTP handler with rate limiting
func (rl *RateLimiter) Limit(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get client IP
		ip := getClientIP(r)

		// Get or create limiter for this IP
		limiter := rl.getVisitor(ip)

		// Check if request is allowed
		if !limiter.Allow() {
			http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
			return
		}

		// Call next handler
		next(w, r)
	}
}

// getClientIP extracts the client IP from the request
// Checks X-Forwarded-For and X-Real-IP headers for proxies
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (for proxies/load balancers)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		// Handle comma-separated list
		if idx := strings.Index(forwarded, ","); idx >= 0 {
			forwarded = strings.TrimSpace(forwarded[:idx])
		}
		return forwarded
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	// Remove port if present (format: "ip:port" or "[ipv6]:port")
	if strings.Contains(ip, ":") {
		// Handle IPv6 format [::1]:port
		if strings.HasPrefix(ip, "[") {
			if idx := strings.Index(ip, "]:"); idx >= 0 {
				ip = ip[1:idx] // Remove [ and ]
			}
		} else {
			// Handle IPv4 format 192.168.1.1:port
			if idx := strings.LastIndex(ip, ":"); idx >= 0 {
				ip = ip[:idx]
			}
		}
	}
	return ip
}
