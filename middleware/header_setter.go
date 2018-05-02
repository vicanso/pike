package custommiddleware

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/vicanso/dash"
	"github.com/vicanso/pike/cache"
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
			if config.Skipper(c) {
				return next(c)
			}
			cr, ok := c.Get(vars.Response).(*cache.Response)
			if !ok {
				return vars.ErrResponseNotSet
			}
			h := c.Response().Header()
			for k, values := range cr.Header {
				if dash.IncludesString(ignoreHeaderKeys, k) {
					continue
				}
				for _, v := range values {
					h.Add(k, v)
				}
			}
			return next(c)
		}
	}
}
