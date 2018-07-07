package middleware

import (
	"github.com/vicanso/pike/cache"
	"github.com/vicanso/pike/pike"
)

type (
	// CacheFetcherConfig cache fetcher配置
	CacheFetcherConfig struct {
	}
)

// CacheFetcher 从缓存中获取数据
func CacheFetcher(config CacheFetcherConfig, client *cache.Client) pike.Middleware {
	return func(c *pike.Context, next pike.Next) error {
		done := c.ServerTiming.Start(pike.ServerTimingCacheFetcher)
		status := c.Status
		if status == 0 {
			done()
			return ErrRequestStatusNotSet
		}
		// 如果非cache的
		if status != cache.Cacheable {
			done()
			return next()
		}
		identity := c.Identity
		if identity == nil {
			done()
			return ErrIdentityNotSet
		}
		resp, err := client.GetResponse(identity)
		if err != nil {
			done()
			return err
		}
		c.Resp = resp
		done()
		return next()
	}
}
