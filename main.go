package main

import (
	"github.com/vicanso/cod"

	"github.com/vicanso/hes"
	"github.com/vicanso/pike/config"
	"github.com/vicanso/pike/server"
	"github.com/vicanso/pike/upstream"
	"github.com/vicanso/pike/util"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	c := zap.NewProductionConfig()
	c.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	// 只针对panic 以上的日志增加stack trace
	logger, err := c.Build(zap.AddStacktrace(zap.DPanicLevel))
	if err != nil {
		panic(err)
	}
	us := upstream.NewUpstreamsFromConfig()
	us.StartHealthCheck()

	d := server.New(us)

	// 出错时输出相关出错日志
	d.OnError(func(c *cod.Context, err error) {
		he := hes.Wrap(err)
		logger.Error(he.Message,
			zap.String("category", he.Category),
			zap.String("method", c.Request.Method),
			zap.String("host", c.Request.Host),
			zap.String("url", c.Request.RequestURI),
			zap.Strings("stack", util.GetStack(4, 9)),
		)
	})

	listen := config.GetListenAddress()
	logger.Info("pike is starting",
		zap.String("listen", listen),
	)
	err = d.ListenAndServe(listen)
	if err != nil {
		panic(err)
	}
}
