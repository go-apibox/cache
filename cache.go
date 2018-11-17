// REF: https://github.com/wunderlist/ttlcache

package cache

import (
	"sync"
	"time"
)

// Cache is a synchronised map of items that auto-expire once stale
type Cache struct {
	mutex           sync.RWMutex
	ttl             time.Duration
	cleanupInterval time.Duration
	items           map[string]*Item
}

// Set is a thread-safe way to add new items to the map
func (cache *Cache) Set(key string, data interface{}) {
	cache.mutex.Lock()
	item := &Item{data: data}
	item.touch(cache.ttl)
	cache.items[key] = item
	cache.mutex.Unlock()
}

// SetIfNotExist is a thread-safe way to add new items to the map.
// Add successfully only when item is not exists.
func (cache *Cache) SetIfNotExist(key string, data interface{}) (ok bool) {
	cache.mutex.Lock()

	item, exists := cache.items[key]
	if exists && !item.expired() {
		cache.mutex.Unlock()
		return false
	}

	item = &Item{data: data}
	item.touch(cache.ttl)
	cache.items[key] = item

	cache.mutex.Unlock()

	return true
}

// Get is a thread-safe way to lookup items
// Every lookup, also touches the item, hence extending it's life
func (cache *Cache) Get(key string) (item *Item, found bool) {
	cache.mutex.Lock()
	item, exists := cache.items[key]
	if !exists || item.expired() {
		item = nil
		found = false
	} else {
		item.touch(cache.ttl)
		found = true
	}
	cache.mutex.Unlock()
	return
}

// Has is a thread-safe way to check if item exists.
func (cache *Cache) Has(key string) (found bool) {
	cache.mutex.Lock()
	item, exists := cache.items[key]
	if !exists || item.expired() {
		found = false
	} else {
		found = true
	}
	cache.mutex.Unlock()
	return
}

// Count returns the number of items in the cache
// (helpful for tracking memory leaks)
func (cache *Cache) Count() int {
	cache.mutex.RLock()
	count := len(cache.items)
	cache.mutex.RUnlock()
	return count
}

func (cache *Cache) cleanup() {
	cache.mutex.Lock()
	for key, item := range cache.items {
		if item.expired() {
			delete(cache.items, key)
		}
	}
	cache.mutex.Unlock()
}

func (cache *Cache) startCleanupTimer() {
	duration := cache.cleanupInterval
	if duration < time.Second {
		duration = time.Second
	}
	ticker := time.Tick(duration)
	go (func() {
		for {
			select {
			case <-ticker:
				cache.cleanup()
			}
		}
	})()
}

// NewCache is a helper to create instance of the Cache struct
func NewCache(expire time.Duration) *Cache {
	cache := &Cache{
		ttl:             expire,
		cleanupInterval: expire,
		items:           map[string]*Item{},
	}
	cache.startCleanupTimer()
	return cache
}

// NewCacheEx is a helper to create instance of the Cache struct
// with specified cleanup interval
func NewCacheEx(expire, cleanupInterval time.Duration) *Cache {
	cache := &Cache{
		ttl:             expire,
		cleanupInterval: cleanupInterval,
		items:           map[string]*Item{},
	}
	cache.startCleanupTimer()
	return cache
}
