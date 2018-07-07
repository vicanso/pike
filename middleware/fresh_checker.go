package middleware

import (
	"net/http"

	"github.com/vicanso/pike/pike"

	"github.com/vicanso/fresh"
)

type (
	// FreshCheckerConfig freshChecker配置
	FreshCheckerConfig struct {
	}
)

// FreshChecker 判断请求是否fresh(304)
func FreshChecker(config FreshCheckerConfig) pike.Middleware {

	return func(c *pike.Context, next pike.Next) error {
		done := c.ServerTiming.Start(pike.ServerTimingFreshChecker)
		cr := c.Resp
		if cr == nil {
			done()
			return ErrResponseNotSet
		}
		statusCode := int(cr.StatusCode)
		method := c.Request.Method
		c.Fresh = false
		if method != http.MethodGet && method != http.MethodHead {
			done()
			return next()
		}
		if statusCode < http.StatusOK || statusCode >= http.StatusBadRequest {
			done()
			return next()
		}

		reqHeader := c.Request.Header
		resHeader := c.Response.Header()

		ifModifiedSince := []byte(reqHeader.Get(pike.HeaderIfModifiedSince))
		ifNoneMatch := []byte(reqHeader.Get(pike.HeaderIfNoneMatch))
		cacheControl := []byte(reqHeader.Get(pike.HeaderCacheControl))
		eTag := []byte(resHeader.Get(pike.HeaderETag))
		lastModified := []byte(resHeader.Get(pike.HeaderLastModified))

		// 如果请求还是fresh，则后续处理可返回304
		if fresh.Check(ifModifiedSince, ifNoneMatch, cacheControl, lastModified, eTag) {
			c.Fresh = true
		}

		done()
		return next()
	}
}
