package server

import (
	"bytes"
	"os"

	"github.com/gobuffalo/packr/v2"
	"github.com/vicanso/cod"
	basicauth "github.com/vicanso/cod-basic-auth"
	errorhandler "github.com/vicanso/cod-error-handler"
	recover "github.com/vicanso/cod-recover"
	responder "github.com/vicanso/cod-responder"
	staticServe "github.com/vicanso/cod-static-serve"

	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/stats"
	"github.com/vicanso/pike/upstream"
)

type (
	staticFile struct {
		box *packr.Box
	}
)

var (
	box = packr.New("asset", "../web/build")
)

func (sf *staticFile) Exists(file string) bool {
	return sf.box.Has(file)
}
func (sf *staticFile) Get(file string) ([]byte, error) {
	return sf.box.Find(file)
}
func (sf *staticFile) Stat(file string) os.FileInfo {
	return nil
}
func sendFile(c *cod.Context, file string) (err error) {
	buf, err := box.Find(file)
	if err != nil {
		return
	}
	c.SetContentTypeByExt(file)
	c.BodyBuffer = bytes.NewBuffer(buf)
	return
}

// NewAdminServer create an admin server
func NewAdminServer(opts Options) *cod.Cod {
	cfg := opts.Config
	insStats := opts.Stats
	director := opts.Director
	dsp := opts.Dispatcher

	d := cod.New()
	d.Use(recover.New())
	d.Use(errorhandler.NewDefault())
	d.Use(responder.NewDefault())

	adminHandlerList := make([]cod.Handler, 0)

	adminUser := cfg.GetAdminUser()
	adminPwd := cfg.GetAdminPassword()
	// 设置 basic auth 认证
	if adminUser != "" && adminPwd != "" {
		adminHandlerList = append(adminHandlerList, basicauth.New(basicauth.Config{
			Validate: func(account, pwd string, _ *cod.Context) (bool, error) {
				if account == adminUser && pwd == adminPwd {
					return true, nil
				}
				return false, nil
			},
		}))
	}

	g := cod.NewGroup("", adminHandlerList...)
	// 获取系统状态统计
	g.GET("/stats", func(c *cod.Context) error {
		c.Body = &struct {
			Stats *stats.Info `json:"stats,omitempty"`
		}{
			insStats.GetInfo(),
		}
		return nil
	})
	// 获取 upstream 列表
	g.GET("/upstreams", func(c *cod.Context) error {
		c.Body = &struct {
			Upstreams []upstream.Info `json:"upstreams,omitempty"`
		}{
			director.GetUpstreamInfos(),
		}
		return nil
	})

	// 获取缓存列表
	g.GET("/caches", func(c *cod.Context) error {
		c.Body = &struct {
			Caches []*cache.HTTPCacheInfo `json:"caches,omitempty"`
		}{
			dsp.GetCacheList(),
		}
		return nil
	})

	// 删除缓存
	g.DELETE("/caches", func(c *cod.Context) error {
		dsp.Expire([]byte(c.QueryParam("key")))
		c.NoContent()
		return nil
	})

	// 获取配置列表
	g.GET("/configs", func(c *cod.Context) (err error) {
		basicYaml, err := opts.Config.ToYAML()
		if err != nil {
			return
		}
		directorYaml, err := opts.DirectorConfig.ToYAML()
		if err != nil {
			return
		}
		c.Body = &struct {
			Basic        map[string]interface{} `json:"basic,omitempty"`
			BasicYaml    string                 `json:"basicYaml,omitempty"`
			Director     map[string]interface{} `json:"director,omitempty"`
			DirectorYaml string                 `json:"directorYaml,omitempty"`
		}{
			opts.Config.Viper.AllSettings(),
			string(basicYaml),
			opts.DirectorConfig.Viper.AllSettings(),
			string(directorYaml),
		}
		return
	})

	sf := &staticFile{
		box: box,
	}
	// 静态文件

	g.GET("/", func(c *cod.Context) error {
		c.CacheMaxAge("10s")
		return sendFile(c, "index.html")
	})
	g.GET("/static/*file", staticServe.New(sf, staticServe.Config{
		Path: "/static",
		// 客户端缓存一年
		MaxAge: 365 * 24 * 3600,
		// 缓存服务器缓存一个小时
		SMaxAge:             60 * 60,
		DenyQueryString:     true,
		DisableLastModified: true,
	}))

	d.AddGroup(g)
	return d
}
