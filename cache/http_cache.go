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

// HTTP缓存模块，返回当前对应的缓存状态，是获取中、hit for pass等等。
// 以及对缓存数据压缩、智能匹配返回格式等处理。

package cache

import (
	"fmt"
	"sync"
	"time"
)

const (
	// StatusUnknown unknown status
	StatusUnknown = iota
	// StatusFetching fetching status
	StatusFetching
	// StatusHitForPass hit-for-pass status
	StatusHitForPass
	// StatusCachable cachable status
	StatusCachable
	// StatusPassed pass status
	StatusPassed
)

// HTTPCache cache status
type HTTPCache struct {
	mu     sync.Mutex
	status int
	chans  []chan bool
}

// NewHTTPCache new http cache
func NewHTTPCache() *HTTPCache {
	return &HTTPCache{}
}

// GetStatus get http status
func (hc *HTTPCache) GetStatus() (status int, done chan bool) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	fmt.Println(time.Now().Unix())
	// TODO 如果缓存已过期，设置为StatusUnknown

	if hc.status == StatusUnknown {
		hc.status = StatusFetching
		hc.chans = make([]chan bool, 0, 5)
	}
	// 如果是fetching，则相同的请求需要等待完成
	// 通过chan bool返回完成
	if hc.status == StatusFetching {
		done = make(chan bool)
		hc.chans = append(hc.chans, done)
	}
	status = hc.status
	return
}
