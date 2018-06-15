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
			util.AddStartTiming(c)
			if config.Skipper(c) {
				return next(c)
			}
			rid := c.Get(vars.RID).(string)
			timing, _ := c.Get(vars.Timing).(*servertiming.Header)
			req := c.Request()
			method := req.Method
			debug := c.Logger().Debug
			// 只有get与head请求可缓存
			if method != echo.GET && method != echo.HEAD {
				c.Set(vars.Status, cache.Pass)
				debug(rid, " is pass")
				return next(c)
			}
			key := []byte(method + " " + req.Host + " " + req.RequestURI)
			var m *servertiming.Metric
			if timing != nil {
				m = timing.NewMetric(vars.GetRequestStatusMetric)
				m.WithDesc("get request status").Start()
			}
			status, ch := client.GetRequestStatus(key)
			if m != nil {
				m.Stop()
				m = nil
			}
			if ch != nil {
				debug(rid, " is waitting")
				if timing != nil {
					m = timing.NewMetric(vars.WaitForRequestStatusMetric)
					m.WithDesc("wait for request status").Start()
				}
				// TODO 是否需要增加超时处理
				status = <-ch
				if m != nil {
					m.Stop()
				}
			}
			c.Set(vars.Status, status)
			c.Set(vars.Identity, key)
			debug(rid, " status:", status)
			return next(c)
		}
	}
}
