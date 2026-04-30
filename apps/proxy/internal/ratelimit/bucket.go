package ratelimit

import (
	"context"
	"sync"
	"time"
)

// TokenBucket implements the token bucket rate limiting algorithm.
// Each tenant+domain gets an isolated bucket in Redis.
type TokenBucket struct {
	// redisClient redis.Client
	RPS     int           // Tokens added per second
	Burst   int           // Max tokens in bucket
	Window  time.Duration // Refill window
	mu      sync.Mutex
	buckets map[string]*bucketState
}

type bucketState struct {
	tokens     float64
	lastRefill time.Time
}

func NewTokenBucket(rps, burst int) *TokenBucket {
	if rps <= 0 {
		rps = 1
	}
	if burst <= 0 {
		burst = rps
	}

	return &TokenBucket{
		RPS:     rps,
		Burst:   burst,
		Window:  time.Second,
		buckets: make(map[string]*bucketState),
	}
}

// Allow checks if a request is allowed for the given tenant+domain.
// Returns (allowed bool, position int, waitTime time.Duration)
func (tb *TokenBucket) Allow(ctx context.Context, tenantID, domain string) (bool, int, time.Duration) {
	_ = ctx

	tb.mu.Lock()
	defer tb.mu.Unlock()

	if tb.buckets == nil {
		tb.buckets = make(map[string]*bucketState)
	}

	key := tenantID + ":" + domain
	now := time.Now()
	state, ok := tb.buckets[key]
	if !ok {
		tb.buckets[key] = &bucketState{
			tokens:     float64(tb.Burst - 1),
			lastRefill: now,
		}
		return true, 0, 0
	}

	elapsed := now.Sub(state.lastRefill).Seconds()
	if elapsed > 0 {
		state.tokens += elapsed * float64(tb.RPS)
		if state.tokens > float64(tb.Burst) {
			state.tokens = float64(tb.Burst)
		}
		state.lastRefill = now
	}

	if state.tokens >= 1 {
		state.tokens--
		return true, 0, 0
	}

	deficit := 1 - state.tokens
	wait := time.Duration(deficit / float64(tb.RPS) * float64(time.Second))
	if wait < 0 {
		wait = 0
	}

	// TODO: implement Redis EVAL script for atomic token bucket
	// Key: tenant:{tenantID}:bucket:{domain}
	// Lua script: check tokens, decrement if available, return result
	return false, 1, wait
}
