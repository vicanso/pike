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
	"net/http/httptest"
	"testing"

	"github.com/vicanso/pike/config"

	"github.com/vicanso/elton"
	"github.com/vicanso/pike/util"

	"github.com/stretchr/testify/assert"
)

func TestCompressForNotCacheable(t *testing.T) {
	t.Run("br compress", func(t *testing.T) {
		assert := assert.New(t)
		req := httptest.NewRequest("GET", "/", nil)
		resp := httptest.NewRecorder()
		c := elton.NewContext(resp, req)
		c.SetRequestHeader(elton.HeaderAcceptEncoding, "br")
		buf := []byte("abcd")
		c.BodyBuffer = bytes.NewBuffer(buf)
		compressForNotCacheable(c, 0)
		assert.Equal(c.GetHeader(elton.HeaderContentEncoding), "br")
		assert.NotEmpty(c.BodyBuffer)
		dst, _ := util.BrotliDecode(c.BodyBuffer.Bytes())
		assert.Equal(buf, dst)
	})

	t.Run("gzip compress", func(t *testing.T) {
		assert := assert.New(t)
		req := httptest.NewRequest("GET", "/", nil)
		resp := httptest.NewRecorder()
		c := elton.NewContext(resp, req)
		c.SetRequestHeader(elton.HeaderAcceptEncoding, "gzip")
		buf := []byte("abcd")
		c.BodyBuffer = bytes.NewBuffer(buf)
		compressForNotCacheable(c, 0)
		assert.Equal(c.GetHeader(elton.HeaderContentEncoding), "gzip")
		assert.NotEmpty(c.BodyBuffer)
		dst, _ := util.Gunzip(c.BodyBuffer.Bytes())
		assert.Equal(buf, dst)
	})
}

