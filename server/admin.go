package server

import (
	"github.com/vicanso/cod"
	basicauth "github.com/vicanso/cod-basic-auth"
	errorhandler "github.com/vicanso/cod-error-handler"
	recover "github.com/vicanso/cod-recover"
	responder "github.com/vicanso/cod-responder"

	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/stats"
	"github.com/vicanso/pike/upstream"
)

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

	d.AddGroup(g)
	return d
}
