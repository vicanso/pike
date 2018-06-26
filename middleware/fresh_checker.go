package custommiddleware

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/vicanso/fresh"
	"github.com/vicanso/pike/vars"
)

type (
	// FreshCheckerConfig freshChecker配置
	FreshCheckerConfig struct {
		Skipper middleware.Skipper
	}
)

// FreshChecker 判断请求是否fresh(304)
func FreshChecker(config FreshCheckerConfig) echo.MiddlewareFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = middleware.DefaultSkipper
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Logger().Debug("fresh checker middleware")
			if config.Skipper(c) {
				return next(c)
			}
			pc := c.(*Context)
			done := pc.serverTiming.Start(ServerTimingFreshChecker)
			cr := pc.resp
			if cr == nil {
				done()
				return vars.ErrResponseNotSet
			}
			statusCode := int(cr.StatusCode)
			method := c.Request().Method
			pc.fresh = false
			if method != echo.GET && method != echo.HEAD {
				done()
				return next(pc)
			}
			if statusCode < http.StatusOK || statusCode >= http.StatusBadRequest {
				done()
				return next(c)
			}

			reqHeader := pc.Request().Header
			resHeader := pc.Response().Header()

			ifModifiedSince := []byte(reqHeader.Get(echo.HeaderIfModifiedSince))
			ifNoneMatch := []byte(reqHeader.Get(vars.IfNoneMatch))
			cacheControl := []byte(reqHeader.Get(vars.CacheControl))
			eTag := []byte(resHeader.Get(vars.ETag))
			lastModified := []byte(resHeader.Get(echo.HeaderLastModified))

			// 如果请求还是fresh，则后续处理可返回304
			if fresh.Check(ifModifiedSince, ifNoneMatch, cacheControl, lastModified, eTag) {
				pc.fresh = true
			}

			done()
			return next(pc)
		}
	}
}
