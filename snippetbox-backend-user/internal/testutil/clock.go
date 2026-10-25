package testutil

import (
	"sync"
	"time"
)

var Epoch = time.Date(2026, 10, 15, 8, 0, 0, 0, time.UTC)

type FixedClock struct {
	mu  sync.Mutex
	now time.Time
}

func NewFixedClock(now time.Time) *FixedClock {
	return &FixedClock{now: now}
}

func (c *FixedClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *FixedClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}
