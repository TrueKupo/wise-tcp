package hashcash_test

import (
	"testing"
	"time"

	"wise-tcp/internal/pow/hashcash"
)

func TestCache_AddAndRemove(t *testing.T) {
	cache := hashcash.NewCache(10 * time.Second)
	defer cache.Stop()

	testFingerprint := "test123"

	cache.Add(testFingerprint, 1*time.Hour)

	err := cache.Remove(testFingerprint)
	if err != nil {
		t.Fatalf("expected to remove fingerprint successfully, but got error: %v", err)
	}

	err = cache.Remove(testFingerprint)
	if err == nil || err.Error() != "fingerprint not found in cache" {
		t.Fatalf("expected 'fingerprint not found' error, but got: %v", err)
	}
}

func TestCache_Expiration(t *testing.T) {
	cache := hashcash.NewCache(10 * time.Second)
	defer cache.Stop()

	testFingerprint := "expiring123"

	cache.Add(testFingerprint, 100*time.Millisecond)

	err := cache.Remove(testFingerprint)
	if err != nil {
		t.Fatalf("expected fingerprint to exist before expiration, but got: %v", err)
	}

	cache.Add(testFingerprint, 100*time.Millisecond)

	time.Sleep(150 * time.Millisecond)

	err = cache.Remove(testFingerprint)
	if err == nil || err.Error() != "fingerprint expired" {
		t.Fatalf("expected 'fingerprint expired' error, but got: %v", err)
	}
}

func TestCache_Cleanup(t *testing.T) {
	cache := hashcash.NewCache(50 * time.Millisecond)
	defer cache.Stop()

	testFingerprint := "cleanup123"

	cache.Add(testFingerprint, 20*time.Millisecond)

	time.Sleep(100 * time.Millisecond)

	err := cache.Remove(testFingerprint)
	if err == nil || err.Error() != "fingerprint not found in cache" {
		t.Fatalf("expected 'fingerprint not found' after cleanup, but got: %v", err)
	}
}

func TestCache_Stop(t *testing.T) {
	cache := hashcash.NewCache(10 * time.Millisecond)
	cache.Stop()

	fingerprint := "stopTest"
	cache.Add(fingerprint, 10*time.Millisecond)
	err := cache.Remove(fingerprint)
	if err != nil {
		t.Fatalf("expected to remove fingerprint successfully after stopping the cleanup worker, but got: %v", err)
	}
}

func TestCache_Concurrency(t *testing.T) {
	cache := hashcash.NewCache(10 * time.Second)
	defer cache.Stop()

	testFingerprint := "concurrent123"

	cache.Add(testFingerprint, 1*time.Second)

	done := make(chan error, 2)

	go func() {
		done <- cache.Remove(testFingerprint)
	}()
	go func() {
		done <- cache.Remove(testFingerprint)
	}()

	err1 := <-done
	err2 := <-done

	if err1 != nil {
		t.Fatalf("unexpected error in first goroutine: %v", err1)
	}
	if err2 == nil {
		t.Fatalf("expected error in second goroutine, but got no error")
	}
	if err2.Error() != "fingerprint not found in cache" {
		t.Fatalf("unexpected error in second goroutine: %v", err2)
	}
}
