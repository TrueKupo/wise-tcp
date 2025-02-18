package hashcash

import (
	"errors"
	"sync"
	"time"
)

type Cache struct {
	fingerprints map[string]time.Time
	mu           sync.RWMutex
	ticker       *time.Ticker
	stop         chan struct{}
}

func NewCache(cleanupInterval time.Duration) *Cache {
	c := &Cache{
		fingerprints: make(map[string]time.Time),
		ticker:       time.NewTicker(cleanupInterval),
		stop:         make(chan struct{}),
	}

	go c.startCleanupWorker()

	return c
}

func (c *Cache) Add(fingerprint string, expiry time.Duration) {
	expirationTime := time.Now().Add(expiry)

	c.mu.Lock()
	defer c.mu.Unlock()

	c.fingerprints[fingerprint] = expirationTime
}

func (c *Cache) Remove(fingerprint string) error {
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

func (c *Cache) startCleanupWorker() {
	for {
		select {
		case <-c.ticker.C:
			c.Cleanup()
		case <-c.stop:
			c.ticker.Stop()
			return
		}
	}
}

func (c *Cache) Cleanup() {
	now := time.Now()

	c.mu.Lock()
	defer c.mu.Unlock()

	for fingerprint, expirationTime := range c.fingerprints {
		if now.After(expirationTime) {
			delete(c.fingerprints, fingerprint)
		}
	}
}

func (c *Cache) Stop() {
	close(c.stop)
}
