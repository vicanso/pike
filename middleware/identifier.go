package custommiddleware

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/util"
	"github.com/vicanso/pike/vars"
)

type (
	// IdentifierConfig 定义配置
	IdentifierConfig struct {
		Skipper middleware.Skipper
	}
)

// Identifier 对请求的参数校验，生成各类状态值
/*
- 判断请求状态，生成status
- 对于状态非Pass的请求，根据request url 生成identity
*/
func Identifier(config IdentifierConfig, client *cache.Client) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = middleware.DefaultSkipper
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}
			rid := c.Get(vars.RID).(string)
			done := util.CreateTiming(c, vars.MetricIdentifier)
			// timing, _ := c.Get(vars.Timing).(*servertiming.Header)
			req := c.Request()
			method := req.Method
			debug := c.Logger().Debug
			// 只有get与head请求可缓存
			if method != echo.GET && method != echo.HEAD {
				c.Set(vars.Status, cache.Pass)
				debug(rid, " is pass")
				done()
				return next(c)
			}
			key := []byte(method + " " + req.Host + " " + req.RequestURI)
			status, ch := client.GetRequestStatus(key)
			if ch != nil {
				status = <-ch
			}
			c.Set(vars.Status, status)
			c.Set(vars.Identity, key)
			debug(rid, " status:", status)
			done()
			return next(c)
		}
	}
}
