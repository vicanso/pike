package custommiddleware

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/vicanso/fresh"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/vars"
)

// FreshChecker 判断请求是否fresh(304)
func FreshChecker() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cr, ok := c.Get(vars.Response).(*cache.Response)
			if !ok {
				return vars.ErrResponseNotSet
			}
			statusCode := int(cr.StatusCode)
			method := c.Request().Method
			c.Set(vars.Fresh, false)
			if method != echo.GET && method != echo.HEAD {
				return next(c)
			}
			if statusCode < http.StatusOK || statusCode >= http.StatusBadRequest {
				return next(c)
			}
			reqHeader := c.Request().Header
			resHeader := c.Response().Header()

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
				c.Set(vars.Fresh, true)
			}
			return next(c)
		}
	}
}
