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

// 针对同一个请求，在状态未知时，控制只允许一个请求转发至后续流程
// 在获取状态之后，支持hit for pass 与 hit 两种处理，其中hit for pass表示该请求不可缓存，
// 直接转发至后端程序，而hit则返回当前缓存的响应数据

package cache

import (
	"sync"
	"time"

	"github.com/vicanso/pike/compress"
)

type Status int

const (
	// StatusUnknown unknown status
	StatusUnknown Status = iota
	// StatusFetching fetching status
	StatusFetching
	// StatusHitForPass hit-for-pass status
	StatusHitForPass
	// StatusHit hit cache status
	StatusHit
	// StatusPassed pass status
	StatusPassed
)

// defaultHitForPassSeconds default hit for pass: 300 seconds
const defaultHitForPassSeconds = 300

type (
	// httpCache http cache (only for same request method+host+uri)
	httpCache struct {
		mu        *sync.RWMutex
		status    Status
		chanList  []chan struct{}
		response  *HTTPResponse
		createdAt int
		expiredAt int
	}
)

func nowUnix() int {
	return int(time.Now().Unix())
}

func (i Status) String() string {
	switch i {
	case StatusFetching:
		return "fetching"
	case StatusHitForPass:
		return "hitForPass"
	case StatusHit:
		return "hit"
	case StatusPassed:
		return "passed"
	default:
		return "unknown"
	}
}

// NewHTTPCache new a http cache
func NewHTTPCache() *httpCache {
	return &httpCache{
		mu: &sync.RWMutex{},
	}
}

// Get get http cache
func (hc *httpCache) Get() (status Status, response *HTTPResponse) {
	hc.mu.Lock()
	status, done, response := hc.get()
	hc.mu.Unlock()
	// 如果done不为空，表示需要等待确认当前请求状态
	if done != nil {
		// TODO 后续再考虑是否需要添加timeout（proxy部分有超时，因此暂时可不添加)
		<-done
		// 完成后重新获取当前状态与响应
		// 此时状态只可能是hit for pass 或者 hit
		// 而此两种状态的数据缓存均不会立即失效，因此可以从hc中获取
		status = hc.status
		response = hc.response
	}
	return
}

func (hc *httpCache) get() (status Status, done chan struct{}, data *HTTPResponse) {
	now := nowUnix()
	// 如果缓存已过期，设置为StatusUnknown
	if hc.expiredAt != 0 && hc.expiredAt < now {
		hc.status = StatusUnknown
		// 将有效期重置（若不重置则导致hs.status每次都被重置为Unknown)
		hc.expiredAt = 0
	}

	// 仅有同类请求为fetching，才会需要等待
	// 如果是fetching，则相同的请求需要等待完成
	// 通过chan返回完成
	if hc.status == StatusFetching {
		done = make(chan struct{})
		hc.chanList = append(hc.chanList, done)
	}

	if hc.status == StatusUnknown {
		hc.status = StatusFetching
		hc.chanList = make([]chan struct{}, 0, 5)
	}

	status = hc.status
	// 为什么需要返回status与data
	// 因为有可能在函数调用完成后，刚好缓存过期了，如果此时不返回status与data
	// 当其它goroutine获取锁之后，有可能刚好重置数据
	if status == StatusHit {
		data = hc.response
	}
	return
}

// HitForPass set the http cache hit for pass
func (hc *httpCache) HitForPass(ttl int) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	if ttl <= 0 {
		ttl = defaultHitForPassSeconds
	}
	hc.expiredAt = nowUnix() + ttl
	hc.status = StatusHitForPass
	list := hc.chanList
	hc.chanList = nil
	for _, ch := range list {
		ch <- struct{}{}
	}
}

// Cacheable set http cache cacheable and compress it
func (hc *httpCache) Cacheable(resp *HTTPResponse, ttl int) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	// 如果是可缓存数据，则选择默认的best compression
	resp.CompressSrv = compress.BestCompression
	_ = resp.Compress()
	hc.createdAt = nowUnix()
	hc.expiredAt = hc.createdAt + ttl
	hc.status = StatusHit
	hc.response = resp
	list := hc.chanList
	hc.chanList = nil
	for _, ch := range list {
		ch <- struct{}{}
	}
}

// Age get http cache's age
func (hc *httpCache) Age() int {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return nowUnix() - hc.createdAt
}

// GetStatus get http cache status
func (hc *httpCache) GetStatus() Status {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.status
}

// IsExpired the cache is expired
func (hc *httpCache) IsExpired() bool {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	if hc.expiredAt == 0 {
		return false
	}
	return hc.expiredAt < nowUnix()
}
