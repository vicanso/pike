package main

import (
	"github.com/vicanso/cod"

	recover "github.com/vicanso/cod-recover"
	"github.com/vicanso/hes"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/config"
	"github.com/vicanso/pike/df"
	"github.com/vicanso/pike/log"
	"github.com/vicanso/pike/server"
	"github.com/vicanso/pike/upstream"
	"github.com/vicanso/pike/util"
	"go.uber.org/zap"
)

var (
	// BuildedAt application builded at ???
	BuildedAt = ""
	// CommitID git commit id
	CommitID = ""
)

func init() {
	df.BuildedAt = BuildedAt
	df.CommitID = CommitID
}

func main() {
	logger := log.Default()
	director := &upstream.Director{}
	director.Fetch()
	director.StartHealthCheck()
	dsp := cache.NewDispatcher(cache.GetOptionsFromConfig())
	d := server.New(director, dsp)

	// 出错时输出相关出错日志
	d.OnError(func(c *cod.Context, err error) {
		he := hes.Wrap(err)
		// 如果是recover，记录统计
		if he.Category == recover.ErrCategory {

		}
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
	err := d.ListenAndServe(listen)
	if err != nil {
		panic(err)
	}
}
