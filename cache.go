package gocache

import (
	"sync"
	"time"
)

// Cache represents an in-memory cache with expiration.
type Cache struct {
	items             map[string]Item
	mu                sync.RWMutex
	defaultExpiration time.Duration
	cleanupInterval   time.Duration
	stopCleanup       chan bool
}

// Options contains configuration options for creating a new cache.
type Options struct {
	// DefaultExpiration is the default duration after which cache items expire.
	// If 0, items never expire by default.
	DefaultExpiration time.Duration

	// CleanupInterval is the interval between automatic cleanup of expired items.
	// If 0, expired items are not cleaned up automatically.
	CleanupInterval time.Duration
}

// New creates a new Cache with the specified default expiration and cleanup interval.
// If cleanupInterval > 0, a background goroutine will be started to clean up expired
// items at the specified interval.
func New(options Options) *Cache {
	c := &Cache{
		items:             make(map[string]Item),
		defaultExpiration: options.DefaultExpiration,
		cleanupInterval:   options.CleanupInterval,
		stopCleanup:       make(chan bool),
	}

	// Start cleanup routine if cleanup interval is specified
	if options.CleanupInterval > 0 {
		go c.startCleanupRoutine()
	}

	return c
}

// startCleanupRoutine starts a background goroutine that will periodically
// delete expired items from the cache.
func (c *Cache) startCleanupRoutine() {
	ticker := time.NewTicker(c.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.DeleteExpired()
		case <-c.stopCleanup:
			return
		}
	}
}

// Set adds an item to the cache with the specified key and value.
// The item will expire after the DefaultExpiration time has passed.
func (c *Cache) Set(key string, value interface{}) error {
	return c.SetWithExpiration(key, value, c.defaultExpiration)
}

// SetWithExpiration adds an item to the cache with the specified key, value, and expiration duration.
// If duration is 0, the item never expires.
func (c *Cache) SetWithExpiration(key string, value interface{}, duration time.Duration) error {
	if value == nil {
		return ErrNilValue
	}

	var expiration int64
	if duration > 0 {
		expiration = time.Now().Add(duration).UnixNano()
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = Item{
		Value:      value,
		Expiration: expiration,
	}

	return nil
}

// Get returns the value stored in the cache for the given key.
// Returns ErrKeyNotFound if the key does not exist or ErrKeyExpired if the key has expired.
func (c *Cache) Get(key string) (interface{}, error) {
	c.mu.RLock()
	item, found := c.items[key]
	c.mu.RUnlock()

	if !found {
		return nil, ErrKeyNotFound
	}

	if item.Expired() {
		// Delete the key if it's expired
		c.mu.Lock()
		// Check again after acquiring write lock to prevent race condition
		if item, found := c.items[key]; found && item.Expired() {
			delete(c.items, key)
		}
		c.mu.Unlock()
		return nil, ErrKeyExpired
	}

	return item.Value, nil
}

// GetOrSet gets the value from the cache if it exists and is not expired.
// Otherwise, it sets the value using the provided function and returns it.
func (c *Cache) GetOrSet(key string, fn func() (interface{}, error)) (interface{}, error) {
	// Try to get the value from the cache first
	value, err := c.Get(key)
	if err == nil {
		// Value found and not expired
		return value, nil
	}

	// Value not found or expired, compute it
	value, err = fn()
	if err != nil {
		return nil, err
	}

	// Store the computed value in the cache
	err = c.Set(key, value)
	if err != nil {
		return nil, err
	}

	return value, nil
}

// Delete removes the item with the given key from the cache.
// It returns true if the key was found and deleted.
func (c *Cache) Delete(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, found := c.items[key]
	if found {
		delete(c.items, key)
		return true
	}
	return false
}

// DeleteExpired removes all expired items from the cache.
func (c *Cache) DeleteExpired() {
	now := time.Now().UnixNano()
	c.mu.Lock()
	defer c.mu.Unlock()

	for k, v := range c.items {
		if v.Expiration > 0 && now > v.Expiration {
			delete(c.items, k)
		}
	}
}

// Items returns a copy of all unexpired items in the cache.
func (c *Cache) Items() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	items := make(map[string]interface{}, len(c.items))
	now := time.Now().UnixNano()

	for k, v := range c.items {
		if v.Expiration == 0 || now < v.Expiration {
			items[k] = v.Value
		}
	}

	return items
}

// ItemCount returns the number of items in the cache, including expired items.
func (c *Cache) ItemCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// Flush removes all items from the cache.
func (c *Cache) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]Item)
}

// Stop stops the automatic cleanup goroutine.
func (c *Cache) Stop() {
	if c.cleanupInterval > 0 {
		c.stopCleanup <- true
	}
}
