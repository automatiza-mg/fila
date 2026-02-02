package cache

import (
	"errors"
	"testing"
	"time"
)

func TestMemoryCache_PutAndGet(t *testing.T) {
	t.Parallel()
	c := NewMemoryCache()

	key := "test-key"
	val := []byte("hello world")

	err := c.Put(t.Context(), key, val, 1*time.Minute)
	if err != nil {
		t.Fatalf("failed to put to cache: %v", err)
	}

	got, err := c.Get(t.Context(), key)
	if err != nil {
		t.Fatalf("failed to get from cache: %v", err)
	}

	if string(got) != string(val) {
		t.Errorf("want %s, got %s", val, got)
	}
}

func TestMemoryCache_CacheMiss(t *testing.T) {
	t.Parallel()
	c := NewMemoryCache()

	_, err := c.Get(t.Context(), "non-existent")
	if !errors.Is(err, ErrCacheMiss) {
		t.Errorf("expected ErrCacheMiss, got %v", err)
	}
}

func TestMemoryCache_Expiration(t *testing.T) {
	t.Parallel()
	c := NewMemoryCache()

	key := "expiring-key"

	// Usamos um TTL curto para o teste e esperamos tempo suficiente passar
	c.Put(t.Context(), key, []byte("data"), 10*time.Millisecond)
	time.Sleep(20 * time.Millisecond)

	_, err := c.Get(t.Context(), key)
	if !errors.Is(err, ErrCacheMiss) {
		t.Error("key should have expired")
	}
}

func TestMemoryCache_Remember(t *testing.T) {
	t.Parallel()
	c := NewMemoryCache()

	key := "remember-key"
	calls := 0

	fn := func() ([]byte, error) {
		calls++
		return []byte("fresh-data"), nil
	}

	data, _ := c.Remember(t.Context(), key, 1*time.Minute, fn)
	if string(data) != "fresh-data" || calls != 1 {
		t.Errorf("failed to remember. calls: %d", calls)
	}

	data, _ = c.Remember(t.Context(), key, 1*time.Minute, fn)
	if string(data) != "fresh-data" || calls != 1 {
		t.Errorf("failed to return correct data, calls: %d", calls)
	}
}

func TestMemoryCache_Del(t *testing.T) {
	t.Parallel()
	c := NewMemoryCache()

	key := "delete-me"
	val := []byte("temporary data")

	c.Put(t.Context(), key, val, 1*time.Minute)
	err := c.Del(t.Context(), key)
	if err != nil {
		t.Fatalf("failed to delete key: %v", err)
	}

	_, err = c.Get(t.Context(), key)
	if !errors.Is(err, ErrCacheMiss) {
		t.Errorf("expected ErrCacheMiss, got: %v", err)
	}

	err = c.Del(t.Context(), "non-existent-key")
	if err != nil {
		t.Fatal(err)
	}
}
