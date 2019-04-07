package main

import (
	"github.com/vicanso/cod"

	"github.com/vicanso/hes"
	"github.com/vicanso/pike/config"
	"github.com/vicanso/pike/df"
	"github.com/vicanso/pike/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	compress "github.com/vicanso/cod-compress"
	etag "github.com/vicanso/cod-etag"
	fresh "github.com/vicanso/cod-fresh"
	recover "github.com/vicanso/cod-recover"
)

func main() {
	c := zap.NewProductionConfig()
	c.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	// 只针对panic 以上的日志增加stack trace
	logger, err := c.Build(zap.AddStacktrace(zap.DPanicLevel))
	if err != nil {
		panic(err)
	}

	d := cod.New()
	d.EnableTrace = config.IsEnableServerTiming()
	if d.EnableTrace {
		prefix := df.APP + "-"
		d.OnTrace(func(c *cod.Context, traceInfos cod.TraceInfos) {
			c.SetHeader(cod.HeaderServerTiming, string(traceInfos.ServerTiming(prefix)))
		})
	}

	d.OnError(func(c *cod.Context, err error) {
		he := hes.Wrap(err)
		logger.Error(he.Message,
			zap.String("category", he.Category),
			zap.String("method", c.Request.Method),
			zap.String("host", c.Request.Host),
			zap.String("url", c.Request.RequestURI),
		)
	})

	d.Use(recover.New())

	fn := middleware.NewInitialization()
	d.Use(fn)
	d.SetFunctionName(fn, "Initialization")

	fn = fresh.NewDefault()
	d.Use(fn)
	d.SetFunctionName(fn, "Fresh")

	// 可缓存数据在缓存时会生成gzip 与br
	fn = compress.NewWithDefaultCompressor(compress.Config{
		MinLength: config.GetCompressMinLength(),
		Level:     config.GetCompressLevel(),
		Checker:   config.GetTextFilter(),
	})
	d.Use(fn)
	d.SetFunctionName(fn, "Compress")

	fn = middleware.NewResponder()
	d.Use(fn)
	d.SetFunctionName(fn, "Responder")

	fn = middleware.NewCacheIdentifier()
	d.Use(fn)
	d.SetFunctionName(fn, "CacheIdentifier")

	fn = etag.NewDefault()
	d.Use(fn)
	d.SetFunctionName(fn, "ETag")

	fn = middleware.NewProxy()
	d.Use(fn)
	d.SetFunctionName(fn, "Proxy")

	noop := func(c *cod.Context) (err error) {
		return
	}
	d.ALL("/*url", noop)
	d.SetFunctionName(noop, "-")

	listen := config.GetListenAddress()
	logger.Info("pike is starting",
		zap.String("listen", listen),
	)
	err = d.ListenAndServe(listen)
	if err != nil {
		panic(err)
	}
}
