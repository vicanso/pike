package middleware

import (
	"time"

	"github.com/vicanso/pike/pike"

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
func Logger(config LoggerConfig) pike.Middleware {
	writer := config.Writer
	tags := httplog.Parse([]byte(config.LogFormat))
	enabledLogger := writer != nil && len(tags) != 0
	return func(c *pike.Context, next pike.Next) (err error) {
		if !enabledLogger {
			return next()
		}
		startedAt := time.Now()
		err = next()
		str := httplog.Format(c, tags, startedAt)
		go func() {
			writer.Write([]byte(str))
		}()
		return
	}
}
