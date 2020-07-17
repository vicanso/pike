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

package server

import (
	"net/http"
	"strconv"

	"github.com/vicanso/elton"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/config"
	"github.com/vicanso/pike/util"
)

const (
	headerStatusKey = "X-Status"
)

func requestIsPass(req *http.Request) bool {
	method := req.Method
	return method != http.MethodGet && method != http.MethodHead
}

// 根据Cache-Control的信息，获取s-maxage 或者max-age的值
func getCacheAge(header http.Header) int {
	// 如果有设置cookie，则为不可缓存
	if len(header.Get(elton.HeaderSetCookie)) != 0 {
		return 0
	}
	// 如果没有设置cache-control，则不可缓存
	cc := header.Get(elton.HeaderCacheControl)
	if len(cc) == 0 {
		return 0
	}

	// 如果设置不可缓存，返回0
	match := noCacheReg.MatchString(cc)
	if match {
		return 0
	}
	// 优先从s-maxage中获取
	var maxAge = 0
	result := sMaxAgeReg.FindStringSubmatch(cc)
	if len(result) == 2 {
		maxAge, _ = strconv.Atoi(result[1])
	} else {
		// 从max-age中获取缓存时间
		result = maxAgeReg.FindStringSubmatch(cc)
		if len(result) == 2 {
			maxAge, _ = strconv.Atoi(result[1])
		}
	}

	// 如果有设置了 age 字段，则最大缓存时长减少
	age := header.Get(headerAge)
	if age != "" {
		v, _ := strconv.Atoi(age)
		maxAge -= v
	}

	return maxAge
}

// newCacheDispatchMiddleware create a cache dispatch middleware
func newCacheDispatchMiddleware(dispatcher *cache.Dispatcher, compress *config.Compress, generateEtag bool) elton.Handler {

	compressHandler := createCompressHandler(compress)
	return func(c *elton.Context) (err error) {
		status := cache.StatusUnknown
		cacheable := false
		passed := false
		var httpData *cache.HTTPData
		var httpCache *cache.HTTPCache
		// 如果设置了dispatcher，而且不是pass类的请求
		// 则表示有可能可缓存请求
		if dispatcher != nil && !requestIsPass(c.Request) {
			key := util.GetIdentity(c.Request)
			httpCache = dispatcher.GetHTTPCache(key)

			status, httpData = httpCache.Get()
			c.Set(statusKey, status)
			// 如果获取到缓存，则直接返回
			if status == cache.StatusCacheable {
				httpData.SetResponse(c)
				// 设置Age
				age := httpCache.Age()
				if age > 0 {
					c.SetHeader(headerAge, strconv.Itoa(age))
				}
				c.SetHeader(headerStatusKey, cache.StatusString(status))
				return
			}
			c.SetHeader(headerStatusKey, cache.StatusString(status))
			c.Set(httpCacheKey, httpCache)
		} else {
			passed = true
			c.Set(statusKey, cache.StatusPassed)
			c.SetHeader(headerStatusKey, cache.StatusString(cache.StatusPassed))
		}

		// 对于fetching类的请求，如果最终是不可缓存的，则设置hit for pass
		if status == cache.StatusFetching {
			defer func() {
				if !cacheable {
					httpCache.HitForPass(dispatcher.HitForPass)
				}
			}()
		}

		err = c.Next()
		if err != nil {
			return
		}

		// 执行proxy成功之后
		headers := c.Header()
		encoding := headers.Get(elton.HeaderContentEncoding)
		var body []byte
		if c.BodyBuffer != nil {
			body = c.BodyBuffer.Bytes()
		}

		// 生成etag
		if generateEtag {
			etag := headers.Get(elton.HeaderETag)
			if etag == "" {
				etag = util.GenerateETag(body)
				headers.Set(elton.HeaderETag, etag)
			}
		}

		// 如果是不支持的encoding，则直接返回
		if encoding != "" && encoding != elton.Gzip {
			return
		}

		// 如果是fetching状态的，在成功获取数据后，要根据返回数据设置缓存状态
		cacheAge := 0
		// 如果是pass的请求，都不可以缓存
		if !passed {
			if status == cache.StatusFetching {
				cacheAge = getCacheAge(c.Header())
			}
			// 缓存时长大于0
			if cacheAge != 0 {
				cacheable = true
			}
		}

		httpData = compressHandler(c, cacheable)
		// 如果是可缓存的，则缓存数据
		if cacheable {
			httpCache.Cachable(cacheAge, httpData)
		}

		return
	}
}
