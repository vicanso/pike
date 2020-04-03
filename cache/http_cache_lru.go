// Copyright 2019 tree xie
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// copy from https://github.com/golang/groupcache/blob/master/lru/lru.go
// 此模块中主要实现lru方式的HTTP缓存，避免缓存过大占用过多内存

package cache

import (
	"container/list"
	"sync"
)

// HTTPCacheLRU is an LRU cache. It is not safe for concurrent access.
type HTTPCacheLRU struct {
	// 用于保证每个dispatcher的操作，避免同时操作lru cache
	mu sync.Mutex
	// MaxEntries is the maximum number of cache entries before
	// an item is evicted. Zero means no limit.
	MaxEntries int

	ll    *list.List
	cache map[string]*list.Element
}

// Iterator iterator function
type Iterator func(key string, value *HTTPCache)

type entry struct {
	key   string
	value *HTTPCache
}

// NewHTTPCacheLRU creates a new Cache.
// If maxEntries is zero, the cache has no limit and it's assumed
// that eviction is done by the caller.
func NewHTTPCacheLRU(maxEntries int) *HTTPCacheLRU {
	return &HTTPCacheLRU{
		MaxEntries: maxEntries,
		ll:         list.New(),
	}
}

// Lock mutex lock
func (c *HTTPCacheLRU) Lock() {
	c.mu.Lock()
}

// Unlock mutex unlock
func (c *HTTPCacheLRU) Unlock() {
	c.mu.Unlock()
}

// FindOrCreate find or create http cache
// 需要注意，lru其它方法中并没有锁的处理
func (c *HTTPCacheLRU) FindOrCreate(key string) *HTTPCache {
	c.Lock()
	defer c.Unlock()
	cache, ok := c.Get(key)
	if !ok {
		cache = NewHTTPCache()
		c.Add(key, cache)
	}
	return cache
}

// Add adds a value to the cache.
func (c *HTTPCacheLRU) Add(key string, value *HTTPCache) {
	if c.cache == nil {
		c.cache = make(map[string]*list.Element)
		c.ll = list.New()
	}
	if ee, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ee)
		ee.Value.(*entry).value = value
		return
	}
	ele := c.ll.PushFront(&entry{key, value})
	c.cache[key] = ele
	if c.MaxEntries != 0 && c.ll.Len() > c.MaxEntries {
		c.RemoveOldest()
	}
}

// Get looks up a key's value from the cache.
func (c *HTTPCacheLRU) Get(key string) (value *HTTPCache, ok bool) {
	if c.cache == nil {
		return
	}
	if ele, hit := c.cache[key]; hit {
		c.ll.MoveToFront(ele)
		return ele.Value.(*entry).value, true
	}
	return
}

// Remove removes the provided key from the cache.
func (c *HTTPCacheLRU) Remove(key string) {
	if c.cache == nil {
		return
	}
	if ele, hit := c.cache[key]; hit {
		c.removeElement(ele)
	}
}

// RemoveOldest removes the oldest item from the cache.
func (c *HTTPCacheLRU) RemoveOldest() {
	if c.cache == nil {
		return
	}
	ele := c.ll.Back()
	if ele != nil {
		c.removeElement(ele)
	}
}

func (c *HTTPCacheLRU) removeElement(e *list.Element) {
	c.ll.Remove(e)
	kv := e.Value.(*entry)
	delete(c.cache, kv.key)
}

// Len returns the number of items in the cache.
func (c *HTTPCacheLRU) Len() int {
	if c.cache == nil {
		return 0
	}
	return c.ll.Len()
}

// Clear purges all stored items from the cache.
func (c *HTTPCacheLRU) Clear() {
	c.ll = nil
	c.cache = nil
}

// ForEach for each
func (c *HTTPCacheLRU) ForEach(fn Iterator) {
	for _, e := range c.cache {
		kv := e.Value.(*entry)
		fn(kv.key, kv.value)
	}
}

// RemoveExpired remove expired cache
func (c *HTTPCacheLRU) RemoveExpired() int {
	c.Lock()
	defer c.Unlock()
	count := 0
	c.ForEach(func(key string, item *HTTPCache) {
		if item.IsExpired() {
			c.Remove(key)
			count++
		}
	})
	return count
}
