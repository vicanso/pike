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

	e.Use(responder.NewDefault())

	e.GET("/ping", func(c *elton.Context) error {
		c.BodyBuffer = bytes.NewBufferString("pong")
		return nil
	})

	var cacheCount int32
	e.GET("/cache-count", func(c *elton.Context) error {
		c.CacheMaxAge("1m")
		v := atomic.AddInt32(&cacheCount, 1)
		c.Body = strconv.Itoa(int(v))
		return nil
	})

	var nocacheCount int32
	e.GET("/nocache-count", func(c *elton.Context) error {
		v := atomic.AddInt32(&nocacheCount, 1)
		c.Body = strconv.Itoa(int(v))
		return nil
	})

	var postCount int32
	e.POST("/post-count", func(c *elton.Context) error {
		v := atomic.AddInt32(&postCount, 1)
		c.Body = strconv.Itoa(int(v))
		return nil
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

	eltonConfig := &EltonConfig{
		maxConcurrency: 1024,
		eTag:           true,
		locations:      locations,
		upstreams:      upstreams,
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
	}

	e := NewElton(eltonConfig)

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
				assert.Equal("1", resp.Body.String())
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
				assert.Equal("passed", resp.Header().Get(headerStatusKey))
				dataList.Add(resp.Body.String())
				wg.Done()
			}()
		}
		wg.Wait()
		dataList.Sort()
		assert.Equal("1,10,2,3,4,5,6,7,8,9", strings.Join(dataList.data, ","))
	})
}
