package middleware

import (
	funk "github.com/thoas/go-funk"
	"github.com/vicanso/pike/pike"
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
	}
)

// HeaderSetter 设置响应头
func HeaderSetter(config HeaderSetterConfig) pike.Middleware {
	return func(c *pike.Context, next pike.Next) error {
		done := c.ServerTiming.Start(pike.ServerTimingHeaderSetter)
		cr := c.Resp
		if cr == nil {
			done()
			return ErrResponseNotSet
		}
		h := c.Response.Header()
		for k, values := range cr.Header {
			if funk.ContainsString(ignoreHeaderKeys, k) {
				continue
			}
			for _, v := range values {
				h.Add(k, v)
			}
		}
		done()
		return next()
	}
}
