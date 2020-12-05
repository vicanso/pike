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
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vicanso/elton"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/config"
)

func TestRequestIsPass(t *testing.T) {
	assert := assert.New(t)
	assert.True(requestIsPass(httptest.NewRequest("POST", "/", nil)))
	assert.True(requestIsPass(httptest.NewRequest("PATCH", "/", nil)))
	assert.True(requestIsPass(httptest.NewRequest("PUT", "/", nil)))
	assert.False(requestIsPass(httptest.NewRequest("GET", "/", nil)))
	assert.False(requestIsPass(httptest.NewRequest("HEAD", "/", nil)))
}

func TestGetKey(t *testing.T) {
	assert := assert.New(t)

	req := httptest.NewRequest("GET", "http://test.com/users/me?type=1", nil)
	assert.Equal("GET test.com http://test.com/users/me?type=1", string(getKey(req)))
}

func TestCacheMiddleware(t *testing.T) {
	assert := assert.New(t)

	cacheableContext := elton.NewContext(
		httptest.NewRecorder(),
		httptest.NewRequest("GET", "/", nil),
	)
	// 设置可缓存有效期为10
	setHTTPCacheMaxAge(cacheableContext, 10)
	setHTTPResp(cacheableContext, &cache.HTTPResponse{})

	tests := []struct {
		c      *elton.Context
		status cache.Status
	}{
		// 直接pass的请求
		{
			c: elton.NewContext(
				httptest.NewRecorder(),
				httptest.NewRequest("POST", "/users/login", nil),
			),
			status: cache.StatusPassed,
		},
		// 首次fetching，返回不可缓存
		{
			c: elton.NewContext(
				httptest.NewRecorder(),
				httptest.NewRequest("GET", "/users/me", nil),
			),
			status: cache.StatusFetching,
		},
		// 第二次hit for pass
		{
			c: elton.NewContext(
				httptest.NewRecorder(),
				httptest.NewRequest("GET", "/users/me", nil),
			),
			status: cache.StatusHitForPass,
		},
		// 首次fetching，返回可缓存
		{
			c:      cacheableContext,
			status: cache.StatusFetching,
		},
		// 第二次则从缓存获取
		{
			c: elton.NewContext(
				httptest.NewRecorder(),
				httptest.NewRequest("GET", "/", nil),
			),
			status: cache.StatusCacheable,
		},
	}

	cacheName := "test"
	cache.ResetDispatchers([]config.CacheConfig{
		{
			Name: cacheName,
			Size: 100,
		},
	})
	s := NewServer(ServerOption{
		Cache: cacheName,
	})

	fn := NewCache(s)
	for _, tt := range tests {
		tt.c.Next = func() error {
			return nil
		}
		err := fn(tt.c)
		assert.Nil(err)
		assert.Equal(tt.status, getCacheStatus(tt.c))
	}

}
