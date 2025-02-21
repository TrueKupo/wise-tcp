package hashcash

import (
	"context"
	"errors"
	"sync"
	"time"
)

type MemoryCache struct {
	fingerprints map[string]time.Time
	mu           sync.RWMutex
	ticker       *time.Ticker
	stop         chan struct{}
}

func NewMemoryCache(cleanupInterval time.Duration) *MemoryCache {
	c := &MemoryCache{
		fingerprints: make(map[string]time.Time),
		ticker:       time.NewTicker(cleanupInterval),
		stop:         make(chan struct{}),
	}

	go c.startCleanupWorker()

	return c
}

func (c *MemoryCache) Start(_ context.Context) error {
	go c.startCleanupWorker()
	return nil
}

func (c *MemoryCache) Stop(_ context.Context) error {
	c.stop <- struct{}{}
	return nil
}

func (c *MemoryCache) Add(fingerprint string, _ string, expiry time.Duration) error {
	expirationTime := time.Now().Add(expiry)

	c.mu.Lock()
	defer c.mu.Unlock()

	c.fingerprints[fingerprint] = expirationTime
	return nil
}

func (c *MemoryCache) Remove(fingerprint string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	expirationTime, exists := c.fingerprints[fingerprint]
	if !exists {
		return errors.New("fingerprint not found in cache")
	}

	if time.Now().After(expirationTime) {
		delete(c.fingerprints, fingerprint)
		return errors.New("fingerprint expired")
	}

	delete(c.fingerprints, fingerprint)
	return nil
}

func (c *MemoryCache) startCleanupWorker() {
	for {
		select {
		case <-c.ticker.C:
			c.cleanup()
		case <-c.stop:
			c.ticker.Stop()
			return
		}
	}
}

func (c *MemoryCache) cleanup() {
	now := time.Now()

	c.mu.Lock()
	defer c.mu.Unlock()

	for fingerprint, expirationTime := range c.fingerprints {
		if now.After(expirationTime) {
			delete(c.fingerprints, fingerprint)
		}
	}
}
