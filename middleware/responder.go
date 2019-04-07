package middleware

import (
	"bytes"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/vicanso/hes"

	"github.com/vicanso/cod"
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
func NewResponder() cod.Handler {
	return func(c *cod.Context) (err error) {
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
		acceptEncoding := c.GetRequestHeader(cod.HeaderAcceptEncoding)

		for k, value := range hc.Headers {
			for _, v := range value {
				c.SetHeader(k, v)
			}
		}
		var encoding string
		var buf *bytes.Buffer
		c.StatusCode = hc.StatusCode
		age := time.Now().Unix() - hc.CreatedAt
		if age > 0 {
			c.SetHeader(df.HeaderAge, strconv.Itoa(int(age)))
		}
		if hc.BrBody != nil &&
			strings.Contains(acceptEncoding, df.BR) {
			buf = hc.BrBody
			encoding = df.BR
		} else if hc.GzipBody != nil && strings.Contains(acceptEncoding, df.GZIP) {
			buf = hc.GzipBody
			encoding = df.GZIP
		} else if hc.GzipBody != nil {
			// 缓存了压缩数据，但是客户端不支持，需要解压
			rawData, e := cache.Gunzip(hc.GzipBody.Bytes())
			if e != nil {
				err = hes.NewWithError(e)
				return
			}
			buf = bytes.NewBuffer(rawData)
		} else {
			buf = hc.Body
		}
		c.BodyBuffer = buf
		if encoding != "" {
			c.SetHeader(cod.HeaderContentEncoding, encoding)
		}
		return
	}
}
