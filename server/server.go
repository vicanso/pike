package server

import (
	"bytes"
	"net"
	"strconv"
	"sync/atomic"

	"github.com/vicanso/cod"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/config"
	"github.com/vicanso/pike/df"
	"github.com/vicanso/pike/middleware"
	"github.com/vicanso/pike/upstream"

	compress "github.com/vicanso/cod-compress"
	etag "github.com/vicanso/cod-etag"
	fresh "github.com/vicanso/cod-fresh"
	recover "github.com/vicanso/cod-recover"
)

// New create a cod server
func New(director *upstream.Director, dsp *cache.Dispatcher) *cod.Cod {
	d := cod.New()
	d.EnableTrace = config.IsEnableServerTiming()
	// 如果启动 server timing
	// 则在回调中调整响应头
	if d.EnableTrace {
		prefix := df.APP + "-"
		d.OnTrace(func(c *cod.Context, traceInfos cod.TraceInfos) {
			c.SetHeader(cod.HeaderServerTiming, string(traceInfos.ServerTiming(prefix)))
		})
	}

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

	fn = middleware.NewCacheIdentifier(dsp)
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

func newTestServer() (ln net.Listener, err error) {
	ln, err = net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		return
	}

	d := cod.New()

	inc := func(p *int32) *bytes.Buffer {
		v := atomic.AddInt32(p, 1)
		return bytes.NewBufferString(strconv.Itoa(int(v)))
	}

	genBuffer := func(size int) *bytes.Buffer {
		buf := new(bytes.Buffer)
		for i := 0; i < size; i++ {
			buf.WriteString("a")
		}
		return buf
	}

	// 响应未压缩
	notCompressHandler := func(c *cod.Context) error {
		c.SetHeader("Content-Type", "text/html")
		c.BodyBuffer = genBuffer(4096)
		return nil
	}
	// 响应数据已压缩
	compressHandler := func(c *cod.Context) error {
		c.SetHeader("Content-Type", "text/html")
		c.SetHeader("Content-Encoding", "gzip")
		buf, _ := cache.Gzip(genBuffer(4096).Bytes())

		c.BodyBuffer = bytes.NewBuffer(buf)
		return nil
	}

	setCacheNext := func(c *cod.Context) error {
		c.CacheMaxAge("10s")
		return c.Next()
	}

	d.GET("/ping", func(c *cod.Context) error {
		c.BodyBuffer = bytes.NewBufferString("pong")
		return nil
	})

	var postResponseID int32
	d.POST("/post", func(c *cod.Context) error {
		c.BodyBuffer = inc(&postResponseID)
		return nil
	})

	// 非文本类数据
	d.GET("/image-cache", func(c *cod.Context) error {
		c.CacheMaxAge("10s")
		c.SetHeader("Content-Type", "image/png")
		c.BodyBuffer = genBuffer(4096)
		return nil
	})

	d.POST("/post-not-compress", notCompressHandler)
	d.GET("/get-not-compress", notCompressHandler)

	d.POST("/post-compress", compressHandler)
	d.POST("/get-compress", compressHandler)

	d.GET("/get-cache-not-compress", setCacheNext, notCompressHandler)
	d.GET("/get-cache-compress", setCacheNext, compressHandler)

	d.GET("/get-without-etag", notCompressHandler)

	d.GET("/get-with-etag", func(c *cod.Context) error {
		c.SetHeader("ETag", `"123"`)
		return notCompressHandler(c)
	})

	var noCacheResponseID int32
	d.GET("/no-cache", func(c *cod.Context) error {
		c.BodyBuffer = inc(&noCacheResponseID)
		return nil
	})

	go d.Serve(ln)

	return
}
