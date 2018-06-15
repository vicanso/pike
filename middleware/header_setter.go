package custommiddleware

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	servertiming "github.com/mitchellh/go-server-timing"
	funk "github.com/thoas/go-funk"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/util"
	"github.com/vicanso/pike/vars"
)

var (
	ignoreHeaderKeys = []string{
		"Date",
		"Connection",
		"Server",
	}
)

type (
	// HeaderSetterConfig header setter的配置
	HeaderSetterConfig struct {
		Skipper middleware.Skipper
	}
)

// HeaderSetter 设置响应头
func HeaderSetter(config HeaderSetterConfig) echo.MiddlewareFunc {
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
			debug := c.Logger().Debug
			cr, ok := c.Get(vars.Response).(*cache.Response)
			if !ok {
				debug(rid, " response not set")
				return vars.ErrResponseNotSet
			}
			timing, _ := c.Get(vars.Timing).(*servertiming.Header)
			var m *servertiming.Metric
			if timing != nil {
				m = timing.NewMetric(vars.HeaderSetterMetric)
				m.WithDesc("header setter").Start()
			}
			h := c.Response().Header()
			for k, values := range cr.Header {
				if funk.ContainsString(ignoreHeaderKeys, k) {
					continue
				}
				for _, v := range values {
					h.Add(k, v)
				}
			}
			if m != nil {
				m.Stop()
			}
			debug(rid, " set header done")
			return next(c)
		}
	}
}
