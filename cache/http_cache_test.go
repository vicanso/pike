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

package cache

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/vicanso/elton"
	"github.com/vicanso/pike/util"

	"github.com/stretchr/testify/assert"
)

func TestHTTPHeaders(t *testing.T) {
	assert := assert.New(t)
	header := make(http.Header)
	header.Add("A", "1")
	header.Add("A", "2")
	header.Add("B", "3")
	header.Add("D", "4")
	headers := NewHTTPHeaders(header, "D")
	assert.Equal(3, len(headers))
	fieldA := ""
	fieldB := ""
	for _, item := range headers {
		k := string(item[0])
		v := string(item[1])
		if k == "A" {
			fieldA += v
		} else {
			fieldB += v
		}
	}
	assert.Equal("12", fieldA)
	assert.Equal("3", fieldB)
}

func TestHTTPCacheGet(t *testing.T) {
	t.Run("fetching", func(t *testing.T) {
		assert := assert.New(t)
		hc := NewHTTPCache()
		status, data := hc.Get()
		assert.Equal(StatusFetching, status)
		assert.Nil(data)
	})

	t.Run("hit for pass", func(t *testing.T) {
		assert := assert.New(t)
		hc := NewHTTPCache()
		status, data := hc.Get()
		assert.Equal(StatusFetching, status)
		assert.Nil(data)
		go func() {
			time.Sleep(time.Millisecond)
			hc.HitForPass(300)
		}()
		// 此时会因为fetching而等待
		status, data = hc.Get()
		assert.Equal(StatusHitForPass, status)
		assert.Equal(StatusHitForPass, hc.status)
		assert.Nil(data)

		// 此次不会再因为fetching而等待
		status, data = hc.Get()
		assert.Equal(StatusHitForPass, status)
		assert.Equal(StatusHitForPass, hc.status)
		assert.Nil(data)
	})

	t.Run("cachable", func(t *testing.T) {
		assert := assert.New(t)
		hc := NewHTTPCache()
		status, data := hc.Get()
		assert.Equal(StatusFetching, status)
		assert.Nil(data)
		statusCode := 200
		rawBody := []byte("raw body")
		gzipBody := []byte("gzip body")
		brBody := []byte("br body")
		go func() {
			time.Sleep(time.Millisecond)
			hc.Cachable(300, &HTTPData{
				StatusCode: statusCode,
				RawBody:    rawBody,
				GzipBody:   gzipBody,
				BrBody:     brBody,
			})
		}()
		// 此时会因为fetching而等待
		status, data = hc.Get()
		assert.Equal(StatusCacheable, status)
		assert.Equal(StatusCacheable, hc.status)
		assert.Equal(statusCode, data.StatusCode)
		assert.Equal(rawBody, data.RawBody)
		assert.Equal(gzipBody, data.GzipBody)
		assert.Equal(brBody, data.BrBody)

		// 此次不会再因为fetching而等待
		status, data = hc.Get()
		assert.Equal(StatusCacheable, status)
		assert.Equal(StatusCacheable, hc.status)
		assert.Equal(statusCode, data.StatusCode)

		// 设置为过期(expiredAt < now)
		hc.expiredAt = 1
		status, data = hc.Get()
		assert.Equal(StatusFetching, status)
		assert.Nil(data)
	})

	t.Run("get age", func(t *testing.T) {
		assert := assert.New(t)
		age := 10
		hc := HTTPCache{
			createdAt: int(time.Now().Unix()) - age,
		}
		assert.Equal(age, hc.Age())
	})
}

func TestSetResponse(t *testing.T) {
	t.Run("br", func(t *testing.T) {
		assert := assert.New(t)
		resp := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set(elton.HeaderAcceptEncoding, "gzip, deflate, br")
		c := elton.NewContext(resp, req)
		brBody := []byte("br data")
		headers := make(HTTPHeaders, 1)
		headers[0] = HTTPHeader{
			[]byte("a"),
			[]byte("1"),
		}
		httpData := HTTPData{
			BrBody:   brBody,
			GzipBody: []byte("gzip data"),
			Headers:  headers,
		}
		httpData.SetResponse(c)
		assert.Equal(brBody, c.BodyBuffer.Bytes())
		assert.Equal("7", c.GetHeader(elton.HeaderContentLength))
		assert.Equal("br", c.GetHeader(elton.HeaderContentEncoding))
		assert.Equal("1", c.GetHeader("a"))
	})

	t.Run("gzip", func(t *testing.T) {
		assert := assert.New(t)
		resp := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set(elton.HeaderAcceptEncoding, "gzip, deflate, br")
		c := elton.NewContext(resp, req)
		gzipBody := []byte("gzip data")
		httpData := HTTPData{
			GzipBody: gzipBody,
		}
		httpData.SetResponse(c)
		assert.Equal(gzipBody, c.BodyBuffer.Bytes())
		assert.Equal("9", c.GetHeader(elton.HeaderContentLength))
		assert.Equal("gzip", c.GetHeader(elton.HeaderContentEncoding))
	})

	t.Run("raw body", func(t *testing.T) {
		assert := assert.New(t)
		resp := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		c := elton.NewContext(resp, req)
		rawBody := []byte("raw data")
		httpData := HTTPData{
			RawBody: rawBody,
		}
		httpData.SetResponse(c)
		assert.Equal(rawBody, c.BodyBuffer.Bytes())
		assert.Equal("8", c.GetHeader(elton.HeaderContentLength))
		assert.Empty(c.GetHeader(elton.HeaderContentEncoding))
	})

	t.Run("get raw body from gzip", func(t *testing.T) {
		assert := assert.New(t)
		resp := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		c := elton.NewContext(resp, req)
		rawBody := []byte("raw data")
		gzipBody, _ := util.Gzip(rawBody, 9)
		httpData := HTTPData{
			GzipBody: gzipBody,
		}
		httpData.SetResponse(c)
		assert.Equal(rawBody, c.BodyBuffer.Bytes())
		assert.Equal("8", c.GetHeader(elton.HeaderContentLength))
		assert.Empty(c.GetHeader(elton.HeaderContentEncoding))
	})

	t.Run("get status", func(t *testing.T) {
		assert := assert.New(t)
		hc := HTTPCache{
			status: StatusHitForPass,
		}
		assert.Equal(hc.status, hc.GetStatus())
	})
}
