package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// SimpleCache provides a simple in-memory and file-based cache
type SimpleCache struct {
	data map[string]cacheEntry
	dir  string
}

type cacheEntry struct {
	Data      interface{}
	ExpiresAt time.Time
}

// NewSimpleCache creates a new cache instance
func NewSimpleCache(dir string) *SimpleCache {
	return &SimpleCache{
		data: make(map[string]cacheEntry),
		dir:  dir,
	}
}

// Get retrieves an item from the cache
func (c *SimpleCache) Get(key string) (interface{}, bool) {
	// Check memory first
	if entry, ok := c.data[key]; ok {
		if time.Now().Before(entry.ExpiresAt) {
			return entry.Data, true
		}
		// Expired, remove from memory
		delete(c.data, key)
	}

	// Try to load from file
	if c.dir != "" {
		path := c.filePath(key)
		data, err := os.ReadFile(path)
		if err == nil {
			var entry cacheEntry
			if err := json.Unmarshal(data, &entry); err == nil && time.Now().Before(entry.ExpiresAt) {
				// Restore to memory
				c.data[key] = entry
				return entry.Data, true
			}
			// Remove expired or invalid cache file
			_ = os.Remove(path)
		}
	}

	return nil, false
}

// Set stores an item in the cache with TTL
func (c *SimpleCache) Set(key string, value interface{}, ttl time.Duration) {
	entry := cacheEntry{
		Data:      value,
		ExpiresAt: time.Now().Add(ttl),
	}

	// Store in memory
	c.data[key] = entry

	// Persist to file if directory is set
	if c.dir != "" {
		path := c.filePath(key)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err == nil {
			if data, err := json.Marshal(entry); err == nil {
				_ = os.WriteFile(path, data, 0644)
			}
		}
	}
}

// filePath returns the file path for a cache key
func (c *SimpleCache) filePath(key string) string {
	// Use first 2 chars of key as subdirectory
	if len(key) < 2 {
		return filepath.Join(c.dir, key)
	}
	return filepath.Join(c.dir, key[:2], key)
}

// CacheDir returns the cache directory path
func CacheDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".cache", "usm")
}
