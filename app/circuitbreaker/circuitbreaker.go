package circuitbreaker

import (
	"errors"
	"time"

	"golang.org/x/time/rate"
)

const (
	timeout = 5 * time.Second
	success = 5  // 5 success before the circuit breaker becomes closed.
	failure = 10 // 10 failures before the circuit breaker becomes open.
)

// ErrUnavailable returns the error when the circuit breaker is not available.
var ErrUnavailable = errors.New("circuit: unavailable")

type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

var stateText = map[State]string{
	StateClosed:   "closed",
	StateOpen:     "open",
	StateHalfOpen: "half-open",
}

func (s State) String() string {
	text := stateText[s]
	return text
}

// CircuitBreaker represents the circuit breaker.
type CircuitBreaker struct {
	// Private.
	state    State
	counter  int
	deadline time.Time

	// Public.
	Success  int
	Failure  int
	Timeout  func() time.Duration
	Now      func() time.Time
	Sampling rate.Sometimes
}

// New returns a pointer to CircuitBreaker.
func New() *CircuitBreaker {
	return &CircuitBreaker{
		Timeout: func() time.Duration {
			return timeout
		},
		Success:  success,
		Failure:  failure,
		Now:      time.Now,
		Sampling: rate.Sometimes{Every: 1},
	}
}

// Exec updates the circuit breaker state based on the returned error.
func (c *CircuitBreaker) Exec(fn func() error) error {
	if !c.allow() {
		return ErrUnavailable
	}

	err := fn()
	c.exec(err == nil)

	return err
}

// ResetIn returns the wait time before the service can be called again.
func (c *CircuitBreaker) ResetIn() time.Duration {
	return c.deadline.Sub(c.Now())
}

func (c *CircuitBreaker) State() string {
	return c.state.String()
}

func (c *CircuitBreaker) IsOpen() bool {
	return c.state == StateOpen
}

func (c *CircuitBreaker) IsClosed() bool {
	return c.state == StateClosed
}

func (c *CircuitBreaker) IsHalfOpen() bool {
	return c.state == StateHalfOpen
}

func (c *CircuitBreaker) allow() bool {
	if c.IsOpen() {
		c.exec(true)
	}

	return !c.IsOpen()
}

func (c *CircuitBreaker) exec(ok bool) {
	c.Sampling.Do(func() {
		c.update(ok)
	})
}

func (c *CircuitBreaker) update(ok bool) {
	switch c.state {
	case StateOpen:
		if c.Now().After(c.deadline) {
			c.counter = 0
			c.state = StateHalfOpen
		}
	case StateHalfOpen:
		// The service is still unhealthy
		// Reset the counter and revert to Open.
		if !ok {
			c.counter = 0
			c.state = StateOpen

			return
		}

		// The service is healthy.
		// After a certain threshold, circuit breaker becomes Closed.
		c.counter++
		if c.counter >= c.Success {
			c.counter = 0
			c.state = StateClosed
		}
	case StateClosed:
		// The service is healthy.
		if ok {
			return
		}

		// The service is unhealthy.
		// After a certain threshold, circuit breaker becomes Open.
		c.counter++
		if c.counter >= c.Failure {
			c.deadline = c.Now().Add(c.Timeout())
			c.state = StateOpen
		}
	}
}
