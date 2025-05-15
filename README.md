# In-Memory Cache System in Go

## Overview
This project implements a simple, performant in-memory cache system in Go. The cache supports storing key-value pairs, retrieving values by keys, and automatically removing expired items. It is designed to be thread-safe, handle errors appropriately, and follow Go idiomatic practices. The implementation is modular, well-documented, and easy to extend.

### Features
- **Key-Value Storage**: Store and retrieve key-value pairs with `Set` and `Get`.
- **Expiration**: Items can have a default or custom expiration time. Expired items are automatically removed.
- **Thread-Safety**: Uses `sync.RWMutex` to ensure safe concurrent access.
- **Error Handling**: Returns custom errors for common cases (e.g., `ErrKeyNotFound`, `ErrKeyExpired`).
- **Automatic Cleanup**: A background goroutine periodically removes expired items based on a configurable interval.
- **Additional Functionality**: Includes methods like `GetOrSet` for lazy computation, `Items` to list unexpired items, and `Flush` to clear the cache.

## Project Structure
- `cache.go`: Core cache implementation with methods like `Set`, `Get`, `Delete`, etc.
- `item.go`: Defines the `Item` struct for storing values and expiration times.
- `errors.go`: Defines custom errors used by the cache.
- `cache_test.go`: Unit tests and benchmarks for the cache implementation.
- `go.mod`: Module definition for the project.
- `main/main.go`: Example usage of the cache, demonstrating its features.

## Setup and Usage

### Prerequisites
- Go 1.21 or later installed.

### Running the Example
1. Clone the repository:
   ```bash
   git clone https://github.com/ashishjsharda/gocache.git
   cd gocache
   ```

2. Run the example program:
   ```bash
   go run main/main.go
   ```
   This will execute the example program, which demonstrates the cache's functionality, including setting values, expiration, lazy computation, and more.

### Running Tests
To run the unit tests and benchmarks:
```bash
go test .
```
This will execute all tests in cache_test.go, verifying the cache's functionality, thread-safety, and performance.

## Example Usage

Here's an example of how to use the cache (similar to the provided main/main.go file):

```go
package main

import (
    "fmt"
    "log"
    "time"
    "gocache"
)

func main() {
    // Create a new cache with default expiration of 5 minutes and cleanup every minute
    c := gocache.New(gocache.Options{
        DefaultExpiration: 5 * time.Minute,
        CleanupInterval:   1 * time.Minute,
    })
    defer c.Stop() // Stop the cleanup goroutine when done

    // Store some values in the cache
    c.Set("string", "hello world")
    c.Set("number", 42)
    c.Set("bool", true)

    // Store a value with a custom expiration
    c.SetWithExpiration("short-lived", "I'll expire soon", 2*time.Second)

    // Retrieve values from the cache
    printValue(c, "string")
    printValue(c, "number")
    printValue(c, "bool")
    printValue(c, "short-lived")
    printValue(c, "non-existent")

    fmt.Println("\nWaiting for the short-lived value to expire...")
    time.Sleep(3 * time.Second)

    printValue(c, "short-lived") // This should be expired now
    printValue(c, "string")      // This should still exist

    // Use GetOrSet for lazy computation
    _, err := c.GetOrSet("computed", func() (interface{}, error) {
        fmt.Println("Computing value...")
        // Simulate expensive computation
        time.Sleep(100 * time.Millisecond)
        return "computed value", nil
    })
    if err != nil {
        log.Fatal(err)
    }

    // First call should have computed the value, second call should use the cache
    printValue(c, "computed")
    
    // Second call will use cached value
    _, err = c.GetOrSet("computed", func() (interface{}, error) {
        fmt.Println("Computing value again... (this shouldn't be displayed)")
        return "new computed value", nil
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Get all remaining items
    fmt.Println("\nAll items in cache:")
    items := c.Items()
    for k, v := range items {
        fmt.Printf("%s: %v\n", k, v)
    }

    // Flush the cache
    fmt.Println("\nFlushing cache...")
    c.Flush()
    fmt.Printf("Items in cache after flush: %d\n", c.ItemCount())
}

func printValue(c *gocache.Cache, key string) {
    value, err := c.Get(key)
    if err != nil {
        fmt.Printf("Key '%s': %v\n", key, err)
        return
    }
    fmt.Printf("Key '%s': %v\n", key, value)
}
```

## Testing

The project includes comprehensive tests in cache_test.go:
* Unit tests for setting/getting values, expiration, deletion, concurrency, and more.
* Benchmarks for Get and Set operations to measure performance.

Run `go test` to execute all tests.

## Reflection

For a detailed discussion of what I would do differently or add with more time, please see [Reflection.md](Reflection.md).

This document outlines potential improvements in:
- Performance enhancements
- Advanced features
- Code quality improvements
- Advanced capabilities

## Notes
* The gocache module is local to this repository and does not require external dependencies.
* The implementation prioritizes simplicity and performance while ensuring thread-safety and proper error handling.
* All operations are thread-safe and can be called from multiple goroutines.
