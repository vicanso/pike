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
	"net/http"
	"regexp"
	"strconv"

	"github.com/vicanso/elton"
	"github.com/vicanso/hes"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/location"
	"github.com/vicanso/pike/upstream"
	"github.com/vicanso/pike/util"
	"golang.org/x/net/context"
)

var (
	noCacheReg = regexp.MustCompile(`no-cache|no-store|private`)
	sMaxAgeReg = regexp.MustCompile(`s-maxage=(\d+)`)
	maxAgeReg  = regexp.MustCompile(`max-age=(\d+)`)
)

// 根据Cache-Control的信息，获取s-maxage 或者max-age的值
func getCacheMaxAge(header http.Header) int {
	// 如果有设置cookie，则为不可缓存
	if header.Get(elton.HeaderSetCookie) != "" {
		return 0
	}
	// 如果没有设置cache-control，则不可缓存
	cc := header.Get(elton.HeaderCacheControl)
	if cc == "" {
		return 0
	}

	// 如果设置不可缓存，返回0
	if noCacheReg.MatchString(cc) {
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
	if age := header.Get(headerAge); age != "" {
		v, _ := strconv.Atoi(age)
		maxAge -= v
	}

	return maxAge
}

// NewProxy create proxy middleware
func NewProxy(s *server) elton.Handler {
	return func(c *elton.Context) (err error) {
		originalNext := c.Next
		// 由于proxy中间件会调用next，因此直接覆盖，
		// 避免导致先执行了后续的中间件（保证在函数调用next前是已完成此中间件处理）
		c.Next = func() error {
			return nil
		}

		l := location.Get(c.Request.Host, c.Request.RequestURI, s.GetLocations()...)
		if l == nil {
			err = ErrLocationNotFound
			return
		}

		upstream := upstream.Get(l.Upstream)
		if upstream == nil {
			err = ErrUpstreamNotFound
			return
		}

		reqHeader := c.Request.Header
		var ifModifiedSince, ifNoneMatch string
		status := getCacheStatus(c)
		// 针对fetching的请求，由于其最终状态未知，因此需要删除有可能导致304的请求，避免无法生成缓存
		if status == cache.StatusFetching {
			ifModifiedSince = reqHeader.Get(elton.HeaderIfModifiedSince)
			ifNoneMatch = reqHeader.Get(elton.HeaderIfNoneMatch)
			if ifModifiedSince != "" {
				reqHeader.Del(elton.HeaderIfModifiedSince)
			}
			if ifNoneMatch != "" {
				reqHeader.Del(elton.HeaderIfNoneMatch)
			}
		}

		// url rewrite
		var originalPath string
		if l.URLRewriter != nil {
			originalPath = c.Request.URL.Path
			l.URLRewriter(c.Request)
		}
		// 添加额外的请求头
		l.AddRequestHeader(reqHeader)

		// 添加query string
		var originRawQuery string
		if l.ShouldModifyQuery() {
			originRawQuery = c.Request.URL.RawQuery
			l.AddQuery(c.Request)
		}

		var acceptEncoding string

		// 根据upstream设置可接受压缩编码调整
		acceptEncodingChanged := upstream.Option.AcceptEncoding != ""
		if acceptEncodingChanged {
			acceptEncoding = reqHeader.Get(elton.HeaderAcceptEncoding)
			reqHeader.Set(elton.HeaderAcceptEncoding, upstream.Option.AcceptEncoding)
		}

		if l.ProxyTimeout != 0 {
			ctx, cancel := context.WithTimeout(c.Context(), l.ProxyTimeout)
			defer cancel()
			c.WithContext(ctx)
		}

		// clone当前header，用于后续恢复
		originalHeader := c.Header().Clone()
		c.ResetHeader()
		err = upstream.Proxy(c)
		// 如果出错超时，则转换为504 timeout，category:pike
		if err != nil {
			if he, ok := err.(*hes.Error); ok {
				if he.Err == context.DeadlineExceeded {
					err = util.NewError("Timeout", http.StatusGatewayTimeout)
				}
			}
		}

		// 恢复请求头
		if ifModifiedSince != "" {
			reqHeader.Set(elton.HeaderIfModifiedSince, ifModifiedSince)
		}
		if ifNoneMatch != "" {
			reqHeader.Set(elton.HeaderIfNoneMatch, ifNoneMatch)
		}
		if acceptEncodingChanged {
			reqHeader.Set(elton.HeaderAcceptEncoding, acceptEncoding)
		}

		// 恢复query
		if originRawQuery != "" {
			c.Request.URL.RawQuery = originRawQuery
		}

		header := c.Header()
		// 添加额外的响应头
		l.AddResponseHeader(header)
		// 恢复原始url path
		if originalPath != "" {
			c.Request.URL.Path = originalPath
		}

		if err != nil {
			return
		}

		var data []byte
		if c.BodyBuffer != nil {
			data = c.BodyBuffer.Bytes()
		}

		// 对于fetching的请求，从响应头中判断该请求缓存的有效期
		if status == cache.StatusFetching {
			maxAge := getCacheMaxAge(header)
			if maxAge > 0 {
				setHTTPCacheMaxAge(c, maxAge)
			}
		}

		// 初始化http response时，如果已压缩，而且非gzip br，则会解压
		httpResp, err := cache.NewHTTPResponse(c.StatusCode, header, header.Get(elton.HeaderContentEncoding), data)
		if err != nil {
			return
		}

		compressSrv, minLength, filter := s.GetCompress()
		httpResp.CompressSrv = compressSrv
		httpResp.CompressMinLength = minLength
		httpResp.CompressContentTypeFilter = filter
		setHTTPResp(c, httpResp)

		// 重置context中由于proxy中间件影响的状态 statusCode, header, body
		// 因为最终响应会从http response中生成，该响应会包括http响应头，
		// 因此清除现在的header并恢复原来的header
		c.ResetHeader()
		c.MergeHeader(originalHeader)
		c.BodyBuffer = nil
		c.StatusCode = 0
		c.Next = originalNext
		return c.Next()
	}
}
