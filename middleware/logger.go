package custommiddleware

import (
	"time"

	"github.com/labstack/echo"
	"github.com/vicanso/pike/httplog"
)

type (
	// LoggerConfig logger配置
	LoggerConfig struct {
		Writer    httplog.Writer
		LogFormat string
	}
)

// Logger logger中间件
func Logger(config LoggerConfig) echo.MiddlewareFunc {
	writer := config.Writer
	tags := httplog.Parse([]byte(config.LogFormat))
	enabledLogger := writer != nil && len(tags) != 0
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if !enabledLogger {
				return next(c)
			}
			startedAt := time.Now()
			err = next(c)
			go func() {
				str := httplog.Format(c, tags, startedAt)
				writer.Write([]byte(str))
			}()
			return
		}
	}
}
