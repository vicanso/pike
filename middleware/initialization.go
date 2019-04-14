package middleware

import (
	"net/http"

	"github.com/vicanso/cod"
	"github.com/vicanso/hes"
	"github.com/vicanso/pike/config"

	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/df"
	"github.com/vicanso/pike/performance"
)

var (
	errTooManyRequest = &hes.Error{
		StatusCode: http.StatusTooManyRequests,
		Message:    "too many request is handling",
		Category:   df.APP,
		Exception:  true,
	}
)

// NewInitialization create an initialization middleware
func NewInitialization() cod.Handler {
	maxConcurrency := config.GetConcurrency()
	header := config.GetHeader()
	requestHeader := config.GetRequestHeader()

	return func(c *cod.Context) (err error) {
		defer performance.DecreaseConcurrency()
		performance.IncreaseRequestCount()
		count := performance.IncreaseConcurrency()

		// 设置请求头
		reqHeader := c.Request.Header
		for key, values := range requestHeader {
			for _, value := range values {
				reqHeader.Add(key, value)
			}
		}

		// 如果请求数超过最大限制
		if count > maxConcurrency {
			err = errTooManyRequest
			return
		}

		err = c.Next()
		// 设置响应头 （最后再设置，避免在后续缓存中将全局响应头缓存）
		for key, values := range header {
			for _, value := range values {
				c.AddHeader(key, value)
			}
		}
		v, _ := c.Get(df.Status).(int)
		c.SetHeader(df.HeaderStatus, cache.GetStatusDesc(v))
		return
	}
}
