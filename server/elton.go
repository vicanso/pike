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
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/vicanso/elton"
	fresh "github.com/vicanso/elton-fresh"
	proxy "github.com/vicanso/elton-proxy"
	recover "github.com/vicanso/elton-recover"
	"github.com/vicanso/hes"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/config"
	"github.com/vicanso/pike/log"
	"github.com/vicanso/pike/upstream"
	"github.com/vicanso/pike/util"
	"go.uber.org/zap"
)

const (
	statusKey    = "status"
	httpCacheKey = "httpCache"

	// 默认的 admin 目录
	defaultAdminPath = "/pike"

	headerAge = "Age"
)

var (
	// 需要清除的header
	clearHeaders = []string{
		"Date",
		"Connection",
		elton.HeaderContentLength,
	}
	noCacheReg = regexp.MustCompile(`no-cache|no-store|private`)
	sMaxAgeReg = regexp.MustCompile(`s-maxage=(\d+)`)
	maxAgeReg  = regexp.MustCompile(`max-age=(\d+)`)
)

var (
	errTooManyRequests = &hes.Error{
		StatusCode: http.StatusTooManyRequests,
		Message:    "Too Many Requests",
	}
	errServiceUnavailable = &hes.Error{
		StatusCode: http.StatusServiceUnavailable,
		Message:    "Service Unavailable",
	}
)

type (
	// CompressHandler compress handler
	CompressHandler func(*elton.Context, []byte, bool) (httpData *cache.HTTPData)
)

// EltonConfig elton config
type EltonConfig struct {
	maxConcurrency uint32
	eTag           bool
	locations      config.Locations
	upstreams      *upstream.Upstreams
	dispatcher     *cache.Dispatcher
	compress       *config.Compress
}

// createProxyMiddlewares create proxy middleware functions
func createProxyMiddlewares(locations config.Locations, upstreams *upstream.Upstreams) map[string]elton.Handler {
	proxyMids := make(map[string]elton.Handler)
	for _, item := range locations {
		up := upstreams.Get(item.Upstream)
		if up == nil {
			continue
		}
		proxyMids[item.Name] = proxy.New(proxy.Config{
			TargetPicker: func(c *elton.Context) (*url.URL, proxy.Done, error) {
				httpUpstream, done := up.Next()
				if httpUpstream == nil {
					return nil, nil, errServiceUnavailable
				}
				// 如果不需要设置done
				if done == nil {
					return httpUpstream.URL, nil, nil
				}
				// 返回了done（如最少连接数的策略）
				return httpUpstream.URL, func(_ *elton.Context) {
					done()
				}, nil
			},
		})

	}
	return proxyMids
}

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

func createCompressHandler(compressConfig *config.Compress) CompressHandler {
	// 前置已针对未支持的encoding过滤，因此data只可能为未压缩或者gzip数据
	var filter *regexp.Regexp
	if compressConfig != nil && compressConfig.Filter != "" {
		filter, _ = regexp.Compile(compressConfig.Filter)
	}
	return func(c *elton.Context, data []byte, ignoreHeader bool) (httpData *cache.HTTPData) {
		headers := c.Headers
		httpData = &cache.HTTPData{
			StatusCode: c.StatusCode,
			RawBody:    data,
		}
		// 如果忽略header（对于post等不可缓存的请求）
		if !ignoreHeader {
			httpData.Headers = cache.NewHTTPHeaders(headers, elton.HeaderContentEncoding, elton.HeaderContentLength)
		}
		// 如果未指定filter的数据，则认为都不需要压缩
		if filter == nil || len(data) == 0 {
			return
		}
		if len(data) < compressConfig.MinLength {
			return
		}
		contentType := headers.Get(elton.HeaderContentType)
		if !filter.MatchString(contentType) {
			return
		}
		httpData.GzipBody, _ = util.Gzip(data, compressConfig.Level)
		// 如果gzip压缩成功，则删除rawBody（因为基本所有客户端都支持gzip，若不支持则从gzip解压获取），减少内存占用
		if httpData.GzipBody != nil {
			httpData.RawBody = nil
		}
		httpData.BrBody, _ = util.Brotli(data, compressConfig.Level)
		return
	}
}

