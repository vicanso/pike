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

package server

import (
	"net/http"

	"github.com/vicanso/elton"
	"github.com/vicanso/pike/cache"
)

const (
	// spaceByte 空格
	spaceByte = byte(' ')
)

// requestIsPass check request is passed
func requestIsPass(req *http.Request) bool {
	// 非GET HEAD 的请求均直接pass
	return req.Method != http.MethodGet &&
		req.Method != http.MethodHead
}

// getKey get key of request
func getKey(req *http.Request) []byte {
	methodLen := len(req.Method)
	hostLen := len(req.Host)
	uriLen := len(req.RequestURI)
	buffer := make([]byte, methodLen+hostLen+uriLen+2)
	len := 0

	copy(buffer[len:], req.Method)
	len += methodLen

	buffer[len] = spaceByte
	len++

	copy(buffer[len:], req.Host)
	len += hostLen

	buffer[len] = spaceByte
	len++

	copy(buffer[len:], req.RequestURI)
	return buffer
}

// NewCache new a cache middleware
func NewCache(s *server) elton.Handler {
	return func(c *elton.Context) (err error) {
		// 不可缓存请求，直接pass至upstream
		if requestIsPass(c.Request) {
			setCacheStatus(c, cache.StatusPassed)
			return c.Next()
		}
		disp := cache.GetDispatcher(s.GetCache())
		if disp == nil {
			err = ErrCacheDispatcherNotFound
			return
		}

		key := getKey(c.Request)
		httpCache := disp.GetHTTPCache(key)
		cacheStatus, httpResp := httpCache.Get()

		cacheable := false
		// 对于fetching类的请求，如果最终是不可缓存的，则设置hit for pass
		// 保证只要不是panic，fetching的请求非可缓存的都为hit for pass
		if cacheStatus == cache.StatusFetching {
			defer func() {
				if !cacheable {
					httpCache.HitForPass(disp.GetHitForPass())
				}
			}()
		}

		setCacheStatus(c, cacheStatus)
		// 缓存中读取的可缓存数据，不需要next
		if cacheStatus == cache.StatusHit {
			// 设置缓存数据
			setHTTPResp(c, httpResp)
			// 设置缓存数据的age
			setHTTPRespAge(c, httpCache.Age())
			return nil
		}

		err = c.Next()
		if err != nil {
			return err
		}

		// TODO 如果是hit for pass，但是缓存有效期不为0
		if cacheStatus == cache.StatusFetching {
			// 获取缓存有效期
			if maxAge := getHTTPCacheMaxAge(c); maxAge > 0 {
				// 只有有响应数据可缓存时才设置为cacheable
				if httpResp = getHTTPResp(c); httpResp != nil {
					cacheable = true
					httpCache.Cacheable(httpResp, maxAge)
				}
			}
		}
		return nil
	}
}
