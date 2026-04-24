package ratelimit

import (
	"context"
	"time"
)

// TokenBucket implements the token bucket rate limiting algorithm.
// Each tenant+domain gets an isolated bucket in Redis.
type TokenBucket struct {
	// redisClient redis.Client
	RPS      int           // Tokens added per second
	Burst    int           // Max tokens in bucket
	Window   time.Duration // Refill window
}

// Allow checks if a request is allowed for the given tenant+domain.
// Returns (allowed bool, position int, waitTime time.Duration)
func (tb *TokenBucket) Allow(ctx context.Context, tenantID, domain string) (bool, int, time.Duration) {
	// TODO: implement Redis EVAL script for atomic token bucket
	// Key: tenant:{tenantID}:bucket:{domain}
	// Lua script: check tokens, decrement if available, return result
	return true, 0, 0
}
