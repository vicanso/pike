package middleware

import (
	"net/http"
	"sync/atomic"

	"github.com/vicanso/pike/pike"
)

var (
	pong = []byte("pong")
)

type (
	// PingConfig ping的配置
	PingConfig struct {
		URL          string
		DisabledPing *int32
	}
)

// Ping ping middleware function
func Ping(config PingConfig) pike.Middleware {
	url := config.URL
	return func(c *pike.Context, next pike.Next) error {
		if c.Request.RequestURI == url {
			// 0表示非禁用，非0表示禁用
			v := atomic.LoadInt32(config.DisabledPing)
			if v != 0 {
				return pike.ErrDisableServer
			}
			c.Response.WriteHeader(http.StatusOK)
			c.Response.Write(pong)
			return nil
		}
		return next()
	}
}
