package graphql

import (
	"net/http"
	"sync"
	"time"
)

func RateLimiter(rate int, interval time.Duration) func(next http.Handler) http.Handler {
	rl := &limiter{
		visitors: make(map[string]*visitor),
		rate:     rate,
		interval: interval,
	}
	go rl.cleanup()
	return rl.middleware
}

type limiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	rate     int
	interval time.Duration
}

type visitor struct {
	tokens    int
	lastCheck time.Time
}

func (l *limiter) cleanup() {
	for {
		time.Sleep(time.Minute)
		l.mu.Lock()
		for ip, v := range l.visitors {
			if time.Since(v.lastCheck) > l.interval {
				delete(l.visitors, ip)
			}
		}
		l.mu.Unlock()
	}
}

func (l *limiter) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		l.mu.Lock()
		v, exists := l.visitors[ip]
		if !exists {
			v = &visitor{tokens: l.rate, lastCheck: time.Now()}
			l.visitors[ip] = v
		}
		now := time.Now()
		elapsed := now.Sub(v.lastCheck)
		newTokens := int(elapsed.Seconds() * float64(l.rate) / l.interval.Seconds())
		if newTokens > 0 {
			v.tokens += newTokens
			if v.tokens > l.rate {
				v.tokens = l.rate
			}
			v.lastCheck = now
		}
		if v.tokens > 0 {
			v.tokens--
			l.mu.Unlock()
			next.ServeHTTP(w, r)
			return
		}
		l.mu.Unlock()
		http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
	})
}