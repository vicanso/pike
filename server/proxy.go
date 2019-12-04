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
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/vicanso/elton"
	proxy "github.com/vicanso/elton-proxy"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/config"
	"github.com/vicanso/pike/upstream"
	"github.com/vicanso/pike/util"
)

func newProxyHandlers(locations config.Locations, upstreams *upstream.Upstreams) map[string]elton.Handler {
	proxyMids := make(map[string]elton.Handler)
	defaultTransport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		// 与backend的请求都是内网，不使用http2
		ForceAttemptHTTP2: false,
		MaxIdleConns:      1000,
		// 调整默认的每个host的最大连接因为缓存服务与backend调用较多
		MaxIdleConnsPerHost:   50,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	for _, item := range locations {
		up := upstreams.Get(item.Upstream)
		if up == nil {
			continue
		}
		proxyMids[item.Name] = proxy.New(proxy.Config{
			Transport: defaultTransport,
			TargetPicker: func(c *elton.Context) (*url.URL, proxy.Done, error) {
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
	return proxyMids
}

// createProxyMiddleware create proxy middleware handler
func createProxyMiddleware(locations config.Locations, upstreams *upstream.Upstreams) elton.Handler {
	proxyMids := newProxyHandlers(locations, upstreams)
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
		if l.RequestHeader != nil {
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

		fn := proxyMids[l.Name]
		if fn == nil {
			err = errServiceUnavailable
			return
		}

		reqHeader := c.Request.Header
		var ifModifiedSince, ifNoneMatch, acceptEncoding string
		status, _ := c.Get(statusKey).(int)
		// 针对fetching的请求，由于其最终状态未知，因此需要删除有可能导致304的请求，避免无法生成缓存
		if status == cache.StatusFetching {
			acceptEncoding = reqHeader.Get(elton.HeaderAcceptEncoding)
			ifModifiedSince = reqHeader.Get(elton.HeaderIfModifiedSince)
			ifNoneMatch = reqHeader.Get(elton.HeaderIfNoneMatch)
			if ifModifiedSince != "" {
				reqHeader.Del(elton.HeaderIfModifiedSince)
			}
			if ifNoneMatch != "" {
				reqHeader.Del(elton.HeaderIfNoneMatch)
			}

			if strings.Contains(acceptEncoding, elton.Gzip) {
				reqHeader.Set(elton.HeaderAcceptEncoding, elton.Gzip)
			} else {
				reqHeader.Del(elton.HeaderAcceptEncoding)
			}
		}

		err = fn(c)

		// 将原有的请求头恢复（就算出错也需要恢复）
		if acceptEncoding != "" {
			reqHeader.Set(elton.HeaderAcceptEncoding, acceptEncoding)
		}
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
