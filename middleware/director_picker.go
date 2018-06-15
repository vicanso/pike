package custommiddleware

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	servertiming "github.com/mitchellh/go-server-timing"

	"github.com/vicanso/pike/proxy"
	"github.com/vicanso/pike/util"
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
			util.AddStartTiming(c)
			if config.Skipper(c) {
				return next(c)
			}
			req := c.Request()
			host := req.Host
			uri := req.RequestURI
			found := false
			timing, _ := c.Get(vars.Timing).(*servertiming.Header)
			var m *servertiming.Metric
			if timing != nil {
				m = timing.NewMetric(vars.GetMatchDirectorMetric)
				m.WithDesc("get match director").Start()
			}

			var director *proxy.Director
			for _, d := range directors {
				if d.Match(host, uri) {
					c.Set(vars.Director, d)
					director = d
					found = true
					break
				}
			}
			if m != nil {
				m.Stop()
			}
			rid := c.Get(vars.RID).(string)
			debug := c.Logger().Debug
			if !found {
				debug(rid, " the director is not found")
				return vars.ErrDirectorNotFound
			}
			debug(rid, " the director is ", director.Name)
			return next(c)
		}
	}
}
