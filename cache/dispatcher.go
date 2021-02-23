// MIT License

// Copyright (c) 2020 Tree Xie

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// 创建缓存分发组件，初始化时创建128长度的lru缓存数组，每次根据缓存的key生成hash，
// 根据hash的值判断使用对应的lru，减少锁的冲突，提升性能

package cache

import (
	"sync"

	"github.com/golang/groupcache/lru"
	"github.com/vicanso/pike/log"
	"github.com/vicanso/pike/store"
	"github.com/vicanso/pike/util"
	"go.uber.org/zap"
)

// defaultZoneSize default zone size
const defaultZoneSize = 128

type (
	// httpLRUCache http lru cache
	httpLRUCache struct {
		cache *lru.Cache
		mu    *sync.Mutex
	}
	// dispatcher http cache dispatcher
	dispatcher struct {
		zoneSize   uint64
		hitForPass int
		list       []*httpLRUCache
		store      store.Store
	}
	// dispatchers http cache dispatchers
	dispatchers struct {
		m *sync.Map
	}
	// DispatcherOption dispatcher option
	DispatcherOption struct {
		Name       string
		Size       int
		HitForPass int
		Store      string
	}
)

func newHTTPLRUCache(size int) *httpLRUCache {
	c := &httpLRUCache{
		cache: lru.New(size),
		mu:    &sync.Mutex{},
	}
	return c
}

// getCache get http cache by key
func (lru *httpLRUCache) getCache(key []byte) (*httpCache, bool) {
	value, ok := lru.cache.Get(byteSliceToString(key))
	if !ok {
		return nil, false
	}
	if hc, ok := value.(*httpCache); ok {
		return hc, true
	}
	return nil, false
}

// addCache add http cache by key
func (lru *httpLRUCache) addCache(key []byte, hc *httpCache) {
	lru.cache.Add(byteSliceToString(key), hc)
}

// removeCache remove http cache by key
func (lru *httpLRUCache) removeCache(key []byte) {
	lru.cache.Remove(byteSliceToString(key))
}

// NewDispatcher new a http cache dispatcher
func NewDispatcher(option DispatcherOption) *dispatcher {
	zoneSize := defaultZoneSize
	size := option.Size
	if option.Size <= 0 {
		size = zoneSize * 100
	}
	// 如果配置lru缓存数量较小，则zone的空间调小
	if size < 1024 {
		zoneSize = 8
	}

	// 按zoneSize与size创建二维缓存，存放的是LRU缓存实例
	lruSize := size / zoneSize
	list := make([]*httpLRUCache, zoneSize)
	// 根据zone size生成一个缓存对列
	for i := 0; i < zoneSize; i++ {
		list[i] = newHTTPLRUCache(lruSize)
	}
	disp := &dispatcher{
		zoneSize:   uint64(zoneSize),
		list:       list,
		hitForPass: option.HitForPass,
	}
	// 如果有配置store
	if option.Store != "" {
		store, err := store.NewStore(option.Store)
		if err != nil {
			log.Default().Error("new store fail",
				zap.String("url", option.Store),
				zap.Error(err),
			)
		}
		if store != nil {
			disp.store = store
		}
	}
	return disp
}

func (d *dispatcher) getLRU(key []byte) *httpLRUCache {
	// 计算hash值
	index := MemHash(key) % d.zoneSize
	// 从预定义的列表中取对应的缓存
	return d.list[index]
}

// GetHTTPCache get http cache through key
func (d *dispatcher) GetHTTPCache(key []byte) *httpCache {
	// 锁只在public的方法在使用，public方法之间不互相调用
	lru := d.getLRU(key)
	lru.mu.Lock()
	defer lru.mu.Unlock()
	hc, ok := lru.getCache(key)
	if ok {
		return hc
	}
	if d.store != nil {
		hc = NewHTTPStoreCache(key, d.store)
	} else {
		hc = NewHTTPCache()
	}
	lru.addCache(key, hc)
	return hc
}

// RemoveHTTPCache remove http cache
func (d *dispatcher) RemoveHTTPCache(key []byte) {
	lru := d.getLRU(key)
	lru.mu.Lock()
	defer lru.mu.Unlock()
	lru.removeCache(key)
	if d.store != nil {
		err := d.store.Delete(key)
		if err != nil {
			log.Default().Error("delete from store fail",
				zap.String("key", string(key)),
				zap.Error(err),
			)
		}
	}
}

// GetHitForPass get hit for pass
func (d *dispatcher) GetHitForPass() int {
	return d.hitForPass
}

// NewDispatchers new dispatchers
func NewDispatchers(opts []DispatcherOption) *dispatchers {
	ds := &dispatchers{
		m: &sync.Map{},
	}
	for _, opt := range opts {
		ds.m.Store(opt.Name, NewDispatcher(opt))
	}
	return ds
}

// Get get dispatcher by name
func (ds *dispatchers) Get(name string) *dispatcher {
	value, ok := ds.m.Load(name)
	if !ok {
		return nil
	}
	d, ok := value.(*dispatcher)
	if !ok {
		return nil
	}
	return d
}

// RemoveHTTPCache remove http cache
func (ds *dispatchers) RemoveHTTPCache(name string, key []byte) {
	if name != "" {
		d := ds.Get(name)
		if d == nil {
			return
		}
		d.RemoveHTTPCache(key)
		return
	}
	// 如果未指定名称，则从所有缓存中删除
	ds.m.Range(func(_, v interface{}) bool {
		d, ok := v.(*dispatcher)
		if ok {
			d.RemoveHTTPCache(key)
		}
		return true
	})
}

// Reset reset the dispatchers, remove not exists dispatchers and create new dispatcher. If the dispatcher is exists, then use the old one.
func (ds *dispatchers) Reset(opts []DispatcherOption) {
	// 删除不再使用的dispatcher
	_ = util.MapDelete(ds.m, func(key string) bool {
		// 如果不存在的，则删除
		exists := false
		for _, opt := range opts {
			if opt.Name == key {
				exists = true
				break
			}
		}
		return !exists
	})

	for _, opt := range opts {
		_, ok := ds.m.Load(opt.Name)
		// 如果当前dispatcher不存在，则创建
		// 如果存在，对原来的size不调整
		if !ok {
			ds.m.Store(opt.Name, NewDispatcher(opt))
		}
	}
}
