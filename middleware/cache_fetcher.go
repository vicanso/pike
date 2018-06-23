package custommiddleware

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/vars"
)

type (
	// CacheFetcherConfig cache fetcher配置
	CacheFetcherConfig struct {
		Skipper middleware.Skipper
	}
)

// CacheFetcher 从缓存中获取数据
func CacheFetcher(config CacheFetcherConfig, client *cache.Client) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = middleware.DefaultSkipper
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Logger().Debug("cache fetcher middleware")
			if config.Skipper(c) {
				return next(c)
			}
			pc := c.(*Context)
			// status, ok := c.Get(vars.Status).(int)
			status := pc.status
			if status == 0 {
				return vars.ErrRequestStatusNotSet
			}
			// 如果非cache的
			if status != cache.Cacheable {
				return next(pc)
			}
			identity := pc.identity
			if identity == nil {
				return vars.ErrIdentityNotSet
			}
			serverTiming := pc.serverTiming
			serverTiming.CacheFetchStart()
			resp, err := client.GetResponse(identity)
			serverTiming.CacheFetchEnd()
			if err != nil {
				return err
			}
			pc.resp = resp
			return next(pc)
		}
	}
}
