package gocache

import (
	"testing"
	"time"
)

func TestCacheSetGet(t *testing.T) {
	cache := New(Options{
		DefaultExpiration: 5 * time.Minute,
		CleanupInterval:   1 * time.Minute,
	})
	defer cache.Stop()

	// Test setting and getting values
	err := cache.Set("key1", "value1")
	if err != nil {
		t.Errorf("Failed to set key1: %v", err)
	}

	value, err := cache.Get("key1")
	if err != nil {
		t.Errorf("Failed to get key1: %v", err)
	}
	if value != "value1" {
		t.Errorf("Expected 'value1', got '%v'", value)
	}

	// Test getting a non-existent key
	_, err = cache.Get("nonexistent")
	if err != ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound, got %v", err)
	}
}

func TestCacheExpiration(t *testing.T) {
	cache := New(Options{
		DefaultExpiration: 100 * time.Millisecond,
		CleanupInterval:   0, // Disable automatic cleanup for this test
	})

	// Set a key with a short expiration
	cache.Set("short", "expiring-value")

	// Set a key with a longer expiration
	cache.SetWithExpiration("long", "persistent-value", 1*time.Hour)

	// Set a key with no expiration
	cache.SetWithExpiration("forever", "eternal-value", 0)

	// Verify all keys exist
	_, err := cache.Get("short")
	if err != nil {
		t.Errorf("Failed to get 'short' key: %v", err)
	}

	// Wait for the short key to expire
	time.Sleep(200 * time.Millisecond)

	// The short key should be expired now
	_, err = cache.Get("short")
	if err != ErrKeyExpired {
		t.Errorf("Expected ErrKeyExpired for 'short' key, got %v", err)
	}

	// The long and forever keys should still exist
	_, err = cache.Get("long")
	if err != nil {
		t.Errorf("Failed to get 'long' key: %v", err)
	}

	_, err = cache.Get("forever")
	if err != nil {
		t.Errorf("Failed to get 'forever' key: %v", err)
	}
}

func TestCacheDelete(t *testing.T) {
	cache := New(Options{DefaultExpiration: time.Minute})

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	// Delete key1
	deleted := cache.Delete("key1")
	if !deleted {
		t.Error("Delete returned false, expected true")
	}

	// Try to get the deleted key
	_, err := cache.Get("key1")
	if err != ErrKeyNotFound {
		t.Errorf("Expected ErrKeyNotFound, got %v", err)
	}

	// key2 should still exist
	value, err := cache.Get("key2")
	if err != nil {
		t.Errorf("Failed to get key2: %v", err)
	}
	if value != "value2" {
		t.Errorf("Expected 'value2', got '%v'", value)
	}
}

func TestCacheGetOrSet(t *testing.T) {
	cache := New(Options{DefaultExpiration: time.Minute})

	// First call should compute the value
	computeCount := 0
	getValue := func() (interface{}, error) {
		computeCount++
		return "computed-value", nil
	}

	value, err := cache.GetOrSet("compute-key", getValue)
	if err != nil {
		t.Errorf("Failed to GetOrSet: %v", err)
	}
	if value != "computed-value" {
		t.Errorf("Expected 'computed-value', got '%v'", value)
	}
	if computeCount != 1 {
		t.Errorf("Expected compute function to be called once, got %d", computeCount)
	}

	// Second call should use the cached value
	value, err = cache.GetOrSet("compute-key", getValue)
	if err != nil {
		t.Errorf("Failed to GetOrSet: %v", err)
	}
	if value != "computed-value" {
		t.Errorf("Expected 'computed-value', got '%v'", value)
	}
	if computeCount != 1 {
		t.Errorf("Expected compute function to be called once, got %d", computeCount)
	}
}

func TestCacheConcurrency(t *testing.T) {
	cache := New(Options{DefaultExpiration: time.Minute})
	done := make(chan bool)

	// Concurrent reads and writes
	for i := 0; i < 10; i++ {
		go func(index int) {
			for j := 0; j < 100; j++ {
				key := "key"
				cache.Set(key, j)
				cache.Get(key)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestCacheDeleteExpired(t *testing.T) {
	cache := New(Options{
		DefaultExpiration: 100 * time.Millisecond,
		CleanupInterval:   0, // Disable automatic cleanup for this test
	})

	// Add items with expiration
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.SetWithExpiration("key3", "value3", 1*time.Hour)

	// Wait for default expiration
	time.Sleep(200 * time.Millisecond)

	// Verify count before deletion
	count := cache.ItemCount()
	if count != 3 {
		t.Errorf("Expected 3 items, got %d", count)
	}

	// Delete expired items
	cache.DeleteExpired()

	// Verify count after deletion
	count = cache.ItemCount()
	if count != 1 {
		t.Errorf("Expected 1 item after expiration, got %d", count)
	}

	// Verify key3 still exists
	value, err := cache.Get("key3")
	if err != nil {
		t.Errorf("Failed to get key3: %v", err)
	}
	if value != "value3" {
		t.Errorf("Expected 'value3', got '%v'", value)
	}
}

func TestCacheItems(t *testing.T) {
	cache := New(Options{DefaultExpiration: time.Minute})

	cache.Set("key1", "value1")
	cache.Set("key2", 123)
	cache.Set("key3", true)

	items := cache.Items()
	if len(items) != 3 {
		t.Errorf("Expected 3 items, got %d", len(items))
	}

	if items["key1"] != "value1" {
		t.Errorf("Expected key1='value1', got '%v'", items["key1"])
	}
	if items["key2"] != 123 {
		t.Errorf("Expected key2=123, got '%v'", items["key2"])
	}
	if items["key3"] != true {
		t.Errorf("Expected key3=true, got '%v'", items["key3"])
	}
}

func TestCacheFlush(t *testing.T) {
	cache := New(Options{DefaultExpiration: time.Minute})

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	count := cache.ItemCount()
	if count != 2 {
		t.Errorf("Expected 2 items, got %d", count)
	}

	cache.Flush()

	count = cache.ItemCount()
	if count != 0 {
		t.Errorf("Expected 0 items after flush, got %d", count)
	}
}

func TestCacheNilValue(t *testing.T) {
	cache := New(Options{DefaultExpiration: time.Minute})

	err := cache.Set("nil-key", nil)
	if err != ErrNilValue {
		t.Errorf("Expected ErrNilValue, got %v", err)
	}
}

func BenchmarkCacheGet(b *testing.B) {
	cache := New(Options{DefaultExpiration: time.Minute})
	cache.Set("key", "value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get("key")
	}
}

func BenchmarkCacheSet(b *testing.B) {
	cache := New(Options{DefaultExpiration: time.Minute})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set("key", "value")
	}
}
