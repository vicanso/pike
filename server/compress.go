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
	"regexp"
	"strings"

	"github.com/vicanso/elton"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/config"
	"github.com/vicanso/pike/util"
)

type (
	// CompressHandler compress handler
	CompressHandler func(*elton.Context, bool) (httpData *cache.HTTPData)
	compressFn      func(buf []byte, level int) ([]byte, error)
)

func compressForNotCacheable(c *elton.Context, level int) {
	acceptEncoding := c.GetRequestHeader(elton.HeaderAcceptEncoding)
	var fn compressFn
	encoding := ""
	// 如果支持br
	if strings.Contains(acceptEncoding, elton.Br) {
		fn = util.Brotli
		encoding = elton.Br
	} else if strings.Contains(acceptEncoding, elton.Gzip) {
		fn = util.Gzip
		encoding = elton.Gzip
	}
	if encoding == "" {
		return
	}
	buf, _ := fn(c.BodyBuffer.Bytes(), level)
	if len(buf) != 0 {
		c.SetHeader(elton.HeaderContentEncoding, encoding)
		c.BodyBuffer = bytes.NewBuffer(buf)
	}
}

func createCompressHandler(compressConfig *config.Compress) CompressHandler {
	var filter *regexp.Regexp
	if compressConfig != nil && compressConfig.Filter != "" {
		filter, _ = regexp.Compile(compressConfig.Filter)
	}
	// 如果是可缓存数据会返回httpData，如果不可缓存数据，直接生成响应数据则返回
	return func(c *elton.Context, cacheable bool) (httpData *cache.HTTPData) {
		// 前置已针对未支持的encoding过滤，因此data只可能为未压缩或者gzip数据
		encoding := c.GetHeader(elton.HeaderContentEncoding)

		// 不可缓存的，响应仅用于本次
		if !cacheable {
			// 如果未指定filter，或者已压缩过，或者数据长度较小，则直接返回
			// 对于不可缓存数据，不做动态压缩匹配（br gzip)
			if filter == nil ||
				encoding != "" ||
				c.BodyBuffer == nil || c.BodyBuffer.Len() < compressConfig.MinLength {
				return
			}
			// 未压缩的则使用压缩处理
			compressForNotCacheable(c, compressConfig.Level)
			return
		}
		var rawBody []byte
		var gzipBody []byte
		bufLength := 0
		if c.BodyBuffer != nil {
			if encoding == elton.Gzip {
				gzipBody = c.BodyBuffer.Bytes()
				bufLength = len(gzipBody)
			} else {
				rawBody = c.BodyBuffer.Bytes()
				bufLength = len(rawBody)
			}
		}

		headers := c.Header()
		httpData = &cache.HTTPData{
			StatusCode: c.StatusCode,
			RawBody:    rawBody,
			GzipBody:   gzipBody,
		}
		httpData.Headers = cache.NewHTTPHeaders(headers, elton.HeaderContentEncoding, elton.HeaderContentLength)

		// 如果未指定filter的数据或少于最小压缩长度
		if filter == nil || bufLength < compressConfig.MinLength {
			return
		}
		contentType := headers.Get(elton.HeaderContentType)
		if !filter.MatchString(contentType) {
			return
		}
		// 如果有gzip数据
		if len(gzipBody) != 0 {
			rawBody, _ = util.Gunzip(gzipBody)
		} else {
			// 从原始数据中压缩gzip
			httpData.GzipBody, _ = util.Gzip(rawBody, compressConfig.Level)
			// 如果gzip压缩成功，则删除rawBody（因为基本所有客户端都支持gzip，若不支持则从gzip解压获取），减少内存占用
			if len(httpData.GzipBody) != 0 {
				httpData.RawBody = nil
			}
		}
		if len(rawBody) != 0 {
			httpData.BrBody, _ = util.Brotli(rawBody, compressConfig.Level)
		}
		httpData.SetResponse(c)
		return
	}
}