// NewElton new an elton instance
func NewElton(eltonConfig *EltonConfig) *elton.Elton {
	logger := log.Default()
	locations := eltonConfig.locations
	upstreams := eltonConfig.upstreams
	dispatcher := eltonConfig.dispatcher
	compress := eltonConfig.compress
	e := elton.New()

	adminElton := NewAdmin(defaultAdminPath, eltonConfig)

	proxyMids := createProxyMiddlewares(locations, upstreams)
	compressHandler := createCompressHandler(compress)

	// 未处理错误
	e.OnError(func(c *elton.Context, err error) {
		logger.Error("uncaught exception",
			zap.String("url", c.Request.RequestURI),
			zap.Error(err),
		)
		// 如果没有设置dispatcher，则无需要处理以下流程
		if dispatcher == nil {
			return
		}
		status, _ := c.Get(statusKey).(int)
		if status == cache.StatusFetching {
			httpCache, _ := c.Get(httpCacheKey).(*cache.HTTPCache)
			if httpCache != nil {
				httpCache.HitForPass(dispatcher.HitForPass)
			}
		}
	})
	e.Use(recover.New())

	var concurrency uint32
	maxConcurrency := eltonConfig.maxConcurrency
	if maxConcurrency > 0 {
		e.Use(func(c *elton.Context) error {
			v := atomic.AddUint32(&concurrency, 1)
			defer atomic.AddUint32(&concurrency, ^uint32(0))
			if v > maxConcurrency {
				return errTooManyRequests
			}
			return c.Next()
		})
	}

	// 如果是admin路径，则转发至admin elton
	e.Use(func(c *elton.Context) error {
		if !strings.HasPrefix(c.Request.RequestURI, defaultAdminPath) {
			return c.Next()
		}
		c.Pass(adminElton)
		return nil
	})

	e.Use(fresh.NewDefault())

	// get http cache
	e.Use(func(c *elton.Context) (err error) {
		// 如果dispatcher未设置，所有请求都直接pass
		if dispatcher == nil || requestIsPass(c.Request) {
			return c.Next()
		}
		key := util.GetIdentity(c.Request)
		httpCache := dispatcher.GetHTTPCache(key)

		status, httpData := httpCache.Get()
		if status == cache.StatusCachable {
			httpData.SetResponse(c)
			// 设置Age
			age := httpCache.Age()
			if age > 0 {
				c.SetHeader(headerAge, strconv.Itoa(age))
			}
			return
		}

		c.Set(statusKey, status)
		c.Set(httpCacheKey, httpCache)

		err = c.Next()

		headers := c.Headers
		encoding := headers.Get(elton.HeaderContentEncoding)
		var body []byte
		if c.BodyBuffer != nil {
			body = c.BodyBuffer.Bytes()
		}

		done := func() {
			if status == cache.StatusFetching {
				httpCache.HitForPass(dispatcher.HitForPass)
			}
		}
		if err != nil {
			done()
			return
		}

		// 生成etag
		if eltonConfig.eTag {
			etag := headers.Get(elton.HeaderETag)
			if etag == "" {
				etag = util.GenerateETag(body)
				headers.Set(elton.HeaderETag, etag)
			}
		}

		// 如果是不支持的encoding，则直接返回
		if encoding != "" && encoding != elton.Gzip {
			done()
			return
		}

		// 如果是fetching状态的，在成功获取数据后，要根据返回数据设置缓存状态
		if status == cache.StatusFetching {
			age := getCacheAge(c.Headers)
			if age == 0 {
				done()
				return
			}

			// 如果返回的数据已压缩，解压
			if strings.EqualFold(encoding, elton.Gzip) {
				body, err = util.Gunzip(body)
				if err != nil {
					done()
					return
				}
				headers.Del(elton.HeaderContentEncoding)
			}
			httpData = compressHandler(c, body, false)
			httpCache.Cachable(age, httpData)
			httpData.SetResponse(c)
		} else {
			// 如果返回的数据已压缩，解压
			if strings.EqualFold(encoding, elton.Gzip) {
				body, err = util.Gunzip(body)
				if err != nil {
					return
				}
			}
			httpData := compressHandler(c, body, true)
			httpData.SetResponse(c)
		}
		return
	})

	// http request proxy
	e.Use(func(c *elton.Context) (err error) {
		originalNext := c.Next
		// 由于proxy中间件会调用next，因此直接覆盖，
		// 避免导致先执行了后续的中间件（保证在函数调用next前是已完成此中间件处理）
		c.Next = func() error {
			return nil
		}

		host := c.Request.Host
		url := c.Request.RequestURI
		l := locations.GetMatch(host, url)
		if l == nil {
			err = errServiceUnavailable
			return
		}
		// 设置请求头
		if l.RequestHeader != nil {
			util.MergeHeader(c.Request.Header, l.ReqHeader)
			host := l.ReqHeader.Get("Host")
			// 如果有配置Host请求头，则设置request host
			if host != "" {
				c.Request.Host = host
			}
		}
		// 设置响应头
		if l.ResHeader != nil {
			util.MergeHeader(c.Header(), l.ResHeader)
		}

		fn := proxyMids[l.Name]
		if fn == nil {
			err = errServiceUnavailable
			return
		}

		reqHeader := c.Request.Header
		var ifModifiedSince, ifNoneMatch, acceptEncoding string
		status, _ := c.Get(statusKey).(int)
		// 针对fetching的请求，由于其最终状态未知，因此需要删除有可能导致304的请求，避免无法生成缓存
		if status == cache.StatusFetching {
			acceptEncoding = reqHeader.Get(elton.HeaderAcceptEncoding)
			ifModifiedSince = reqHeader.Get(elton.HeaderIfModifiedSince)
			ifNoneMatch = reqHeader.Get(elton.HeaderIfNoneMatch)
			if ifModifiedSince != "" {
				reqHeader.Del(elton.HeaderIfModifiedSince)
			}
			if ifNoneMatch != "" {
				reqHeader.Del(elton.HeaderIfNoneMatch)
			}

			if strings.Contains(acceptEncoding, elton.Gzip) {
				reqHeader.Set(elton.HeaderAcceptEncoding, elton.Gzip)
			} else {
				reqHeader.Del(elton.HeaderAcceptEncoding)
			}
		}

		err = fn(c)

		// 将原有的请求头恢复（就算出错也需要恢复）
		if acceptEncoding != "" {
			reqHeader.Set(elton.HeaderAcceptEncoding, acceptEncoding)
		}
		if ifModifiedSince != "" {
			reqHeader.Set(elton.HeaderIfModifiedSince, ifModifiedSince)
		}
		if ifNoneMatch != "" {
			reqHeader.Set(elton.HeaderIfNoneMatch, ifNoneMatch)
		}
		if err != nil {
			return
		}
		for _, key := range clearHeaders {
			// 清除header
			c.SetHeader(key, "")
		}

		return originalNext()
	})

	e.ALL("/*url", func(c *elton.Context) error {
		return nil
	})
	return e
}
