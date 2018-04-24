package customMiddleware

import (
	"github.com/labstack/echo"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/vars"
)

// CacheFetcher 从缓存中获取数据
func CacheFetcher(client *cache.Client) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			status := c.Get(vars.Status)
			if status == nil {
				return vars.ErrRequestStatusNotSet
			}
			// 如果非cache的
			if status.(int) != cache.Cacheable {
				return next(c)
			}
			identity := c.Get(vars.Identity)
			resp, err := client.GetResponse(identity.([]byte))
			if err != nil {
				return err
			}
			c.Set(vars.Response, resp)
			return next(c)
		}
	}
}
