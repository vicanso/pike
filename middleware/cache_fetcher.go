package custommiddleware

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/mitchellh/go-server-timing"
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
			util.AddStartTiming(c)
			if config.Skipper(c) {
				return next(c)
			}
			status, ok := c.Get(vars.Status).(int)
			if !ok {
				return vars.ErrRequestStatusNotSet
			}
			rid := c.Get(vars.RID).(string)
			debug := c.Logger().Debug
			// 如果非cache的
			if status != cache.Cacheable {
				debug(rid, " pass cache fetcher")
				return next(c)
			}
			identity, ok := c.Get(vars.Identity).([]byte)
			if !ok {
				return vars.ErrIdentityStatusNotSet
			}
			timing, _ := c.Get(vars.Timing).(*servertiming.Header)
			var m *servertiming.Metric
			if timing != nil {
				m = timing.NewMetric(vars.GetResponseFromCacheMetric)
				m.WithDesc("get response from cache").Start()
			}
			resp, err := client.GetResponse(identity)
			if m != nil {
				m.Stop()
			}
			if err != nil {
				debug(rid, " get cache response fail")
				return err
			}
			c.Set(vars.Response, resp)
			debug(rid, " get from cache")
			return next(c)
		}
	}
}
