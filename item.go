package gocache

import (
	"time"
)

// Item represents a value stored in the cache along with its expiration time.
type Item struct {
	Value      interface{}
	Expiration int64 // Unix timestamp in nanoseconds
}

// Expired returns true if the item has expired.
func (item *Item) Expired() bool {
	if item.Expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > item.Expiration
}
