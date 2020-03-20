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
	"strings"
	"sync/atomic"
	"time"

	"github.com/vicanso/elton"
	fresh "github.com/vicanso/elton-fresh"
	recover "github.com/vicanso/elton-recover"
	"github.com/vicanso/hes"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/config"
	"github.com/vicanso/pike/log"
	"github.com/vicanso/pike/upstream"
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

// EltonOptions elton options
type EltonOptions struct {
	name      string
	influxSrv *InfluxSrv
	cfg       *config.Config
	// adminConfig    *config.Admin
	maxConcurrency uint32
	eTag           bool
	locations      config.Locations
	upstreams      *upstream.Upstreams
	dispatcher     *cache.Dispatcher
	compress       *config.Compress
}

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

// NewElton new an elton instance
func NewElton(eltonOptions *EltonOptions) *elton.Elton {
	cfg := eltonOptions.cfg
	logger := log.Default()
	locations := eltonOptions.locations
	upstreams := eltonOptions.upstreams
	dispatcher := eltonOptions.dispatcher
	e := elton.New()

	adminPath, adminElton := NewAdmin(cfg)

	// 未处理错误
	e.OnError(newErrorListener(dispatcher, logger))
	e.Use(recover.New())
	influxSrv := eltonOptions.influxSrv
	if influxSrv != nil {
		e.Use(func(c *elton.Context) error {
			startedAt := time.Now()
			req := c.Request
			fields := map[string]interface{}{
				"url":  req.RequestURI,
				"ip":   c.RealIP(),
				"path": req.URL.Path,
			}
			tags := map[string]string{
				"method": req.Method,
				"host":   req.Host,
				"server": eltonOptions.name,
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
			influxSrv.Write("pike-http", fields, tags)
			return err
		})
	}

	var concurrency uint32
	maxConcurrency := eltonOptions.maxConcurrency
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

	e.Use(fresh.NewDefault())

	// get http cache
	e.Use(newCacheDispatchMiddleware(dispatcher, eltonOptions.compress, eltonOptions.eTag))

	// http request proxy
	e.Use(createProxyMiddleware(locations, upstreams))

	e.ALL("/*url", func(c *elton.Context) error {
		return nil
	})
	return e
}
