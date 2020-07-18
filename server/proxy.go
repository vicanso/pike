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
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/vicanso/elton"
	"github.com/vicanso/elton/middleware"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/config"
	"github.com/vicanso/pike/upstream"
	"github.com/vicanso/pike/util"
	"golang.org/x/net/http2"
)

var (
	// 需要清除的header
	clearHeaders = []string{
		"Date",
		"Connection",
		elton.HeaderContentLength,
	}
)

const (
	snappyEncoding = "snz"
	lz4Encoding    = "lz4"
)

// 解压函数
type Decompression func([]byte) ([]byte, error)

func newProxyHandlers(locations config.Locations, upstreams *upstream.Upstreams) (proxyMids map[string]elton.Handler, acceptEncodings map[string]string) {
	proxyMids = make(map[string]elton.Handler)
	acceptEncodings = make(map[string]string)

	for _, item := range locations {
		up := upstreams.Get(item.Upstream)
		if up == nil {
			continue
		}
		var transport http.RoundTripper
		// 判断该upstream是否支持h2c的形式
		if upstreams.H2CIsEnabled(item.Upstream) {
			transport = &http2.Transport{
				// 允许使用http的方式
				AllowHTTP: true,
				// tls的dial覆盖
				DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
					return net.Dial(network, addr)
				},
			}
		} else {
			transport = &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
					DualStack: true,
				}).DialContext,
				ForceAttemptHTTP2: true,
				MaxIdleConns:      1000,
				// 调整默认的每个host的最大连接因为缓存服务与backend调用较多
				MaxIdleConnsPerHost:   50,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			}
		}
		// location 对应的upstream接受的编码
		acceptEncoding := upstreams.GetAcceptEncoding(item.Upstream)
		// 如果配置了则设置该location对应的编码
		if acceptEncoding != "" {
			acceptEncodings[item.Name] = acceptEncoding
		}

		proxyMids[item.Name] = middleware.NewProxy(middleware.ProxyConfig{
			Rewrites:  item.Rewrites,
			Transport: transport,
			TargetPicker: func(c *elton.Context) (*url.URL, middleware.ProxyDone, error) {
				httpUpstream, done := up.Next()
				if httpUpstream == nil {
					return nil, nil, errServiceUnavailable
				}
				// 如果不需要设置done
				if done == nil {
					return httpUpstream.URL, nil, nil
				}
				// 返回了done（如最少连接数的策略）
				return httpUpstream.URL, func(_ *elton.Context) {
					done()
				}, nil
			},
		})
	}
	return
}

// decompress 对于snappy与lz4的压缩解压
func decompress(c *elton.Context) (err error) {
	var fn Decompression
	switch c.GetHeader(elton.HeaderContentEncoding) {
	case snappyEncoding:
		fn = util.DecodeSnappy
	case lz4Encoding:
		fn = util.DecodeLz4
	}
	// 如果无匹配的压缩处理，则返回
	if fn == nil {
		return
	}
	buf, err := fn(c.BodyBuffer.Bytes())
	if err != nil {
		return err
	}
	c.BodyBuffer = bytes.NewBuffer(buf)
	c.Header().Del(elton.HeaderContentEncoding)
	return
}

// createProxyMiddleware create proxy middleware handler
func createProxyMiddleware(locations config.Locations, upstreams *upstream.Upstreams) elton.Handler {
	proxyMids, acceptEncodings := newProxyHandlers(locations, upstreams)
	return func(c *elton.Context) (err error) {
		originalNext := c.Next
		// 由于proxy中间件会调用next，因此直接覆盖，
		// 避免导致先执行了后续的中间件（保证在函数调用next前是已完成此中间件处理）
		c.Next = func() error {
			return nil
		}

		host := c.Request.Host
		url := c.Request.RequestURI
		l := locations.GetMatch(host, url)
		if l == nil {
			err = errServiceUnavailable
			return
		}
		// 设置请求头
		if l.ReqHeader != nil {
			util.MergeHeader(c.Request.Header, l.ReqHeader)
			host := l.ReqHeader.Get("Host")
			// 如果有配置Host请求头，则设置request host
			if host != "" {
				c.Request.Host = host
			}
		}
		// 设置响应头
		if l.ResHeader != nil {
			util.MergeHeader(c.Header(), l.ResHeader)
		}

		fn, ok := proxyMids[l.Name]
		if !ok || fn == nil {
			err = errServiceUnavailable
			return
		}

		reqHeader := c.Request.Header
		var ifModifiedSince, ifNoneMatch string
		status := c.GetInt(statusKey)
		acceptEncoding := reqHeader.Get(elton.HeaderAcceptEncoding)
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
		// 设置默认支持snappy lz4与gzip压缩
		reqAcceptEncoding, ok := acceptEncodings[l.Name]
		if !ok {
			reqAcceptEncoding = elton.Gzip
		}
		reqHeader.Set(elton.HeaderAcceptEncoding, reqAcceptEncoding)

		err = fn(c)

		// 解压处理，解压出错的忽略
		_ = decompress(c)

		// 将原有的请求头恢复（就算出错也需要恢复）
		reqHeader.Set(elton.HeaderAcceptEncoding, acceptEncoding)
		if ifModifiedSince != "" {
			reqHeader.Set(elton.HeaderIfModifiedSince, ifModifiedSince)
		}
		if ifNoneMatch != "" {
			reqHeader.Set(elton.HeaderIfNoneMatch, ifNoneMatch)
		}
		if err != nil {
			return
		}
		for _, key := range clearHeaders {
			// 清除header
			c.SetHeader(key, "")
		}

		return originalNext()
	}
}
