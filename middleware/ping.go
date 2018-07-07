package middleware

import (
	"net/http"

	"github.com/vicanso/pike/pike"
)

var (
	pong = []byte("pong")
)

type (
	// PingConfig ping的配置
	PingConfig struct {
		URL string
	}
)

// Ping ping middleware function
func Ping(config PingConfig) pike.Middleware {
	url := config.URL
	return func(c *pike.Context, next pike.Next) error {
		if c.Request.RequestURI == url {
			c.Response.WriteHeader(http.StatusOK)
			c.Response.Write(pong)
			return nil
		}
		return next()
	}
}
