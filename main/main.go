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
	// Using _ to ignore the returned value since we don't use it
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

	// Using _ again to ignore the returned value
	_, err = c.GetOrSet("computed", func() (interface{}, error) {
		fmt.Println("Computing value again... (this shouldn't be displayed)")
		return "new computed value", nil
	})
	if err != nil {
		log.Fatal(err)
	}
	printValue(c, "computed") // Should still show the original computed value

	// Delete a value
	fmt.Println("\nDeleting 'string' from cache")
	c.Delete("string")
	printValue(c, "string") // Should be not found

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
