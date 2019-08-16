package middleware

import (
	"strings"

	"github.com/vicanso/elton"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/df"

	"github.com/vicanso/pike/upstream"
)

var (
	// 需要清除的header
	clearHeaders = []string{
		"Date",
		"Connection",
		elton.HeaderContentLength,
	}
)

// NewProxy create a proxy middleware
func NewProxy(director *upstream.Director) elton.Handler {
	return func(c *elton.Context) (err error) {
		// 如果请求是从缓存读取Cacheable ，则直接跳过
		status, ok := c.Get(df.Status).(int)
		if ok && status == cache.Cacheable {
			return c.Next()
		}
		originalNext := c.Next
		// 由于proxy中间件会调用next，因此直接覆盖，
		// 避免导致先执行了后续的中间件
		c.Next = func() error {
			return nil
		}

		var ifModifiedSince, ifNoneMatch, acceptEncoding string
		reqHeader := c.Request.Header
		// proxy时为了避免304的出现，因此调用时临时删除header
		// 如果是 pass 则无需删除，其它的需要删除（因为此时无法确认数据是否可缓存）
		if status != cache.Pass &&
			status != cache.HitForPass {
			acceptEncoding = reqHeader.Get(elton.HeaderAcceptEncoding)
			ifModifiedSince = reqHeader.Get(elton.HeaderIfModifiedSince)
			ifNoneMatch = reqHeader.Get(elton.HeaderIfNoneMatch)
			if ifModifiedSince != "" {
				reqHeader.Del(elton.HeaderIfModifiedSince)
			}
			if ifNoneMatch != "" {
				reqHeader.Del(elton.HeaderIfNoneMatch)
			}

			if strings.Contains(acceptEncoding, df.GZIP) {
				reqHeader.Set(elton.HeaderAcceptEncoding, df.GZIP)
			} else {
				reqHeader.Del(elton.HeaderAcceptEncoding)
			}
		}

		err = director.Proxy(c)

		// 将原有的请求头恢复
		if acceptEncoding != "" {
			reqHeader.Set(elton.HeaderAcceptEncoding, acceptEncoding)
		}
		if ifModifiedSince != "" {
			reqHeader.Set(elton.HeaderIfModifiedSince, ifModifiedSince)
		}
		if ifNoneMatch != "" {
			reqHeader.Set(elton.HeaderIfNoneMatch, ifNoneMatch)
		}

		if err != nil {
			return
		}

		for _, key := range clearHeaders {
			// 清除header
			c.SetHeader(key, "")
		}
		return originalNext()
	}
}
