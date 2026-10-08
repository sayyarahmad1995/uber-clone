package httpapi

import (
	"sync"
	"time"
)

type userWindowLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	entries map[string]rateWindow
}

type rateWindow struct {
	start time.Time
	count int
}

func newUserWindowLimiter(limit int) *userWindowLimiter {
	if limit <= 0 {
		limit = 60
	}
	return &userWindowLimiter{
		limit:   limit,
		window:  time.Minute,
		entries: make(map[string]rateWindow),
	}
}

func (l *userWindowLimiter) allow(userID string, now time.Time) bool {
	if l == nil {
		return true
	}
	l.mu.Lock()
	defer l.mu.Unlock()

	current := l.entries[userID]
	if current.start.IsZero() || now.Sub(current.start) >= l.window {
		l.entries[userID] = rateWindow{start: now, count: 1}
		return true
	}
	if current.count >= l.limit {
		return false
	}
	current.count++
	l.entries[userID] = current
	return true
}