func TestCreateCompressHandler(t *testing.T) {
	t.Run("not cacheable pass compress", func(t *testing.T) {
		assert := assert.New(t)
		// no filter
		fn := createCompressHandler(&config.Compress{})
		req := httptest.NewRequest("GET", "/", nil)
		resp := httptest.NewRecorder()
		c := elton.NewContext(resp, req)
		c.SetRequestHeader(elton.HeaderAcceptEncoding, "gzip")
		buf := []byte("abcd")
		c.BodyBuffer = bytes.NewBuffer(buf)
		httpData := fn(c, false)
		assert.Nil(httpData)
		assert.Equal(buf, c.BodyBuffer.Bytes())

		// content encoding is set
		fn = createCompressHandler(&config.Compress{
			Filter: "text|json|javascript",
		})
		req = httptest.NewRequest("GET", "/", nil)
		resp = httptest.NewRecorder()
		c = elton.NewContext(resp, req)
		c.SetHeader(elton.HeaderContentEncoding, "gzip")
		c.SetRequestHeader(elton.HeaderAcceptEncoding, "gzip")
		buf = []byte("abcd")
		c.BodyBuffer = bytes.NewBuffer(buf)
		httpData = fn(c, false)
		assert.Nil(httpData)
		assert.Equal(buf, c.BodyBuffer.Bytes())

		// buffer's length less than min compress length
		fn = createCompressHandler(&config.Compress{
			Filter:    "text|json|javascript",
			MinLength: 1024,
		})
		req = httptest.NewRequest("GET", "/", nil)
		resp = httptest.NewRecorder()
		c = elton.NewContext(resp, req)
		c.SetRequestHeader(elton.HeaderAcceptEncoding, "gzip")
		buf = []byte("abcd")
		c.BodyBuffer = bytes.NewBuffer(buf)
		httpData = fn(c, false)
		assert.Nil(httpData)
		assert.Equal(buf, c.BodyBuffer.Bytes())
	})

	t.Run("cacheable", func(t *testing.T) {
		assert := assert.New(t)
		fn := createCompressHandler(&config.Compress{
			Filter:    "text|json|javascript",
			MinLength: 1,
			Level:     6,
		})
		req := httptest.NewRequest("GET", "/", nil)
		resp := httptest.NewRecorder()
		c := elton.NewContext(resp, req)
		buf := []byte("abcd")
		c.SetContentTypeByExt(".txt")
		c.BodyBuffer = bytes.NewBuffer(buf)
		c.SetRequestHeader(elton.HeaderAcceptEncoding, "gzip")
		httpData := fn(c, true)
		assert.NotNil(httpData)
		assert.NotEmpty(httpData.GzipBody)
		assert.NotEmpty(httpData.BrBody)
		assert.Empty(httpData.RawBody)
		assert.Equal(httpData.GzipBody, c.BodyBuffer.Bytes())
		assert.Equal(1, len(httpData.Headers))
		dst, _ := util.Gunzip(httpData.GzipBody)
		assert.Equal(buf, dst)
		dst, _ = util.BrotliDecode(httpData.BrBody)
		assert.Equal(buf, dst)
	})

	t.Run("cacheable(204)", func(t *testing.T) {
		assert := assert.New(t)
		fn := createCompressHandler(&config.Compress{
			Filter:    "text|json|javascript",
			MinLength: 1,
			Level:     6,
		})
		req := httptest.NewRequest("GET", "/", nil)
		resp := httptest.NewRecorder()
		c := elton.NewContext(resp, req)
		c.NoContent()
		httpData := fn(c, true)
		assert.NotNil(httpData)
		assert.Nil(c.BodyBuffer)
		assert.Equal(204, httpData.StatusCode)
		assert.Nil(httpData.RawBody)
		assert.Nil(httpData.GzipBody)
		assert.Nil(httpData.BrBody)
	})
	t.Run("cacheable(not match filter)", func(t *testing.T) {
		assert := assert.New(t)
		fn := createCompressHandler(&config.Compress{
			Filter:    "text|json|javascript",
			MinLength: 1,
			Level:     6,
		})
		req := httptest.NewRequest("GET", "/", nil)
		resp := httptest.NewRecorder()
		c := elton.NewContext(resp, req)
		buf := []byte("abc")
		c.SetContentTypeByExt(".png")
		c.BodyBuffer = bytes.NewBuffer(buf)
		httpData := fn(c, true)
		assert.NotNil(httpData)
		assert.Equal(buf, c.BodyBuffer.Bytes())
		assert.Equal(buf, httpData.RawBody)
		assert.Nil(httpData.GzipBody)
		assert.Nil(httpData.BrBody)
	})
	t.Run("cacheable(response body is gzip)", func(t *testing.T) {
		assert := assert.New(t)
		fn := createCompressHandler(&config.Compress{
			Filter:    "text|json|javascript",
			MinLength: 1,
			Level:     6,
		})
		req := httptest.NewRequest("GET", "/", nil)
		resp := httptest.NewRecorder()
		c := elton.NewContext(resp, req)
		buf := []byte("abcd")
		gzipBuf, _ := util.Gzip(buf, 0)
		c.SetContentTypeByExt(".txt")
		c.BodyBuffer = bytes.NewBuffer(gzipBuf)
		c.SetHeader(elton.HeaderContentEncoding, "gzip")
		c.SetRequestHeader(elton.HeaderAcceptEncoding, "gzip")
		httpData := fn(c, true)
		assert.NotNil(httpData)
		assert.NotEmpty(httpData.GzipBody)
		assert.NotEmpty(httpData.BrBody)
		assert.Empty(httpData.RawBody)
		assert.Equal(httpData.GzipBody, c.BodyBuffer.Bytes())
		assert.Equal(1, len(httpData.Headers))
		dst, _ := util.Gunzip(httpData.GzipBody)
		assert.Equal(buf, dst)
		dst, _ = util.BrotliDecode(httpData.BrBody)
		assert.Equal(buf, dst)
	})
}
