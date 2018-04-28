package custommiddleware

import (
	"github.com/labstack/echo"
	"github.com/mitchellh/go-server-timing"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/vars"
)

// CacheFetcher 从缓存中获取数据
func CacheFetcher(client *cache.Client) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			status, ok := c.Get(vars.Status).(int)
			if !ok {
				return vars.ErrRequestStatusNotSet
			}
			// 如果非cache的
			if status != cache.Cacheable {
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
				return err
			}
			c.Set(vars.Response, resp)
			return next(c)
		}
	}
}
