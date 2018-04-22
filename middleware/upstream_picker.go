package customMiddleware

import (
	"github.com/labstack/echo"
	"github.com/vicanso/pike/config"
)

// UpstreamPicker 根据请求的参数获取相应的upstream
func UpstreamPicker(directors []*config.Director) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return next(c)
		}
	}
}
