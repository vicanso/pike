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
	"bytes"
	"errors"
	"net"
	"net/http/httptest"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vicanso/elton"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/config"
	"github.com/vicanso/pike/log"
	"github.com/vicanso/pike/upstream"
	"github.com/vicanso/pike/util"

	responder "github.com/vicanso/elton-responder"
)

func TestNewErrorListener(t *testing.T) {
	assert := assert.New(t)
	dispatcher := new(cache.Dispatcher)
	logger := log.Default()
	fn := newErrorListener(dispatcher, logger)
	httpCache := new(cache.HTTPCache)

	e := elton.New()
	e.OnError(fn)
	e.GET("/", func(c *elton.Context) error {
		c.Set(statusKey, cache.StatusFetching)
		c.Set(httpCacheKey, httpCache)
		return errors.New("custom error")
	})
	req := httptest.NewRequest("GET", "/", nil)
	resp := httptest.NewRecorder()
	e.ServeHTTP(resp, req)
	assert.Equal(cache.StatusHitForPass, httpCache.GetStatus())
}

func newTestServer(ln net.Listener) {
	e := elton.New()

	inc := func(p *int32) *bytes.Buffer {
		v := atomic.AddInt32(p, 1)
		return bytes.NewBufferString(strconv.Itoa(int(v)))
	}
	genBuffer := func(size int) *bytes.Buffer {
		buf := new(bytes.Buffer)
		for i := 0; i < size; i++ {
			buf.WriteString("a")
		}
		return buf
	}

	// 响应未压缩
	notCompressHandler := func(c *elton.Context) error {
		c.SetHeader("Content-Type", "text/html")
		c.BodyBuffer = genBuffer(4096)
		return nil
	}
	// 响应数据已压缩
	compressHandler := func(c *elton.Context) error {
		c.SetHeader("Content-Type", "text/html")
		c.SetHeader("Content-Encoding", "snz")
		buf, _ := util.Gzip(genBuffer(4096).Bytes(), 0)

		c.BodyBuffer = bytes.NewBuffer(buf)
		return nil
	}
	setCacheNext := func(c *elton.Context) error {
		c.CacheMaxAge("10s")
		return c.Next()
	}

	e.Use(responder.NewDefault())

	e.GET("/ping", func(c *elton.Context) error {
		c.BodyBuffer = bytes.NewBufferString("pong")
		return nil
	})

	var cacheCount int32
	e.GET("/cache-count", func(c *elton.Context) error {
		c.CacheMaxAge("1m")
		c.BodyBuffer = inc(&cacheCount)
		return nil
	})

	var nocacheCount int32
	e.GET("/nocache-count", func(c *elton.Context) error {
		c.BodyBuffer = inc(&nocacheCount)
		return nil
	})

	var postCount int32
	e.POST("/post-count", func(c *elton.Context) error {
		c.BodyBuffer = inc(&postCount)
		return nil
	})

	// 非文本类数据
	e.GET("/image-cache", func(c *elton.Context) error {
		c.CacheMaxAge("10s")
		c.SetHeader("Content-Type", "image/png")
		c.BodyBuffer = genBuffer(4096)
		return nil
	})

	e.POST("/post-not-compress", notCompressHandler)
	e.GET("/get-not-compress", notCompressHandler)

	e.POST("/post-compress", compressHandler)
	e.GET("/get-compress", compressHandler)

	e.GET("/get-cache-not-compress", setCacheNext, notCompressHandler)
	e.GET("/get-cache-compress", setCacheNext, compressHandler)

	e.GET("/get-with-etag", func(c *elton.Context) error {
		c.SetHeader(elton.HeaderETag, `"123"`)
		return notCompressHandler(c)
	})

	// nolint
	e.Serve(ln)
}

type stringList struct {
	sync.Mutex
	data []string
}

func (sl *stringList) Add(v string) {
	sl.Lock()
	if sl.data == nil {
		sl.data = make([]string, 0)
	}
	sl.data = append(sl.data, v)
	sl.Unlock()
}

func (sl *stringList) Sort() {
	sl.Lock()
	sort.Strings(sl.data)
	sl.Unlock()
}

