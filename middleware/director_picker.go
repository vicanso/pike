package customMiddleware

import (
	"github.com/labstack/echo"

	"github.com/vicanso/pike/proxy"
	"github.com/vicanso/pike/vars"
)

// DirectorPicker 根据请求的参数获取相应的director
// 判断director是否符合是顺序查询，因此需要将directors先根据优先级排好序
func DirectorPicker(directors proxy.Directors) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			host := req.Host
			uri := req.RequestURI
			found := false
			for _, d := range directors {
				if d.Match(host, uri) {
					c.Set(vars.Director, d)
					found = true
					break
				}
			}
			if !found {
				return vars.ErrDirectorNotFound
			}
			return next(c)
		}
	}
}
