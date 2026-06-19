package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
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

func (rl *RateLimiter) Middleware(c *fiber.Ctx) error {
	ip := c.IP()

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
		return c.Next()
	}
	rl.mu.Unlock()

	return c.Status(http.StatusTooManyRequests).JSON(fiber.Map{"error": "rate limit exceeded"})
}
