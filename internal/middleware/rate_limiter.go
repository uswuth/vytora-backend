package middleware

import (
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     int           // requests per interval
	interval time.Duration
}

type visitor struct {
	tokens    int
	lastCheck time.Time
}

func NewRateLimiter(rate int, interval time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		interval: interval,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(time.Minute)
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastCheck) > rl.interval {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr

		rl.mu.Lock()
		v, exists := rl.visitors[ip]
		if !exists {
			v = &visitor{tokens: rl.rate, lastCheck: time.Now()}
			rl.visitors[ip] = v
		}

		// Refill tokens
		now := time.Now()
		elapsed := now.Sub(v.lastCheck)
		newTokens := int(elapsed.Seconds() * float64(rl.rate) / rl.interval.Seconds())
		if newTokens > 0 {
			v.tokens += newTokens
			if v.tokens > rl.rate {
				v.tokens = rl.rate
			}
			v.lastCheck = now
		}

		if v.tokens > 0 {
			v.tokens--
			rl.mu.Unlock()
			next.ServeHTTP(w, r)
			return
		}
		rl.mu.Unlock()

		w.Header().Set("Retry-After", "10")
		http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
	})
}