package custommiddleware

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/vicanso/pike/proxy"
	"github.com/vicanso/pike/vars"
)

type (
	// DirectorPickerConfig director配置
	DirectorPickerConfig struct {
		Skipper middleware.Skipper
	}
)

// DirectorPicker 根据请求的参数获取相应的director
// 判断director是否符合是顺序查询，因此需要将directors先根据优先级排好序
func DirectorPicker(config DirectorPickerConfig, directors proxy.Directors) echo.MiddlewareFunc {
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
			done := pc.serverTiming.Start(ServerTimingDirectorPicker)
			req := pc.Request()
			host := req.Host
			uri := req.RequestURI
			found := false

			for _, d := range directors {
				if d.Match(host, uri) {
					pc.director = d
					found = true
					break
				}
			}
			if !found {
				done()
				return vars.ErrDirectorNotFound
			}
			done()
			return next(pc)
		}
	}
}
