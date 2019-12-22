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
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/vicanso/elton"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/config"

	"github.com/stretchr/testify/assert"
)

func TestRequestIsPass(t *testing.T) {
	assert := assert.New(t)
	req := &http.Request{
		Method: "GET",
	}
	assert.False(requestIsPass(req))

	req.Method = "HEAD"
	assert.False(requestIsPass(req))

	req.Method = "POST"
	assert.True(requestIsPass(req))
}

func TestGetCacheAge(t *testing.T) {
	assert := assert.New(t)
	h := make(http.Header)

	h.Set(elton.HeaderSetCookie, "abc")
	assert.Equal(0, getCacheAge(h))
	h.Del(elton.HeaderSetCookie)

	assert.Equal(0, getCacheAge(h))

	h.Set(elton.HeaderCacheControl, "no-cache")
	assert.Equal(0, getCacheAge(h))

	h.Set(elton.HeaderCacheControl, "public, max-age=10, s-maxage=2")
	assert.Equal(2, getCacheAge(h))

	h.Set(elton.HeaderCacheControl, "public, max-age=10")
	assert.Equal(10, getCacheAge(h))

	h.Set(headerAge, "2")
	assert.Equal(8, getCacheAge(h))
}

func TestCacheDispatchMiddleware(t *testing.T) {

	dispatcher := cache.NewDispatcher(&config.Cache{
		Size:       10,
		Zone:       10,
		HitForPass: 30,
	})

	compressConfig := &config.Compress{
		Filter:    "text|json|javascript",
		MinLength: 1,
	}
	fn := newCacheDispatchMiddleware(dispatcher, compressConfig, true)

	t.Run("no cache", func(t *testing.T) {
		assert := assert.New(t)
		req := httptest.NewRequest("GET", "https://aslant.site/users/me", nil)
		req.Header.Set(elton.HeaderAcceptEncoding, elton.Gzip)
		resp := httptest.NewRecorder()

		c := elton.NewContext(resp, req)
		count := 0
		c.Next = func() error {
			count++
			c.SetHeader(elton.HeaderContentType, "text/plain")
			c.BodyBuffer = bytes.NewBufferString("abcd")
			return nil
		}
		err := fn(c)
		assert.Nil(err)
		assert.Equal(1, count)
		assert.Equal(cache.StatusFetching, c.Get(statusKey))
		assert.NotEmpty(c.BodyBuffer)
		assert.NotEmpty(c.GetHeader(elton.HeaderETag))
		assert.Equal(elton.Gzip, c.GetHeader(elton.HeaderContentEncoding))

		// 第二次请求hit for pass
		c.Response = httptest.NewRecorder()
		err = fn(c)
		assert.Nil(err)
		assert.Equal(2, count)
		assert.Equal(cache.StatusHitForPass, c.Get(statusKey))
		assert.NotEmpty(c.BodyBuffer)
		assert.NotEmpty(c.GetHeader(elton.HeaderETag))
		assert.Equal(elton.Gzip, c.GetHeader(elton.HeaderContentEncoding))
	})

	t.Run("cacheable", func(t *testing.T) {
		assert := assert.New(t)
		req := httptest.NewRequest("GET", "https://aslant.site/users", nil)
		req.Header.Set(elton.HeaderAcceptEncoding, elton.Br)
		resp := httptest.NewRecorder()

		c := elton.NewContext(resp, req)
		count := 0
		c.Next = func() error {
			count++
			c.CacheMaxAge("10s")
			c.SetHeader(elton.HeaderContentType, "application/json")
			c.BodyBuffer = bytes.NewBufferString(`{
				"users": [
					{
						"account": "foo"
					},
					{
						"account": "bar"
					}
				]
			}`)
			return nil
		}
		err := fn(c)
		assert.Nil(err)
		assert.Equal(1, count)
		assert.Equal(cache.StatusFetching, c.Get(statusKey))
		assert.NotEmpty(c.BodyBuffer)
		assert.NotEmpty(c.GetHeader(elton.HeaderETag))
		assert.Equal(elton.Br, c.GetHeader(elton.HeaderContentEncoding))

		// 第二次读取缓存
		time.Sleep(time.Second)
		c.Response = httptest.NewRecorder()
		err = fn(c)
		assert.Nil(err)
		assert.Equal(1, count)
		assert.NotEmpty(c.GetHeader(headerAge))
		assert.Equal(cache.StatusCacheable, c.Get(statusKey))
		assert.NotEmpty(c.BodyBuffer)
		assert.NotEmpty(c.GetHeader(elton.HeaderETag))
		assert.Equal(elton.Br, c.GetHeader(elton.HeaderContentEncoding))
	})

	t.Run("pass", func(t *testing.T) {
		assert := assert.New(t)
		req := httptest.NewRequest("POST", "https://aslant.site/users/login", nil)
		req.Header.Set(elton.HeaderAcceptEncoding, elton.Gzip)
		resp := httptest.NewRecorder()

		c := elton.NewContext(resp, req)
		count := 0
		c.Next = func() error {
			count++
			c.SetHeader(elton.HeaderContentType, "text/plain")
			c.BodyBuffer = bytes.NewBufferString("abcd")
			return nil
		}
		err := fn(c)
		assert.Nil(err)
		assert.Equal(1, count)
		assert.Equal(cache.StatusPassed, c.Get(statusKey))
		assert.NotEmpty(c.BodyBuffer)
		assert.NotEmpty(c.GetHeader(elton.HeaderETag))
		assert.Equal(elton.Gzip, c.GetHeader(elton.HeaderContentEncoding))

		// 第二次请求还是pass
		c.Response = httptest.NewRecorder()
		err = fn(c)
		assert.Nil(err)
		assert.Equal(2, count)
		assert.Equal(cache.StatusPassed, c.Get(statusKey))
		assert.NotEmpty(c.BodyBuffer)
		assert.NotEmpty(c.GetHeader(elton.HeaderETag))
		assert.Equal(elton.Gzip, c.GetHeader(elton.HeaderContentEncoding))
	})
}
