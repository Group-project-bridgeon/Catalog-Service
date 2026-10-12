package cache

import (
	"sync"
	"time"
)

type TTL[T any] struct {
	mu        sync.RWMutex
	value     T
	hasValue  bool
	expiresAt time.Time
	ttl       time.Duration
}

func NewTTL[T any](
	ttl time.Duration,
) *TTL[T] {
	return &TTL[T]{
		ttl: ttl,
	}
}

func (c *TTL[T]) Get() (T, bool) {
	c.mu.RLock()

	if !c.hasValue ||
		time.Now().After(c.expiresAt) {
		c.mu.RUnlock()

		var zero T
		return zero, false
	}

	value := c.value

	c.mu.RUnlock()

	return value, true
}

func (c *TTL[T]) Set(
	value T,
) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.value = value
	c.hasValue = true
	c.expiresAt = time.Now().Add(c.ttl)
}

func (c *TTL[T]) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()

	var zero T

	c.value = zero
	c.hasValue = false
	c.expiresAt = time.Time{}
}
