// Package ratelimit provides an in-process per-key sliding-window rate limiter.
// For multi-instance deployments, swap in a Redis-backed implementation behind
// the same Allow(key) signature.
package ratelimit

import (
	"sync"
	"time"
)

type bucket struct {
	count int
	resetAt time.Time
}

type Limiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	rate    int
	window  time.Duration
}

func New(rate int, window time.Duration) *Limiter {
	return &Limiter{
		buckets: make(map[string]*bucket),
		rate:    rate,
		window:  window,
	}
}

func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	b, ok := l.buckets[key]
	if !ok || now.After(b.resetAt) {
		l.buckets[key] = &bucket{count: 1, resetAt: now.Add(l.window)}
		return true
	}
	if b.count >= l.rate {
		return false
	}
	b.count++
	return true
}
