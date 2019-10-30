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
	"github.com/vicanso/pike/util"

	"github.com/minio/highwayhash"
)

const (
	defaultSize     = 10
	defaultZoneSize = 1024
)

var (
	hashKey = []byte("2fKEes0u2jpZhJpfjVeAsmUE2RW7Ab2I")
)

type (
	// Dispatcher http cache dispatcher
	Dispatcher struct {
		size int
		list []*HTTPCacheLRU
	}
)

// NewDispatcher create a new dispatcher
func NewDispatcher(size, zoneSize int) *Dispatcher {
	if size <= 0 {
		size = defaultSize
	}
	if zoneSize <= 0 {
		zoneSize = defaultZoneSize
	}
	list := make([]*HTTPCacheLRU, size)
	for i := 0; i < size; i++ {
		list[i] = NewHTTPCacheLRU(zoneSize)
	}

	return &Dispatcher{
		size: size,
		list: list,
	}
}

// GetHTTPCache get http cache through key
func (d *Dispatcher) GetHTTPCache(key []byte) *HTTPCache {
	// 计算hash值
	index := int(highwayhash.Sum64(key, hashKey)) % d.size
	// 从预定义的列表中取对应的缓存
	lru := d.list[index]
	return lru.FindOrCreate(util.ByteSliceToString(key))
}
