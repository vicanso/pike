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
	"bytes"
	"net"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vicanso/elton"
	"github.com/vicanso/elton/middleware"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/config"
	"github.com/vicanso/pike/location"
	"github.com/vicanso/pike/upstream"
)

func TestGetCacheMaxAge(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		key       string
		value     string
		age       int
		existsAge int
	}{
		// 设置了 set cookie
		{
			key:   elton.HeaderSetCookie,
			value: "set cookie",
			age:   0,
		},
		// 未设置cache control
		{
			age: 0,
		},
		// 设置了cache control 为 no cache
		{
			key:   elton.HeaderCacheControl,
			value: "no-cache",
			age:   0,
		},
		// 设置了cache control 为 no store
		{
			key:   elton.HeaderCacheControl,
			value: "no-store",
			age:   0,
		},
		// 设置了cache control 为 private
		{
			key:   elton.HeaderCacheControl,
			value: "private, max-age=10",
			age:   0,
		},
		// 设置了max-age
		{
			key:   elton.HeaderCacheControl,
			value: "max-age=10",
			age:   10,
		},
		// 设置了s-maxage
		{
			key:   elton.HeaderCacheControl,
			value: "max-age=10, s-maxage=1 ",
			age:   1,
		},
		// 设置了age
		{
			key:       elton.HeaderCacheControl,
			value:     "max-age=10",
			age:       8,
			existsAge: 2,
		},
	}

	for _, tt := range tests {
		h := http.Header{}
		h.Add(tt.key, tt.value)
		if tt.existsAge != 0 {
			h.Add("Age", strconv.Itoa(tt.existsAge))
		}
		age := getCacheMaxAge(h)
		assert.Equal(tt.age, age)
	}
}

func TestProxyMiddleware(t *testing.T) {
	assert := assert.New(t)
	ln, err := net.Listen("tcp", "127.0.0.1:")
	assert.Nil(err)
	defer ln.Close()

	cacheResp := []byte("cache response")

	go func() {
		e := elton.New()
		e.Use(middleware.NewDefaultBodyParser())
		e.GET("/ping", func(c *elton.Context) error {
			c.BodyBuffer = bytes.NewBufferString("pong")
			return nil
		})

		e.GET("/cache", func(c *elton.Context) error {
			c.CacheMaxAge(time.Minute)
			c.BodyBuffer = bytes.NewBuffer(cacheResp)
			return nil
		})

		e.GET("/accept-encoding", func(c *elton.Context) error {
			c.BodyBuffer = bytes.NewBufferString(c.GetRequestHeader(elton.HeaderAcceptEncoding))
			return nil
		})
		e.GET("/remove-304-header", func(c *elton.Context) error {
			values := make([]string, 0)
			for _, key := range []string{
				elton.HeaderIfModifiedSince,
				elton.HeaderIfNoneMatch,
				"X-Custom",
			} {
				values = append(values, c.GetRequestHeader(key))
			}
			c.BodyBuffer = bytes.NewBufferString(strings.Join(values, ","))
			return nil
		})

		// e.POST("/")
		_ = e.Serve(ln)
	}()
	time.Sleep(50 * time.Millisecond)
	reqHeader := http.Header{
		"X-Request-ID": []string{
			"1",
		},
	}
	respHeader := http.Header{
		"X-Response-ID": []string{
			"2",
		},
	}
	location.Reset([]config.LocationConfig{
		{
			Name:     "test",
			Upstream: "test",
			ReqHeaders: []string{
				"X-Request-ID:1",
			},
			RespHeaders: []string{
				"X-Response-ID:2",
			},
			Rewrites: []string{
				"/api/*:/$1",
			},
			ProxyTimeout: "1s",
		},
	})
	upstream.Reset([]config.UpstreamConfig{
		{
			Name: "test",
			Servers: []config.UpstreamServerConfig{
				{
					Addr: "http://" + ln.Addr().String(),
				},
			},
			AcceptEncoding: "snappy",
		},
	})

	tests := []struct {
		create                 func() *elton.Context
		body                   string
		age                    int
		originalAcceptEncoding string
	}{
		// 正常fetching，可缓存请求
		{
			create: func() *elton.Context {
				req := httptest.NewRequest("GET", "/cache", nil)
				c := elton.NewContext(httptest.NewRecorder(), req)
				setCacheStatus(c, cache.StatusFetching)
				return c
			},
			body: string(cacheResp),
			age:  60,
		},
		// url rewrite
		{
			create: func() *elton.Context {
				req := httptest.NewRequest("GET", "/api/cache", nil)
				return elton.NewContext(httptest.NewRecorder(), req)
			},
			body: string(cacheResp),
		},
		// 修改accept encoding
		{
			create: func() *elton.Context {
				req := httptest.NewRequest("GET", "/accept-encoding", nil)
				req.Header.Set(elton.HeaderAcceptEncoding, "lz4")
				return elton.NewContext(httptest.NewRecorder(), req)
			},
			body:                   "snappy",
			originalAcceptEncoding: "lz4",
		},
		// remove 304 header
		{
			create: func() *elton.Context {
				req := httptest.NewRequest("GET", "/remove-304-header", nil)
				c := elton.NewContext(httptest.NewRecorder(), req)
				// 设置为fetching的才会影响304的请求头
				setCacheStatus(c, cache.StatusFetching)
				c.SetRequestHeader(elton.HeaderIfModifiedSince, "if modified since")
				c.SetRequestHeader(elton.HeaderIfNoneMatch, "if none match")
				c.SetRequestHeader("X-Custom", "1")
				return c
			},
			body: ",,1",
		},
	}

	serverOption := ServerOption{
		Locations: []string{
			"test",
		},
		Compress:                  "test-compress",
		CompressMinLength:         100,
		CompressContentTypeFilter: regexp.MustCompile(`text|json`),
	}
	fn := NewProxy(NewServer(serverOption))

	for _, tt := range tests {
		c := tt.create()
		c.Next = func() error {
			return nil
		}
		err := fn(c)
		assert.Nil(err)
		httpResp := getHTTPResp(c)
		for key, value := range reqHeader {
			assert.Equal(value[0], c.GetRequestHeader(key))
		}
		for key, value := range respHeader {
			assert.Equal(value[0], httpResp.Header.Get(key))
			// 确认context中的header并没有设置
			assert.Empty(c.GetHeader(key))
		}
		assert.Equal(tt.originalAcceptEncoding, c.GetRequestHeader(elton.HeaderAcceptEncoding))
		assert.Equal(tt.body, string(httpResp.RawBody))
		assert.Equal(tt.age, getHTTPCacheMaxAge(c))
		assert.Equal(serverOption.CompressContentTypeFilter, httpResp.CompressContentTypeFilter)
		assert.Equal(serverOption.CompressMinLength, httpResp.CompressMinLength)
		assert.Equal(serverOption.Compress, httpResp.CompressSrv)
	}
}
