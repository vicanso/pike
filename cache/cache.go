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

// http cache

package cache

import (
	"strings"
	"time"

	"github.com/vicanso/pike/config"
	"github.com/vicanso/pike/util"
)

const (
	defaultSize       = 10
	defaultZoneSize   = 1024
	defaultHitForPass = 300
)

type (
	// Dispatcher http cache dispatcher
	Dispatcher struct {
		HitForPass int
		size       uint64
		list       []*HTTPCacheLRU
	}
	// Dispatchers http cache dispatcher list
	Dispatchers struct {
		dispatchers map[string]*Dispatcher
	}
	// CacheSummary cache summary
	CacheSummary struct {
		Key       string `json:"key,omitempty"`
		CreatedAt int    `json:"createdAt,omitempty"`
		ExpiredAt int    `json:"expiredAt,omitempty"`
	}
)

// NewDispatcher new a dispatcher
func NewDispatcher(cacheConfig *config.Cache) *Dispatcher {
	size := defaultSize
	zoneSize := defaultZoneSize
	hitForPass := defaultHitForPass
	if cacheConfig != nil {
		if cacheConfig.Size > 0 {
			size = cacheConfig.Size
		}
		if cacheConfig.Zone > 0 {
			zoneSize = cacheConfig.Zone
		}
		if cacheConfig.HitForPass > 0 {
			hitForPass = cacheConfig.HitForPass
		}
	}

	// 按zoneSize与size创建二维缓存，存放的是LRU缓存实例
	list := make([]*HTTPCacheLRU, size)
	for i := 0; i < size; i++ {
		list[i] = NewHTTPCacheLRU(zoneSize)
	}

	return &Dispatcher{
		HitForPass: hitForPass,
		size:       uint64(size),
		list:       list,
	}
}

func (d *Dispatcher) getLRU(key []byte) *HTTPCacheLRU {
	// 计算hash值
	index := MemHash(key) % d.size
	// 从预定义的列表中取对应的缓存
	return d.list[index]
}

// GetHTTPCache get http cache through key
func (d *Dispatcher) GetHTTPCache(key []byte) *HTTPCache {
	lru := d.getLRU(key)
	return lru.FindOrCreate(util.ByteSliceToString(key))
}

// RemoveExpired remove expired cache
func (d *Dispatcher) RemoveExpired() int {
	count := 0
	for _, lruCache := range d.list {
		count += lruCache.RemoveExpired()
	}
	return count
}

// List list the cache
func (d *Dispatcher) List(status int, limit int, keyword string) (summaryList []*CacheSummary) {
	count := 0
	done := false
	summaryList = make([]*CacheSummary, 0, limit)
	now := int(time.Now().Unix())
	for _, lruCache := range d.list {
		if done {
			break
		}
		lruCache.ForEach(func(key string, item *HTTPCache) {
			if done {
				return
			}
			// 状态不符合或者未过期
			if item.status != status || item.expiredAt < now {
				return
			}
			// 如果指定了筛选关键字
			if keyword != "" && !strings.Contains(key, keyword) {
				return
			}
			summaryList = append(summaryList, &CacheSummary{
				Key:       key,
				CreatedAt: item.createdAt,
				ExpiredAt: item.expiredAt,
			})
			count++
		})
	}
	return
}

// Remove remove the cache
func (d *Dispatcher) Remove(key []byte) {
	lru := d.getLRU(key)
	lru.Lock()
	defer lru.Unlock()
	lru.Remove(util.ByteSliceToString(key))
}

// Get get dispatcher
func (ds *Dispatchers) Get(name string) *Dispatcher {
	return ds.dispatchers[name]
}

// NewDispatchers new a dispatcher list
func NewDispatchers(cachesConfig config.Caches) (ds *Dispatchers) {
	ds = &Dispatchers{}
	dispatchers := make(map[string]*Dispatcher)
	for _, item := range cachesConfig {
		dispatchers[item.Name] = NewDispatcher(item)
	}
	ds.dispatchers = dispatchers
	return
}
