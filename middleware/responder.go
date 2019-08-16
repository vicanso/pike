package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/vicanso/hes"

	"github.com/vicanso/elton"
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/df"
)

var (
	errCacheInvalid = &hes.Error{
		StatusCode: http.StatusInternalServerError,
		Category:   df.APP,
		Message:    "http cache is invalid",
		Exception:  true,
	}
)

// NewResponder create a respond middleware
func NewResponder() elton.Handler {
	return func(c *elton.Context) (err error) {
		err = c.Next()
		// 出错或者已设置响应数据
		if err != nil || c.BodyBuffer != nil {
			return
		}
		v := c.Get(df.Cache)
		if v == nil {
			return
		}
		hc, ok := v.(*cache.HTTPCache)
		if !ok {
			err = errCacheInvalid
			return
		}
		if hc.Status != cache.Cacheable {
			return
		}
		// 获取客户端可接受的 encoding
		acceptEncoding := c.GetRequestHeader(elton.HeaderAcceptEncoding)

		for k, value := range hc.Headers {
			for _, v := range value {
				c.SetHeader(k, v)
			}
		}
		c.StatusCode = hc.StatusCode
		// 计算缓存已存在时长
		age := time.Now().Unix() - hc.CreatedAt
		if age > 0 {
			c.SetHeader(df.HeaderAge, strconv.Itoa(int(age)))
		}
		buf, encoding, err := hc.Response(acceptEncoding)
		if err != nil {
			return
		}
		c.BodyBuffer = buf
		if encoding != "" {
			c.SetHeader(elton.HeaderContentEncoding, encoding)
		}
		return
	}
}
