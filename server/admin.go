package server

import (
	"github.com/vicanso/cod"
	basicauth "github.com/vicanso/cod-basic-auth"
	recover "github.com/vicanso/cod-recover"
	responder "github.com/vicanso/cod-responder"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/config"
	"github.com/vicanso/pike/stats"
	"github.com/vicanso/pike/upstream"
)

// NewAdminServer create an admin server
func NewAdminServer(cfg *config.Config, director *upstream.Director, dsp *cache.Dispatcher, insStats *stats.Stats) *cod.Cod {
	d := cod.New()
	d.Use(recover.New())
	d.Use(responder.NewDefault())

	adminHandlerList := make([]cod.Handler, 0)

	adminUser := cfg.GetAdminUser()
	adminPwd := cfg.GetAdminPassword()
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
	g.GET("/stats", func(c *cod.Context) error {
		c.Body = &struct {
			Stats  *stats.Info            `json:"stats,omitempty"`
			Caches []*cache.HTTPCacheInfo `json:"caches,omitempty"`
		}{
			insStats.GetInfo(),
			dsp.GetCacheList(),
		}
		return nil
	})
	g.GET("/upstreams", func(c *cod.Context) error {
		c.Body = &struct {
			Upstreams []upstream.Info `json:"upstreams,omitempty"`
		}{
			director.GetUpstreamInfos(),
		}
		return nil
	})
	d.AddGroup(g)
	return d
}
