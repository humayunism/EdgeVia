package circuitbreaker

import "time"

type State int

const (
	StateClosed   State = iota // Normal operation
	StateOpen                  // Blocking all requests
	StateHalfOpen              // Testing with 10% traffic
)

type CircuitBreaker struct {
	FailureThreshold float64       // e.g. 0.5 = 50% error rate triggers OPEN
	RecoveryTimeout  time.Duration // Time before trying HALF_OPEN
	SampleWindow     time.Duration // Time window to measure error rate
}

func (cb *CircuitBreaker) GetState(domain string) State {
	// TODO: read from Redis key: breaker:{domain}
	return StateClosed
}

func (cb *CircuitBreaker) RecordSuccess(domain string) {
	// TODO: decrement failure counter in Redis
}

func (cb *CircuitBreaker) RecordFailure(domain string) {
	// TODO: increment failure counter, check threshold, flip state
}
