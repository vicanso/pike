package custommiddleware

import (
	"github.com/labstack/echo"
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

// HeaderSetter 设置响应头
func HeaderSetter() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
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
