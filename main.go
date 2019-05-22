package main

import (
	"net"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/vicanso/cod"

	recover "github.com/vicanso/cod-recover"
	"github.com/vicanso/hes"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/config"
	"github.com/vicanso/pike/df"
	"github.com/vicanso/pike/log"
	"github.com/vicanso/pike/server"
	"github.com/vicanso/pike/stats"
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

func getListen() string {
	v := os.Getenv("LISTEN")
	if v == "" {
		v = ":3015"
	}
	return v
}

func main() {
	logger := log.Default()

	insStats := stats.New()
	cfg := config.New()
	err := cfg.Fetch()
	if err != nil {
		panic(err)
	}
	textFilter := regexp.MustCompile(cfg.GetTextFilter())
	dsp := cache.NewDispatcher(cache.Options{
		Size:              cfg.GetCacheZoneSize(),
		ZoneSize:          cfg.GetCacheZoneSize(),
		CompressLevel:     cfg.GetCompressLevel(),
		CompressMinLength: cfg.GetCompressMinLength(),
		HitForPassTTL:     cfg.GetHitForPassTTL(),
		TextFilter:        textFilter,
	})
	director := &upstream.Director{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   cfg.GetConnectTimeout(),
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       cfg.GetIdleConnTimeout(),
			TLSHandshakeTimeout:   cfg.GetTLSHandshakeTimeout(),
			ExpectContinueTimeout: cfg.GetExpectContinueTimeout(),
			ResponseHeaderTimeout: cfg.GetResponseHeaderTimeout(),
		},
	}
	directorConfig := config.NewFileConfig("director")
	err = directorConfig.Fetch()
	if err != nil {
		panic(err)
	}
	director.SetBackends(directorConfig.GetBackends())

	director.StartHealthCheck()

	d := server.New(server.Options{
		Config:         cfg,
		DirectorConfig: directorConfig,
		Director:       director,
		Dispatcher:     dsp,
		Stats:          insStats,
	})

	// 出错时输出相关出错日志
	d.OnError(func(c *cod.Context, err error) {
		he := hes.Wrap(err)
		// 如果是recover，记录统计
		if he.Category == recover.ErrCategory {
			insStats.IncreaseRecoverCount()
		}
		logger.Error(he.Message,
			zap.String("category", he.Category),
			zap.String("method", c.Request.Method),
			zap.String("host", c.Request.Host),
			zap.String("url", c.Request.RequestURI),
			zap.Strings("stack", util.GetStack(4, 9)),
		)
	})

	listen := getListen()

	logger.Info("pike is starting",
		zap.String("listen", listen),
	)
	err = d.ListenAndServe(listen)
	if err != nil {
		panic(err)
	}
}
