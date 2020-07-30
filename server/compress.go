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

		var rawBody []byte
		var gzipBody []byte
		bufLength := 0
		if c.BodyBuffer != nil {
			// 如果后端程序已压缩，则简单使用压缩后的长度来判断
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
		httpData.CacheHeaders(headers, elton.HeaderContentEncoding, elton.HeaderContentLength)

		// 如果指定压缩数据类型的filter
		// 数据长度大于等于最小压缩长度
		// 而且数据类型符合filter
		if filter != nil &&
			bufLength >= compressConfig.MinLength &&
			filter.MatchString(headers.Get(elton.HeaderContentType)) {
			// TODO 对于不可缓存请求，是否可以pipe的形式返回
			// 不可缓存的请求，只按需压缩
			if !cacheable {
				acceptEncoding := c.GetRequestHeader(elton.HeaderAcceptEncoding)
				if strings.Contains(acceptEncoding, elton.Br) {
					_ = httpData.DoBrotli(compressConfig.Level)
				} else if strings.Contains(acceptEncoding, elton.Gzip) {
					_ = httpData.DoGzip(compressConfig.Level)
				}
			} else {
				// 可缓存的则预压缩
				// 如果出错忽略
				_ = httpData.DoGzip(compressConfig.Level)
				_ = httpData.DoBrotli(compressConfig.Level)
			}

			// 如果有压缩gzip成功，则删除raw body，因为绝大部分客户端都支持gzip，不支持也可以通过gzip解压
			if len(httpData.GzipBody) != 0 {
				httpData.RawBody = nil
			}
		}
		httpData.SetResponse(c, true)
		return
	}
}
