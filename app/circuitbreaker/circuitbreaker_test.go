package circuitbreaker_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alextanhongpin/go-service-oriented-package/app/circuitbreaker"
	"github.com/stretchr/testify/assert"
)

var wantErr = errors.New("want error")

func TestCircuitBreaker(t *testing.T) {
	assert := assert.New(t)

	every := 1
	success := 3
	failure := 3

	ctx := context.Background()
	m := newMockCircuitBreaker(every)
	assert.Nil(m.Handle(ctx))
	assert.True(m.cb.IsClosed())

	// Service unhealthy.
	m.ok = false
	for i := 0; i < success-1; i++ {
		assert.ErrorIs(m.Handle(ctx), wantErr)
		assert.True(m.cb.IsClosed())
	}

	// Closed on the third failure.
	assert.ErrorIs(m.Handle(ctx), wantErr)
	assert.True(m.cb.IsOpen())

	assert.ErrorIs(m.Handle(ctx), circuitbreaker.ErrUnavailable)
	assert.True(m.cb.IsOpen())
	assert.True(m.cb.ResetIn() > 0)

	time.Sleep(m.cb.ResetIn())

	// Service healthy again.
	m.ok = true
	for i := 0; i < failure-1; i++ {
		assert.Nil(m.Handle(ctx))
		assert.True(m.cb.IsHalfOpen())
	}

	assert.Nil(m.Handle(ctx))
	assert.True(m.cb.IsClosed())
}

func TestCircuitSampling(t *testing.T) {
	assert := assert.New(t)

	ctx := context.Background()

	// Skipping 1 call for every 2 calls.
	every := 2
	success := 3
	failure := 3

	m := newMockCircuitBreaker(every)
	assert.Nil(m.Handle(ctx))
	assert.True(m.cb.IsClosed())

	// Service unhealthy.
	m.ok = false
	for i := 0; i < success*every-1; i++ {
		assert.ErrorIs(m.Handle(ctx), wantErr)
		assert.True(m.cb.IsClosed())
	}

	// Closed on the third failure.
	assert.ErrorIs(m.Handle(ctx), wantErr)
	assert.True(m.cb.IsOpen())

	assert.ErrorIs(m.Handle(ctx), circuitbreaker.ErrUnavailable)
	assert.True(m.cb.IsOpen())
	assert.True(m.cb.ResetIn() > 0)

	time.Sleep(m.cb.ResetIn())

	// Service healthy again.
	m.ok = true

	for i := 0; i < failure*every-1; i++ {
		assert.Nil(m.Handle(ctx))
		assert.True(m.cb.IsHalfOpen())
	}

	assert.Nil(m.Handle(ctx))
	assert.True(m.cb.IsClosed())
}

type circuitBreaker interface {
	Exec(func() error) error
	ResetIn() time.Duration
	IsOpen() bool
	IsHalfOpen() bool
	IsClosed() bool
}

type mockCircuitBreaker struct {
	cb circuitBreaker
	ok bool
}

func newMockCircuitBreaker(every int) *mockCircuitBreaker {
	cb := circuitbreaker.New()
	cb.Timeout = func() time.Duration {
		return 1 * time.Second
	}
	cb.Sampling.Every = every
	cb.Success = 3
	cb.Failure = 3
	m := &mockCircuitBreaker{
		cb: cb,
		ok: true,
	}

	return m
}

func (m *mockCircuitBreaker) Handle(ctx context.Context) error {
	return m.cb.Exec(func() error {
		if !m.ok {
			return wantErr
		}

		return nil
	})
}
