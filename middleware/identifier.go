package custommiddleware

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/util"
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
			c.Logger().Debug("identifier middleware")
			if config.Skipper(c) {
				return next(c)
			}
			pc := c.(*Context)
			serverTiming := pc.serverTiming
			req := pc.Request()
			method := req.Method
			// 只有get与head请求可缓存
			if method != echo.GET && method != echo.HEAD {
				pc.status = cache.Pass
				return next(pc)
			}
			key := util.GetIdentity(req)
			serverTiming.GetRequestStatusStart()
			status, ch := client.GetRequestStatus(key)
			serverTiming.GetRequestStatusEnd()
			// TODO是否应该增加超时机制
			if ch != nil {
				status = <-ch
			}
			pc.status = status
			pc.identity = key
			return next(pc)
		}
	}
}
