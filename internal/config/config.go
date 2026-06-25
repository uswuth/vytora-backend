package config

import "time"

// Config holds application configuration values loaded from environment/.env
type Config struct {
	DatabaseURL       string
	JWTSecret         string
	JWTExpiryHours    int
	AllowedOrigins    []string
	HealthAllowedIPs  []string
	RateLimitRequests int
	RateLimitInterval time.Duration
	MaxBodySize       int64
	Host              string
	Port              string
	AppEnv            string
	ConnectionName    string
	DBHost            string
	DBPort            string
	DBUser            string
	DBPassword        string
	DBName            string
}
