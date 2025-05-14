package gocache

import (
	"errors"
)

// Common errors returned by the cache.
var (
	ErrKeyNotFound = errors.New("key not found in cache")
	ErrKeyExpired  = errors.New("key has expired")
	ErrNilValue    = errors.New("nil value is not allowed")
)
