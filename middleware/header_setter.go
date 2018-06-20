package custommiddleware

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	funk "github.com/thoas/go-funk"
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
			pc := c.(*Context)
			cr := pc.resp
			if cr == nil {
				return vars.ErrResponseNotSet
			}
			h := pc.Response().Header()
			for k, values := range cr.Header {
				if funk.ContainsString(ignoreHeaderKeys, k) {
					continue
				}
				for _, v := range values {
					h.Add(k, v)
				}
			}
			return next(pc)
		}
	}
}
