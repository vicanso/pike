package middleware

import (
	"strings"

	"github.com/vicanso/cod"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/df"

	"github.com/vicanso/pike/upstream"
)

var (
	// 需要清除的header
	clearHeaders = []string{
		"Date",
		"Connection",
		cod.HeaderContentLength,
	}
)

// NewProxy create a proxy middleware
func NewProxy(director *upstream.Director) cod.Handler {
	return func(c *cod.Context) (err error) {
		// 如果请求是从缓存读取Cacheable ，则直接跳过
		status, ok := c.Get(df.Status).(int)
		if ok && status == cache.Cacheable {
			return c.Next()
		}
		originalNext := c.Next
		c.Next = func() error {
			c.Next = originalNext
			return c.Next()
		}

		var ifModifiedSince, ifNoneMatch, acceptEncoding string
		reqHeader := c.Request.Header
		// proxy时为了避免304的出现，因此调用时临时删除header
		// 如果是 pass 则无需删除，其它的需要删除（因为此时无法确认数据是否可缓存）
		if status != cache.Pass &&
			status != cache.HitForPass {
			acceptEncoding = reqHeader.Get(cod.HeaderAcceptEncoding)
			ifModifiedSince = reqHeader.Get(cod.HeaderIfModifiedSince)
			ifNoneMatch = reqHeader.Get(cod.HeaderIfNoneMatch)
			if ifModifiedSince != "" {
				reqHeader.Del(cod.HeaderIfModifiedSince)
			}
			if ifNoneMatch != "" {
				reqHeader.Del(cod.HeaderIfNoneMatch)
			}

			if strings.Contains(acceptEncoding, df.GZIP) {
				reqHeader.Set(cod.HeaderAcceptEncoding, df.GZIP)
			} else {
				reqHeader.Del(cod.HeaderAcceptEncoding)
			}
		}

		err = director.Proxy(c)
		callback := c.Get(df.ProxyDoneCallback)
		if callback != nil {
			fn, _ := callback.(func())
			if fn != nil {
				fn()
			}
		}
		if acceptEncoding != "" {
			reqHeader.Set(cod.HeaderAcceptEncoding, acceptEncoding)
		}
		if ifModifiedSince != "" {
			reqHeader.Set(cod.HeaderIfModifiedSince, ifModifiedSince)
		}
		if ifNoneMatch != "" {
			reqHeader.Set(cod.HeaderIfNoneMatch, ifNoneMatch)
		}
		if err != nil {
			return
		}

		for _, key := range clearHeaders {
			// 清除header
			c.SetHeader(key, "")
		}
		return c.Next()
	}
}
