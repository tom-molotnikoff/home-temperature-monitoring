package service

import (
	"sync"
	"time"
)

type simpleBlocker struct {
	mu         sync.Mutex
	m          map[string]time.Time
	allowOnce  map[string]time.Time
	allowCount map[string]int
}

func newSimpleBlocker() *simpleBlocker {
	return &simpleBlocker{m: make(map[string]time.Time), allowOnce: make(map[string]time.Time), allowCount: make(map[string]int)}
}

func (b *simpleBlocker) getRemainingSeconds(key string) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	t, ok := b.m[key]
	if !ok {
		return 0
	}
	now := time.Now()
	if now.After(t) {
		delete(b.m, key)
		return 0
	}
	return int(t.Sub(now).Seconds())
}

func (b *simpleBlocker) blockFor(key string, seconds int) {
	if seconds <= 0 {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.m[key] = time.Now().Add(time.Duration(seconds) * time.Second)
	b.allowOnce[key] = b.m[key]
	b.allowCount[key] = 3
}

func (b *simpleBlocker) consumeAllowOnceIfReady(key string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	avail, ok := b.allowOnce[key]
	if !ok {
		return false
	}
	now := time.Now()
	if now.Before(avail) {
		return false
	}
	count := b.allowCount[key]
	if count <= 0 {
		delete(b.allowOnce, key)
		delete(b.allowCount, key)
		return false
	}
	b.allowCount[key] = count - 1
	if b.allowCount[key] <= 0 {
		delete(b.allowOnce, key)
		delete(b.allowCount, key)
	}
	return true
}

func (b *simpleBlocker) forceClearAllowOnce(key string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.allowOnce, key)
	delete(b.allowCount, key)
}

var (
	ipBlocker   = newSimpleBlocker()
	userBlocker = newSimpleBlocker()
)