func TestNewElton(t *testing.T) {
	cfg := config.NewTestConfig()
	assert := assert.New(t)
	ln, err := net.Listen("tcp", "127.0.0.1:")
	assert.Nil(err)
	defer ln.Close()
	go newTestServer(ln)

	upstreamsConfig := make(config.Upstreams, 0)
	upstreamsConfig = append(upstreamsConfig, &config.Upstream{
		Name:        "testUpstream",
		HealthCheck: "/ping",
		Servers: []config.UpstreamServer{
			config.UpstreamServer{
				Addr: "http://" + ln.Addr().String(),
			},
		},
	})
	upstreams := upstream.NewUpstreams(upstreamsConfig)

	locations := make(config.Locations, 0)
	locations = append(locations, &config.Location{
		Name:     "testLocation",
		Upstream: "testUpstream",
	})

	e := NewElton(&ServerOptions{
		cfg: cfg,
		server: &config.Server{
			Concurrency: 1024,
			ETag:        true,
		},
		locations: locations,
		upstreams: upstreams,
		dispatcher: cache.NewDispatcher(&config.Cache{
			Size:       100,
			Zone:       100,
			HitForPass: 60,
		}),
		compress: &config.Compress{
			Level:     6,
			MinLength: 1024,
			Filter:    "text|json|javascript",
		},
	})

	// 可缓存请求，count不变
	t.Run("get cache count", func(t *testing.T) {
		// 可缓存的请求，返回的数据一致
		wg := sync.WaitGroup{}
		sl := new(stringList)
		max := 10
		for i := 0; i < max; i++ {
			wg.Add(1)
			go func(index int) {
				req := httptest.NewRequest("GET", "/cache-count", nil)
				resp := httptest.NewRecorder()
				e.ServeHTTP(resp, req)
				assert.NotEmpty(resp.Header().Get(elton.HeaderETag))
				assert.Equal("1", resp.Body.String())
				assert.Equal(200, resp.Code)
				sl.Add(resp.Header().Get(headerStatusKey))
				wg.Done()
			}(i)
		}
		wg.Wait()
		sl.Sort()
		// 只有一个状态为fetching，其它为cacheable
		assert.Equal("fetching", sl.data[max-1])
		for _, value := range sl.data[:max-1] {
			assert.Equal("cacheable", value)
		}
	})

	// 不可缓存请求(get nocache)，count +1
	t.Run("get nocache count", func(t *testing.T) {
		// 不可缓存请求，返回的数据不一致
		statusList := new(stringList)
		dataList := new(stringList)
		wg := sync.WaitGroup{}
		max := 10
		for i := 0; i < max; i++ {
			wg.Add(1)
			go func() {
				req := httptest.NewRequest("GET", "/nocache-count", nil)
				resp := httptest.NewRecorder()
				e.ServeHTTP(resp, req)
				assert.Equal(200, resp.Code)
				statusList.Add(resp.Header().Get(headerStatusKey))
				dataList.Add(resp.Body.String())
				wg.Done()
			}()
		}
		wg.Wait()
		statusList.Sort()
		// 第一个状态为fetching，其它为hitForPass
		assert.Equal("fetching", statusList.data[0])
		for _, value := range statusList.data[1:] {
			assert.Equal("hitForPass", value)
		}
		dataList.Sort()
		assert.Equal("1,10,2,3,4,5,6,7,8,9", strings.Join(dataList.data, ","))
	})

	// 不可缓存请求(POST)，每次count +1
	t.Run("post count", func(t *testing.T) {
		// pass的请求，返回的数据不一致
		dataList := new(stringList)
		wg := sync.WaitGroup{}
		max := 10
		for i := 0; i < max; i++ {
			wg.Add(1)
			go func() {
				req := httptest.NewRequest("POST", "/post-count", nil)
				resp := httptest.NewRecorder()
				e.ServeHTTP(resp, req)
				assert.Equal(200, resp.Code)
				assert.Equal("passed", resp.Header().Get(headerStatusKey))
				dataList.Add(resp.Body.String())
				wg.Done()
			}()
		}
		wg.Wait()
		dataList.Sort()
		assert.Equal("1,10,2,3,4,5,6,7,8,9", strings.Join(dataList.data, ","))
	})

	// 获取图片数据（可缓存）
	t.Run("get image cache", func(t *testing.T) {
		wg := sync.WaitGroup{}
		sl := new(stringList)
		max := 10
		for i := 0; i < max; i++ {
			wg.Add(1)
			go func(index int) {
				req := httptest.NewRequest("GET", "/image-cache", nil)
				resp := httptest.NewRecorder()
				e.ServeHTTP(resp, req)
				assert.Equal(200, resp.Code)
				assert.Equal(4096, resp.Body.Len())
				sl.Add(resp.Header().Get(headerStatusKey))
				wg.Done()
			}(i)
		}
		wg.Wait()
		sl.Sort()
		// 只有一个状态为fetching，其它为cacheable
		assert.Equal("fetching", sl.data[max-1])
		for _, value := range sl.data[:max-1] {
			assert.Equal("cacheable", value)
		}
	})

	// 未压缩数据自动压缩(POST)
	t.Run("post not compress", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/post-not-compress", nil)
		req.Header.Set(elton.HeaderAcceptEncoding, "gzip")
		resp := httptest.NewRecorder()
		e.ServeHTTP(resp, req)
		assert.Equal(elton.Gzip, resp.Header().Get(elton.HeaderContentEncoding))
		assert.Equal(200, resp.Code)
	})

	// 未压缩数据自动压缩(GET)
	t.Run("get not compress", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/get-not-compress", nil)
		req.Header.Set(elton.HeaderAcceptEncoding, "gzip")
		resp := httptest.NewRecorder()
		e.ServeHTTP(resp, req)
		assert.Equal(elton.Gzip, resp.Header().Get(elton.HeaderContentEncoding))
		assert.Equal(200, resp.Code)
	})

	// 已压缩数据不作压缩处理(POST)
	t.Run("post compress", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/post-compress", nil)
		req.Header.Set(elton.HeaderAcceptEncoding, "gzip")
		resp := httptest.NewRecorder()
		e.ServeHTTP(resp, req)
		assert.Equal("snz", resp.Header().Get(elton.HeaderContentEncoding))
		assert.Equal(200, resp.Code)
	})

	// 已压缩数据不作压缩处理(GET)
	t.Run("get compress", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/get-compress", nil)
		req.Header.Set(elton.HeaderAcceptEncoding, "gzip")
		resp := httptest.NewRecorder()
		e.ServeHTTP(resp, req)
		assert.Equal(200, resp.Code)
		assert.Equal("snz", resp.Header().Get(elton.HeaderContentEncoding))
	})

	// 已经生成etag的无需重新生成
	t.Run("get with etag", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/get-with-etag", nil)
		resp := httptest.NewRecorder()
		e.ServeHTTP(resp, req)
		assert.Equal(200, resp.Code)
		assert.Equal(`"123"`, resp.Header().Get(elton.HeaderETag))
	})
}
