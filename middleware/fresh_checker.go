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
			if config.Skipper(c) {
				return next(c)
			}
			pc := c.(*Context)
			cr := pc.resp
			if cr == nil {
				return vars.ErrResponseNotSet
			}
			statusCode := int(cr.StatusCode)
			method := c.Request().Method
			pc.fresh = false
			if method != echo.GET && method != echo.HEAD {
				return next(pc)
			}
			if statusCode < http.StatusOK || statusCode >= http.StatusBadRequest {
				return next(c)
			}

			reqHeader := pc.Request().Header
			resHeader := pc.Response().Header()

			ifModifiedSince := reqHeader.Get(echo.HeaderIfModifiedSince)
			ifNoneMatch := reqHeader.Get(vars.IfNoneMatch)
			cacheControl := reqHeader.Get(vars.CacheControl)
			reqHeaderData := &fresh.RequestHeader{
				IfModifiedSince: []byte(ifModifiedSince),
				IfNoneMatch:     []byte(ifNoneMatch),
				CacheControl:    []byte(cacheControl),
			}
			eTag := resHeader.Get(vars.ETag)
			lastModified := resHeader.Get(echo.HeaderLastModified)
			resHeaderData := &fresh.ResponseHeader{
				ETag:         []byte(eTag),
				LastModified: []byte(lastModified),
			}

			// 如果请求还是fresh，则后续处理可返回304
			if fresh.Fresh(reqHeaderData, resHeaderData) {
				pc.fresh = true
			}
			return next(pc)
		}
	}
}
