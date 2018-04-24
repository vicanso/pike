package customMiddleware

import (
	"github.com/labstack/echo"

	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/vars"
)

// Identifier 对请求的参数校验，生成各类状态值
/*
- 判断请求状态，生成status
- 对于状态非Pass的请求，根据request url 生成identity
*/
func Identifier(client *cache.Client) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			method := req.Method
			// 只有get与head请求可缓存
			if method != echo.GET && method != echo.HEAD {
				c.Set(vars.Status, cache.Pass)
				return next(c)
			}
			key := []byte(method + " " + req.Host + " " + req.RequestURI)
			status, ch := client.GetRequestStatus(key)
			if ch != nil {
				// TODO 是否需要增加超时处理
				status = <-ch
			}
			c.Set(vars.Status, status)
			c.Set(vars.Identity, key)
			c.Logger().Debug(req.RequestURI, " status:", status)
			return next(c)
		}
	}
}
