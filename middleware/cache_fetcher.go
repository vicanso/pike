package custommiddleware

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/util"
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
			if config.Skipper(c) {
				return next(c)
			}
			status, ok := c.Get(vars.Status).(int)
			done := util.CreateTiming(c, vars.MetricCacheFetcher)
			if !ok {
				done()
				return vars.ErrRequestStatusNotSet
			}
			rid := c.Get(vars.RID).(string)
			debug := c.Logger().Debug
			// 如果非cache的
			if status != cache.Cacheable {
				debug(rid, " pass cache fetcher")
				done()
				return next(c)
			}
			identity, ok := c.Get(vars.Identity).([]byte)
			if !ok {
				done()
				return vars.ErrIdentityStatusNotSet
			}
			resp, err := client.GetResponse(identity)
			if err != nil {
				debug(rid, " get cache response fail")
				done()
				return err
			}
			c.Set(vars.Response, resp)
			debug(rid, " get from cache")
			done()
			return next(c)
		}
	}
}
