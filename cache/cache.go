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

// GetHTTPCache get http cache through key
func (d *Dispatcher) GetHTTPCache(key []byte) *HTTPCache {
	// 计算hash值
	index := MemHash(key) % d.size
	// 从预定义的列表中取对应的缓存
	lru := d.list[index]
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
