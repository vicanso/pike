package server

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/vicanso/cod"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/config"
	"github.com/vicanso/pike/df"
	"github.com/vicanso/pike/middleware"
	"github.com/vicanso/pike/stats"
	"github.com/vicanso/pike/upstream"

	compress "github.com/vicanso/cod-compress"
	etag "github.com/vicanso/cod-etag"
	fresh "github.com/vicanso/cod-fresh"
	recover "github.com/vicanso/cod-recover"
)

type (
	// Options server options
	Options struct {
		BasicConfig    *config.Config
		DirectorConfig *config.DirectorConfig
		Director       *upstream.Director
		Dispatcher     *cache.Dispatcher
		Stats          *stats.Stats
	}
)

// NewServer create a cod server
func NewServer(opts Options) *cod.Cod {
	d := cod.NewWithoutServer()
	cfg := opts.BasicConfig
	insStats := opts.Stats
	director := opts.Director
	dsp := opts.Dispatcher

	d.EnableTrace = cfg.Data.EnableServerTiming
	// 如果启动 server timing
	// 则在回调中调整响应头
	if d.EnableTrace {
		prefix := df.APP + "-"
		d.OnTrace(func(c *cod.Context, traceInfos cod.TraceInfos) {
			c.SetHeader(cod.HeaderServerTiming, string(traceInfos.ServerTiming(prefix)))
		})
	}

	// 如果有配置admin，则添加管理后台处理
	adminPath := cfg.Data.Admin.Prefix
	if adminPath != "" {
		adminServer := NewAdminServer(opts)
		d.Use(func(c *cod.Context) error {
			path := c.Request.URL.Path
			if strings.HasPrefix(path, adminPath) {
				c.Request.URL.Path = path[len(adminPath):]
				c.Pass(adminServer)
				return nil
			}
			return c.Next()
		})
	}

	d.Use(recover.New())

	ping := func(c *cod.Context) error {
		if c.Request.RequestURI == "/ping" {
			c.BodyBuffer = bytes.NewBufferString("pong")
			return nil
		}
		return c.Next()
	}
	d.Use(ping)
	d.SetFunctionName(ping, "-")

	fn := middleware.NewInitialization(cfg.Data, insStats)
	d.Use(fn)
	d.SetFunctionName(fn, "Initialization")

	fn = fresh.NewDefault()
	d.Use(fn)
	d.SetFunctionName(fn, "Fresh")

	// 可缓存数据在缓存时会生成gzip 与br
	compressConfig := cfg.Data.Compress
	textFilter := regexp.MustCompile(compressConfig.Filter)
	fn = compress.NewWithDefaultCompressor(compress.Config{
		MinLength: compressConfig.MinLength,
		Level:     compressConfig.Level,
		Checker:   textFilter,
	})
	d.Use(fn)
	d.SetFunctionName(fn, "Compress")

	fn = middleware.NewResponder()
	d.Use(fn)
	d.SetFunctionName(fn, "Responder")

	fn = middleware.NewCacheIdentifier(cfg.Data, dsp)
	d.Use(fn)
	d.SetFunctionName(fn, "CacheIdentifier")

	fn = etag.NewDefault()
	d.Use(fn)
	d.SetFunctionName(fn, "ETag")

	fn = middleware.NewProxy(director)
	d.Use(fn)
	d.SetFunctionName(fn, "Proxy")

	noop := func(c *cod.Context) (err error) {
		return
	}
	d.ALL("/*url", noop)
	d.SetFunctionName(noop, "-")
	return d
}
