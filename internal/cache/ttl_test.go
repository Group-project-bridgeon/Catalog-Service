package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTTL_MissBeforeSet(t *testing.T) {
	c := NewTTL[string](time.Minute)

	val, ok := c.Get()
	assert.False(t, ok)
	assert.Empty(t, val)
}

func TestTTL_HitAfterSet(t *testing.T) {
	c := NewTTL[string](time.Minute)

	c.Set("hello")

	val, ok := c.Get()
	assert.True(t, ok)
	assert.Equal(t, "hello", val)
}

func TestTTL_Expiration(t *testing.T) {
	c := NewTTL[string](10 * time.Millisecond)

	c.Set("expiring")

	val, ok := c.Get()
	assert.True(t, ok)
	assert.Equal(t, "expiring", val)

	time.Sleep(20 * time.Millisecond)

	val, ok = c.Get()
	assert.False(t, ok)
	assert.Empty(t, val)
}

func TestTTL_Invalidate(t *testing.T) {
	c := NewTTL[string](time.Minute)

	c.Set("valid")

	val, ok := c.Get()
	assert.True(t, ok)
	assert.Equal(t, "valid", val)

	c.Invalidate()

	val, ok = c.Get()
	assert.False(t, ok)
	assert.Empty(t, val)
}
