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
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/vicanso/elton"
	"github.com/vicanso/elton/middleware"
	"github.com/vicanso/hes"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/log"
	"go.uber.org/zap"
)

const (
	statusKey    = "status"
	httpCacheKey = "httpCache"

	// 默认的 admin 目录
	defaultAdminPath = "/pike"

	headerAge = "Age"
)

var (
	noCacheReg = regexp.MustCompile(`no-cache|no-store|private`)
	sMaxAgeReg = regexp.MustCompile(`s-maxage=(\d+)`)
	maxAgeReg  = regexp.MustCompile(`max-age=(\d+)`)
)

var (
	errTooManyRequests = &hes.Error{
		StatusCode: http.StatusTooManyRequests,
		Message:    "Too Many Requests",
	}
	errServiceUnavailable = &hes.Error{
		StatusCode: http.StatusServiceUnavailable,
		Message:    "Service Unavailable",
	}
)

func newErrorListener(dispatcher *cache.Dispatcher, logger *zap.Logger) elton.ErrorListener {
	return func(c *elton.Context, err error) {
		logger.Error("uncaught exception",
			zap.String("host", c.Request.Host),
			zap.String("url", c.Request.RequestURI),
			zap.Error(err),
		)
		// 如果没有设置dispatcher，则无需要处理以下流程
		if dispatcher == nil {
			return
		}
		status := c.GetInt(statusKey)
		if status == cache.StatusFetching {
			v, ok := c.Get(httpCacheKey)
			if !ok {
				return
			}
			httpCache, _ := v.(*cache.HTTPCache)
			if httpCache != nil {
				httpCache.HitForPass(dispatcher.HitForPass)
			}
		}
	}
}

// onStats on stats function
type onStats func(map[string]interface{}, map[string]string)

// newStatsHandler create a new stats handler milldeware
func newStatsHandler(name string, fn onStats) elton.Handler {
	return func(c *elton.Context) error {
		startedAt := time.Now()
		req := c.Request
		fields := map[string]interface{}{
			"url":  req.RequestURI,
			"ip":   c.RealIP(),
			"path": req.URL.Path,
		}
		cacheStatus := c.GetInt(statusKey)
		tags := map[string]string{
			"method":      req.Method,
			"host":        req.Host,
			"server":      name,
			"cacheStatus": strconv.Itoa(cacheStatus),
		}
		err := c.Next()
		fields["use"] = time.Since(startedAt).Milliseconds()
		if err != nil {
			he := hes.Wrap(err)
			fields["statusCode"] = he.StatusCode
		} else {
			fields["statusCode"] = c.StatusCode
		}
		if c.BodyBuffer != nil {
			fields["size"] = c.BodyBuffer.Len()
		}
		fn(fields, tags)
		return err
	}
}

// NewElton new an elton instance
func NewElton(opts *ServerOptions) *elton.Elton {
	logger := log.Default()
	locations := opts.locations
	upstreams := opts.upstreams
	dispatcher := opts.dispatcher
	e := elton.New()

	adminPath, adminElton := NewAdmin(opts)

	// 未处理错误
	e.OnError(newErrorListener(dispatcher, logger))
	e.Use(middleware.NewRecover())
	influxSrv := opts.influxSrv
	if influxSrv != nil {
		e.Use(newStatsHandler(opts.name, func(fields map[string]interface{}, tags map[string]string) {
			influxSrv.Write("pike-http", fields, tags)
		}))
	}

	var concurrency uint32
	maxConcurrency := opts.server.Concurrency
	if maxConcurrency > 0 {
		e.Use(func(c *elton.Context) error {
			v := atomic.AddUint32(&concurrency, 1)
			defer atomic.AddUint32(&concurrency, ^uint32(0))
			if v > maxConcurrency {
				return errTooManyRequests
			}
			return c.Next()
		})
	}

	// 如果是admin路径，则转发至admin elton
	if adminElton != nil {
		e.Use(func(c *elton.Context) error {
			if !strings.HasPrefix(c.Request.RequestURI, adminPath) {
				return c.Next()
			}
			c.Pass(adminElton)
			return nil
		})
	}

	e.Use(middleware.NewDefaultFresh())

	// get http cache
	e.Use(newCacheDispatchMiddleware(dispatcher, opts.compress, opts.server.ETag))

	// http request proxy
	e.Use(createProxyMiddleware(locations, upstreams))

	e.ALL("/*", func(c *elton.Context) error {
		return nil
	})
	return e
}
